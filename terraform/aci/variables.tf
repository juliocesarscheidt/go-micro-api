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

variable "api_replicas_count" {
  type        = number
  default     = 1
  description = "API replicas count"
}

variable "api_liveness_path" {
  type        = string
  default     = "/api/v1/health/live"
  description = "API liveness path"
}

variable "api_message" {
  type        = string
  default     = "Hello World From ACI with Terraform"
  description = "API variable MESSAGE"
}

variable "api_environment" {
  type        = string
  default     = "production"
  description = "API variable ENVIRONMENT"
}

variable "api_cpu" {
  type        = string
  default     = "0.5" # 0.5 vCPU
  description = "API CPU Amount"
}

variable "api_memory" {
  type        = string
  default     = "1" # 1GiB => 1024MiB
  description = "API Memory Amount"
}

variable "api_port" {
  type        = number
  default     = 9000
  description = "API TCP Port"
}

variable "registry_username" {
  type        = string
  description = "Registry Username"
  sensitive   = true
}

variable "registry_password" {
  type        = string
  description = "Registry Password"
  sensitive   = true
}
