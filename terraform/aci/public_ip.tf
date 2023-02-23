resource "azurerm_public_ip" "pub_ip_app_gw" {
  name                = "${var.api_name}-pub-ip"
  resource_group_name = data.azurerm_resource_group.resource_group.name
  location            = data.azurerm_resource_group.resource_group.location
  sku                 = "Standard"
  allocation_method   = "Static"
}

output "public_ip" {
  value = azurerm_public_ip.pub_ip_app_gw.ip_address
}
