variable "exist_endpoint" {
  description = "Base URL of the self-hosted exist.io forge instance"
  type        = string
  default     = "http://127.0.0.1:8080"
}

variable "exist_token" {
  description = "Personal access token or admin API key with full platform management scopes"
  type        = string
  sensitive   = true
}
