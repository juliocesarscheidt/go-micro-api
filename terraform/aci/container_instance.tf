resource "azurerm_container_group" "container_api" {
  name                = var.api_name
  location            = data.azurerm_resource_group.resource_group.location
  resource_group_name = data.azurerm_resource_group.resource_group.name
  ip_address_type     = "Private"
  subnet_ids          = [azurerm_subnet.subnet_container.id]
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
        path   = var.api_health_path
        port   = var.api_port
        scheme = "Http"
      }
      initial_delay_seconds = 10
      period_seconds        = 15
      failure_threshold     = 10
      success_threshold     = 1
      timeout_seconds       = 10
    }
    environment_variables = {
      MESSAGE = var.message
    }
    ports {
      port     = var.api_port
      protocol = "TCP"
    }
  }
  tags = {}
  depends_on = [
    azurerm_subnet.subnet_container,
    azurerm_log_analytics_workspace.log_analytics_workspace,
  ]
}