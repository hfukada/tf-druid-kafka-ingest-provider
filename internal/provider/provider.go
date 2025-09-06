package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Druid router endpoint URL with port (e.g., http://localhost:8080)",
				DefaultFunc: schema.EnvDefaultFunc("DRUID_ENDPOINT", nil),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username for Druid authentication",
				DefaultFunc: schema.EnvDefaultFunc("DRUID_USERNAME", ""),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password for Druid authentication",
				DefaultFunc: schema.EnvDefaultFunc("DRUID_PASSWORD", ""),
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "HTTP client timeout in seconds",
				Default:     30,
			},
		},
		ConfigureContextFunc: configure,
		ResourcesMap: map[string]*schema.Resource{
			"druid_kafka_supervisor": resourceKafkaSupervisor(),
		},
	}
}

func configure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := &Config{
		Endpoint: d.Get("endpoint").(string),
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
		Timeout:  d.Get("timeout").(int),
	}

	client, err := config.Client()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return client, nil
}