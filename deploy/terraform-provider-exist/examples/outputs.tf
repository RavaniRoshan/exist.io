output "namespace_id" {
  description = "Assigned unique ULID identifier of the created namespace"
  value       = exist_namespace.platform.id
}

output "repository_clone_url_http" {
  description = "Smart HTTP clone URL for the created repository"
  value       = "${var.exist_endpoint}/${exist_namespace.platform.slug}/${exist_repository.payment_gateway.name}.git"
}

output "repository_clone_url_ssh" {
  description = "Pure-Go SSH clone URL for the created repository"
  value       = "git@${replace(replace(var.exist_endpoint, "http://", ""), "https://", "")}:${exist_namespace.platform.slug}/${exist_repository.payment_gateway.name}.git"
}
