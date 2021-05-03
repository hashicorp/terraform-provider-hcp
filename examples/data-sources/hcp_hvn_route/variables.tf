variable "hvn" {
  description = "The `self_link` of the HashiCorp Virtual Network (HVN)."
  type        = string
}

variable "destination_cidr" {
  description = "The destination CIDR of the HVN route."
  type        = string
}
