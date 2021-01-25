variable "hvn_id" {
  description = "The ID of the HCP HVN."
  type        = string
}

variable "region" {
  description = "The region of the HCP HVN and Consul cluster."
  type        = string
}

variable "cloud_provider" {
  description = "The cloud provider of the HCP HVN and Consul cluster."
  type        = string
}