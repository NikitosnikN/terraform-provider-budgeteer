# Terraform Provider for OpenRouter

This Terraform provider allows you to manage OpenRouter API keys programmatically using the OpenRouter Key Management API.

## Requirements

- Terraform 0.13.x or later
- Go 1.21 or later (for building from source)
- An OpenRouter Provisioning API key

## Getting Started

### 1. Create a Provisioning API Key

Before using this provider, you need to create a Provisioning API key:

1. Go to the [Provisioning API Keys page](https://openrouter.ai/settings/provisioning-keys)
2. Click "Create New Key"
3. Complete the key creation process

**Important**: Provisioning keys cannot be used to make API calls to OpenRouter's completion endpoints - they are exclusively for key management operations.

### 2. Configure the Provider

```hcl
terraform {
  required_providers {
    openrouter = {
      source  = "hashicorp.com/dev/openrouter"
      version = "~> 1.0"
    }
  }
}

provider "openrouter" {
  # Optional: Base URL for OpenRouter API (defaults to https://openrouter.ai/api/v1)
  # base_url = "https://openrouter.ai/api/v1"

  # Required: Provisioning API key for managing OpenRouter API keys
  provisioning_api_key = var.openrouter_provisioning_api_key
}
```

### 3. Environment Variables

You can set the following environment variables instead of hardcoding values:

```bash
export OPENROUTER_PROVISIONING_API_KEY="your-provisioning-key"
export OPENROUTER_BASE_URL="https://openrouter.ai/api/v1"  # Optional
```

## Resources

### `openrouter_api_key`

Manages an OpenRouter API key.

#### Example Usage

```hcl
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
  name = "Development Key"
}

# Create a key that counts BYOK usage toward the limit
resource "openrouter_api_key" "byok_key" {
  name                  = "BYOK Key"
  limit                 = 500.0
  limit_reset           = "weekly"
  include_byok_in_limit = true
}
```

#### Argument Reference

- `name` - (Required) The name of the API key.
- `limit` - (Optional) Spending limit for the API key in USD. Omit for unlimited.
- `limit_reset` - (Optional) How often the credit limit resets. Valid values: `daily`, `weekly`, `monthly`. Resets happen at midnight UTC; weeks are Monday–Sunday.
- `expires_at` - (Optional) ISO 8601 UTC timestamp when the key expires (e.g. `"2026-12-31T23:59:59Z"`). Changing this forces a new resource.
- `disabled` - (Optional) Whether the API key is disabled. Default: `false`.
- `include_byok_in_limit` - (Optional) Whether to include BYOK (Bring Your Own Key) usage toward the credit limit.

#### Attribute Reference

In addition to the arguments above, the following computed attributes are exported:

- `label` - Human-readable label assigned by OpenRouter.
- `hash` - The unique hash identifier of the API key.
- `key_value` - The actual API key string. **Only populated on creation and stored in state — never shown again by the API.** Marked sensitive.
- `limit_remaining` - Remaining credit in USD.
- `usage` - Total OpenRouter credit usage in USD.
- `usage_daily` - Credit usage for the current UTC day.
- `usage_weekly` - Credit usage for the current UTC week (Monday–Sunday).
- `usage_monthly` - Credit usage for the current UTC month.
- `byok_usage` - Total external BYOK usage in USD.
- `byok_usage_daily` - BYOK usage for the current UTC day.
- `byok_usage_weekly` - BYOK usage for the current UTC week.
- `byok_usage_monthly` - BYOK usage for the current UTC month.
- `created_at` - ISO 8601 timestamp of key creation.
- `updated_at` - ISO 8601 timestamp of last update (null if never updated).

## Use Cases

### SaaS Applications

Automatically create unique API keys for each customer instance:

```hcl
resource "openrouter_api_key" "customer" {
  for_each = var.customers

  name        = "Customer ${each.key}"
  limit       = each.value.credit_limit
  limit_reset = "monthly"
}

output "customer_api_keys" {
  value = {
    for k, v in openrouter_api_key.customer : k => v.key_value
  }
  sensitive = true
}
```

### Key Rotation

Implement automatic key rotation for security compliance:

```hcl
resource "openrouter_api_key" "rotated_key" {
  name  = "Production Key ${formatdate("YYYY-MM", timestamp())}"
  limit = var.production_limit

  lifecycle {
    create_before_destroy = true
  }
}
```

## Building from Source

```bash
git clone https://github.com/knowit-solutions-cocreate/terraform-provider-openrouter
cd terraform-provider-openrouter
go build
```

## Development

### Running Tests

```bash
go test ./...
```

### Local Development

For local development, you can build and install the provider locally:

```bash
go build -o terraform-provider-openrouter
mkdir -p ~/.terraform.d/plugins/hashicorp.com/dev/openrouter/1.0.0/darwin_arm64/
cp terraform-provider-openrouter ~/.terraform.d/plugins/hashicorp.com/dev/openrouter/1.0.0/darwin_arm64/
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for your changes
5. Run tests and ensure they pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
