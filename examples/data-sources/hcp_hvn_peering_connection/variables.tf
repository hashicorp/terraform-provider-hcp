variable "peering_id" {
  description = "The ID of the network peering."
  type        = string
}

variable "hvn_1" {
  description = "The unique URL of one of the HVNs being peered."
  type        = string
}

variable "hvn_2" {
  description = "The unique URL of one of the HVNs being peered."
  type        = string
}
