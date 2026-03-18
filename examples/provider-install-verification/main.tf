terraform {
  required_providers {
    openrouter = {
      source  = "hashicorp.com/dev/openrouter"
      version = "1.0.0"
    }
  }
}

provider "openrouter" {
  # Base URL for OpenRouter API (optional, defaults to https://openrouter.ai/api/v1)
  # base_url = "https://openrouter.ai/api/v1"

  # Provisioning API key — use env var OPENROUTER_PROVISIONING_API_KEY instead of hardcoding
  provisioning_api_key = var.openrouter_provisioning_api_key
}

variable "openrouter_provisioning_api_key" {
  description = "OpenRouter provisioning API key"
  type        = string
  sensitive   = true
}

# Create an API key with a monthly credit limit
resource "openrouter_api_key" "customer_key" {
  name        = "Customer Instance Key"
  limit       = 100.0
  limit_reset = "monthly"
}

# Create an API key with a daily budget and expiry
resource "openrouter_api_key" "short_lived_key" {
  name        = "Short-Lived Key"
  limit       = 10.0
  limit_reset = "daily"
  expires_at  = "2026-12-31T23:59:59Z"
}

# Create an unlimited API key
resource "openrouter_api_key" "unlimited_key" {
  name = "Unlimited Usage Key"
}

# Create a disabled API key
resource "openrouter_api_key" "disabled_key" {
  name     = "Disabled Key"
  disabled = true
  limit    = 50.0
}

# Create a key that includes BYOK usage in the limit
resource "openrouter_api_key" "byok_key" {
  name                 = "BYOK Key"
  limit                = 500.0
  limit_reset          = "weekly"
  include_byok_in_limit = true
}

output "customer_api_key" {
  description = "The actual API key value (only populated on creation)"
  value       = openrouter_api_key.customer_key.key_value
  sensitive   = true
}

output "customer_key_hash" {
  description = "The hash identifier of the customer key"
  value       = openrouter_api_key.customer_key.hash
}

output "customer_key_label" {
  description = "The label assigned by OpenRouter"
  value       = openrouter_api_key.customer_key.label
}
