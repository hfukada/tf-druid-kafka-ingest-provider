package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type druidClient struct{
	username string
	password string
	endpoint string
	sslcert string
	http *http.Client
}


func init() {
	schema.DescriptionKind = schema.StringMarkdown

	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		if s.Computed {
			desc += " Generated."
		}
		return strings.TrimSpace(desc)
	}
}

func Provider(version string, testing bool) *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type: schema.TypeString,
				Required: true,
				Description: "Router endpoint server with port",
				DefaultFunc: schema.EnvDefaultFunc("DRUID_ENDPOINT", "http://localhost:8080"),
			},
			"username": {
				Type: schema.TypeString,
				Required: true,
				Description: "Druid Username to authenticate as",
				DefaultFunc: schema.EnvDefaultFunc("DRUID_USERNAME", "username"),
			},
			"password": {
				Type: schema.TypeString,
				Required: true,
				Description: "Druid Password to authenticate with",
				DefaultFunc: schema.EnvDefaultFunc("DRUID_ENDPOINT", "password"),
			},
			"sslcert": {
				Type: schema.TypeString,
				Required: false,
				Description: "Use this specific certificate for TLS comms for authentication",
				DefaultFunc: schema.EnvDefaultFunc("DRUID_CERT", ""),
			},
		},
		ConfigureContextFunc: configureProvider,
		ResourcesMap: map[string]*schema.Resource{
			"druid_kafka_supervisor": resourceKafkaSupervisor(),
		},
	}
}

func configureProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
    endpoint := d.Get("endpoint").(string)
    username := d.Get("username").(string)
    password := d.Get("password").(string)
		cert := d.Get("sslcert").(string)

    client := &druidClient{
        endpoint: endpoint,
        username: username,
        password: password,
				sslcert: cert,
        http:     &http.Client{},
    }
    return client, nil
}


// -------------------- resource implementation --------------------

type kafkaSupervisorResource struct{
    client *druidClient
}

func resourceKafkaSupervisor() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "jsonSpec": {
                Type:        schema.TypeString,
                Required:    true,
                Description: "Full Kafka supervisor JSON spec",
            },
        },

        CreateContext: resourceKafkaSupervisorCreate,
        ReadContext:   resourceKafkaSupervisorRead,
        UpdateContext: resourceKafkaSupervisorUpdate,
        DeleteContext: resourceKafkaSupervisorDelete,

        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
    }
}

func resourceKafkaSupervisorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := meta.(*druidClient)
    spec := d.Get("spec").(string)

    id, err := client.CreateOrUpdate(ctx, spec)
    if err != nil {
        return diag.FromErr(err)
    }

    d.SetId(id)
    return resourceKafkaSupervisorRead(ctx, d, meta)
}

func resourceKafkaSupervisorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := meta.(*druidClient)

    found, spec, err := client.Get(ctx, d.Id())
    if err != nil {
        return diag.FromErr(err)
    }
    if !found {
        d.SetId("")
        return nil
    }

    _ = d.Set("spec", spec)
    return nil
}

func resourceKafkaSupervisorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    if d.HasChange("spec") {
        client := meta.(*druidClient)
        spec := d.Get("spec").(string)
        if _, err := client.CreateOrUpdate(ctx, spec); err != nil {
            return diag.FromErr(err)
        }
    }
    return resourceKafkaSupervisorRead(ctx, d, meta)
}

func resourceKafkaSupervisorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := meta.(*druidClient)
    if err := client.Delete(ctx, d.Id()); err != nil {
        return diag.FromErr(err)
    }
    d.SetId("")
    return nil
}

// -------------------- druidClient helpers --------------------

func (c *druidClient) newReq(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
    req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.endpoint, path), body)
    if err != nil {
        return nil, err
    }
    if c.username != "" {
        req.SetBasicAuth(c.username, c.password)
    }
    req.Header.Set("Content-Type", "application/json")
    return req, nil
}

func (c *druidClient) CreateOrUpdate(ctx context.Context, spec string) (string, error) {
    req, err := c.newReq(ctx, http.MethodPost, "/druid/indexer/v1/supervisor", bytes.NewBufferString(spec))
    if err != nil {
        return "", err
    }
    resp, err := c.http.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= http.StatusBadRequest {
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("druid error %d: %s", resp.StatusCode, body)
    }

    var out struct{ ID string `json:"id"` }
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return "", err
    }
    return out.ID, nil
}

func (c *druidClient) Get(ctx context.Context, id string) (bool, string, error) {
    req, err := c.newReq(ctx, http.MethodGet, "/druid/indexer/v1/supervisor/"+id, nil)
    if err != nil {
        return false, "", err
    }
    resp, err := c.http.Do(req)
    if err != nil {
        return false, "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return false, "", nil
    }
    if resp.StatusCode >= http.StatusBadRequest {
        body, _ := io.ReadAll(resp.Body)
        return false, "", fmt.Errorf("druid error %d: %s", resp.StatusCode, body)
    }

    body, _ := io.ReadAll(resp.Body)
    return true, string(body), nil
}

func (c *druidClient) Delete(ctx context.Context, id string) error {
    req, err := c.newReq(ctx, http.MethodDelete, "/druid/indexer/v1/supervisor/"+id, nil)
    if err != nil {
        return err
    }
    resp, err := c.http.Do(req)
    if err != nil {
        return err
    }
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("druid error %d: %s", resp.StatusCode, body)
    }
    return nil
}
