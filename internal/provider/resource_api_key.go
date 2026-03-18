package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// OpenRouter API key structures
type OpenRouterKeyResponse struct {
	Data OpenRouterKey `json:"data"`
	Key  string        `json:"key,omitempty"` // Only returned on creation
}

type OpenRouterKeyListResponse struct {
	Data []OpenRouterKey `json:"data"`
}

type OpenRouterKey struct {
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          *string  `json:"updated_at"`
	ExpiresAt          *string  `json:"expires_at"`
	Hash               string   `json:"hash"`
	Label              string   `json:"label"`
	Name               string   `json:"name"`
	Disabled           bool     `json:"disabled"`
	Limit              *float64 `json:"limit"`
	LimitRemaining     *float64 `json:"limit_remaining"`
	LimitReset         *string  `json:"limit_reset"`
	Usage              float64  `json:"usage"`
	UsageDaily         float64  `json:"usage_daily"`
	UsageWeekly        float64  `json:"usage_weekly"`
	UsageMonthly       float64  `json:"usage_monthly"`
	ByokUsage          float64  `json:"byok_usage"`
	ByokUsageDaily     float64  `json:"byok_usage_daily"`
	ByokUsageWeekly    float64  `json:"byok_usage_weekly"`
	ByokUsageMonthly   float64  `json:"byok_usage_monthly"`
	IncludeByokInLimit *bool    `json:"include_byok_in_limit,omitempty"`
}

type CreateKeyRequest struct {
	Name               string   `json:"name"`
	Limit              *float64 `json:"limit,omitempty"`
	LimitReset         string   `json:"limit_reset,omitempty"`
	ExpiresAt          string   `json:"expires_at,omitempty"`
	IncludeByokInLimit *bool    `json:"include_byok_in_limit,omitempty"`
}

type UpdateKeyRequest struct {
	Name               *string  `json:"name,omitempty"`
	Disabled           *bool    `json:"disabled,omitempty"`
	Limit              *float64 `json:"limit,omitempty"`
	LimitReset         *string  `json:"limit_reset,omitempty"`
	IncludeByokInLimit *bool    `json:"include_byok_in_limit,omitempty"`
}

func resourceApiKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApiKeyCreate,
		ReadContext:   resourceApiKeyRead,
		UpdateContext: resourceApiKeyUpdate,
		DeleteContext: resourceApiKeyDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the API key",
			},
			"label": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Human-readable label for the API key (assigned by the API)",
			},
			"limit": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Credit limit for the API key",
			},
			"limit_reset": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "How often the credit limit resets: 'daily', 'weekly', or 'monthly'. Resets happen at midnight UTC; weeks are Monday-Sunday.",
			},
			"expires_at": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ISO 8601 UTC timestamp when the API key expires (cannot be changed after creation, requires replacement)",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the API key is disabled",
			},
			"include_byok_in_limit": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to include BYOK usage in the limit",
			},
			// Computed fields
			"hash": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The hash identifier of the API key",
			},
			"key_value": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The actual API key value (only available on creation)",
			},
			"limit_remaining": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Remaining spending limit in USD",
			},
			"usage": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Total OpenRouter credit usage in USD",
			},
			"usage_daily": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "OpenRouter credit usage for the current UTC day",
			},
			"usage_weekly": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "OpenRouter credit usage for the current UTC week (Monday-Sunday)",
			},
			"usage_monthly": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "OpenRouter credit usage for the current UTC month",
			},
			"byok_usage": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Total external BYOK usage in USD",
			},
			"byok_usage_daily": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "External BYOK usage for the current UTC day",
			},
			"byok_usage_weekly": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "External BYOK usage for the current UTC week (Monday-Sunday)",
			},
			"byok_usage_monthly": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "External BYOK usage for the current UTC month",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update timestamp",
			},
		},
	}
}

func setKeyAttributes(d *schema.ResourceData, key OpenRouterKey) {
	d.Set("hash", key.Hash)
	d.Set("name", key.Name)
	d.Set("label", key.Label)
	d.Set("disabled", key.Disabled)
	d.Set("include_byok_in_limit", key.IncludeByokInLimit != nil && *key.IncludeByokInLimit)
	d.Set("usage", key.Usage)
	d.Set("usage_daily", key.UsageDaily)
	d.Set("usage_weekly", key.UsageWeekly)
	d.Set("usage_monthly", key.UsageMonthly)
	d.Set("byok_usage", key.ByokUsage)
	d.Set("byok_usage_daily", key.ByokUsageDaily)
	d.Set("byok_usage_weekly", key.ByokUsageWeekly)
	d.Set("byok_usage_monthly", key.ByokUsageMonthly)
	d.Set("created_at", key.CreatedAt)

	if key.Limit != nil {
		d.Set("limit", *key.Limit)
	}
	if key.LimitRemaining != nil {
		d.Set("limit_remaining", *key.LimitRemaining)
	}
	if key.LimitReset != nil {
		d.Set("limit_reset", *key.LimitReset)
	}
	if key.UpdatedAt != nil {
		d.Set("updated_at", *key.UpdatedAt)
	}
	if key.ExpiresAt != nil {
		d.Set("expires_at", *key.ExpiresAt)
	}
}

func resourceApiKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	createReq := CreateKeyRequest{
		Name: d.Get("name").(string),
	}

	if limit, ok := d.GetOk("limit"); ok {
		v := limit.(float64)
		createReq.Limit = &v
	}

	if limitReset, ok := d.GetOk("limit_reset"); ok {
		createReq.LimitReset = limitReset.(string)
	}

	if expiresAt, ok := d.GetOk("expires_at"); ok {
		createReq.ExpiresAt = expiresAt.(string)
	}

	if v, ok := d.GetOkExists("include_byok_in_limit"); ok {
		b := v.(bool)
		createReq.IncludeByokInLimit = &b
	}

	jsonData, err := json.Marshal(createReq)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/keys", client.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.provisioningAPIKey))
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return diag.Errorf("Failed to create API key: %s - %s", resp.Status, string(body))
	}

	var keyResp OpenRouterKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&keyResp); err != nil {
		return diag.FromErr(err)
	}

	key := keyResp.Data
	d.SetId(key.Hash)

	setKeyAttributes(d, key)

	// The key value is only returned on creation and is at the root level
	if keyResp.Key != "" {
		d.Set("key_value", keyResp.Key)
	}

	return nil
}

func resourceApiKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)
	keyHash := d.Id()

	url := fmt.Sprintf("%s/keys/%s", client.baseURL, keyHash)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.provisioningAPIKey))
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Key doesn't exist anymore, remove from state
		d.SetId("")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return diag.Errorf("Failed to read API key: %s - %s", resp.Status, string(body))
	}

	var keyResp OpenRouterKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&keyResp); err != nil {
		return diag.FromErr(err)
	}

	key := keyResp.Data

	setKeyAttributes(d, key)

	return nil
}

func resourceApiKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)
	keyHash := d.Id()

	updateReq := UpdateKeyRequest{}
	hasChanges := false

	if d.HasChange("name") {
		name := d.Get("name").(string)
		updateReq.Name = &name
		hasChanges = true
	}

	if d.HasChange("disabled") {
		disabled := d.Get("disabled").(bool)
		updateReq.Disabled = &disabled
		hasChanges = true
	}

	if d.HasChange("limit") {
		limit := d.Get("limit").(float64)
		updateReq.Limit = &limit
		hasChanges = true
	}

	if d.HasChange("limit_reset") {
		limitReset := d.Get("limit_reset").(string)
		updateReq.LimitReset = &limitReset
		hasChanges = true
	}

	if d.HasChange("include_byok_in_limit") {
		includeByok := d.Get("include_byok_in_limit").(bool)
		updateReq.IncludeByokInLimit = &includeByok
		hasChanges = true
	}

	if !hasChanges {
		return resourceApiKeyRead(ctx, d, m)
	}

	jsonData, err := json.Marshal(updateReq)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/keys/%s", client.baseURL, keyHash)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.provisioningAPIKey))
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return diag.Errorf("Failed to update API key: %s - %s", resp.Status, string(body))
	}

	return resourceApiKeyRead(ctx, d, m)
}

func resourceApiKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)
	keyHash := d.Id()

	url := fmt.Sprintf("%s/keys/%s", client.baseURL, keyHash)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.provisioningAPIKey))
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return diag.Errorf("Failed to delete API key: %s - %s", resp.Status, string(body))
	}

	d.SetId("")
	return nil
}
