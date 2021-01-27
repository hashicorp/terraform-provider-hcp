variable "hvn_id" {
  description = "The ID of the HCP HVN."
  type        = string
}

variable "cluster_id" {
  description = "The ID of the HCP Consul cluster."
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

variable "snapshot_name" {
  description = "The name of the Consul snapshot."
  type        = string
}