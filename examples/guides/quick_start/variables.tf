variable "hvn_id" {
  description = "The ID of the HCP HVN."
  type        = string
}

variable "cloud_provider" {
  description = "The cloud provider of the HCP HVN, Peering connection, and Consul cluster."
  type        = string
}

variable "region" {
  description = "The region of the HCP HVN, Peering connection, and Consul cluster."
  type        = string
}