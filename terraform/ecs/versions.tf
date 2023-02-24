terraform {
  required_version = ">= 1.2"
  required_providers {
    aws = {
      source  = "registry.terraform.io/hashicorp/aws"
      version = "=4.55.0"
    }
  }
}
