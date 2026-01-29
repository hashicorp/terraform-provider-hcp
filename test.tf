// Pin the version
terraform {
  required_providers {
    hcp = {
      # source  = "hashicorp/hcp"
      # version = "~> 0.106.0"
      # to build locally
      source  = "localhost/providers/hcp"
      version = "0.0.1"
    }
  }
}
// Configure the provider
provider "hcp" {
  project_id = "9e7c2967-1622-4724-adba-fd50a12e3db9"
}

data "hcp_organization" "example" {}

variable "roles" {
  default = {
    ## General
    #"client.hcp.general" = {
    #description = "General access"
    #role        = "roles/viewer"
    #}
    ##Privileged
    "client.hcp.owner" = {
      description = "Global owner"
      role        = "roles/owner" # Has all of the admin's permissions, and also the ability to delete the organization and promote/demote other owners.
    }
    "client.hcp.admin" = {
      description = "Full org admin"
      role        = "roles/admin" # Has full access to all resources including the right to edit IAM, invite users, edit roles.
    }
    #"client.hcp.contributor" = {
    #description = "Full viewer with project create"
    #role        = "roles/contributor" # Can create and manage all types of resources. Can only view IAM.
    #}
    #"client.hcp.viewer" = {
    #description = "Full viewer"
    #role        = "roles/viewer" # Can view existing resources. Cannot create new resources or edit existing ones. Cannot view the org IAM page.
    #}
    #"client.hcp.browser" = {
    #description = "Project viewer"
    #role        = "roles/resource-manager.browser" # Allows browsing the resource hierarchy of the organization including listing and getting all projects. This does not allow accessing resources inside a project.
    #}
    ## IAM
    #"client.hcp.org.iam" = {
    #description = "Organization IAM Policies Administrator"
    #role        = "roles/resource-manager.org-iam-policies-admin" # Allows performing CRUD operations on organization policies.
    #}
    #"client.hcp.project.iam" = {
    #description = "Project IAM Policies Administrator"
    #role        = "roles/resource-manager.project-iam-policies-admin" # Allows performing CRUD operations on project policies.
    #}
    #"client.hcp.project.creator" = {
    #description = "Project Creator"
    #role        = "roles/resource-manager.project-creator" # Allows creating projects.
    #}
    #"client.hcp.billing" = {
    #description = "Billing administrator"
    #role        = "roles/billing.billing-admin" # Allows managing billing resources.
    #}
    #"client.hcp.iam.group" = {
    #description = "Group Administrator"
    #role        = "roles/iam.group-admin" # Allows performing CRUD operations on groups and managing group members.
    #}
    #"client.hcp.iam.sso" = {
    #description = "SSO Administrator"
    #role        = "roles/iam.sso-admin" # Allows performing CRUD operations on SSO and SCIM configuration for an organization.
    #}
    ## Infragraph
    #"client.hcp.infragraph.admin" = {
    #description = "Infragraph Admin"
    #role        = "roles/infragraph.infragraph-admin" # This role is a collection of permissions that allows principals to create/read/update/delete connections, saved queries, query the graph, view inventory, and assign the Querier role to a principal.
    #}
    #"client.hcp.infragraph.viewer" = {
    #description = "Infragraph Querier"
    #role        = "roles/infragraph.infragraph-querier" # This role is a collection of permissions that allows principals to read/list connections and saved queries, query the graph, and view inventory.
    #}
    ## Vault
    #"client.hcp.vault.app.manager" = {
    #description = "Vault Secrets App Manager"
    #role        = "roles/secrets.app-manager" # This role is a collection of permissions that allows principals to create/read/update/delete secrets and add sync mappings inside a Vault Secrets app
    #}
    #"client.hcp.vault.app.reader" = {
    #description = "Vault Secrets App Secret Reader"
    #role        = "roles/secrets.app-secret-reader" # This role is a collection of permissions that allows principals to read secret values inside an Vault Secrets app
    #}
    #"client.hcp.vault.integration.manager" = {
    #description = "Vault Secrets Integration Manager"
    #role        = "roles/secrets.integration-manager" # This role is a collection of permissions that allows principals to create/read/update/delete integrations inside a Vault Secrets project
    #}
    #"client.hcp.vault.integration.reader" = {
    #description = "Vault Secrets Integration Secret Reader"
    #role        = "roles/secrets.integration-reader" # This role is a collection of permissions that allows principals to read integrations inside a Vault Secrets project
    #}
    #"client.hcp.vault.inventory.reader" = {
    #description = "Vault Secrets Inventory Report Reader"
    #role        = "roles/vault-reporting.secrets-inventory-reader" # This role is a collection of permissions that allows fetching secrets inventory reports
    #}
  }
}
# Create groups
resource "hcp_group" "org_role" {
  for_each     = var.roles
  display_name = each.key
  description  = each.value.description
}
# Add IAM bindings to groups
resource "hcp_organization_iam_binding" "binding" {
  for_each     = { for k, v in var.roles : k => v if v.role != null }
  principal_id = hcp_group.org_role[each.key].resource_id
  role         = each.value.role
}

