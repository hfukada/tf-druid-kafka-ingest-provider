package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	provider := New()
	
	// Test that provider is not nil
	require.NotNil(t, provider)
	
	// Test provider schema
	assert.NotNil(t, provider.Schema)
	assert.NotNil(t, provider.ResourcesMap)
	assert.NotNil(t, provider.ConfigureContextFunc)
	
	// Test required provider configurations
	endpointSchema := provider.Schema["endpoint"]
	assert.True(t, endpointSchema.Required)
	assert.Equal(t, schema.TypeString, endpointSchema.Type)
	
	// Test optional provider configurations
	usernameSchema := provider.Schema["username"]
	assert.True(t, usernameSchema.Optional)
	assert.Equal(t, schema.TypeString, usernameSchema.Type)
	
	passwordSchema := provider.Schema["password"]
	assert.True(t, passwordSchema.Optional)
	assert.True(t, passwordSchema.Sensitive)
	assert.Equal(t, schema.TypeString, passwordSchema.Type)
	
	timeoutSchema := provider.Schema["timeout"]
	assert.True(t, timeoutSchema.Optional)
	assert.Equal(t, schema.TypeInt, timeoutSchema.Type)
	assert.Equal(t, 30, timeoutSchema.Default)
	
	// Test that resources are registered
	assert.Contains(t, provider.ResourcesMap, "druid_kafka_supervisor")
	assert.NotNil(t, provider.ResourcesMap["druid_kafka_supervisor"])
}

func TestProviderConfigure(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid configuration",
			config: map[string]interface{}{
				"endpoint": "http://localhost:8080",
				"username": "testuser",
				"password": "testpass",
				"timeout":  60,
			},
			expectError: false,
		},
		{
			name: "minimal configuration",
			config: map[string]interface{}{
				"endpoint": "http://localhost:8080",
			},
			expectError: false,
		},
		{
			name: "configuration with defaults",
			config: map[string]interface{}{
				"endpoint": "https://druid.example.com:8080",
				"username": "",
				"password": "",
				"timeout":  30,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := New()
			d := schema.TestResourceDataRaw(t, provider.Schema, tt.config)
			
			client, diags := provider.ConfigureContextFunc(nil, d)
			
			if tt.expectError {
				assert.True(t, diags.HasError())
				assert.Nil(t, client)
			} else {
				assert.False(t, diags.HasError())
				assert.NotNil(t, client)
				
				// Verify client configuration
				c, ok := client.(*Client)
				require.True(t, ok)
				assert.Equal(t, tt.config["endpoint"].(string), c.Endpoint)
				assert.NotNil(t, c.HTTPClient)
			}
		})
	}
}

func TestProviderValidation(t *testing.T) {
	provider := New()
	
	// Test provider schema validation
	err := provider.InternalValidate()
	assert.NoError(t, err)
	
	// Test resource validation
	resource := provider.ResourcesMap["druid_kafka_supervisor"]
	err = resource.InternalValidate(nil, true)
	assert.NoError(t, err)
}