variable "resource_group" {
  type        = string
  default     = "go-micro-api-rg"
  description = "Resource Group"
}

variable "api_name" {
  type        = string
  default     = "go-micro-api"
  description = "API name"
}

variable "api_version" {
  type        = string
  default     = "v1.0.0"
  description = "API tag version"
}

variable "api_health_path" {
  type        = string
  default     = "/api/v1/health/live"
  description = "API health path"
}

variable "api_message" {
  type        = string
  default     = "Hello World From ACI with Terraform"
  description = "API environment variable MESSAGE"
}

variable "api_cpu" {
  type        = string
  default     = "1"
  description = "API CPU Amount"
}

variable "api_memory" {
  type        = string
  default     = "1"
  description = "API Memory Amount"
}

variable "api_port" {
  type        = number
  default     = 9000
  description = "API TCP Port"
}

variable "registry_username" {
  type        = string
  description = "ACR Registry Username"
  sensitive   = true
}

variable "registry_password" {
  type        = string
  description = "ACR Registry Password"
  sensitive   = true
}
