locals {
  bot_common_name = "godindiscbot"
  custom_location = "eastus"
}

resource "azurerm_log_analytics_workspace" "godinbot" {
  resource_group_name = azurerm_resource_group.rg.name
  name                = local.bot_common_name
  location            = local.custom_location
}

resource "azurerm_application_insights" "godinbot" {
  application_type    = "web"
  resource_group_name = azurerm_resource_group.rg.name
  location            = local.custom_location
  name                = local.bot_common_name
  workspace_id        = azurerm_log_analytics_workspace.godinbot.id
}

resource "azurerm_storage_account" "godinbot" {
  name                     = local.bot_common_name
  account_kind             = "Storage"
  account_replication_type = "LRS"
  account_tier             = "Standard"
  resource_group_name      = azurerm_resource_group.rg.name
  location                 = local.custom_location
}

resource "azurerm_service_plan" "godinbot" {
  location            = local.custom_location
  name                = local.bot_common_name
  resource_group_name = azurerm_resource_group.rg.name
  sku_name            = "Y1"
  os_type             = "Windows"
}

resource "azurerm_windows_function_app" "godinbot" {
  app_settings = {
    SCALE_CONTROLLER_LOGGING_ENABLED = "AppInsights:Verbose"
    WEBSITE_MOUNT_ENABLED            = 1
    AZURE_SUBSCRIPTION_ID            = data.azurerm_client_config.current.subscription_id
    BASE64_SERVER_KEY                = var.base64_server_key
    DISCORD_BOT_TOKEN                = var.discord_bot_token
    DISCORD_CHANNEL_ID               = var.discord_channel_id
    DISCORD_PUBLIC_KEY               = var.discord_public_key
    STEAM_API_KEY                    = var.steam_api_key
    VMSS_NAME                        = azurerm_linux_virtual_machine_scale_set.compute.name
    VMSS_RESOURCE_GROUP_NAME         = azurerm_resource_group.rg.name
    VMSS_SUBSCRIPTION_ID             = data.azurerm_client_config.current.subscription_id
    WORLD_NAME                       = var.world_name
    STATE_STORAGE_NAME               = azurerm_storage_table.valheim_state.name
    FUNCTIONS_WORKER_RUNTIME         = "custom"
  }
  functions_extension_version = "~4"
  location                    = local.custom_location
  name                        = local.bot_common_name
  resource_group_name         = azurerm_resource_group.rg.name
  service_plan_id             = azurerm_service_plan.godinbot.id
  storage_account_access_key  = azurerm_storage_account.godinbot.primary_access_key
  storage_account_name        = azurerm_storage_account.godinbot.name
  tags = {
    "hidden-link: /app-insights-conn-string"         = azurerm_application_insights.godinbot.connection_string
    "hidden-link: /app-insights-instrumentation-key" = azurerm_application_insights.godinbot.instrumentation_key
    "hidden-link: /app-insights-resource-id"         = azurerm_application_insights.godinbot.id
  }
  identity {
    type = "SystemAssigned"
  }
  site_config {
    application_insights_key = azurerm_application_insights.godinbot.instrumentation_key
    ftps_state               = "FtpsOnly"
    application_stack {
      use_custom_runtime = true
    }
  }
  lifecycle {
    ignore_changes = [
      app_settings["WEBSITE_RUN_FROM_PACKAGE"]
    ]
  }
}

resource "azurerm_role_assignment" "funcapp_2_vmss" {
  role_definition_name = "Virtual Machine Contributor"
  scope                = azurerm_linux_virtual_machine_scale_set.compute.id
  principal_id         = azurerm_windows_function_app.godinbot.identity[0].principal_id
}

resource "azurerm_storage_queue" "events" {
  name                 = "events"
  storage_account_name = azurerm_storage_account.godinbot.name
}

resource "azurerm_role_assignment" "vmss_2_events_queue" {
  scope                = azurerm_storage_account.godinbot.id
  principal_id         = azurerm_linux_virtual_machine_scale_set.compute.identity[0].principal_id
  role_definition_name = "Storage Queue Data Contributor"
}

resource "azurerm_storage_table" "valheim_state" {
  name                 = "valheimstate"
  storage_account_name = azurerm_storage_account.godinbot.name
}
