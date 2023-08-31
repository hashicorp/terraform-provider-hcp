# New Resource Checklist

Implementing a new resource is a good way to learn more about how Terraform interacts with upstream APIs. There are plenty of examples to draw from in the existing resources, but you still get to implement something completely new.

**_NOTE:_** Please put all new resources under the /internal/provider folder and utilize the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) when developing a new resource or data source.

- [ ] __Minimal LOC__: It can be inefficient for both the reviewer and author to go through long feedback cycles on a big PR with many resources. We therefore encourage you to only submit **1 resource at a time**.
- [ ] __Acceptance Tests__: New resources should include acceptance tests covering their behavior. See [Writing Acceptance Tests](writing-tests.md) below for a detailed guide on how to approach these.
- [ ] __Documentation__: Each resource gets a page in the [Terraform Registry documentation](https://registry.terraform.io/providers/hashicorp/hcp/latest/docs). For a new resource, you'll want to add an example and field descriptions. A guide is required if the new feature requires multiple dependent resources to use.
- [ ] __Well-formed Code__: Do your best to follow existing conventions you see in the codebase, and ensure your code is formatted with `go fmt`. The PR reviewers can help out on this front, and may provide comments with suggestions on how to improve the code.

## Schema

- [ ] __Uses Globally Unique ID__: The `id` field needs to be globally unique. Since many of the HCP services use IDs that are only unique within a particular project, you may need to create an `id` for Terraform using the `linkURL()` helper function. This function will produce an `id` of the following format: `/project/<project_id>/<resource_type>/<resource_id>`. If the service uses `ID` for a resource ID that is not globally unique, the resource ID should be specified in the Terraform schema as `<resource_type>_id`.
- [ ] __Validates Fields Where Possible__: All fields that can be validated client-side should include a `ValidateFunc` or `ValidateDiagFunc`.
These validations should favor validators provided by this project, or [Terraform `helper/validation` package](https://godoc.org/github.com/hashicorp/terraform/helper/validation) functions.
- [ ] __Supports Optional Project ID Input__: If the resource has multi-project support, there should be an optional input field for `project_id` in the schema. If the user does not provide a `project_id` for the resource, `client.Config.ProjectID` should be used to retrieve the implied project ID from the provider. The user may provide a `project_id` to the provider, otherwise the provider uses the authentication scope to determine the oldest accessible project. This prevents the user from needing to locate and provide their project ID for single-project organizations.

## CRUD Operations

- [ ] __Uses Context-Aware CRUD Functions__: The context-aware CRUD functions (eg. `CreateContext`, `ReadContext`, etc.) should be used over their deprecated counterparts (eg. `Create`, `Read`, etc.).
- [ ] __Uses Context For API Calls__: The `context.Context` that is passed into the CRUD functions should be passed into all API calls, most often by setting the `Context` field of a `*Params` object. This allows the API calls to be cancelled properly by Terraform.
- [ ] __Uses DRY API Calls__: API calls that are needed in more than one resource should be encapsulated in a wrapper function located in `internal/clients/`. This prevents verbose request blocks from being duplicated across the codebase. Wrapper functions can also be implemented for single-use API requests, but are not required.
- [ ] __Handles Existing Resource Prior To Create__: Before calling the API creation function, there should be a check to ensure that the resource does not already exist. If it does exist, the user should see a helpful log message that they may need to import the resource.
- [ ] __Implements Immediate Resource ID Set During Create__: Immediately after calling the API creation function, the resource ID should be set with [`d.SetId()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.SetId) before other API operations or returning.
- [ ] __Refreshes Attributes During Read__: All attributes available in the API should have [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) called to set their values in the Terraform state during the `Read` function.
- [ ] __Handles Removed Resource During Read__: If a resource is removed outside of Terraform (e.g. via different tool, API, or web UI), `d.SetId("")` and `return nil` can be used in the resource `Read` function to trigger resource recreation. When this occurs, a warning log message should be printed beforehand.
- [ ] __Handles Failed State During Read__: If a resource fails during an operation and ends up in a failed state, `d.SetId("")` and `return nil` can be used in the resource `Read` function to trigger resource recreation. When this occurs, a warning log message should be printed beforehand.
- [ ] __Handles Nonexistent Resource Prior To Delete__: Before calling the API deletion function, there should be a check to ensure that the resource exists. If it does not exist, the user should see a helpful log message that no action was taken.

## Documentation

- [ ] __Includes Descriptions For Resource And Fields__: The resource and all fields in the schema should include a `Description`, which will be used when generating the docs.
- [ ] __Includes Example__: The resource should include an example in `examples/resources/<resource>/resource.tf`, which will be used when generating the docs.
- [ ] __Includes Generated Docs__: The docs should be regenerated using `go generate`, which will update files in the `docs/` directory.
