# Breaking Changes Guide

Breaking changes are changes to resource schema that result in unexpected changes or the destruction of existing resources. They can be especially dangerous when the `--auto-approve` flag is passed in.

## Common Breaking Changes

- adding ForceNew to an existing required property

- adding ForceNew to an optional property without `Computed = true`

## How to Test Existing Resources

Our acceptance tests create new resources on each run, so there is no way to verify the effects of a change on existing resources. This requires manual testing on your branch:

1. Run `terraform init && terraform apply` with the provider configured to any previous version and whichever resource(s) you wish to test.

```hcl
terraform {
  required_providers {
    hcp = {
      source  = "hashicorp/hcp"
      version = "~> 0.25.0"
    }
  }
}
```

1. Run `rm .terraform.lock.hcl && make dev` to build your local binary. Configure the provider to your local binary.

```hcl
terraform {
  required_providers {
    // hcp = {
    //   source  = "hashicorp/hcp"
    //   version = "~> 0.25.0"
    // }

    hcp = {
     source  = "localhost/providers/hcp"
     version = "0.0.1"
    }
  }
}
```

1. Run `terraform init && terraform apply`. Verify that there are no changes to apply. If there are changes, you may need to follow one of the guides below to safely make the change.

## How to Make Breaking Changes Safely

Follow the [Deprecations guide](https://www.terraform.io/plugin/sdkv2/best-practices/deprecations) when renaming or removing a resource's attribute or the resource itself.

Follow the [State Migration guide](https://www.terraform.io/plugin/sdkv2/resources/state-migration) when adjusting the value of an attribute on existing resources.
