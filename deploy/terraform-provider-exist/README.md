# Terraform Provider for exist.io

The official Terraform & OpenTofu provider for managing infrastructure, security rulesets, network allowlists, and repositories on self-hosted `exist.io` enterprise clusters.

## Features Managed via Infrastructure-as-Code

- **`exist_namespace`**: Organization & team namespaces.
- **`exist_repository`**: Bare Git repositories with default branch configurations, merge policies (squash, rebase, merge commits), and branch pruning.
- **`exist_ruleset`**: Enterprise branch protection policies:
  - Enforce signed commits (GPG & OpenSSH)
  - Require pull request reviews & minimum approvals
  - Prevent branch deletions and force pushes
  - Require status check runs (CI / security scanners)
- **`exist_ip_allowlist`**: Network perimeter IP CIDRs (VPNs, office egress).
- **`exist_ssh_ca`**: Enrolled enterprise SSH Certificate Authorities for keyless developer authentication.

## Authentication

Set the provider credentials via arguments or environment variables:

```bash
export EXIST_ENDPOINT="https://git.internal.corp"
export EXIST_TOKEN="exist_pat_01k..."
```

## Running with OpenTofu / Terraform

```bash
cd deploy/terraform-provider-exist/examples
tofu init
tofu plan
tofu apply
```
