resource "hcp_waypoint_template" "example" {
  name                            = "waypoint-template-name"
  summary                         = "short summary for waypoint template"
  terraform_no_code_module_source = "private/tfcorg/modulename/providername"
  terraform_project_id            = "prj-j5pmTUfmstDi6okP"
  variable_options = [
    {
      "name" : "storage_account_name",
      "options" : [
        "azure_storage_account"
      ],
      "user_editable" : false,
      "variable_type" : "string"
    },
    {
      "name" : "web_app_name",
      "options" : [],
      "user_editable" : true,
      "variable_type" : "string"
  }]
}
