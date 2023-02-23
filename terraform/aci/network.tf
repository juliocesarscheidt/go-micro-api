resource "azurerm_virtual_network" "vnet" {
  name                = "${var.api_name}-vnet"
  location            = data.azurerm_resource_group.resource_group.location
  resource_group_name = data.azurerm_resource_group.resource_group.name
  address_space       = ["10.100.0.0/20"]
  depends_on          = []
}

resource "azurerm_subnet" "subnet_a" {
  name                 = "${var.api_name}-subnet-a"
  virtual_network_name = azurerm_virtual_network.vnet.name
  resource_group_name  = data.azurerm_resource_group.resource_group.name
  address_prefixes     = ["10.100.0.0/24"]
  depends_on = [
    azurerm_virtual_network.vnet,
  ]
}

resource "azurerm_subnet" "subnet_b" {
  name                 = "${var.api_name}-subnet-b"
  virtual_network_name = azurerm_virtual_network.vnet.name
  resource_group_name  = data.azurerm_resource_group.resource_group.name
  address_prefixes     = ["10.100.1.0/24"]
  delegation {
    name = "delegation"
    service_delegation {
      name = "Microsoft.ContainerInstance/containerGroups"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/action",
      ]
    }
  }
  depends_on = [
    azurerm_virtual_network.vnet,
  ]
}

