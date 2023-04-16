resource "azurerm_container_group" "container_api" {
  count               = var.api_replicas_count
  name                = "${var.api_name}-${count.index}"
  location            = data.azurerm_resource_group.resource_group.location
  resource_group_name = data.azurerm_resource_group.resource_group.name
  ip_address_type     = "Private"
  subnet_ids          = [azurerm_subnet.subnet_b.id]
  os_type             = "Linux"
  restart_policy      = "Always"
  image_registry_credential {
    server   = "${var.registry_username}.azurecr.io"
    username = var.registry_username
    password = var.registry_password
  }
  exposed_port = [
    {
      port     = var.api_port
      protocol = "TCP"
    }
  ]
  diagnostics {
    log_analytics {
      workspace_id  = azurerm_log_analytics_workspace.log_analytics_workspace.workspace_id
      workspace_key = azurerm_log_analytics_workspace.log_analytics_workspace.primary_shared_key
      # log_type      = "ContainerInstanceLogs" # for some reason, this doesn't work with Terraform
    }
  }
  container {
    name   = var.api_name
    image  = "${var.registry_username}.azurecr.io/${var.api_name}:${var.api_version}"
    cpu    = var.api_cpu
    memory = var.api_memory
    liveness_probe {
      http_get {
        path   = var.api_liveness_path
        port   = var.api_port
        scheme = "Http"
      }
      initial_delay_seconds = 10
      period_seconds        = 15
      timeout_seconds       = 10
      failure_threshold     = 5
      success_threshold     = 2
    }
    environment_variables = {
      MESSAGE     = var.api_message
      ENVIRONMENT = var.api_environment
    }
    ports {
      port     = var.api_port
      protocol = "TCP"
    }
  }
  tags = {
    "Name"    = "${var.api_name}-${count.index}"
    "Cluster" = var.api_name
  }
  depends_on = [
    azurerm_subnet.subnet_b,
    azurerm_log_analytics_workspace.log_analytics_workspace,
  ]
}
