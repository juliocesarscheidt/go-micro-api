locals {
  gateway_ip_configuration       = "appGatewayIpConfig"
  backend_address_pool_name      = "appGatewayBackendPool"
  frontend_port_name             = "appGatewayFrontendPort"
  frontend_ip_configuration_name = "appGatewayFrontendIP"
  http_setting_name              = "appGatewayHttpSettings"
  listener_name                  = "appGatewayListener"
  request_routing_rule_name      = "appGatewayRoutingRule"
}

resource "azurerm_application_gateway" "app_gw" {
  name                = "${var.api_name}-app-gw"
  resource_group_name = data.azurerm_resource_group.resource_group.name
  location            = data.azurerm_resource_group.resource_group.location
  sku {
    name     = "Standard_v2"
    tier     = "Standard_v2"
    capacity = 2
  }
  gateway_ip_configuration {
    name      = local.gateway_ip_configuration
    subnet_id = azurerm_subnet.subnet_a.id
  }
  frontend_port {
    name = local.frontend_port_name
    port = 80
  }
  frontend_ip_configuration {
    name                 = local.frontend_ip_configuration_name
    public_ip_address_id = azurerm_public_ip.pub_ip_app_gw.id
  }
  backend_address_pool {
    name         = local.backend_address_pool_name
    ip_addresses = tolist(azurerm_container_group.container_api[*].ip_address)
  }
  backend_http_settings {
    name                  = local.http_setting_name
    cookie_based_affinity = "Disabled"
    path                  = "/"
    port                  = var.api_port
    protocol              = "Http"
    request_timeout       = 60
    probe_name            = "healthProbe"
  }
  http_listener {
    name                           = local.listener_name
    frontend_ip_configuration_name = local.frontend_ip_configuration_name
    frontend_port_name             = local.frontend_port_name
    protocol                       = "Http"
  }
  probe {
    name                                      = "healthProbe"
    protocol                                  = "Http"
    path                                      = var.api_liveness_path
    interval                                  = 15
    timeout                                   = 10
    unhealthy_threshold                       = 5
    port                                      = var.api_port
    host                                      = "127.0.0.1"
    pick_host_name_from_backend_http_settings = false
  }
  request_routing_rule {
    name                       = local.request_routing_rule_name
    rule_type                  = "Basic"
    http_listener_name         = local.listener_name
    backend_address_pool_name  = local.backend_address_pool_name
    backend_http_settings_name = local.http_setting_name
    priority                   = 1000
  }
  depends_on = [
    azurerm_subnet.subnet_a,
    azurerm_public_ip.pub_ip_app_gw,
    azurerm_container_group.container_api,
  ]
}
