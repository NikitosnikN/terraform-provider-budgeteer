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
# Create an API key with a credit limit
resource "openrouter_api_key" "customer_key" {
  name  = "Customer Instance Key"
  label = "customer-123"
  limit = 1000.0
}

# Create an unlimited API key
resource "openrouter_api_key" "unlimited_key" {
  name = "Development Key"
  label = "dev-environment"
}

# Create a disabled API key
resource "openrouter_api_key" "disabled_key" {
  name     = "Disabled Key"
  disabled = true
  limit    = 50.0
}
```

#### Argument Reference

- `name` - (Required) The name of the API key
- `label` - (Optional) Optional label for the API key
- `limit` - (Optional) Credit limit for the API key
- `disabled` - (Optional) Whether the API key is disabled (default: false)
- `include_byok_in_limit` - (Optional) Whether to include BYOK usage in the limit

#### Attribute Reference

- `hash` - The hash identifier of the API key
- `key_value` - The actual API key value (only available on creation, sensitive)
- `usage` - Current usage of the API key
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

## Use Cases

### SaaS Applications

Automatically create unique API keys for each customer instance:

```hcl
resource "openrouter_api_key" "customer" {
  for_each = var.customers

  name  = "Customer ${each.key}"
  label = "customer-${each.key}"
  limit = each.value.credit_limit
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
  label = "prod-${formatdate("YYYY-MM", timestamp())}"
  limit = var.production_limit

  lifecycle {
    create_before_destroy = true
  }
}

# Disable old key when new one is created
resource "openrouter_api_key" "old_key" {
  count = var.disable_old_key ? 1 : 0

  name     = "Production Key ${formatdate("YYYY-MM", timeadd(timestamp(), "-720h"))}"
  disabled = true

  depends_on = [openrouter_api_key.rotated_key]
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
