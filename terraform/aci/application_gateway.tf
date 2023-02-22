resource "azurerm_public_ip" "pub_ip_app_gw" {
  name                = "${var.api_name}-pub-ip"
  resource_group_name = data.azurerm_resource_group.resource_group.name
  location            = data.azurerm_resource_group.resource_group.location
  sku                 = "Standard"
  allocation_method   = "Static"
}

output "app_gw_public_ip" {
  value = azurerm_public_ip.pub_ip_app_gw.ip_address
}

resource "azurerm_lb" "app_gw" {
  name                = "${var.api_name}-app-gw"
  location            = data.azurerm_resource_group.resource_group.location
  resource_group_name = data.azurerm_resource_group.resource_group.name
  sku                 = "Standard"
  frontend_ip_configuration {
    name                 = "HttpServerFrontendPool"
    public_ip_address_id = azurerm_public_ip.pub_ip_app_gw.id
  }
  depends_on = [
    azurerm_public_ip.pub_ip_app_gw,
  ]
}

# backend pool
resource "azurerm_lb_backend_address_pool" "app_gw_backend_pool" {
  loadbalancer_id = azurerm_lb.app_gw.id
  name            = "HttpServerBackendPool"
  depends_on = [
    azurerm_lb.app_gw,
  ]
}

resource "azurerm_lb_backend_address_pool_address" "app_gw_backend_pool_address_1" {
  name                    = "app_gw_backend_pool_address_1"
  backend_address_pool_id = azurerm_lb_backend_address_pool.app_gw_backend_pool.id
  virtual_network_id      = azurerm_virtual_network.vnet.id
  ip_address              = azurerm_container_group.container_api.ip_address
  depends_on = [
    azurerm_lb_backend_address_pool.app_gw_backend_pool,
    azurerm_virtual_network.vnet,
    azurerm_container_group.container_api,
  ]
}

# Load balancing rules
resource "azurerm_lb_rule" "app_gw_rule_1" {
  loadbalancer_id                = azurerm_lb.app_gw.id
  name                           = "HTTPAccessRule"
  protocol                       = "Tcp"
  frontend_port                  = 80
  backend_port                   = var.api_port
  frontend_ip_configuration_name = "HttpServerFrontendPool"
  backend_address_pool_ids = [
    azurerm_lb_backend_address_pool.app_gw_backend_pool.id,
  ]
  probe_id = azurerm_lb_probe.app_gw_probe_1.id
  depends_on = [
    azurerm_lb_probe.app_gw_probe_1,
    azurerm_lb_backend_address_pool.app_gw_backend_pool,
  ]
}

# probes
resource "azurerm_lb_probe" "app_gw_probe_1" {
  name                = "healthProbe"
  loadbalancer_id     = azurerm_lb.app_gw.id
  protocol            = "Http"
  port                = var.api_port
  request_path        = var.api_health_path
  interval_in_seconds = "15"
  number_of_probes    = "3"
  depends_on = [
    azurerm_lb.app_gw,
  ]
}