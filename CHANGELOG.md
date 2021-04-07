## 0.4.0 (Unreleased)

FEATURES:

- **New resource** `hcp_vault_cluster` (#97)
- **New resource** `hcp_vault_cluster_admin_token` (#97)

IMPROVEMENTS:

- all: Log import ID used when an import fails due to parsing (#82)
- all: Add comment to clarify that Links can be sent in API requests (#82)
- ci: Add github checks (#90)
- docs: Add pull request lifecycle docs (#89)
- docs: Add issue lifecycle docs (#93)
- datasource/consul_agent_helm_config: Remove extraneous protocol from FQDN string (#95)
- resource/consul_cluster: Add VM size to Consul cluster (#77)
- resource/aws_network_peering: Update comments, docs, and messages to use correct capitalization for network peering (#82)
- resource/aws_network_peering: Update peering wait function to use helper (#82)

FIXES:
- all: Ensure context is being passed for all HCP API calls (#82)

## 0.3.0 (March 25, 2021)

IMPROVEMENTS:
* all: Improve error messages for requests made to all HCP services (#83)
* ci: Run unit tests instead of acceptance tests on Pull Requests (#73)
* docs: Add contribution guidelines (#71)
* docs: Update contribution docs to include guidance on acceptance tests (#79)
* docs: Add CODEOWNERS (#76)
* docs: Add PR template (#80)
* provider: Bump `hcp-go-sdk` dependency (#83)
* provider: Bump `uuid` dependency (#49)
* provider: Bump `testify` dependency (#51)
* resource/hcp_consul_cluster: Add basic acceptance test (#78)
* resource/hcp_hvn: Add basic acceptance test (#74)

## 0.2.0 (February 22, 2021)

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
