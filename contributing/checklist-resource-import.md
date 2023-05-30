# Adding Resource Import Support Checklist

Adding import support for Terraform resources will allow existing infrastructure to be managed within Terraform. This type of enhancement generally requires a small to moderate amount of code changes.

Comprehensive code examples and information about resource import support can be found in the [Extending Terraform documentation](https://www.terraform.io/docs/extend/resources/import.html).

- [ ] __Uses Context-Aware Import Function__: The context-aware `StateContext` function should be used over the deprecated `State` function.
- [ ] __Supports Optional Project ID In Import Identifier__: If the resource has multi-project support, there should be an optional `{project_id}:` prefix in the import identifier. If the user does not provide a project ID explicitly, the provider uses the authentication scope to determine the oldest accessible project. This prevents the user from needing to locate and provide their project ID for single-project organizations. `client.Config.ProjectID` should be used to retrieve the implied project ID.
- [ ] __Uses Passthrough If Possible__: If the import identifier can match the `id` of the resource, and this does not violate any other guidelines, the `ImportStatePassthroughContext` passthrough should be used.
- [ ] __Specifies Minimal Import Identifier__: If more than one value needs to be specified in the import identifier, the minimal number of values should be used, and those values should be colon (`:`) separated.
- [ ] __Includes Import Documentation__: There should be an import example at `examples/resources/<resource>/import.sh`, which will be used when generating the docs. The docs should then be regenerated using `go generate`, which will update files in the `docs/` directory.
