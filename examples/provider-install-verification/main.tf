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

  # Provisioning API key for managing OpenRouter API keys
  provisioning_api_key = "your-provisioning-key-here"
  # provisioning_api_key = var.openrouter_provisioning_api_key
}

# Example: Create an API key with a credit limit
resource "openrouter_api_key" "customer_key" {
  name  = "Customer Instance Key"
  label = "customer-123"
  limit = 1000.0
}

# Example: Create an API key without a limit
resource "openrouter_api_key" "unlimited_key" {
  name  = "Unlimited Usage Key"
  label = "unlimited-dev"
}

# Example: Create a disabled API key
resource "openrouter_api_key" "disabled_key" {
  name     = "Disabled Key"
  label    = "disabled-test"
  disabled = true
  limit    = 50.0
}

# Output the created API key value (sensitive)
output "customer_api_key" {
  value     = openrouter_api_key.customer_key.key_value
  sensitive = true
}

output "customer_key_hash" {
  value = openrouter_api_key.customer_key.hash
}

resource "budgeteer_api_key" "eriks_key" {
  name   = "eriks-key"
  budget = 10000
}

resource "budgeteer_api_key" "jimmys_key" {
  name   = "jimmys-key"
  budget = 10000
}
