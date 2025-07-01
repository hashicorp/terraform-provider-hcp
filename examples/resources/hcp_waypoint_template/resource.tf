resource "tfe_project" "example" {
  name         = "waypoint-build-destination"
  organization = var.org_name
}

data "tfe_registry_module" "example" {
  organization    = var.org_name
  name            = "my-nocode-example-module"
  module_provider = "aws"
}

resource "hcp_waypoint_template" "example" {
  name                            = "example-aws-template"
  summary                         = "AWS waypoint deployment."
  description                     = "Deploys a nocode module."
  terraform_project_id            = tfe_project.example.id
  labels                          = ["pets"]
  terraform_no_code_module_source = data.tfe_registry_module.example.no_code_module_source
  terraform_no_code_module_id     = data.tfe_registry_module.example.no_code_module_id
  variable_options = [
    {
      name          = "resource_size"
      user_editable = true
      options       = ["small", "medium", "large"]
    },
    {
      name          = "service_port"
      user_editable = false
      options       = ["8080"]
    },
  ]
}
