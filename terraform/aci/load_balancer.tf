# resource "azurerm_lb" "app_gw" {
#   name                = "${var.api_name}-app-gw"
#   location            = data.azurerm_resource_group.resource_group.location
#   resource_group_name = data.azurerm_resource_group.resource_group.name
#   sku                 = "Standard"
#   frontend_ip_configuration {
#     name                 = "appGatewayFrontendIP"
#     public_ip_address_id = azurerm_public_ip.pub_ip_app_gw.id
#   }
#   depends_on = [
#     azurerm_public_ip.pub_ip_app_gw,
#   ]
# }

# # backend pool
# resource "azurerm_lb_backend_address_pool" "app_gw_backend_pool" {
#   loadbalancer_id = azurerm_lb.app_gw.id
#   name            = "appGatewayBackendPool"
#   depends_on = [
#     azurerm_lb.app_gw,
#   ]
# }

# resource "azurerm_lb_backend_address_pool_address" "app_gw_backend_pool_address" {
#   name                    = "appGatewayBackendPoolAddress"
#   backend_address_pool_id = azurerm_lb_backend_address_pool.app_gw_backend_pool.id
#   virtual_network_id      = azurerm_virtual_network.vnet.id
#   ip_address              = azurerm_container_group.container_api.ip_address
#   depends_on = [
#     azurerm_lb_backend_address_pool.app_gw_backend_pool,
#     azurerm_virtual_network.vnet,
#     azurerm_container_group.container_api,
#   ]
# }

# # probes
# resource "azurerm_lb_probe" "app_gw_health_probe" {
#   name                = "healthProbe"
#   loadbalancer_id     = azurerm_lb.app_gw.id
#   protocol            = "Http"
#   port                = var.api_port
#   request_path        = var.api_health_path
#   interval_in_seconds = "15"
#   number_of_probes    = "5" # The number of failed probe attempts
#   depends_on = [
#     azurerm_lb.app_gw,
#   ]
# }

# # Load balancing rules
# resource "azurerm_lb_rule" "app_gw_rule_1" {
#   name                           = "rule1"
#   loadbalancer_id                = azurerm_lb.app_gw.id
#   protocol                       = "Tcp"
#   frontend_port                  = 80
#   backend_port                   = var.api_port
#   frontend_ip_configuration_name = "appGatewayFrontendIP"
#   backend_address_pool_ids = [
#     azurerm_lb_backend_address_pool.app_gw_backend_pool.id,
#   ]
#   probe_id = azurerm_lb_probe.app_gw_health_probe.id
#   depends_on = [
#     azurerm_lb_probe.app_gw_health_probe,
#     azurerm_lb_backend_address_pool.app_gw_backend_pool,
#   ]
# }
