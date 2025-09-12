package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   false,
				DefaultFunc: schema.EnvDefaultFunc("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
				Description: "The base URL for the OpenRouter API",
			},
			"provisioning_api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("OPENROUTER_PROVISIONING_API_KEY", nil),
				Description: "The provisioning API key for OpenRouter key management",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"openrouter_api_key": resourceApiKey(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

type apiClient struct {
	baseURL            string
	provisioningAPIKey string
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	baseURL := d.Get("base_url").(string)
	provisioningAPIKey := d.Get("provisioning_api_key").(string)

	var diags diag.Diagnostics

	if provisioningAPIKey == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing provisioning API key",
			Detail:   "The provisioning_api_key must be provided",
		})
		return nil, diags
	}

	client := &apiClient{
		baseURL:            baseURL,
		provisioningAPIKey: provisioningAPIKey,
	}

	return client, diags
}
