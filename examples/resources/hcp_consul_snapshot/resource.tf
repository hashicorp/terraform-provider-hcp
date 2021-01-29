// Note: Snapshots currently have a retention policy of 30 days. After that time, any Terraform
// state refresh will note that a new snapshot resource will be created.
resource "hcp_consul_snapshot" "example" {
  cluster_id    = "consul-cluster"
  snapshot_name = "my-snapshot"
}