## 0.31.0 (June 7, 2022)

IMPROVEMENTS:

* provider: Bump `google.golang.org/grpc` from 1.46.2 to 1.47.0 ([#316](https://github.com/hashicorp/terraform-provider-hcp/pull/316))
* provider: Bump `github.com/hashicorp/terraform-plugin-sdk/v2 from 2.16.0 to 2.17.0 ([#317](https://github.com/hashicorp/terraform-provider-hcp/pull/317))
* provider: Bump `github.com/hashicorp/terraform-plugin-docs` from 0.8.1 to 0.9.` ([#318](https://github.com/hashicorp/terraform-provider-hcp/pull/318))
* provider: Bump `github.com/stretchr/testify` from 1.7.1 to 1.7.2 ([#321](https://github.com/hashicorp/terraform-provider-hcp/pull/321))

FEATURES: 

* resource/vault: Enable metrics_config and audit_log_config ([#319](https://github.com/hashicorp/terraform-provider-hcp/pull/319))
* resource/consul_cluster: Adds Azure on Consul in public beta ([#320](https://github.com/hashicorp/terraform-provider-hcp/pull/320))


## 0.30.0 (May 26, 2022)

IMPROVEMENTS:

* provider: Bump `github.com/hashicorp/go-version` from 1.4.0 to 1.5.0 ([#313](https://github.com/hashicorp/terraform-provider-hcp/pull/313))
* resource/vault: Docs update secondary tier examples ([#289](https://github.com/hashicorp/terraform-provider-hcp/pull/289))

## 0.29.0 (May 18, 2022)

IMPROVEMENTS:

* provider: Bump `google.golang.org/grpc` from 1.46.0 to 1.46.2 ([#311](https://github.com/hashicorp/terraform-provider-hcp/issues/311))
* provider: Bump  `github.com/go-openapi/runtime` from 0.24.0 to 0.24.1 ([#306](https://github.com/hashicorp/terraform-provider-hcp/issues/306))

## 0.28.0 (May 11, 2022)

IMPROVEMENTS:

* resource/packer_image: bump hcp-sdk-go and fix packer import path ([#307](https://github.com/hashicorp/terraform-provider-hcp/issues/307))
* provider: Bump `terraform-plugin-sdk/v2` from 2.10.1 to 2.16.0 ([#309](https://github.com/hashicorp/terraform-provider-hcp/issues/309))
* provider: Bump `terraform-plugin-docs` from 0.7.0 to 0.8.1 ([#308](https://github.com/hashicorp/terraform-provider-hcp/issues/308))

## 0.27.0 (May 5, 2022)

⚠️ Note: To continue receiving warnings when HCP is reporting degraded performance or an outage, upgrade to this version. ⚠️

* provider: provider reports all HCP component statuses ([303](https://github.com/hashicorp/terraform-provider-hcp/issues/298))
* provider: Bump `actions/upload-artifact` from 2 to 3 ([#288](https://github.com/hashicorp/terraform-provider-hcp/issues/288))
* provider: Bump `google.golang.org/grpc` from 1.45.0 to 1.46.0 ([#296](https://github.com/hashicorp/terraform-provider-hcp/issues/296))
* provider: Bump `github.com/go-openapi/runtime` from 0.23.3 to 0.24.0 ([#300](https://github.com/hashicorp/terraform-provider-hcp/issues/300))
* docs: fix peer_vnet_region in azure_peering example ([303](https://github.com/hashicorp/terraform-provider-hcp/issues/298))
* docs: add contributors guide on breaking changes ([#294](https://github.com/hashicorp/terraform-provider-hcp/issues/294))

## 0.26.0 (April 14, 2022)

FIXES:

* provider: only warn on all platform outage statuses ([#290](https://github.com/hashicorp/terraform-provider-hcp/issues/290))

## 0.25.0 (April 05, 2022)

FEATURES:

* resource/vault_cluster: enable paths_filter and scaling in Plus-tier ([#281](https://github.com/hashicorp/terraform-provider-hcp/issues/281))

FIXES:

* datasource/hcp_packer_iteration: make sure test registry is plus ([#284](https://github.com/hashicorp/terraform-provider-hcp/issues/284))

IMPROVEMENTS:

* provider: Bump `actions/setup-go` from 2.2.0 to 3.0.0 ([#285](https://github.com/hashicorp/terraform-provider-hcp/issues/285))
* provider: Bump `actions/checkout` from 2.2.0 to 3.0.0 ([#285](https://github.com/hashicorp/terraform-provider-hcp/issues/285))
* provider: Bump `google.golang.org/grpc` from 1.44.0 to 1.45.0 ([#285](https://github.com/hashicorp/terraform-provider-hcp/issues/285))
* provider: Bump `terraform-plugin-docs` from 0.5.1 to 0.7.0 ([#285](https://github.com/hashicorp/terraform-provider-hcp/issues/285))

## 0.24.1 (March 23, 2022)

FIXES:

* docs: Remove beta notes from Packer data sources ([#278](https://github.com/hashicorp/terraform-provider-hcp/pull/278))

## 0.24.0 (March 09, 2022)

FEATURES:

* resource/vault_cluster: add support for performance replication in Plus tier clusters ([#266](https://github.com/hashicorp/terraform-provider-hcp/issues/266))

FIXES:

* resource/consul_cluster: Fix min_consul_version on creation not taking affect ([#252](https://github.com/hashicorp/terraform-provider-hcp/issues/252))

## 0.23.1 (March 03, 2022)

FIXES:

* datasource/hcp_packer_image: Remove check for revoked iterations ([#264](https://github.com/hashicorp/terraform-provider-hcp/issues/264))
* datasource/hcp_packer_iteration: Remove check for revoked iterations ([#264](https://github.com/hashicorp/terraform-provider-hcp/issues/264))
* datasource/hcp_packer_image_iteration: Remove check for revoked iterations ([#264](https://github.com/hashicorp/terraform-provider-hcp/issues/264))

## 0.23.0 (March 03, 2022)

:tada: Azure support is coming soon!

FEATURES:

* resource/consul_cluster: adds Azure on Consul (internal only)  ([#247](https://github.com/hashicorp/terraform-provider-hcp/issues/247))
* resource/azure_peering_connection: adds Azure peering resource (internal only)  ([#248](https://github.com/hashicorp/terraform-provider-hcp/issues/248))

FIXES:

* datasource/hcp_packer: Update tests to only set CloudProvider on CreateBuild ([#260](https://github.com/hashicorp/terraform-provider-hcp/issues/260))
* datasource/hcp_packer: Do not fail packer datasources for iteration with revoke_at set to the future ([#262](https://github.com/hashicorp/terraform-provider-hcp/issues/262))

IMPROVEMENTS:

* resource/aws_network_peering: add wait_for_active_state input ([#258](https://github.com/hashicorp/terraform-provider-hcp/issues/258))
* provider: Bump `actions/setup-go` from 2.1.4 to 2.2.0 ([#251](https://github.com/hashicorp/terraform-provider-hcp/issues/251))
* provider: Bump `github.com/go-openapi/strfmt` from 0.21.1 to 0.21.2 ([#253](https://github.com/hashicorp/terraform-provider-hcp/pull/253))
* provider: Bump `google.golang.org/grpc` from 1.42.0 to 1.44.0 ([#253](https://github.com/hashicorp/terraform-provider-hcp/pull/253))
* provider: Bump `github.com/hashicorp/go-version` from 1.3.0 to 1.4.0 ([#253](https://github.com/hashicorp/terraform-provider-hcp/pull/253))
* provider: Bump `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.10.0 to 2.10.1 ([#253](https://github.com/hashicorp/terraform-provider-hcp/pull/253))
* provider: Bump `github.com/go-openapi/runtime` from 0.21.0 to 0.23.1 ([#255](https://github.com/hashicorp/terraform-provider-hcp/pull/253))

## 0.22.0 (January 26, 2022)

IMPROVEMENTS:

* datasource/packer: Add check for revoked iterations to HCP Packer datasources ([#240](https://github.com/hashicorp/terraform-provider-hcp/issues/240))

FIXES:

* docs: Correct root token documentation ([#241](https://github.com/hashicorp/terraform-provider-hcp/issues/241))

## 0.21.1 (December 09, 2021)

FEATURES:

* resource/vault: Implement cluster tier scaling ([#228](https://github.com/hashicorp/terraform-provider-hcp/issues/228))
* docs: Add cluster tier scaling guide ([#228](https://github.com/hashicorp/terraform-provider-hcp/issues/228))

FIXES:

* resource/vault: when changing tiers, do not force new ([#233](https://github.com/hashicorp/terraform-provider-hcp/issues/233))

IMPROVEMENTS:

* datasource/packer: Improve error messages for requests made to HCP Packer. ([#229](https://github.com/hashicorp/terraform-provider-hcp/issues/229))
* provider: Bump `terraform-plugin-sdk/v2` dependency ([#230](https://github.com/hashicorp/terraform-provider-hcp/issues/230))
* provider: Bump `terraform-plugin-docs` from 0.5.0 to 0.5.1 ([#223](https://github.com/hashicorp/terraform-provider-hcp/issues/223))
* provider: Bump `go-openapi/strfmt` from 0.21.0 to 0.21.1 ([#226](https://github.com/hashicorp/terraform-provider-hcp/issues/226))

## 0.21.0 (December 08, 2021)

⚠️ Note: There is an issue with this version of the HCP Provider in which Terraform will incorrectly recommend a rebuild of a Vault cluster if the tier is changed, which could result in data loss. For this reason, the v0.21.0 release and tag is no longer available to use. Please upgrade to the patch v0.21.1 or beyond. ⚠️

## 0.20.0 (November 04, 2021)

IMPROVEMENTS:

* datasource/hcp_packer_image: Add build labels to the hcp_packer_image data source ([#217](https://github.com/hashicorp/terraform-provider-hcp/issues/217))
* provider: Bump `go-openapi/runtime` dependency ([#218](https://github.com/hashicorp/terraform-provider-hcp/issues/218))
* provider: Bump `go-openapi/strfmt` dependency ([#218](https://github.com/hashicorp/terraform-provider-hcp/issues/218))
* provider: Bump `actions/checkout` dependency ([#219](https://github.com/hashicorp/terraform-provider-hcp/issues/219))
* provider: Bump `google.golang.org/grpc` dependency ([#220](https://github.com/hashicorp/terraform-provider-hcp/issues/220))

## 0.19.0 (October 27, 2021)

IMPROVEMENTS:

* resource/hvn: Add CIDR Validator that matches backend validator ([#214](https://github.com/hashicorp/terraform-provider-hcp/pull/214))
* resource/hcp_aws_network_peering: Update source channel with metadata ([#213](https://github.com/hashicorp/terraform-provider-hcp/pull/213))
* docs: Add HCP arch image and documentation link ([#212](https://github.com/hashicorp/terraform-provider-hcp/pull/212))
* docs: Rearrange banners in documentation for consistency ([#211](https://github.com/hashicorp/terraform-provider-hcp/pull/211))

## 0.18.0 (October 20, 2021)

FIXES:

* resource/hcp_consul_cluster: Make cluster_id understand id as well ([#205](https://github.com/hashicorp/terraform-provider-hcp/pull/205))

IMPROVEMENTS:

* datasource/packer: Bump Packer datasources to public beta ([#207](https://github.com/hashicorp/terraform-provider-hcp/pull/207))
* provider: Bump several dependencies ([#208](https://github.com/hashicorp/terraform-provider-hcp/pull/208))
* provider: Add provider meta schema with module_name field ([#197](https://github.com/hashicorp/terraform-provider-hcp/pull/197))

## 0.17.0 (September 23, 2021)

IMPROVEMENTS:

* provider: Bump `hcp-sdk-go` dependency ([#199](https://github.com/hashicorp/terraform-provider-hcp/pull/199))

FEATURES:

* **New data source** `hcp_packer_image` ([#194](https://github.com/hashicorp/terraform-provider-hcp/pull/194))
* **New data source** `hcp_packer_iteration` ([#194](https://github.com/hashicorp/terraform-provider-hcp/pull/194))

## 0.16.0 (September 15, 2021)

IMPROVEMENTS:

* resource/hcp_consul_cluster: Updated 'size' description to specify support for size upgrade. ([#193](https://github.com/hashicorp/terraform-provider-hcp/issues/193))
* provider: Bump `terraform-plugin-docs` dependency ([#195](https://github.com/hashicorp/terraform-provider-hcp/issues/195))

## 0.15.0 (September 01, 2021)

IMPROVEMENTS:

* resource/packer_image_iteration: Change field 'bucket' to 'bucket_name' to remain consistent with Packer ([#188](https://github.com/hashicorp/terraform-provider-hcp/issues/188))
* provider: Bump `terraform-plugin-sdk/v2` dependency ([#191](https://github.com/hashicorp/terraform-provider-hcp/issues/191))
* provider: Bump `go-openapi/runtime` dependency ([#190](https://github.com/hashicorp/terraform-provider-hcp/issues/190))
* provider: Bump `go-openapi/strfmt` dependency ([#187](https://github.com/hashicorp/terraform-provider-hcp/issues/187))
* provider: Bump `actions/setup-go` dependency ([#189](https://github.com/hashicorp/terraform-provider-hcp/issues/189))

## 0.14.0 (August 13, 2021)

FEATURES:

* resource/hcp_consul_cluster: Add size upgrade field for consul cluster update ([#168](https://github.com/hashicorp/terraform-provider-hcp/issues/168))

IMPROVEMENTS:

* provider: Add HCP status check to run before TF commands ([#184](https://github.com/hashicorp/terraform-provider-hcp/issues/184))
* provider: Bump `google.golang.org/grpc` dependency ([#185](https://github.com/hashicorp/terraform-provider-hcp/issues/185))
* provider: Bump `github.com/go-openapi/runtime` dependency ([#183](https://github.com/hashicorp/terraform-provider-hcp/issues/183))

## 0.13.0 (August 06, 2021)

FEATURES:

* **New data source** `packer_image_iteration` ([#169](https://github.com/hashicorp/terraform-provider-hcp/issues/169)) in **private beta**

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
