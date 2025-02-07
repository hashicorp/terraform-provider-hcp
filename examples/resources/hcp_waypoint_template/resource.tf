resource "hcp_waypoint_template" "template" {
  name                            = "go-k8s-microservice"
  summary                         = "A simple Go microservice running on Kubernetes."
  description                     = <<EOF
This template deploys a simple Go microservice to Kubernetes. The microservice
is a simple HTTP server that listens on port 8080 and returns a JSON response.
The template includes a Dockerfile, Kubernetes manifests, and boiler plate code
for a gRPC service written in Go.
EOF
  terraform_project_id            = "prj-123456"
  labels                          = ["go", "kubernetes"]
  terraform_no_code_module_source = "private/fake-org/go-k8s-microservice/kubernetes"
  terraform_no_code_module_id     = "nocode-123456"
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
