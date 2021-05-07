<!--
Adding a new resource or datasource? Use this checklist to get started: https://github.com/hashicorp/terraform-provider-hcp/blob/main/contributing/checklist-resource.md

Hashicorp contributors, please consider: what stage of release is your feature in?
If the feature is for internal Hashicorp usage only, it should be maintained on a feature branch until ready for beta release.
If the feature is for select beta users, it can be merged to main and released as a new minor version. A beta banner must be added to the documentation.
If the feature is ready for all HCP users, it can be merged to main and released as a new minor version. You may wish to coordinate with the release of the feature in the UI.
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
