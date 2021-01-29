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

variable "peer_vpc_region" {
  description = "The region of the peer VPC."
  type        = string
}

variable "peer_id" {
  description = "The ID to use for the HCP Network peering."
  type        = string
}
