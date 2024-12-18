resource "hcp_waypoint_add_on_definition" "add_on_definition" {
  name                            = "postgres"
  summary                         = "An add-on that provisions a PostgreSQL database."
  description                     = <<EOF
This add-on provisions a PostgreSQL database in AWS. The database is provisioned
with a default schema and user.
EOF
  terraform_project_id            = "prj-123456"
  labels                          = ["postgres", "aws", "db"]
  terraform_no_code_module_source = "private/fake-org/postgres-aws/aws"
  terraform_no_code_module_id     = "nocode-abcdef"
  variable_options = [
    {
      name          = "size"
      user_editable = true
      options       = ["small", "medium", "large"]
    }
  ]
}