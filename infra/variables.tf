variable "server_name" {
  description = "Name of the Valheim server"
  type        = string
  default     = "AzureValhalla"
}

variable "world_name" {
  description = "Name of the Valheim world"
  type        = string
  default     = "AzureValhalla"
}

variable "base64_server_key" {
  type        = string
  sensitive   = true
  description = "base64 pem key used to connect to valheim server and execute commands"
}

variable "discord_bot_token" {
  type        = string
  sensitive   = true
  description = "bot token so the function app can send messages reporting valheim server state changes"
}

variable "discord_public_key" {
  type        = string
  sensitive   = true
  description = "public key to verify the request came from the actual discord bot"
}

variable "steam_api_key" {
  type        = string
  sensitive   = true
  description = "steam api key so the bot can fetch user name when a new connection to the server is established"
}

variable "discord_channel_id" {
  type        = string
  sensitive   = false
  description = "id of the channel messages will be sent to"
}
