## 0.2.0 (Unreleased)

FEATURES:
* **New data source** `hcp_aws_transit_gateway_attachment` (#58)
* **New data source** `hcp_consul_versions` (#63)
* **New resource** `hcp_aws_transit_gateway_attachment` (#58)

IMPROVEMENTS:
* all: Improve error messages for requests made to the Consul service (#68)
* data-source/hcp_consul_cluster: Add HCP Consul federation support (#68)
* resource/hcp_aws_transit_gateway_attachment: Support resource import (#64)
* resource/hcp_consul_cluster: Add HCP Consul federation support (#68)

BUGS:
* all: Set resource id before polling operation and re-create failed deployments (#59)
* resource/hcp_consul_cluster: Validate Consul datacenter and lowercase the default (#57)

## 0.1.0 (January 29, 2021)

FEATURES:
* **New resource** `hcp_hvn`.
* **New resource** `hcp_consul_cluster`.
* **New resource** `hcp_aws_network_peering`.
* **New resource** `hcp_consul_cluster_root_token`.
* **New resource** `hcp_consul_snapshot`.

* **New data source** `hcp_hvn`.
* **New data source** `hcp_consul_cluster`.
* **New data source** `hcp_aws_network_peering`.
* **New data source** `hcp_consul_cluster_root_token`.
* **New data source** `hcp_consul_snapshot`.
* **New data source** `hcp_consul_agent_helm_config`.
* **New data source** `hcp_consul_agent_kubernetes_secret`.
