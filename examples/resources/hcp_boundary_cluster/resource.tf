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
  auth_token_time_to_live  = "36h0m0s"
  auth_token_time_to_stale = "12h0m0s"
}
