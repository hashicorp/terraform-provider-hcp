resource "hcp_boundary_cluster" "example" {
  cluster_id = "boundary-cluster"
  username   = "test-user"
  password   = "Password123!"
  maintenance_window_config {
    day          = "TUESDAY"
    start        = 2
    end          = 12
    upgrade_type = "SCHEDULED"
  }
}
