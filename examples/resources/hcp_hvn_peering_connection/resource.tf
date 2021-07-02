resource "hcp_hvn" "hvn_1" {
  hvn_id         = "hvn-1"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_hvn" "hvn_2" {
  hvn_id         = "hvn-2"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.18.16.0/20"
}

resource "hcp_hvn_peering_connection" "peer_1" {
  peering_id = "peer-1"
  hvn_1      = hcp_hvn.hvn_1.self_link
  hvn_2      = hcp_hvn.hvn_2.self_link
}
