data "hcp_consul_agent_helm_config" "example" {
  cluster_id          = var.cluster_id
  kubernetes_endpoint = var.kubernetes_endpoint
}