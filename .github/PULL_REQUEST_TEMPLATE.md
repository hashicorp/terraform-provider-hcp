<!--
Adding a new resource or datasource? Use this checklist to get started: https://github.com/hashicorp/terraform-provider-hcp/blob/main/contributing/checklist-resource.md

Updating a resource? Avoid accidental breaking changes by reviewing this guide: https://github.com/hashicorp/terraform-provider-hcp/blob/main/contributing/breaking-changes.md
-->

### :hammer_and_wrench: Description

<!-- What code changed, and why? If adding a new resource, what is it and what are its key features? If updating an existing resource, what new functionality was added? -->

### :building_construction: Acceptance tests

- [ ] Are there any feature flags that are required to use this functionality?
- [ ] Have you added an acceptance test for the functionality being added?
- [ ] Have you run the acceptance tests on this branch?

Output from acceptance testing:

<!--
Replace TestAccXXX with a pattern that matches the tests affected by this PR. More info on acceptance tests here: https://github.com/hashicorp/terraform-provider-hcp/blob/main/contributing/writing-tests.md

For more information on the `-run` flag, see the `go test` documentation at https://tip.golang.org/cmd/go/#hdr-Testing_flags.
-->
```
$ make testacc TESTARGS='-run=TestAccXXX'

...
```
