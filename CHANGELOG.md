## v0.105.0 (April 29, 2025)

FEATURES:

* Add support for sync resource in HCP Vault Secrets [[GH-1196](https://github.com/hashicorp/terraform-provider-hcp/pull/1196)]

IMPROVEMENTS:

* Added resources that are required for full functionality to the example code for `hcp_aws_transit_gateway_attachment`:  , `aws_subnet`, `aws_ec2_transit_gateway_vpc_attachment, and `aws_ `aws_route` pointing back to the HVN attachment. [[GH-1253](https://github.com/hashicorp/terraform-provider-hcp/pull/1253)]
* Allow users to assign one or more syncs with an HCP Vault Secrets App [[GH-1196](https://github.com/hashicorp/terraform-provider-hcp/pull/1196)]
* Group operations (create, delete, update, update members) now block client-side if an operationID is present in the response. [[GH-1260](https://github.com/hashicorp/terraform-provider-hcp/pull/1260)]
* Increase project deletion timeout [[GH-1246](https://github.com/hashicorp/terraform-provider-hcp/pull/1246)]
* Upgrade terraform-plugin-go to v0.26.0 [[GH-1231](https://github.com/hashicorp/terraform-provider-hcp/pull/1231)]

BUG FIXES:

* Fixed a bug for `hcp_waypoint_add_on_definition.variable_options` which incorrectly required `options`. [[GH-1242](https://github.com/hashicorp/terraform-provider-hcp/pull/1242)]
* Fixed a bug when `tfe_waypoint_template.variable_options` did not specify `options` list. [[GH-1242](https://github.com/hashicorp/terraform-provider-hcp/pull/1242)]
## v0.104.0 (February 27, 2025)

IMPROVEMENTS:

* Add way to assign actions to a Waypoint template or application on create or update. [[GH-1224](https://github.com/hashicorp/terraform-provider-hcp/pull/1224)]
* Enabling failing HCP Vault Dedicates GHA pipeline tests with fix. [[GH-1218](https://github.com/hashicorp/terraform-provider-hcp/pull/1218)]

DEPRECATIONS:

* The following attribute is deprecated for Waypoint the add-on resource and data source:
  - `hcp_waypoint_add_on_created_by` [[GH-1208](https://github.com/hashicorp/terraform-provider-hcp/pull/1208)]
## v0.103.0 (February 19, 2025)

IMPROVEMENTS:

* Add support for GitLab integration in HVS. [[GH-1189](https://github.com/hashicorp/terraform-provider-hcp/pull/1189)]
* Update hcp_project calls to retry when encountering a 502, 503, or 504 error. [[GH-1181](https://github.com/hashicorp/terraform-provider-hcp/pull/1181)]
* Updating the provider for HCP Boundary to allow for controller configuration settings upon hcp_boundary_cluster resource/data-source [[GH-1164](https://github.com/hashicorp/terraform-provider-hcp/pull/1164)]
* hcp_log_streaming_destination: Generally Available [[GH-1207](https://github.com/hashicorp/terraform-provider-hcp/pull/1207)]
* hcp_log_streaming_destination: manage resource with Log Streaming Service client [[GH-1174](https://github.com/hashicorp/terraform-provider-hcp/pull/1174)]

BUG FIXES:

* Fix a bug where updating an HVS secret name or app name would not recreate the resource as expected [[GH-1157](https://github.com/hashicorp/terraform-provider-hcp/pull/1157)]
* diagnostics: Remove statuspage checks [[GH-1216](https://github.com/hashicorp/terraform-provider-hcp/pull/1216)]
## v0.102.0 (January 17, 2025)
BREAKING CHANGES:

* waypoint: Add new required field for no-code module ID to Waypoint template and
add-on definition resources. [[GH-1146](https://github.com/hashicorp/terraform-provider-hcp/pull/1146)]

FEATURES:

* remove 'starter_small' as a valid tier for hcp vault cluster [[GH-1162](https://github.com/hashicorp/terraform-provider-hcp/pull/1162)]

IMPROVEMENTS:

* Update Go version to 1.22. [[GH-1161](https://github.com/hashicorp/terraform-provider-hcp/pull/1161)]
* Updating docs with examples for hcp_waypoint_template resource/data-source [[GH-1156](https://github.com/hashicorp/terraform-provider-hcp/pull/1156)]
## v0.101.0 (December 16, 2024)

FEATURES:

* add 'hcp_vault_secrets_integration' resource as a replacement for the provider-specific resources [[GH-1151](https://github.com/hashicorp/terraform-provider-hcp/pull/1151)]
* add vault_secrets_integration_azure resource and add support for azure secrets to vault_secrets_rotating_secret resource [[GH-1143](https://github.com/hashicorp/terraform-provider-hcp/pull/1143)]

DEPRECATIONS:

* the following resources are deprecated:
  - `hcp_vault_secrets_integration_aws`
  - `hcp_vault_secrets_integration_azure`
  - `hcp_vault_secrets_integration_confluent`
  - `hcp_vault_secrets_integration_gcp`
  - `hcp_vault_secrets_integration_mongodbatlas`
  - 'hcp_vault_secrets_integration_twilio' [[GH-1151](https://github.com/hashicorp/terraform-provider-hcp/pull/1151)]

BUG FIXES:

* Fix a panic when creating HVS app or secret [[GH-1154](https://github.com/hashicorp/terraform-provider-hcp/pull/1154)]
## v0.100.0 (December 02, 2024)

IMPROVEMENTS:

* Enable updates to Vault Radar integration connections. [[GH-1133](https://github.com/hashicorp/terraform-provider-hcp/pull/1133)]
* Enable updates to Vault Radar integration subscriptions. [[GH-1139](https://github.com/hashicorp/terraform-provider-hcp/pull/1139)]
* Enable updates to Vault Radar source token. [[GH-1134](https://github.com/hashicorp/terraform-provider-hcp/pull/1134)]
* Update hcp_group API calls to retry when encountering a 502, 503, or 504 error. [[GH-1140](https://github.com/hashicorp/terraform-provider-hcp/pull/1140)]
* Upgrade the HCP SDK and update Vault Secrets to use API V2 (stable/2023-11-28) [[GH-1142](https://github.com/hashicorp/terraform-provider-hcp/pull/1142)]

BUG FIXES:

* Retry billing service update call when creating a project [[GH-1131](https://github.com/hashicorp/terraform-provider-hcp/pull/1131)]
## v0.99.0 (November 12, 2024)

DEPRECATIONS:

* Added a Deprecation message for the following resources: `hcp_consul_agent_helm_config`,`hcp_consul_agent_kubernetes_secret`,`hcp_consul_cluster`,`hcp_consul_versions` [[GH-1124](https://github.com/hashicorp/terraform-provider-hcp/pull/1124)]
* Added a Deprecation message for the following resources: `hcp_consul_cluster`,`hcp_consul_cluster_root_token`,`hcp_consul_snapshot` [[GH-1124](https://github.com/hashicorp/terraform-provider-hcp/pull/1124)]
## v0.98.1 (November 07, 2024)

BUG FIXES:

* HCP Vault Radar resource documents did not have the subcategory set correctly. [[GH-1129](https://github.com/hashicorp/terraform-provider-hcp/pull/1129)]
## v0.98.0 (November 06, 2024)

FEATURES:

* Add preview of the following Vault Radar connections and subscriptions:
hcp_vault_radar_integration_jira_connection
hcp_vault_radar_integration_jira_subscription
hcp_vault_radar_integration_slack_connection
hcp_vault_radar_integration_slack_subscription [[GH-1116](https://github.com/hashicorp/terraform-provider-hcp/pull/1116)]
* Add preview of vault_radar_source_github_cloud resource. [[GH-1119](https://github.com/hashicorp/terraform-provider-hcp/pull/1119)]
* add vault_secrets_integration_confluent resource and add support for confluent secrets to vault_secrets_rotating_secret resource [[GH-1120](https://github.com/hashicorp/terraform-provider-hcp/pull/1120)]

IMPROVEMENTS:

* Fixed issues where the HVS App resource would behave improperly such as returning an error if using an empty description. [[GH-1118](https://github.com/hashicorp/terraform-provider-hcp/pull/1118)]
* Upgrade the HCP SDK and fix breaking change w/ Vault Secrets [[GH-1113](https://github.com/hashicorp/terraform-provider-hcp/pull/1113)]
* Upgrade the HCP SDK for Vault Secrets [[GH-1122](https://github.com/hashicorp/terraform-provider-hcp/pull/1122)]

BUG FIXES:

* creating a mongodb rotating secret should create a mongodb rotating secret not a twilio one [[GH-1126](https://github.com/hashicorp/terraform-provider-hcp/pull/1126)]
## v0.97.0 (October 08, 2024)

FEATURES:

* Add pagination support for vault OpenAppSecrets preview api. [[GH-1109](https://github.com/hashicorp/terraform-provider-hcp/pull/1109)]
* Add preview of vault_radar_source_github_enterprise resource. [[GH-1092](https://github.com/hashicorp/terraform-provider-hcp/pull/1092)]
* Waypoint: New template resource attribute `use_module_readme` allows users to use the associated Terraform module readme in place of providing a seperate readme for the template. [[GH-1099](https://github.com/hashicorp/terraform-provider-hcp/pull/1099)]
* add vault_secrets_dynamic_secret resource [[GH-1097](https://github.com/hashicorp/terraform-provider-hcp/pull/1097)]
* add vault_secrets_rotating_secret resource [[GH-1101](https://github.com/hashicorp/terraform-provider-hcp/pull/1101)]
* waypoint: Add execution mode and agent pool ID configurations to the
`hcp_waypoint_template` and `hcp_waypoint_add_on_definition` resources, and
the `hcp_waypoint_template` and `hcp_waypoint_add_on_definition` data sources. [[GH-1098](https://github.com/hashicorp/terraform-provider-hcp/pull/1098)]

IMPROVEMENTS:

* Update Vault Radar codeowners patterns. [[GH-1103](https://github.com/hashicorp/terraform-provider-hcp/pull/1103)]
* Update codeowners to include Vault Radar. [[GH-1102](https://github.com/hashicorp/terraform-provider-hcp/pull/1102)]
* Upgrade the HCP SDK and fix a breaking change in the Boundary API. [[GH-1110](https://github.com/hashicorp/terraform-provider-hcp/pull/1110)]

BUG FIXES:

* Waypoint: fix where using the deprecated terraform_cloud_workspace_details.name field could cause an error. [[GH-1093](https://github.com/hashicorp/terraform-provider-hcp/pull/1093)]
* waypoint: Fix `hcp_waypoint_template` and `hcp_waypoint_add_on_definition` data sources. [[GH-1094](https://github.com/hashicorp/terraform-provider-hcp/pull/1094)]
## v0.96.0 (September 13, 2024)

FEATURES:

* add vault_secrets_integration_aws resource [[GH-1078](https://github.com/hashicorp/terraform-provider-hcp/pull/1078)]
* add vault_secrets_integration_gcp resource [[GH-1085](https://github.com/hashicorp/terraform-provider-hcp/pull/1085)]
* add vault_secrets_integration_mongodbatlas resource [[GH-1084](https://github.com/hashicorp/terraform-provider-hcp/pull/1084)]
* add vault_secrets_integration_twilio resource [[GH-1081](https://github.com/hashicorp/terraform-provider-hcp/pull/1081)]

IMPROVEMENTS:

* Added the `version` attribute(s) to `hcp_boundary_cluster` resource and data sources. [[GH-1089](https://github.com/hashicorp/terraform-provider-hcp/pull/1089)]
## v0.95.1 (August 29, 2024)

IMPROVEMENTS:

* Waypoint resources for templates and add-on definitions no longer require setting
a `terraform_cloud_workspace_details` resource, and instead can be set by the
`terraform_project_id` param. [[GH-1052](https://github.com/hashicorp/terraform-provider-hcp/pull/1052)]
* update hcp-sdk-go [[GH-1080](https://github.com/hashicorp/terraform-provider-hcp/pull/1080)]
## v0.95.0 (August 21, 2024)
BREAKING CHANGES:

* waypoint: Remove version number from templates and add-on definition resources and data sources. [[GH-869](https://github.com/hashicorp/terraform-provider-hcp/pull/869)]

FEATURES:

* Allow authenticating the provider using Workload Identity Federation via a
direct token in the provider configuration. [[GH-1053](https://github.com/hashicorp/terraform-provider-hcp/pull/1053)]
* HCP Waypoint Application resource and data source can now read any output values associated with that application [[GH-871](https://github.com/hashicorp/terraform-provider-hcp/pull/871)]

IMPROVEMENTS:

* Updated documentation to communicate the APIs the provider leverages during usage. [[GH-1066](https://github.com/hashicorp/terraform-provider-hcp/pull/1066)]
* waypoint: The `readme_markdown_template` attribute for both template and add-on definition resources now accepts unencoded strings as well as base64 encoded strings. [[GH-894](https://github.com/hashicorp/terraform-provider-hcp/pull/894)]
## v0.94.1 (July 08, 2024)

BUG FIXES:

* Fix the documentation location for hcp_packer_bucket, hcp_packer_bucket_iam_policy, and hcp_packer_bucket_iam_binding. [[GH-904](https://github.com/hashicorp/terraform-provider-hcp/pull/904)]
## v0.94.0 (July 03, 2024)

FEATURES:

* New resource: Add `hcp_packer_bucket` resource for managing HCP Packer buckets
 New Resource: Add `hcp_packer_bucket_iam_policy` resource for assigning a list of policy bindings to multiple principals for a HCP Packer Bucket
 New resource: Add `hcp_packer_bucket_iam_binding` resource for assigning a single role to a principal for a HCP Packer Bucket [[GH-852](https://github.com/hashicorp/terraform-provider-hcp/pull/852)]
* add vault_secrets_dynamic_secret data source [[GH-868](https://github.com/hashicorp/terraform-provider-hcp/pull/868)]

IMPROVEMENTS:

* waypoint: Add support for input variables to the `hcp_wapyoint_add_on` resource
and data source. [[GH-864](https://github.com/hashicorp/terraform-provider-hcp/pull/864)]

BUG FIXES:

* Fix intermittent conflicts during IAM policy updates [[GH-870](https://github.com/hashicorp/terraform-provider-hcp/pull/870)]
## v0.93.0 (June 25, 2024)
BREAKING CHANGES:

* waypoint: Rename `hcp_waypoint_application_template` resource and data source to
`hcp_waypoint_template`. [[GH-861](https://github.com/hashicorp/terraform-provider-hcp/pull/861)]

FEATURES:

* New resource: Add `hcp_waypoint_action_config` resource for managing Waypoint Action Configs.
 New data-source: Add `data.hcp_waypoint_action_config` data-source for Waypoint Action Configs. [[GH-815](https://github.com/hashicorp/terraform-provider-hcp/pull/815)]
## v0.92.0 (June 18, 2024)

FEATURES:

* Add `hcp_group_iam_binding` resource. [[GH-865](https://github.com/hashicorp/terraform-provider-hcp/pull/865)]
* Add `hcp_group_iam_policy` resource. [[GH-865](https://github.com/hashicorp/terraform-provider-hcp/pull/865)]
## v0.91.1 (June 13, 2024)

BUG FIXES:

* Fix the documentation for hcp_vault_secrets_rotating_secret [[GH-867](https://github.com/hashicorp/terraform-provider-hcp/pull/867)]
## v0.91.0 (June 06, 2024)

FEATURES:

* Allows users to fetch rotating secrets using the hcp_vault_secrets_app and hcp_vault_secrets_secret data sources [[GH-850](https://github.com/hashicorp/terraform-provider-hcp/pull/850)]
* add vault_secrets_rotating_secret data source [[GH-854](https://github.com/hashicorp/terraform-provider-hcp/pull/854)]

IMPROVEMENTS:

* Add support for input variables to `hcp_waypoint_application` resource and
data source. [[GH-833](https://github.com/hashicorp/terraform-provider-hcp/pull/833)]
* Update `hcp-sdk-go` [[GH-851](https://github.com/hashicorp/terraform-provider-hcp/pull/851)]
* Update `hcp-sdk-go` to v0.98.0 [[GH-855](https://github.com/hashicorp/terraform-provider-hcp/pull/855)]
## v0.90.0 (May 22, 2024)

FEATURES:

* Added HCP Boundary to statuspage components used to warn about potential issues with HCP. [[GH-835](https://github.com/hashicorp/terraform-provider-hcp/pull/835)]

IMPROVEMENTS:

* Support retrieving output values in `hcp_waypoint_add_on` [[GH-822](https://github.com/hashicorp/terraform-provider-hcp/pull/822)]

BUG FIXES:

* Waypoint: fix issue with importing existing Waypoint resources [[GH-839](https://github.com/hashicorp/terraform-provider-hcp/pull/839)]
## v0.89.0 (May 10, 2024)

IMPROVEMENTS:

* Support setting variable options in `Support setting variable options in `hcp_waypoint_add_on_definition`. [[GH-819](https://github.com/hashicorp/terraform-provider-hcp/pull/819)]

BUG FIXES:

* Fix first apply scenario for hcp_vault_secrets_app_iam_binding and hcp_vault_secrets_app_iam_policy, update docs [[GH-838](https://github.com/hashicorp/terraform-provider-hcp/pull/838)]
## v0.88.0 (May 02, 2024)

IMPROVEMENTS:

* Documentation: call OpenAppSecrets instead of List + Open in hcp_vault_secrets_app data source. This improves an issue with additional unnecessary client API usage. [[GH-826](https://github.com/hashicorp/terraform-provider-hcp/pull/826)]
* Support setting variable options in `hcp_waypoint_application_template`. [[GH-816](https://github.com/hashicorp/terraform-provider-hcp/pull/816)]
* resource/hcp_log_streaming_destination: Label resource as public beta [[GH-830](https://github.com/hashicorp/terraform-provider-hcp/pull/830)]
## v0.87.1 (April 25, 2024)

BUG FIXES:

* Fix documentation of hcp_vault_secrets_app_iam_binding and hcp_vault_secrets_app_iam_policy [[GH-825](https://github.com/hashicorp/terraform-provider-hcp/pull/825)]
## v0.87.0 (April 25, 2024)

FEATURES:

* Add `hcp_vault_secrets_app_iam_binding` resource. [[GH-813](https://github.com/hashicorp/terraform-provider-hcp/pull/813)]
* Add `hcp_vault_secrets_app_iam_policy` resource. [[GH-813](https://github.com/hashicorp/terraform-provider-hcp/pull/813)]
* New resource: Add `hcp_waypoint_add_on` resource for managing Waypoint Add-ons.
 New data-source: Add `data.hcp_waypoint_add_on` data-source for Waypoint Add-ons. [[GH-807](https://github.com/hashicorp/terraform-provider-hcp/pull/807)]

IMPROVEMENTS:

* Add GitHub Action to run identity-specific tests [[GH-810](https://github.com/hashicorp/terraform-provider-hcp/pull/810)]
* CODEOWNERS: Fix the vault-secrets resource ownership to @hashicorp/cloud-vault-secrets team. [[GH-820](https://github.com/hashicorp/terraform-provider-hcp/pull/820)]
* CODEOWNERS: Fix vault-secrets resource ownership to @hashicorp/cloud-vault-secrets team. [[GH-814](https://github.com/hashicorp/terraform-provider-hcp/pull/814)]
* Documentation: Gracefully handle rate limiting error on `hcp_vault_secrets_secret` resource. [[GH-812](https://github.com/hashicorp/terraform-provider-hcp/pull/812)]
* Return resource name when a Vault Secrets App is created [[GH-813](https://github.com/hashicorp/terraform-provider-hcp/pull/813)]
* Vault secret resources can now be created with an optional project ID. If project ID is present, the resource will be created within that project. [[GH-808](https://github.com/hashicorp/terraform-provider-hcp/pull/808)]

BUG FIXES:

* Fixes the case where Vault secret resources ignore provider project changes. [[GH-808](https://github.com/hashicorp/terraform-provider-hcp/pull/808)]
## v0.86.0 (April 04, 2024)

FEATURES:

* Add DataDog config option for hcp_log_streaming_destination resource. [[GH-803](https://github.com/hashicorp/terraform-provider-hcp/pull/803)]
* Add in-place update functionality to `hcp_log_streaming_destination` resource. [[GH-802](https://github.com/hashicorp/terraform-provider-hcp/pull/802)]

BUG FIXES:

* Fixed an issue that caused a HTTP audit-log streaming config to fail with allowed values for http_codec are only "JSON" or "NDJSON" [[GH-781](https://github.com/hashicorp/terraform-provider-hcp/pull/781)]
* Handle the case when multiple organizations are associated with the configured credentials. Now, it instead prompts an error, requiring users to specify a particular organization in the HCP provider config block. [[GH-793](https://github.com/hashicorp/terraform-provider-hcp/pull/793)]
## v0.85.0 (April 03, 2024)

FEATURES:

* New resource: Add `hcp_waypoint_application` resource for managing Waypoint Application.
 New data-source: Add `data.hcp_waypoint_application` data-source for Waypoint Application. [[GH-794](https://github.com/hashicorp/terraform-provider-hcp/pull/794)]

IMPROVEMENTS:

* Documentation: Update hcp_iam_policy docs with HCP Vault Secrets Roles App Manager and App Secret Reader [[GH-801](https://github.com/hashicorp/terraform-provider-hcp/pull/801)]
## v0.84.1 (March 22, 2024)

IMPROVEMENTS:

* Documentation: Adjust the subcategory for hcp_user_principal, hcp_group_members, hcp_iam_workload_identity_provider to Cloud Platform [[GH-797](https://github.com/hashicorp/terraform-provider-hcp/pull/797)]

BUG FIXES:

* Fixes an issue where organization ID was not sent on data.hcp_user_principal lookup [[GH-797](https://github.com/hashicorp/terraform-provider-hcp/pull/797)]
## v0.84.0 (March 22, 2024)
BREAKING CHANGES:

* Removed `hcp_packer_image` data source, previously deprecated and replaced by `hcp_packer_artifact` in [v0.82.0](https://github.com/hashicorp/terraform-provider-hcp/releases/tag/v0.82.0). [[GH-791](https://github.com/hashicorp/terraform-provider-hcp/pull/791)]
* Removed `hcp_packer_iteration` data source, previously deprecated and replaced by `hcp_packer_version` in [v0.82.0](https://github.com/hashicorp/terraform-provider-hcp/releases/tag/v0.82.0). [[GH-791](https://github.com/hashicorp/terraform-provider-hcp/pull/791)]
* Removed `iteration_fingerprint` attribute from `hcp_packer_channel_assignment` resource, previously deprecated and replaced by `version_fingerprint` in [v0.82.0](https://github.com/hashicorp/terraform-provider-hcp/releases/tag/v0.82.0). [[GH-791](https://github.com/hashicorp/terraform-provider-hcp/pull/791)]

FEATURES:

* New data source: Add `hcp_group` data source for retrieving HCP groups. [[GH-787](https://github.com/hashicorp/terraform-provider-hcp/pull/787)]
* New resource: Add `hcp_group_members` resource for managing HCP user principals in a group. [[GH-787](https://github.com/hashicorp/terraform-provider-hcp/pull/787)]
* New resource: Add `hcp_waypoint_add_on_definition` resource for managing Waypoint Add-on Definitions.
 New data-source: Add `data.hcp_waypoint_add_on_definition` data-source for Waypoint Add-on Definitions. [[GH-785](https://github.com/hashicorp/terraform-provider-hcp/pull/785)]
* New resource: Add `hcp_waypoint_application_template` resource for managing Waypoint Application Templates.
 New data-source: Add `data.hcp_waypoint_application_template` data-source for Waypoint Application Templates. [[GH-778](https://github.com/hashicorp/terraform-provider-hcp/pull/778)]
* New resource: Add `hcp_waypoint_tfc_config` resource for managing the TFC
 Configuration used by Waypoint. [[GH-766](https://github.com/hashicorp/terraform-provider-hcp/pull/766)]

BUG FIXES:

* Update Boundary API clients method name to match the latest version [[GH-770](https://github.com/hashicorp/terraform-provider-hcp/pull/770)]
## v0.83.0 (February 23, 2024)

FEATURES:

* Add CloudWatch config option for hcp_log_streaming_destination resource. [[GH-757](https://github.com/hashicorp/terraform-provider-hcp/pull/757)]

IMPROVEMENTS:

* Update example for `hcp_packer_channel_assignment` resource [[GH-749](https://github.com/hashicorp/terraform-provider-hcp/pull/749)]

BUG FIXES:

* Do not panic if provider is configured with credentials with no project access. [[GH-748](https://github.com/hashicorp/terraform-provider-hcp/pull/748)]
* Fixing a panic on errors when opening secrets from HCP Vault Secrets. [[GH-751](https://github.com/hashicorp/terraform-provider-hcp/pull/751)]
* Fixing the validation rules for HCP Vault Secrets app and secret names to match
what we have on the server side. [[GH-750](https://github.com/hashicorp/terraform-provider-hcp/pull/750)]
## v0.82.0 (January 30, 2024)
BREAKING CHANGES:

* `data.hcp_packer_iteration`: Removed the `incremental_version` attribute. Use the `fingerprint`, `id` or `uuid` attributes instead. [[GH-726](https://github.com/hashicorp/terraform-provider-hcp/pull/726)]
* `hcp_packer_channel_assignment`: Removed the `iteration_id` attribute. Use the `version_fingerprint` attribute instead. [[GH-726](https://github.com/hashicorp/terraform-provider-hcp/pull/726)]
* `hcp_packer_channel_assignment`: Removed the `iteration_version` attribute. Use the `version_fingerprint` attribute instead. [[GH-726](https://github.com/hashicorp/terraform-provider-hcp/pull/726)]

FEATURES:

* Add `hcp_notifications_webhook` resource. [[GH-723](https://github.com/hashicorp/terraform-provider-hcp/pull/723)]
* `data.hcp_packer_artifact`: Added the `hcp_packer_artifact` data source, which replaces `data.hcp_packer_image` [[GH-726](https://github.com/hashicorp/terraform-provider-hcp/pull/726)]
* `data.hcp_packer_version`: Added the `hcp_packer_version` data source, which replaces `data.hcp_packer_iteration` [[GH-726](https://github.com/hashicorp/terraform-provider-hcp/pull/726)]

DEPRECATIONS:

* `data.hcp_packer_image`: This data source will be removed in a future release, and is superseded by `data.hcp_packer_artifact` [[GH-735](https://github.com/hashicorp/terraform-provider-hcp/pull/735)]
* `data.hcp_packer_iteration`: This data source will be removed in a future release, and is superseded by `data.hcp_packer_version` [[GH-726](https://github.com/hashicorp/terraform-provider-hcp/pull/726)]
* `hcp_packer_channel_assignment`: The `iteration_fingerprint` attribute is now deprecated and will be removed in a future release. 
Refer to the `version_fingerprint` attribute release notes for more information. [[GH-726](https://github.com/hashicorp/terraform-provider-hcp/pull/726)]
## v0.81.0 (January 24, 2024)

FEATURES:

* Add `ip_allowlist` field for HCP Vault clusters to create or update allowed IP address ranges (CIDRs) for inbound traffic. [[GH-625](https://github.com/hashicorp/terraform-provider-hcp/pull/625)]
## v0.80.0 (January 11, 2023)

FEATURES:

* Add `hcp_log_streaming_destination` resource. [[GH-713](https://github.com/hashicorp/terraform-provider-hcp/pull/713)]
## v0.79.0 (December 27, 2023)

FEATURES:

* Allow authenticating the provider using Workload Identity Federation via a
token_file in the provider configuration. [[GH-691](https://github.com/hashicorp/terraform-provider-hcp/pull/691)]
## v0.78.0 (December 13, 2023)

FEATURES:

* Add optional Azure config block for hvn_route resource to support hub and spoke architectures, `azure_config`.

 Add optional `next_hop_type` parameter.
 Add optional `next_hop_ip_address` parameter.

 Please see the Microsoft learn guide docs for futher info:
 https://learn.microsoft.com/en-us/azure/architecture/reference-architectures/hybrid-networking/hub-spoke [[GH-685](https://github.com/hashicorp/terraform-provider-hcp/pull/685)]
* Add optional attribute `allow_forwarded_traffic` to `hcp_azure_peering_connection` resource. [[GH-681](https://github.com/hashicorp/terraform-provider-hcp/pull/681)]
* Add optional attribute `use_remote_gateways` to `hcp_azure_peering_connection` resource. [[GH-681](https://github.com/hashicorp/terraform-provider-hcp/pull/681)]

IMPROVEMENTS:

* Add custom http endpoints as an observability provider for streaming audit logs and metrics from HCP Vault clusters. [[GH-671](https://github.com/hashicorp/terraform-provider-hcp/pull/671)]

DEPRECATIONS:

* The `data.hcp_packer_iteration.incremental_version` attribute is now deprecated and will be removed in a future release. Use the `fingerprint`, `id` or `uuid` attributes to reference iterations instead. [[GH-678](https://github.com/hashicorp/terraform-provider-hcp/pull/678)]
* The `hcp_packer_channel_assignment.iteration_id` attribute is now deprecated and will be removed in a future release. Use the `iteration_fingerprint` attribute to reference iterations instead. [[GH-678](https://github.com/hashicorp/terraform-provider-hcp/pull/678)]
* The `hcp_packer_channel_assignment.iteration_version` attribute is now deprecated and will be removed in a future release. Use the `iteration_fingerprint` attribute to reference iterations instead. [[GH-678](https://github.com/hashicorp/terraform-provider-hcp/pull/678)]
## v0.77.0 (November 14, 2023)

FEATURES:

* Add `hcp_organization_iam_binding` resource. [[GH-661](https://github.com/hashicorp/terraform-provider-hcp/pull/661)]
* Add `hcp_organization_iam_policy` resource. [[GH-661](https://github.com/hashicorp/terraform-provider-hcp/pull/661)]
* Add `hcp_project_iam_binding` resource. [[GH-661](https://github.com/hashicorp/terraform-provider-hcp/pull/661)]
* Add `hcp_project_iam_policy` resource. [[GH-661](https://github.com/hashicorp/terraform-provider-hcp/pull/661)]
* Add support to authenticate the provider using credential files. [[GH-672](https://github.com/hashicorp/terraform-provider-hcp/pull/672)]

IMPROVEMENTS:

* Add NewRelic as an observability provider for streaming audit logs and metrics from HCP Vault clusters. [[GH-653](https://github.com/hashicorp/terraform-provider-hcp/pull/653)]
* The change introduces the ability to provide configuration via environment variables, which will allow credentials to be provided via environment variable (either directly or via a creds file). [[GH-666](https://github.com/hashicorp/terraform-provider-hcp/pull/666)]
## v0.76.0 (November 01, 2023)

BUG FIXES:

* Fix some documentation mistakes in examples for the `hcp_iam_workload_identity_provider` resource. [[GH-657](https://github.com/hashicorp/terraform-provider-hcp/pull/657)]
* Removed unnecessary validation of a trailing slash on the `oidc.issuer_uri`
field for the `hcp_iam_workload_identity_provider` resource. [[GH-656](https://github.com/hashicorp/terraform-provider-hcp/pull/656)]
## v0.75.0 (October 25, 2023)
BREAKING CHANGES:

* Drops support for Terraform CLI versions prior to v1.1.5, previously deprecated in provider release v0.74.1 [[GH-637](https://github.com/hashicorp/terraform-provider-hcp/pull/637)]

FEATURES:

* Add `hcp_iam_workload_identity_provider` resource. [[GH-647](https://github.com/hashicorp/terraform-provider-hcp/pull/647)]

IMPROVEMENTS:

* Upgrade to Terraform Provider Protocol 6 [[GH-637](https://github.com/hashicorp/terraform-provider-hcp/pull/637)]
## v0.74.1 (October 19, 2023)

DEPRECATIONS:

* An upcoming release will deprecate support for Terraform versions before 1.1.5.
Please upgrade to be able to use the latest releases of the provider. [[GH-644](https://github.com/hashicorp/terraform-provider-hcp/pull/644)]
## v0.74.0 (October 19, 2023)

FEATURES:

* Add `hcp_service_principal_key` resource. [[GH-636](https://github.com/hashicorp/terraform-provider-hcp/pull/636)]
* Add `hcp_service_principal` data source. [[GH-636](https://github.com/hashicorp/terraform-provider-hcp/pull/636)]
* Add `hcp_service_principal` resource. [[GH-636](https://github.com/hashicorp/terraform-provider-hcp/pull/636)]

BUG FIXES:

* Fix hcp_project import [[GH-642](https://github.com/hashicorp/terraform-provider-hcp/pull/642)]
## v0.73.0 (October 12, 2023)

FEATURES:

* Add `hcp_organization` data source. [[GH-629](https://github.com/hashicorp/terraform-provider-hcp/pull/629)]
* Add `hcp_project` data source. [[GH-629](https://github.com/hashicorp/terraform-provider-hcp/pull/629)]
* Add `hcp_project` resource. [[GH-629](https://github.com/hashicorp/terraform-provider-hcp/pull/629)]

BUG FIXES:

* Prevent crash when a vault_secrets_app resource no longer exists. [[GH-619](https://github.com/hashicorp/terraform-provider-hcp/pull/619)]
## v0.72.2 (October 06, 2023)

BUG FIXES:

* Fix documentation duplicate subcategory issue introduced by [#620](https://github.com/hashicorp/terraform-provider-hcp/pull/620) [[GH-622](https://github.com/hashicorp/terraform-provider-hcp/pull/622)]
## v0.72.1 (October 03, 2023)

BUG FIXES:

* Fix documentation formatting issue introduced by [#585](https://github.com/hashicorp/terraform-provider-hcp/pull/585) [[GH-620](https://github.com/hashicorp/terraform-provider-hcp/pull/620)]
## v0.72.0 (September 28, 2023)
BREAKING CHANGES:

* Removed the `hcp_packer_image_iteration` data source, previously deprecated in [v0.68.0](https://github.com/hashicorp/terraform-provider-hcp/releases/tag/v0.68.0) [[GH-609](https://github.com/hashicorp/terraform-provider-hcp/pull/609)]

IMPROVEMENTS:

* Update to Go 1.20 per the [Go support policy](https://go.dev/doc/devel/release#policy). [[GH-610](https://github.com/hashicorp/terraform-provider-hcp/pull/610)]
* Updating docs to use sub-categories [[GH-585](https://github.com/hashicorp/terraform-provider-hcp/pull/585)]
## v0.71.1 (September 25, 2023)

BUG FIXES:

* project_id unauthorized issue when using config client secret [[GH-604](https://github.com/hashicorp/terraform-provider-hcp/pull/604)]

## v0.71.0 (September 20, 2023)

FEATURES:

* Add data source for pulling in a single secret. [[GH-594](https://github.com/hashicorp/terraform-provider-hcp/pull/594)]
* Add resource to create app in Vault Secrets. [[GH-593](https://github.com/hashicorp/terraform-provider-hcp/pull/593)]

IMPROVEMENTS:

* Add ElasticSearch as an observability provider for streaming audit logs and metrics from HCP Vault clusters. [[GH-586](https://github.com/hashicorp/terraform-provider-hcp/pull/586)]
* Update to use the Terraform Multiplexing framework [[GH-565](https://github.com/hashicorp/terraform-provider-hcp/pull/565)]
## v0.70.0 (September 13, 2023)

IMPROVEMENTS:

* Add CloudWatch as an observability provider for streaming audit logs and metrics from HCP Vault clusters. [[GH-580](https://github.com/hashicorp/terraform-provider-hcp/pull/580)]
* Add `hcp_vault_plugin` resource [[GH-579](https://github.com/hashicorp/terraform-provider-hcp/pull/579)]
* Update semver validation to allow specification of pre-release versions [[GH-590](https://github.com/hashicorp/terraform-provider-hcp/pull/590)]

BUG FIXES:

* update hcp_vault_plugin resource test. remove hard coded values that were used for local testing. [[GH-591](https://github.com/hashicorp/terraform-provider-hcp/pull/591)]
## v0.69.0 (August 17, 2023)

IMPROVEMENTS:

* Add `proxy_endpoint` field to enable toggling the proxy option on HCP Vault clusters, along with corresponding read-only `vault_proxy_endpoint_url` field. [[GH-577](https://github.com/hashicorp/terraform-provider-hcp/pull/577)]
* Adjusted wording in the documentation and error messages for `data.hcp_packer_image`. [[GH-571](https://github.com/hashicorp/terraform-provider-hcp/pull/571)]
## v0.68.0 (August 02, 2023)
BREAKING CHANGES:

* Removed the `iteration` block from `hcp_packer_channel`, previously deprecated in v0.63.0 [[GH-555](https://github.com/hashicorp/terraform-provider-hcp/pull/555)]

FEATURES:

* Channel restriction can now be configured with the `restricted` field of `hcp_packer_channel` [[GH-555](https://github.com/hashicorp/terraform-provider-hcp/pull/555)]

IMPROVEMENTS:

* Add support for HCP Packer Managed channels to `hcp_packer_channel`. Also adds a read-only `managed` attribute. [[GH-555](https://github.com/hashicorp/terraform-provider-hcp/pull/555)]

DEPRECATIONS:

* The `hcp_packer_image_iteration` data source is now deprecated.
Use the `hcp_packer_image` or `hcp_packer_iteration` data sources instead. [[GH-569](https://github.com/hashicorp/terraform-provider-hcp/pull/569)]
## v0.67.0 (July 26, 2023)

IMPROVEMENTS:

* Upgrade Go to version 1.19 [[GH-556](https://github.com/hashicorp/terraform-provider-hcp/pull/556)]
## v0.66.0 (July 13, 2023)

FEATURES:

* Adds `hcp_packer_run_task` data source. [[GH-547](https://github.com/hashicorp/terraform-provider-hcp/pull/547)]
* Adds `hcp_packer_run_task` resource. [[GH-547](https://github.com/hashicorp/terraform-provider-hcp/pull/547)]

IMPROVEMENTS:

* Updated Auth Guide to reflect the impact to the provider when using Project-Level Service Principals;  `project_id` within the provider configuration is needed. Fix for previous release in v0.64.0 [[GH-552](https://github.com/hashicorp/terraform-provider-hcp/pull/552)]
* Updated Boundary `Tier` configuration to be case insensitive. [[GH-554](https://github.com/hashicorp/terraform-provider-hcp/pull/554)]
## v0.65.0 (July 10, 2023)

FEATURES:

* Adds `tier` to the `hcp_boundary_cluster` resource to manage the cluster pricing and feature set. [[GH-544](https://github.com/hashicorp/terraform-provider-hcp/pull/544)]
## v0.64.0 (July 07, 2023)

FEATURES:

* Adds `hcp_packer_bucket_names` data source. [[GH-536](https://github.com/hashicorp/terraform-provider-hcp/pull/536)]

IMPROVEMENTS:

* Updated Auth Guide to reflect the impact to the provider when using Project-Level Service Principals;  `project_id` within the provider configuration is needed. [[GH-543](https://github.com/hashicorp/terraform-provider-hcp/pull/543)]

BUG FIXES:

* Capture unknown errors from the HCP Go SDK when getting project or organization, where errors were previously replaced with a static message. [[GH-542](https://github.com/hashicorp/terraform-provider-hcp/pull/542)]
* Resolve panics when Packer API client helpers receive unexpected error formats [[GH-545](https://github.com/hashicorp/terraform-provider-hcp/pull/545)]
## v0.63.0 (June 29, 2023)

FEATURES:

* Adds `hcp_packer_channel_assignment` resource. [[GH-531](https://github.com/hashicorp/terraform-provider-hcp/pull/531)]

IMPROVEMENTS:

* Adds read-only `restricted` field to `hcp_packer_channel` resource. [[GH-531](https://github.com/hashicorp/terraform-provider-hcp/pull/531)]

DEPRECATIONS:

* Setting the `iteration` block on `hcp_packer_channel` is now deprecated in
favor of `hcp_packer_channel_assignment`. [[GH-531](https://github.com/hashicorp/terraform-provider-hcp/pull/531)]
## v0.62.0 (June 23, 2023)

IMPROVEMENTS:

* Add "premium" to descriptions about Consul Cluster tier [[GH-537](https://github.com/hashicorp/terraform-provider-hcp/pull/537)]
* Update the description of the optional `project_id` field in data sources and resources to include details of how project_id is targeted when none is passed in. [[GH-527](https://github.com/hashicorp/terraform-provider-hcp/pull/527)]

BUG FIXES:

* Fixed panic when `hcp_packer_image` points to a channel without an assigned iteration. [[GH-533](https://github.com/hashicorp/terraform-provider-hcp/pull/533)]
## v0.61.0 (June 13, 2023)
FEATURES:

* hcp_vault_secrets_app: New data source that allows bulk retrieval of all secret values in a secrets app [[GH-524](https://github.com/hashicorp/terraform-provider-hcp/pull/524)]
## v0.60.0 (June 07, 2023)

IMPROVEMENTS:

* Documentation: Update contributor documentation and resource-specific import examples to reflect multi-project support. [[GH-517](https://github.com/hashicorp/terraform-provider-hcp/pull/517)]

DEPRECATIONS:

* Setting the `hvn_2` attribute of `data.hcp_hvn_peering_connection` is now 
deprecated. The value of the attribute is not needed to fetch data, and it was 
never validated against the real value for `hvn_2`. The value will now be 
populated automatically. Remove the `hvn_2` attribute from the configuration 
for affected data sources. [[GH-522](https://github.com/hashicorp/terraform-provider-hcp/pull/522)]
* Setting the `project_id` attribute on `hcp_hvn_peering_connection` and 
`data.hcp_hvn_peering_connection` is now deprecated. The value of the field was 
required to match the project ID for `hvn_1` and will now be determined 
automatically. Remove the `project_id` field from the configuration for 
affected resources and data sources. [[GH-522](https://github.com/hashicorp/terraform-provider-hcp/pull/522)]
* Setting the `project_id` attribute on `hcp_hvn_route` and `data.hcp_hvn_route`
is now deprecated. The value of the field was required to match the project ID 
in `hvn_link` and will now be determined automatically. Remove the `project_id` 
field from the configuration for affected resources and data sources. [[GH-522](https://github.com/hashicorp/terraform-provider-hcp/pull/522)]

BUG FIXES:

* Fixed several missing/incorrect implementations for the resource-level 
`project_id` attribute that could lead to undefined or undesirable behavior on 
some resources and data sources when the `project_id` attribute had been used 
and its most recent value was different from the provider-level `project_id`,
whether or not the attribute was still present in the configuration file.
NOTE: See associated PR for caveats on temporary regressions. [[GH-522](https://github.com/hashicorp/terraform-provider-hcp/pull/522)]
* Resolve unintended removal of assigned iteration when `iteration` block is not present on `hcp_packer_channel` [[GH-521](https://github.com/hashicorp/terraform-provider-hcp/pull/521)]
## v0.59.0 (June 01, 2023)

FEATURES:

* Update HVN Route validation for `destination_cidr` attribute to allow RFC6598 CIDR range. [[GH-515](https://github.com/hashicorp/terraform-provider-hcp/pull/515)]

IMPROVEMENTS:

* Documentation: Update HCP Boundary maintenance window documentation to specify that UTC timezone is expected. [[GH-518](https://github.com/hashicorp/terraform-provider-hcp/pull/518)]
* Update prerelease github action workflow to allow bypassing end-to-end tests via input variable. [[GH-520](https://github.com/hashicorp/terraform-provider-hcp/pull/520)]

BUG FIXES:

* Resolve failure to generate apply/destroy plans that include hcp_packer_channel resources that were deleted outside of Terraform [GH-519] [[GH-519](https://github.com/hashicorp/terraform-provider-hcp/pull/519)]
## v0.58.0 (May 26, 2023)

FEATURES:

* Adds `maintenenace_window_config` to the `hcp_boundary_cluster` resource to manage the timeframe for cluster upgrades. [[GH-500](https://github.com/hashicorp/terraform-provider-hcp/pull/500)]

IMPROVEMENTS:

* Documentation: HCP Packer Channel Resource is no longer in beta [[GH-513](https://github.com/hashicorp/terraform-provider-hcp/pull/513)]
## v0.57.0 (May 25, 2023)

FEATURES:

* Enable multi-project at provider level. [[GH-454](https://github.com/hashicorp/terraform-provider-hcp/pull/454)]

IMPROVEMENTS:

* Update CODEOWNERS [[GH-502](https://github.com/hashicorp/terraform-provider-hcp/pull/502)]
* Update `hcp-sdk-go` [[GH-510](https://github.com/hashicorp/terraform-provider-hcp/pull/510)]
## v0.56.0 (March 15, 2023)

IMPROVEMENTS:

* Bump github.com/hashicorp/hcp-sdk-go from 0.35.0 to 0.37.0 [[GH-482](https://github.com/hashicorp/terraform-provider-hcp/pull/482)]

BUG FIXES:

* Update pricing information for vault and consul. [[GH-480](https://github.com/hashicorp/terraform-provider-hcp/pull/480)]
## v0.55.0 (March 08, 2023)

FEATURES:

* New field `ip_allowlist` for `hcp_consul_cluster` to create, or update allowed IP address ranges (CIDRs) for inbound traffic. [[GH-455](https://github.com/hashicorp/terraform-provider-hcp/pull/455)]

IMPROVEMENTS:

* Add cluster scaling acceptance tests for Azure [[GH-465](https://github.com/hashicorp/terraform-provider-hcp/pull/465)]
* Bump github.com/hashicorp/hcp-sdk-go from 0.31.0 to 0.35.0 [[GH-458](https://github.com/hashicorp/terraform-provider-hcp/pull/458)]
* Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.24.1 to 2.25.0 [[GH-459](https://github.com/hashicorp/terraform-provider-hcp/pull/459)]
* Bump google.golang.org/grpc from 1.52.3 to 1.53.0 [[GH-452](https://github.com/hashicorp/terraform-provider-hcp/pull/452)]

BUG FIXES:

* resource/hcp_packer_channel: Fix documentation for incorrectly documented attributes [[GH-462](https://github.com/hashicorp/terraform-provider-hcp/pull/462)]
## v0.54.0 (February 15, 2023)

FEATURES:

* New resource `hcp_packer_channel` to create, or update an existing, channel with or without an assigned iteration. [[GH-435](https://github.com/hashicorp/terraform-provider-hcp/pull/435)]

IMPROVEMENTS:

* Bump github.com/hashicorp/hcp-sdk-go from 0.28.0 to 0.29.0 [[GH-431](https://github.com/hashicorp/terraform-provider-hcp/pull/431)]
* Bump github.com/hashicorp/hcp-sdk-go from 0.29.0 to 0.31.0 [[GH-445](https://github.com/hashicorp/terraform-provider-hcp/pull/445)]
* Bump google.golang.org/grpc from 1.51.0 to 1.52.3 [[GH-444](https://github.com/hashicorp/terraform-provider-hcp/pull/444)]
* Refactor HCP Vault TF acceptance test harness to include test for Azure [[GH-441](https://github.com/hashicorp/terraform-provider-hcp/pull/441)]
* resource/hcp_packer_channel: Label resource as public beta [[GH-457](https://github.com/hashicorp/terraform-provider-hcp/pull/457)]

BUG FIXES:

* Do not exit acceptance test when provider returns a warning [[GH-448](https://github.com/hashicorp/terraform-provider-hcp/pull/448)]
## v0.53.0 (January 20, 2023)

FEATURES:

* Automatically sync the public and internal repos. [[GH-436](https://github.com/hashicorp/terraform-provider-hcp/pull/436)]

IMPROVEMENTS:

* Add linting checks to PR and release pipeline. [[GH-430](https://github.com/hashicorp/terraform-provider-hcp/pull/430)]
* Update auth guide. [[GH-434](https://github.com/hashicorp/terraform-provider-hcp/pull/434)]
* Update hcp_consul_cluster and hcp_consul_cluster_root_token docs [[GH-439](https://github.com/hashicorp/terraform-provider-hcp/pull/439)]
* Use unique clusterIDs in acceptance tests [[GH-437](https://github.com/hashicorp/terraform-provider-hcp/pull/437)]

BUG FIXES:

* Fix issue with E2E tests failing [[GH-440](https://github.com/hashicorp/terraform-provider-hcp/pull/440)]
## v0.52.0 (December 14, 2022)

IMPROVEMENTS:

* Enable automatic changelog creation for dependabot PRs. [[GH-429](https://github.com/hashicorp/terraform-provider-hcp/pull/429)]
## v0.51.0 (December 08, 2022)

IMPROVEMENTS:

* Add E2E tests to auto release pipeline [[GH-421](https://github.com/hashicorp/terraform-provider-hcp/pull/421)]
* Bump github.com/hashicorp/hcp-sdk-go from 0.24.0 to 0.27.0 [[GH-424](https://github.com/hashicorp/terraform-provider-hcp/pull/424)]

BUG FIXES:

* Add check for "v" when compiling changelog [[GH-423](https://github.com/hashicorp/terraform-provider-hcp/pull/423)]
* Increase `hcp_consul_cluster` create timeout to 35 minutes [[GH-427](https://github.com/hashicorp/terraform-provider-hcp/pull/427)]
* The example usage for `hcp_azure_peering_connection` was missing the `vnet`
resource reference. [[GH-425](https://github.com/hashicorp/terraform-provider-hcp/pull/425)]
## v0.50.0 (November 30, 2022)

IMPROVEMENTS:

* Automatically update docs on auto release [[GH-419](https://github.com/hashicorp/terraform-provider-hcp/pull/419)]
* Bump google.golang.org/grpc from 1.50.1 to 1.51.0 [[GH-418](https://github.com/hashicorp/terraform-provider-hcp/pull/418)]
* Bumps github.com/go-openapi/runtime from 0.24.2 to 0.25.0 [[GH-422](https://github.com/hashicorp/terraform-provider-hcp/pull/422)]
* Set up auto release capability [[GH-411](https://github.com/hashicorp/terraform-provider-hcp/pull/411)]
## v0.49.0 (November 16, 2022)

IMPROVEMENTS:

* provider: Bump `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.24.0 to 2.24.1 ([GH-415](https://github.com/hashicorp/terraform-provider-hcp/pull/415))
* provider: Bump `github.com/hashicorp/hcp-sdk-go` from 0.23.0 to 0.24.0 ([GH-413](https://github.com/hashicorp/terraform-provider-hcp/pull/413))
* docs: Update the tutorial links ([GH-414](https://github.com/hashicorp/terraform-provider-hcp/pull/414))
* docs: Updates browser login documentation ([GH-412](https://github.com/hashicorp/terraform-provider-hcp/pull/412))

## 0.48.0 (November 9, 2022)

IMPROVEMENTS:

* provider: Bump `github.com/stretchr/testify` from 1.8.0 to 1.8.1 ([GH-408](https://github.com/hashicorp/terraform-provider-hcp/pull/408))
* provider: Auto detect latest Consul patch version ([GH-406](https://github.com/hashicorp/terraform-provider-hcp/pull/406))

## 0.47.0 (October 21, 2022)

IMPROVEMENTS:

* provider: Bump `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.23.0 to 2.24.0 ([GH-403](https://github.com/hashicorp/terraform-provider-hcp/pull/403))
* provider: Bump `github.com/go-openapi/runtime` from 0.24.1 to 0.24.2 ([GH-404](https://github.com/hashicorp/terraform-provider-hcp/pull/404))
* provider: Bump `google.golang.org/grpc` from 1.50.0 to 1.50.1 ([GH-405](https://github.com/hashicorp/terraform-provider-hcp/pull/405))

## 0.46.0 (October 13, 2022)

IMPROVEMENTS:

* provider: Update Mozilla Public License 2.0 [GH-402](https://github.com/hashicorp/terraform-provider-hcp/pull/402))
* provider: Bump `google.golang.org/grpc` from 1.49.0 to 1.50.0 ([GH-401](https://github.com/hashicorp/terraform-provider-hcp/pull/401))
* provider: Bump `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.21.0 to 2.23.0 ([GH-395](https://github.com/hashicorp/terraform-provider-hcp/pull/395))

FIXES:

* docs: Add note in vault resource that only admins can modify cluster tier and size ([GH-400](https://github.com/hashicorp/terraform-provider-hcp/pull/400))

## 0.45.0 (September 26, 2022)

IMPROVEMENTS:

* provider: Bump `github.com/hashicorp/hcp-sdk-go` from 0.21.0 to 0.23.0. Note this change introduces some breaking changes when using model enums. More information can be found in the [HCP Go SDK Release](https://github.com/hashicorp/hcp-sdk-go/releases/tag/v0.23.0)  ([GH-392](https://github.com/hashicorp/terraform-provider-hcp/pull/392))

FIXES:

* provider: Prevent FAILED peerings/attachments from failing deletes  ([GH-394](https://github.com/hashicorp/terraform-provider-hcp/pull/394))

## 0.44.0 (September 9, 2022)

FIXES:

* Azure/AWS Peering connections: Replace ReadContext with ReadWithoutTimeout  ([GH-389](https://github.com/hashicorp/terraform-provider-hcp/pull/389))
* Documentation: Correct `hvn_link` to `example` instead of `hvn` ([GH-388](https://github.com/hashicorp/terraform-provider-hcp/pull/388))

## 0.43.0 (August 31, 2022)

IMPROVEMENTS:

* provider: Bump `github.com/hashicorp/hcp-sdk-go` from 0.20.0 to 0.21.0 ([GH-381](https://github.com/hashicorp/terraform-provider-hcp/pull/381))
* provider: Added support for HCP Boundary Beta clusters ([GH-378](https://github.com/hashicorp/terraform-provider-hcp/pull/378))

FEATURES:

* Vault: enable changes on mvu configuration  ([GH-383](https://github.com/hashicorp/terraform-provider-hcp/pull/383))

## 0.42.0 (August 24, 2022)

IMPROVEMENTS:

* provider: Bump version of Go to 1.18.5 in `.go-version` ([GH-374](https://github.com/hashicorp/terraform-provider-hcp/pull/374))
* provider: Bump `google.golang.org/grpc` from 1.48.0 to 1.49.0 ([GH-379](https://github.com/hashicorp/terraform-provider-hcp/pull/379))

FIXES:

* all: Prevents the app from crashing when a `*url.Error` is received while retrying HTTP requests. ([GH-376](https://github.com/hashicorp/terraform-provider-hcp/pull/376))

## 0.41.0 (August 18, 2022)

IMPROVEMENTS:

* provider: Upgrade terraform-plugin-sdk to version 2.21.0 ([GH-371](https://github.com/hashicorp/terraform-provider-hcp/pull/371))

FIXES:

* provider: Updates README examples ([GH-368](https://github.com/hashicorp/terraform-provider-hcp/pull/368))
* provider: Fix root token example in documentation ([GH-372](https://github.com/hashicorp/terraform-provider-hcp/pull/372))

## 0.40.0 (August 11, 2022)

FIXES:

* provider: Updates codeowners ([GH-369](https://github.com/hashicorp/terraform-provider-hcp/pull/369))

## 0.39.0 (August 5, 2022)

FIXES:

* provider: Fixes codeowners which showed errors after a recent team name update ([GH-366](https://github.com/hashicorp/terraform-provider-hcp/pull/366))
* vault_cluster: Check type assertions to fix issue #360 ([GH-364](https://github.com/hashicorp/terraform-provider-hcp/pull/364))

IMPROVEMENTS:

* provider: Upgrade Go to version 1.18 ([GH-365](https://github.com/hashicorp/terraform-provider-hcp/pull/365))
* data_source_azure_peering_connection: Log failed peering wait errors ([GH-363](https://github.com/hashicorp/terraform-provider-hcp/pull/363))
* provider: Bump `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.19.0 to 2.20.0 ([GH-362](https://github.com/hashicorp/terraform-provider-hcp/pull/362))

## 0.38.0 (July 28, 2022)

IMPROVEMENTS:

* provider: Bump `hcp-sdk-go` from 0.19.0 to 0.20.0 ([#357](https://github.com/hashicorp/terraform-provider-hcp/pull/357))
* provider: Add retry logic on GET requests when fetching organization and project IDs ([#358](https://github.com/hashicorp/terraform-provider-hcp/pull/358))

## 0.37.0 (July 20,2022)

IMPROVEMENTS:

* provider: Bump `github.com/go-openapi/strfmt` from 0.21.2 to 0.21.3  ([#355](https://github.com/hashicorp/terraform-provider-hcp/pull/355))
* provider: Bump `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.18.0 to 2.19.0 ([#354](https://github.com/hashicorp/terraform-provider-hcp/pull/354))
* resource/vault: Add cross-referencing links to related tutorials ([#353](https://github.com/hashicorp/terraform-provider-hcp/pull/353))
* resource/azure_peering: fix acceptance test ([#349](https://github.com/hashicorp/terraform-provider-hcp/pull/349))

## 0.36.0 (July 13,2022)

IMPROVEMENTS:

* provider: Bump `google.golang.org/grpc` from 1.47.0 to 1.48.0 ([#351](https://github.com/hashicorp/terraform-provider-hcp/pull/351))
* provider: Bump `github.com/hashicorp/terraform-plugin-docs` from 0.12.0 to 0.13.0 ([#350](https://github.com/hashicorp/terraform-provider-hcp/pull/350))
* datasource/hcp_packer_image: Add `component_type` optional argument ([#347](https://github.com/hashicorp/terraform-provider-hcp/pull/347))

## 0.35.0 (July 07,2022)

IMPROVEMENTS:

* provider: Bump `github.com/stretchr/testify` from 1.7.2 to 1.7.4 ([#334](https://github.com/hashicorp/terraform-provider-hcp/pull/334))
* provider: Bump `github.com/hashicorp/go-version` from 1.5.0 to 1.6.0 ([#341](https://github.com/hashicorp/terraform-provider-hcp/pull/341))
* provider: Bump `github.com/hashicorp/terraform-plugin-docs` from 0.10.1 to 0.12.0 ([#342](https://github.com/hashicorp/terraform-provider-hcp/pull/342))
* provider: Bump `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.17.0 to 2.18.0 ([#344](https://github.com/hashicorp/terraform-provider-hcp/pull/344))
* provider: Bump `github.com/stretchr/testify` from 1.7.4 to 1.8.0 ([#346](https://github.com/hashicorp/terraform-provider-hcp/pull/346))

FEATURES:

* datasource/hcp_packer_image: allow `channel` attribute to get an image ([#339](https://github.com/hashicorp/terraform-provider-hcp/pull/339))

FIXES:

* resource_consul_cluster: only WARN on failed client config calls ([#345](https://github.com/hashicorp/terraform-provider-hcp/pull/345))

## 0.34.0 (June 30, 2022)

IMPROVEMENTS:

* docs: Refactor documentation for `hcp_hvn` resource ([337](https://github.com/hashicorp/terraform-provider-hcp/pull/337))

FIXES:

* resource/consul: Store cluster+snapshot state ([326](https://github.com/hashicorp/terraform-provider-hcp/pull/326))
* resource/vault: keep failed clusters, export state ([331](https://github.com/hashicorp/terraform-provider-hcp/pull/331))
* resource/hvn: keep failed networks/peerings, export state ([331](https://github.com/hashicorp/terraform-provider-hcp/pull/331))

## 0.33.0 (June 22, 2022)

IMPROVEMENTS:

* datasource/hcp_packer_image: Include `revoke_at` in the data source output ([330](https://github.com/hashicorp/terraform-provider-hcp/pull/330))
* datasource/hcp_packer_iteration: Include `revoke_at` in the data source output ([330](https://github.com/hashicorp/terraform-provider-hcp/pull/330))
* datasource/hcp_packer_image_iteration: Include `revoke_at` in the data source output ([330](https://github.com/hashicorp/terraform-provider-hcp/pull/330))

FIXES:

* docs: update HVN with Azure & make resource titles consistent ([#333](https://github.com/hashicorp/terraform-provider-hcp/pull/333))

## 0.32.0 (June 15, 2022)

IMPROVEMENTS:

* provider: Bump `github.com/hashicorp/terraform-plugin-docs` from 0.9.0 to 0.10.1 ([#328](https://github.com/hashicorp/terraform-provider-hcp/pull/328))
* provider: Fixes error handling when Terraform cannot connect to status.hashicorp.com ([#325](https://github.com/hashicorp/terraform-provider-hcp/pull/325))

## 0.31.0 (June 8, 2022)

IMPROVEMENTS:

* provider: Bump `google.golang.org/grpc` from 1.46.2 to 1.47.0 ([#316](https://github.com/hashicorp/terraform-provider-hcp/pull/316))
* provider: Bump `github.com/hashicorp/terraform-plugin-sdk/v2` from 2.16.0 to 2.17.0 ([#317](https://github.com/hashicorp/terraform-provider-hcp/pull/317))
* provider: Bump `github.com/hashicorp/terraform-plugin-docs` from 0.8.1 to 0.9.0 ([#318](https://github.com/hashicorp/terraform-provider-hcp/pull/318))
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
