package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
		
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

)

// -------------------- provider entrypoint --------------------
func main() {
    providerserver.Serve(context.Background(), New, providerserver.ServeOpts{
        Address: "registry.terraform.io/example/druid", // change to your namespace
    })
}

func New() provider.Provider { return &druidProvider{} }

// -------------------- provider implementation --------------------

type druidProvider struct {
    endpoint string
    username string
    password string
		useTls   bool
}

func (p *druidProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "druid-supervisor-kafka"
}

func (p *druidProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "endpoint": schema.StringAttribute{
                Required:    true,
                Description: "Base URL of the Druid Overlord API (e.g. https://druid.example.com)",
            },
            "username": schema.StringAttribute{
                Optional:    true,
                Description: "Basic‑auth username (optional)",
            },
            "password": schema.StringAttribute{
                Optional:    true,
                Sensitive:   true,
                Description: "Basic‑auth password (optional)",
            },
        },
    }
}

func (p *druidProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var cfg struct {
        Endpoint types.String `tfsdk:"endpoint"`
        Username types.String `tfsdk:"username"`
        Password types.String `tfsdk:"password"`
    }
    diags := req.Config.Get(ctx, &cfg)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
    p.endpoint = cfg.Endpoint.ValueString()
    p.username = cfg.Username.ValueString()
    p.password = cfg.Password.ValueString()
    resp.DataSourceData = p
    resp.ResourceData = p
}

func (p *druidProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{NewKafkaSupervisorResource}
}

// -------------------- resource implementation --------------------

type kafkaSupervisorResource struct{
    client *druidClient
}

func NewKafkaSupervisorResource() resource.Resource { return &kafkaSupervisorResource{} }

func (r *kafkaSupervisorResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = "druid_kafka_supervisor"
}

func (r *kafkaSupervisorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:    true,
                Description: "Supervisor ID (datasource name)",
            },
            "spec": schema.StringAttribute{
                Required:    true,
                Description: "Full Kafka supervisor JSON spec",
            },
        },
    }
}

type supervisorModel struct {
    ID   types.String `tfsdk:"id"`
    Spec types.String `tfsdk:"spec"`
}

func (r *kafkaSupervisorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
    if req.ProviderData == nil { return }
    p := req.ProviderData.(*druidProvider)
    r.client = &druidClient{endpoint: p.endpoint, username: p.username, password: p.password}
}

// ---- CRUD ----
func (r *kafkaSupervisorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan supervisorModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() { return }

    id, err := r.client.CreateOrUpdate(ctx, plan.Spec.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("creating supervisor", err.Error())
        return
    }
    plan.ID = types.StringValue(id)
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...) // save state
}

func (r *kafkaSupervisorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state supervisorModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // read current state
    if resp.Diagnostics.HasError() { return }

    found, spec, err := r.client.Get(ctx, state.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("reading supervisor", err.Error())
        return
    }
    if !found {
        resp.State.RemoveResource(ctx)
        return
    }
    state.Spec = types.StringValue(spec)
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...) // refresh state
}

func (r *kafkaSupervisorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan supervisorModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() { return }

    _, err := r.client.CreateOrUpdate(ctx, plan.Spec.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("updating supervisor", err.Error())
        return
    }
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...) // persist
}

func (r *kafkaSupervisorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state supervisorModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    if err := r.client.Delete(ctx, state.ID.ValueString()); err != nil {
        resp.Diagnostics.AddError("deleting supervisor", err.Error())
    }
}

// -------------------- simple HTTP client --------------------

type druidClient struct {
    endpoint string
    username string
    password string
}

func (c *druidClient) newReq(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
    req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.endpoint, path), body)
    if err != nil { return nil, err }
    if c.username != "" { req.SetBasicAuth(c.username, c.password) }
    req.Header.Set("Content-Type", "application/json")
    return req, nil
}

func (c *druidClient) CreateOrUpdate(ctx context.Context, spec string) (string, error) {
    req, err := c.newReq(ctx, http.MethodPost, "/druid/indexer/v1/supervisor", bytes.NewBufferString(spec))
    if err != nil { return "", err }
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return "", err }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("druid error %d: %s", resp.StatusCode, body)
    }
    // Druid echoes back {"id":"<datasource>"}
    var out struct{ ID string `json:"id"` }
    _ = json.NewDecoder(resp.Body).Decode(&out)
    return out.ID, nil
}

func (c *druidClient) Get(ctx context.Context, id string) (bool, string, error) {
    req, err := c.newReq(ctx, http.MethodGet, "/druid/indexer/v1/supervisor/"+id, nil)
    if err != nil { return false, "", err }
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return false, "", err }
    defer resp.Body.Close()
    if resp.StatusCode == http.StatusNotFound { return false, "", nil }
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return false, "", fmt.Errorf("druid error %d: %s", resp.StatusCode, body)
    }
    body, _ := io.ReadAll(resp.Body)
    return true, string(body), nil
}

func (c *druidClient) Delete(ctx context.Context, id string) error {
    req, err := c.newReq(ctx, http.MethodDelete, "/druid/indexer/v1/supervisor/"+id, nil)
    if err != nil { return err }
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return err }
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("druid error %d: %s", resp.StatusCode, body)
    }
    return nil
}

