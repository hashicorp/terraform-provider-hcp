## 0.8.0 (Unreleased)

⚠️ Note: This version fixes a bug where the Consul and Vault clusters' `*_endpoint_url` outputs did not return complete URLs. This may result in issues for existing clusters whose endpoint urls are already adjusted by a workaround. ⚠️

FIXES:
* resource/consul_cluster: returns complete endpoint URLs (#145)
* resource/vault_cluster: returns complete endpoint URLs (#145)

## 0.7.0 (June 07, 2021)

⚠️ Note: This version contains breaking changes to the `hcp_aws_transit_gateway_attachment` and `hcp_aws_network_peering` resources and data sources. Please pin to the previous version and follow [this migration guide](https://github.com/hashicorp/terraform-provider-hcp/pull/128) when you're ready to migrate. ⚠️

FEATURES:
* **New resource** `hcp_hvn_route` (#122)

IMPROVEMENTS:
* resource/hcp_aws_transit_gateway_attachment: released as Generally Available (#121)

BREAKING CHANGES:
* resource/hcp_aws_network_peering: now requires `peering_id` to be specified and doesn't accept `peer_vpc_cidr_block` as input (#128)
* datasource/hcp_aws_network_peering: no longer returns `peer_vpc_cidr_block` as output (#128)
* resource/hcp_aws_transit_gateway_attachment: doesn't accept `destination_cidrs` as input (#128)
* datasource/hcp_aws_transit_gateway_attachment: no longer returns `destination_cidrs` as output (#128)

## 0.6.1 (June 03, 2021)

IMPROVEMENTS:
* docs: updates banner on index page to warn of upcoming breaking changes (#134)
* resource/hcp_consul_cluster_snapshot_test: add Consul cluster snapshot acceptance test (#126)

## 0.6.0 (May 10, 2021)

FEATURES:
* **New data source** `hcp_hvn_route` (#115)

IMPROVEMENTS:
* provider: Bump `hcp-go-sdk` dependency (#105)
* provider: Bump `go-openapi/runtime` dependency (#106)
* resource/hvn, peering, tgw attachment: added `self_link` output (#111)
* docs: various doc updates (#117, #119)

## 0.5.0 (April 20, 2021)

IMPROVEMENTS:
* provider: Upgraded to Go 1.16. Binary releases of this provider now include the darwin-arm64 platform (#104, #108)
* provider: Bump `terraform-plugin-sdk/v2` dependency (#86)
* provider: Bump `go-openapi/runtime` dependency (#81)
* provider: Bump `terraform-plugin-docs` dependency (#55)
* provider: Bump `go-openapi/strfmt` dependency (#99)
* docs: Add warnings (#102)
* resource/consul_cluster: Fixed Consul cluster acceptance test (#103)

## 0.4.1 (April 09, 2021)

FIXES:
* resource/consul_cluster: Set "computed=true" option for the vm size (#100)

## 0.4.0 (April 07, 2021)

⚠️ Note: There is an issue with this version of the HCP Provider in which existing Consul clusters that do not specify size will be recommended by Terraform to be recreated on the next terraform apply, resulting in potential data loss. Please upgrade to the patch v0.4.1 or beyond to avoid this issue. ⚠️

FEATURES:
* **New resource** `hcp_vault_cluster` (#97)
* **New resource** `hcp_vault_cluster_admin_token` (#97)

IMPROVEMENTS:
* all: Log import ID used when an import fails due to parsing (#82)
* all: Add comment to clarify that Links can be sent in API requests (#82)
* ci: Add github checks (#90)
* docs: Add pull request lifecycle docs (#89)
* docs: Add issue lifecycle docs (#93)
* datasource/consul_agent_helm_config: Remove extraneous protocol from FQDN string (#95)
* resource/consul_cluster: Add VM size to Consul cluster (#77)
* resource/aws_network_peering: Update comments, docs, and messages to use correct capitalization for network peering (#82)
* resource/aws_network_peering: Update peering wait function to use helper (#82)

FIXES:
* all: Ensure context is being passed for all HCP API calls (#82)

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
