# Resource Control Policy Feature Guide

This guide describes the `hcp_resource_control_policy` feature and its expected behavior.

## Feature Scope

The resource manages organization-level resource control policy constraints as a full document.

- Resource: `hcp_resource_control_policy`
- Scope: organization-level
- Primary APIs:
  - `ListConstraints`
  - `GetResourceControlPolicy`
  - `SetResourceControlPolicy`

## Implemented Behavior

### Create

1. Reads desired `enabled_constraints` from config.
2. Calls `ListConstraints` and validates all provided IDs.
3. Calls `SetResourceControlPolicy` with etag when available.
4. Performs read-after-write via `GetResourceControlPolicy`.
5. Writes canonical state.

### Read

1. Calls `GetResourceControlPolicy`.
2. Syncs `enabled_constraints`, `organization_id`, and `etag` into state.
3. Removes resource from state if policy is not found.

### Update

1. Reads plan and state.
2. Validates constraints using `ListConstraints`.
3. Calls `SetResourceControlPolicy` with the stored etag.
4. Performs read-after-write to refresh state.

### Delete

Delete clears all constraints (it does not delete the organization):

- Calls `SetResourceControlPolicy` with an empty constraints set.

### Import

1. Import ID is the organization ID.
2. Calls `GetResourceControlPolicy` and hydrates state.
3. Optionally cross-checks against `ListConstraints` and emits warnings for stale IDs.

## Etag and Conflict Handling

Writes use etag for optimistic concurrency.

- Conflict responses are detected and surfaced with actionable diagnostics.
- Current implementation returns a conflict diagnostic and instructs the user to rerun `terraform apply`.
- No internal retry loop is performed in the current implementation.

## Constraint Semantics

`enabled_constraints` is treated as a set.

- Order does not affect behavior.
- This avoids plan/apply mismatches when the backend returns constraints in a different order.

## Typical Terraform Usage

```hcl
resource "hcp_resource_control_policy" "example" {
  organization_id = data.hcp_organization.example.resource_id

  enabled_constraints = [
    "constraints/packer.deny-creation",
    "constraints/vault.deny-creation",
  ]
}
```

## Operational Behavior

If the policy changes outside Terraform after a plan is created, apply can return a conflict diagnostic. In that case, rerun `terraform apply` so Terraform refreshes the latest policy state before retrying.

## Validation and Testing

### Unit tests

```sh
go test ./internal/provider/resourcemanager/...
```

### Acceptance tests (feature-focused)

```sh
TF_ACC=1 go test ./internal/provider/resourcemanager -run TestAccOrganizationResourceControlPolicyResource -v
TF_ACC=1 go test ./internal/provider/resourcemanager -run TestAccOrganizationResourceControlPolicyResource_InvalidConstraint -v
```

### Generated docs check

```sh
make gencheck
```

If `gencheck` reports drift, run generation and commit resulting docs updates.
