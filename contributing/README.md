# Contributing to the HCP Terraform Provider

This directory contains documentation about the HCP Terraform Provider codebase, aimed at readers who are interested in making code contributions.

To learn more about how to create issues and pull requests in this repository, and what happens after they are created, you may refer to the resources below:
- [Pull Request submission and lifecycle](pull-request-lifecycle.md)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.18

## Building the Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using `make dev`. This will place the provider onto your system in a [Terraform 0.13-compliant](https://www.terraform.io/upgrade-guides/0-13.html#in-house-providers) manner.

You'll need to ensure that your Terraform file contains the information necessary to find the plugin when running `terraform init`. `make dev` will use a version number of 0.0.1, so the following block will work:

```hcl
terraform {
  required_providers {
    hcp = {
      source  = "localhost/providers/hcp"
      version = "0.0.1"
    }
  }
}
```

## Testing the Provider

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run. Please read [Writing Acceptance Tests](writing-tests.md) in the contribution guidelines for more information on usage.

```sh
$ make testacc
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Generating Docs

To generate or update documentation, run `go generate`.
```shell script
$ go generate
```

## Changelogs

This repo requires that a chagnelog file be added in all pull requests. The name of the file must follow `[PR #].txt` and must reside in the `.changelog` directory. The contents must have the following formatting:

~~~
```release-note:TYPE
ENTRY
```
~~~

Where `TYPE` is the type of release note entry this is. This is one of either: `breaking-change`, `security`, `feature`, `improvement`, `deprecation`, `bug`.

`ENTRY` is the body of the changelog entry, and should describe the changes that were made. This is used as free-text input and will be returned to you as it is entered when generating the changelog.

Sometimes PRs have multiple changelog entries associated with them. In this case, use multiple blocks.

~~~
```release-note:deprecation
Deprecated the `foo` interface, please use the `bar` interface instead.
```

```release-note:improvement
Added the `bar` interface.
```
~~~


## Checklists

The following checklists are meant to be used for PRs to give developers and reviewers confidence that the proper changes have been made:

* [New resource](checklist-resource.md)
* [Adding resource import support](checklist-resource-import.md)

## References

The [reference documentation](references.md) includes more background material on specific functionality. This documentation is intended for developers extending or updating the Terraform HCP Provider. Typical operators writing and applying Terraform configurations do not need to read or understand this material.
