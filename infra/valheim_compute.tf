
resource "random_string" "suffix" {
  length  = 6
  upper   = false
  special = false
}

resource "azurerm_storage_account" "world" {
  name                     = "valheim${random_string.suffix.result}"
  resource_group_name      = azurerm_resource_group.rg.name
  location                 = azurerm_resource_group.rg.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_share" "world" {
  name                 = "world"
  storage_account_name = azurerm_storage_account.world.name
  quota                = 10 # in GB
}

resource "azurerm_virtual_network" "compute" {
  name                = "valheim-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
}

resource "azurerm_subnet" "compute" {
  name                 = "valheim-subnet"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.compute.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_network_security_group" "compute" {
  name                = "valheim-nsg"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  security_rule {
    name                       = "allow_valheim_udp"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Udp"
    source_port_range          = "*"
    destination_port_ranges    = ["2456-2458"]
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "allow_ssh"
    priority                   = 200
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_ranges    = ["22"]
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

resource "azurerm_subnet_network_security_group_association" "subnet_nsg_assoc" {
  subnet_id                 = azurerm_subnet.compute.id
  network_security_group_id = azurerm_network_security_group.compute.id
}

resource "random_string" "valheim_password" {
  length = 6
}

data "template_file" "cloud_init" {
  template = file("${path.module}/cloud-init.yml")

  vars = {
    valheim_worlds_storage_account_name = azurerm_storage_account.world.name
    valheim_worlds_storage_account_key  = azurerm_storage_account.world.primary_access_key
    world_share_name                    = azurerm_storage_share.world.name
    server_name                         = var.server_name
    world_name                          = var.world_name
    server_pass                         = random_string.valheim_password.result
    events_storage_account_name         = azurerm_storage_account.godinbot.name
    events_queue_name                   = azurerm_storage_queue.events.name
  }
}

# Virtual Machine Scale Set
resource "azurerm_linux_virtual_machine_scale_set" "compute" {
  name                = "valheim-server-vmss"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Standard_D2as_v4"
  instances           = 0
  admin_username      = "azureuser"

  disable_password_authentication = true
  admin_ssh_key {
    username   = "azureuser"
    public_key = file("valheim.key.pub")
  }

  identity {
    type = "SystemAssigned"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts"
    version   = "latest"
  }

  upgrade_mode = "Manual"

  priority        = "Spot"
  eviction_policy = "Delete"
  max_bid_price   = -1

  os_disk {
    caching              = "ReadOnly"
    storage_account_type = "Standard_LRS"
    diff_disk_settings {
      option = "Local"
    }
  }

  network_interface {
    name                      = "vmss-nic"
    primary                   = true
    network_security_group_id = azurerm_network_security_group.compute.id

    ip_configuration {
      name      = "vmss-ipconfig"
      primary   = true
      subnet_id = azurerm_subnet.compute.id

      public_ip_address {
        name = "instance-public-ip"
      }
    }
  }

  custom_data = base64encode(data.template_file.cloud_init.rendered)

  depends_on = [azurerm_subnet_network_security_group_association.subnet_nsg_assoc]
}

output "cloud_init" {
  value = data.template_file.cloud_init.rendered
}


