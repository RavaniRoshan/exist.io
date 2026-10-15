terraform {
  required_version = ">= 1.5.0"
  required_providers {
    exist = {
      source  = "exist-io/exist"
      version = "~> 1.0.0"
    }
  }
}

provider "exist" {
  endpoint = var.exist_endpoint
  token    = var.exist_token
}

# 1. Enterprise Namespace
resource "exist_namespace" "platform" {
  slug = "enterprise-platform"
  name = "Enterprise Platform Engineering"
  type = "team"
}

# 2. Production Service Repository
resource "exist_repository" "payment_gateway" {
  namespace_id        = exist_namespace.platform.id
  name                = "payment-gateway"
  default_branch      = "main"
  visibility          = "private"
  allow_merge_commit  = true
  allow_squash_merge  = true
  delete_branch_on_merge = true
}

# 3. Enterprise Branch Protection Ruleset
resource "exist_ruleset" "protect_production" {
  namespace_slug     = exist_namespace.platform.slug
  repository_name    = exist_repository.payment_gateway.name
  name               = "require-signed-commits-and-reviews"
  enforcement        = "active"
  kind               = "push_branch"
  included_ref_names = ["refs/heads/main"]

  rule {
    type = "require_pull_request"
    parameters = {
      required_approvals = 2
    }
  }

  rule {
    type = "require_signed_commits"
  }

  rule {
    type = "prevent_force_push"
  }

  rule {
    type = "require_status_checks"
    parameters = {
      required_status_checks = ["ci/security-scan", "ci/unit-tests"]
    }
  }
}

# 4. Network Perimeter: Corporate CIDR Allowlist
resource "exist_ip_allowlist" "corp_vpn" {
  namespace_slug = exist_namespace.platform.slug
  cidr           = "10.0.0.0/16"
  description    = "Corporate Office & Global VPN Egress"
}

# 5. Enrolled Enterprise SSH Certificate Authority
resource "exist_ssh_ca" "corporate_ca" {
  namespace_slug = exist_namespace.platform.slug
  name           = "Corp-Identity-SSH-CA"
  public_key     = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKgV38qH9h6mQp2jK7lZ8v9u1w2x3y4z5A6B7C8D9E0F corp-ca@enterprise.internal"
}
