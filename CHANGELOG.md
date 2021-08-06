## 0.13.0 (Unreleased)

FEATURES

* **New data source** `packer_image_iteration` [GH-169]

## 0.12.0 (August 04, 2021)

FEATURES:

* resource/hcp_vault_cluster: Add `starter_small` cluster tier ([#178](https://github.com/hashicorp/terraform-provider-hcp/issues/178))

IMPROVEMENTS:

* provider: Bump `terraform-plugin-sdk/v2` dependency ([#157](https://github.com/hashicorp/terraform-provider-hcp/issues/157))
* provider: Bump `go-openapi/runtime` dependency ([#140](https://github.com/hashicorp/terraform-provider-hcp/issues/140))
* provider: Bump `google/uuid` dependency ([#164](https://github.com/hashicorp/terraform-provider-hcp/issues/164))
* docs: Update Consul docs to include hcp_hvn_peering_connection ([#176](https://github.com/hashicorp/terraform-provider-hcp/issues/176))

## 0.11.0 (July 30, 2021)

FEATURES:

* **New resource** `hcp_hvn_peering_connection` ([#156](https://github.com/hashicorp/terraform-provider-hcp/issues/156))
* resource/hcp_consul_cluster: Consul federation released as Generally Available ([#171](https://github.com/hashicorp/terraform-provider-hcp/issues/171))

## 0.10.0 (July 15, 2021)

FIXES:

* resource/hcp_consul_cluster: Fix terraform updates for min_consul_version ([#161](https://github.com/hashicorp/terraform-provider-hcp/pull/161))

IMPROVEMENTS:

* docs: Add CIDR guidance to HVN resource documentation ([#160](https://github.com/hashicorp/terraform-provider-hcp/pull/160))
* docs: Add design doc on networking resources ([#159](https://github.com/hashicorp/terraform-provider-hcp/pull/159))

## 0.9.0 (June 30, 2021)

IMPROVEMENTS:

* resource/hcp_vault_cluster: add update functionality to Vault cluster ([#152](https://github.com/hashicorp/terraform-provider-hcp/pull/152))
* docs: updates to Consul root token doc ([#153](https://github.com/hashicorp/terraform-provider-hcp/pull/153))
* resource/hcp_consul_cluster: add auto peering for Consul Federation ([#154](https://github.com/hashicorp/terraform-provider-hcp/pull/154))

## 0.8.0 (June 18, 2021)

⚠️ Note: This version fixes a bug where the Consul and Vault clusters' `*_endpoint_url` outputs did not return complete URLs. This may result in a breaking change for existing clusters whose endpoint URLs are already adjusted to be a full URL with string helpers.
Please remove any functions that adjust the output of the `vault_private_endpoint_url`, `vault_public_endpoint_url`, `consul_private_endpoint_url`, and `consul_public_endpoint_url` when upgrading to this version. ⚠️

For example, your Vault provider configuration might need to change:

```hcl
# before
provider "vault" {
  address = join("", ["https://", hcp_vault_cluster.example.vault_public_endpoint_url, ":8200"])
}

# after
provider "vault" {
  address = hcp_vault_cluster.example.vault_public_endpoint_url
}
```

IMPROVEMENTS:

* resource/hcp_vault_cluster: `tier` is now an optional input, with the options `dev`, `standard_small`, `standard_medium`, and `standard_large` ([#144](https://github.com/hashicorp/terraform-provider-hcp/pull/144)) (our first open-source contribution - thanks @waxb!)
* resource/hcp_consul_cluster: `plus` is now available as a `tier` option ([#148](https://github.com/hashicorp/terraform-provider-hcp/pull/148))
* tests: expands acceptance test coverage to data sources and dependent resources ([#135](https://github.com/hashicorp/terraform-provider-hcp/pull/135), [#142](https://github.com/hashicorp/terraform-provider-hcp/pull/142), [#150](https://github.com/hashicorp/terraform-provider-hcp/pull/150))

BREAKING CHANGES:

* resource/hcp_consul_cluster: returns complete endpoint URLs ([#145](https://github.com/hashicorp/terraform-provider-hcp/pull/145))
* resource/hcp_vault_cluster: returns complete endpoint URLs ([#145](https://github.com/hashicorp/terraform-provider-hcp/pull/145))

## 0.7.0 (June 07, 2021)

⚠️ Note: This version contains breaking changes to the `hcp_aws_transit_gateway_attachment` and `hcp_aws_network_peering` resources and data sources. Please pin to the previous version and follow [this migration guide](https://github.com/hashicorp/terraform-provider-hcp/pull/128) when you're ready to migrate. ⚠️

FEATURES:

* **New resource** `hcp_hvn_route` ([#122](https://github.com/hashicorp/terraform-provider-hcp/pull/122))

IMPROVEMENTS:

* resource/hcp_aws_transit_gateway_attachment: released as Generally Available ([#121](https://github.com/hashicorp/terraform-provider-hcp/pull/121))

BREAKING CHANGES:

* resource/hcp_aws_network_peering: now requires `peering_id` to be specified and doesn't accept `peer_vpc_cidr_block` as input ([#128](https://github.com/hashicorp/terraform-provider-hcp/pull/128))
* datasource/hcp_aws_network_peering: no longer returns `peer_vpc_cidr_block` as output ([#128](https://github.com/hashicorp/terraform-provider-hcp/pull/128))
* resource/hcp_aws_transit_gateway_attachment: doesn't accept `destination_cidrs` as input ([#128](https://github.com/hashicorp/terraform-provider-hcp/pull/128))
* datasource/hcp_aws_transit_gateway_attachment: no longer returns `destination_cidrs` as output ([#128](https://github.com/hashicorp/terraform-provider-hcp/pull/128))

## 0.6.1 (June 03, 2021)

IMPROVEMENTS:

* docs: updates banner on index page to warn of upcoming breaking changes ([#134](https://github.com/hashicorp/terraform-provider-hcp/pull/134))
* resource/hcp_consul_cluster_snapshot_test: add Consul cluster snapshot acceptance test ([#126](https://github.com/hashicorp/terraform-provider-hcp/pull/126))

## 0.6.0 (May 10, 2021)

FEATURES:

* **New data source** `hcp_hvn_route` ([#115](https://github.com/hashicorp/terraform-provider-hcp/pull/115))

IMPROVEMENTS:

* provider: Bump `hcp-go-sdk` dependency ([#105](https://github.com/hashicorp/terraform-provider-hcp/pull/105))
* provider: Bump `go-openapi/runtime` dependency ([#106](https://github.com/hashicorp/terraform-provider-hcp/pull/106))
* resource/hvn, peering, tgw attachment: added `self_link` output ([#111](https://github.com/hashicorp/terraform-provider-hcp/pull/111))
* docs: various doc updates ([#117](https://github.com/hashicorp/terraform-provider-hcp/pull/117), [#119](https://github.com/hashicorp/terraform-provider-hcp/pull/119))

## 0.5.0 (April 20, 2021)

IMPROVEMENTS:

* provider: Upgraded to Go 1.16. Binary releases of this provider now include the darwin-arm64 platform ([#104](https://github.com/hashicorp/terraform-provider-hcp/pull/104), [#108](https://github.com/hashicorp/terraform-provider-hcp/pull/108))
* provider: Bump `terraform-plugin-sdk/v2` dependency ([#86](https://github.com/hashicorp/terraform-provider-hcp/pull/86))
* provider: Bump `go-openapi/runtime` dependency ([#81](https://github.com/hashicorp/terraform-provider-hcp/pull/81))
* provider: Bump `terraform-plugin-docs` dependency ([#55](https://github.com/hashicorp/terraform-provider-hcp/pull/55))
* provider: Bump `go-openapi/strfmt` dependency ([#99](https://github.com/hashicorp/terraform-provider-hcp/pull/99))
* docs: Add warnings ([#102](https://github.com/hashicorp/terraform-provider-hcp/pull/102))
* resource/consul_cluster: Fixed Consul cluster acceptance test ([#103](https://github.com/hashicorp/terraform-provider-hcp/pull/103))

## 0.4.1 (April 09, 2021)

FIXES:

* resource/consul_cluster: Set "computed=true" option for the vm size ([#100](https://github.com/hashicorp/terraform-provider-hcp/pull/100))

## 0.4.0 (April 07, 2021)

⚠️ Note: There is an issue with this version of the HCP Provider in which existing Consul clusters that do not specify size will be recommended by Terraform to be recreated on the next terraform apply, resulting in potential data loss. Please upgrade to the patch v0.4.1 or beyond to avoid this issue. ⚠️

FEATURES:

* **New resource** `hcp_vault_cluster` ([#97](https://github.com/hashicorp/terraform-provider-hcp/pull/97))
* **New resource** `hcp_vault_cluster_admin_token` ([#97](https://github.com/hashicorp/terraform-provider-hcp/pull/97))

IMPROVEMENTS:

* all: Log import ID used when an import fails due to parsing ([#82](https://github.com/hashicorp/terraform-provider-hcp/pull/82))
* all: Add comment to clarify that Links can be sent in API requests ([#82](https://github.com/hashicorp/terraform-provider-hcp/pull/82))
* ci: Add github checks ([#90](https://github.com/hashicorp/terraform-provider-hcp/pull/90))
* docs: Add pull request lifecycle docs ([#89](https://github.com/hashicorp/terraform-provider-hcp/pull/89))
* docs: Add issue lifecycle docs ([#93](https://github.com/hashicorp/terraform-provider-hcp/pull/93))
* datasource/consul_agent_helm_config: Remove extraneous protocol from FQDN string ([#95](https://github.com/hashicorp/terraform-provider-hcp/pull/95))
* resource/consul_cluster: Add VM size to Consul cluster ([#77](https://github.com/hashicorp/terraform-provider-hcp/pull/77))
* resource/aws_network_peering: Update comments, docs, and messages to use correct capitalization for network peering ([#82](https://github.com/hashicorp/terraform-provider-hcp/pull/82))
* resource/aws_network_peering: Update peering wait function to use helper ([#82](https://github.com/hashicorp/terraform-provider-hcp/pull/82))

FIXES:

* all: Ensure context is being passed for all HCP API calls ([#82](https://github.com/hashicorp/terraform-provider-hcp/pull/82))

## 0.3.0 (March 25, 2021)

IMPROVEMENTS:

* all: Improve error messages for requests made to all HCP services ([#83](https://github.com/hashicorp/terraform-provider-hcp/pull/83))
* ci: Run unit tests instead of acceptance tests on Pull Requests ([#73](https://github.com/hashicorp/terraform-provider-hcp/pull/73))
* docs: Add contribution guidelines ([#71](https://github.com/hashicorp/terraform-provider-hcp/pull/71))
* docs: Update contribution docs to include guidance on acceptance tests ([#79](https://github.com/hashicorp/terraform-provider-hcp/pull/79))
* docs: Add CODEOWNERS ([#76](https://github.com/hashicorp/terraform-provider-hcp/pull/76))
* docs: Add PR template ([#80](https://github.com/hashicorp/terraform-provider-hcp/pull/80))
* provider: Bump `hcp-go-sdk` dependency ([#83](https://github.com/hashicorp/terraform-provider-hcp/pull/83))
* provider: Bump `uuid` dependency ([#49](https://github.com/hashicorp/terraform-provider-hcp/pull/49))
* provider: Bump `testify` dependency ([#51](https://github.com/hashicorp/terraform-provider-hcp/pull/51))
* resource/hcp_consul_cluster: Add basic acceptance test ([#78](https://github.com/hashicorp/terraform-provider-hcp/pull/78))
* resource/hcp_hvn: Add basic acceptance test ([#74](https://github.com/hashicorp/terraform-provider-hcp/pull/74))

## 0.2.0 (February 22, 2021)

FEATURES:

* **New data source** `hcp_aws_transit_gateway_attachment` ([#58](https://github.com/hashicorp/terraform-provider-hcp/pull/58))
* **New data source** `hcp_consul_versions` ([#63](https://github.com/hashicorp/terraform-provider-hcp/pull/63))
* **New resource** `hcp_aws_transit_gateway_attachment` [(#58](https://github.com/hashicorp/terraform-provider-hcp/pull/58))

IMPROVEMENTS:

* all: Improve error messages for requests made to the Consul service ([#68](https://github.com/hashicorp/terraform-provider-hcp/pull/68))
* data-source/hcp_consul_cluster: Add HCP Consul federation support ([#68](https://github.com/hashicorp/terraform-provider-hcp/pull/68))
* resource/hcp_aws_transit_gateway_attachment: Support resource import ([#64](https://github.com/hashicorp/terraform-provider-hcp/pull/64))
* resource/hcp_consul_cluster: Add HCP Consul federation support ([#68](https://github.com/hashicorp/terraform-provider-hcp/pull/68))

BUGS:

* all: Set resource id before polling operation and re-create failed deployments ([#59](https://github.com/hashicorp/terraform-provider-hcp/pull/59))
* resource/hcp_consul_cluster: Validate Consul datacenter and lowercase the default ([#57](https://github.com/hashicorp/terraform-provider-hcp/pull/57))

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
