variable "cluster_id" {
  description = "The ID of the HCP Consul cluster."
  type        = string
}

variable "kubernetes_endpoint" {
  description = "The FQDN for the Kubernetes API."
  type        = string
}