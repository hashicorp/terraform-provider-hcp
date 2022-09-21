# Writing Acceptance Tests

Terraform includes an acceptance test harness that does most of the repetitive
work involved in testing a resource. Please read the [Extending Terraform documentation](https://www.terraform.io/docs/extend/testing/index.html)
first for a general understanding of how this test harness works. The guide
below is meant to augment that documentation with information specific to the
HCP provider.

## Acceptance Tests Often Cost Money to Run

Because acceptance tests create real resources, they often cost money to run.
Because the resources only exist for a short period of time, the total amount
of money required is usually a relatively small. Nevertheless, we don't want
financial limitations to be a barrier to contribution, so if you are unable to
pay to run acceptance tests for your contribution, mention this in your
pull request. We will happily accept "best effort" implementations of
acceptance tests and run them for you on our side. This might mean that your PR
takes a bit longer to merge, but it most definitely is not a blocker for
contributions.

## Acceptance Tests Take a While to Run

These acceptance tests create real resources, some of which can take up to 15-20 minutes each to spin up and tear down. For this reason, we urge contributors to consolidate acceptance tests on related resources in one test file to avoid creating multiples of the same resource over the course of the tests. Data sources and dependent resources can be consolidated into their corresponding resource test.

For example, the `resource_vault_cluster_test.go` reuses one test config to test the Vault cluster resource, the Vault cluster datasource, and the dependent Vault cluster admin token resource. This helps speed up the acceptance test runtime by creating a Vault cluster, the most time-intensive resource, only once.

Exceptions may be made when the HCL required to test a single resource is particularly complex, as in `resource_aws_transit_gateway_attachment_test.go`. In such cases, test readability should be preferred over consolidation.

## Running an Acceptance Test

Acceptance tests can be run using the `testacc` target in the
`GNUmakefile`. The individual tests to run can be controlled using a regular
expression. Prior to running the tests, provider configuration details such as
access keys must be made available as environment variables:

```sh
export HCP_CLIENT_ID=...
export HCP_CLIENT_SECRET=...
```

For some tests AWS credentials are also required in order to create "customer"
resources for testing certain HCP features (network peerings and TGW attachments):

```sh
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_SESSION_TOKEN=...
```

**Note for HCP developers**: this AWS account **MUST NOT** be the same AWS account that is being used by
your HCP organization (dataplane) otherwise tests would fail.

Tests can then be run by specifying a regular expression defining the tests to run:

```sh
$ make testacc TESTARGS='-run=TestAccConsulCluster'
TF_ACC=1 go test ./... -v -run=TestAccConsulCluster -timeout 120m
ok  	github.com/hashicorp/terraform-provider-hcp/internal/consul	(cached) [no tests to run]
=== RUN   TestAccConsulCluster
--- PASS: TestAccConsulCluster (538.48s)
PASS
ok  	github.com/hashicorp/terraform-provider-hcp/internal/provider	539.112s
```

Entire resource test suites can be targeted by using the naming convention to
write the regular expression.

For advanced developers, the acceptance testing framework accepts some additional environment variables that can be used to control Terraform CLI binary selection, logging, and other behaviors. See the [Extending Terraform documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html#environment-variables) for more information.

## Writing an Acceptance Test

Terraform has a framework for writing acceptance tests which minimizes the
amount of boilerplate code necessary to use common testing patterns. The entry
point to the framework is the `resource.Test()` function.

Tests are divided into `TestStep`s. Each `TestStep` proceeds by applying some
Terraform configuration using the provider under test, and then verifying that
results are as expected by making assertions using the provider API. It is
common for a single test function to exercise both the creation of and updates
to a single resource. Most tests follow a similar structure.

1. Pre-flight checks are made to ensure that sufficient provider configuration
   is available to be able to proceed - in the case of HCP, `HCP_CLIENT_ID` and `HCP_CLIENT_SECRET` must be set prior
   to running acceptance tests. This is common to all tests exercising a single
   provider.

Each `TestStep` is defined in the call to `resource.Test()`. Most assertion
functions are defined out of band with the tests. This keeps the tests
readable, and allows reuse of assertion functions across different tests of the
same type of resource. The definition of a complete test looks like this:

```go
func TestAccConsulCluster(t *testing.T) {
    resourceName := "hcp_consul_cluster.test"

    resource.Test(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t) },
        ProviderFactories: providerFactories,
        CheckDestroy:      testAccCheckConsulClusterDestroy,
        Steps: []resource.TestStep{
            {
                Config: testConfig(testAccConsulClusterConfig),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckConsulClusterExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "cluster_id", "test-consul-cluster"),
                ),
            },
        },
    })
}
```

When executing the test, the following steps are taken for each `TestStep`:

1. The Terraform configuration required for the test is applied. This is
   responsible for configuring the resource under test, and any dependencies it
   may have. For example, to test the `hcp_consul_cluster` resource, an
   `hcp_hvn` is required. This results in configuration which
   looks like this:

   ```hcl
   resource "hcp_hvn" "test" {
       hvn_id         = "test-hvn"
       cloud_provider = "aws"
       region         = "us-west-2"
   }

   resource "hcp_consul_cluster" "test" {
       cluster_id = "test-consul-cluster"
       hvn_id     = hcp_hvn.test.hvn_id
       tier       = "development"
   }
   ```

   **Note:** Use spaces instead of tabs in test HCL.

1. Assertions are run using the provider API. These use the provider API
   directly rather than asserting against the resource state. For example, to
   verify that the `hcp_consul_cluster` described above was created
   successfully, a test function like this is used:

   ```go
   func testAccCheckConsulClusterExists(name string) resource.TestCheckFunc {
       return func(s *terraform.State) error {
           rs, ok := s.RootModule().Resources[name]
           if !ok {
               return fmt.Errorf("not found: %s", name)
           }

           id := rs.Primary.ID
           if id == "" {
               return fmt.Errorf("no ID is set")
           }

           client := testAccProvider.Meta().(*clients.Client)

           link, err := buildLinkFromURL(id, ConsulClusterResourceType, client.Config.OrganizationID)
           if err != nil {
               return fmt.Errorf("unable to build link for %q: %v", id, err)
           }

           clusterID := link.ID
           loc := link.Location

           if _, err := clients.GetConsulClusterByID(context.Background(), client, loc, clusterID); err != nil {
               return fmt.Errorf("unable to read Consul cluster %q: %v", id, err)
           }

           return nil
       }
   }
   ```

   Notice that the only information used from the Terraform state is the ID of
   the resource - though in this case it is necessary to split the ID into
   constituent parts in order to use the provider API. For computed properties,
   we instead assert that the value saved in the Terraform state was the
   expected value if possible. The testing framework provides helper functions
   for several common types of check - for example:

   ```go
   resource.TestCheckResourceAttr(resourceName, "cluster_id", "test-consul-cluster"),
   ```

1. The resources created by the test are destroyed. This step happens
   automatically, and is the equivalent of calling `terraform destroy`.

1. Assertions are made against the provider API to verify that the resources
   have indeed been removed. If these checks fail, the test fails and reports
   "dangling resources". The code to ensure that the `hcp_consul_cluster` shown
   above is removed looks like this:

   ```go
   func testAccCheckConsulClusterDestroy(s *terraform.State) error {
       client := testAccProvider.Meta().(*clients.Client)

       for _, rs := range s.RootModule().Resources {
           switch rs.Type {
           case "hcp_consul_cluster":
               id := rs.Primary.ID

               link, err := buildLinkFromURL(id, ConsulClusterResourceType, client.Config.OrganizationID)
               if err != nil {
                   return fmt.Errorf("unable to build link for %q: %v", id, err)
               }

               clusterID := link.ID
               loc := link.Location

               _, err = clients.GetConsulClusterByID(context.Background(), client, loc, clusterID)
               if err == nil || !clients.IsResponseCodeNotFound(err) {
                   return fmt.Errorf("didn't get a 404 when reading destroyed Consul cluster %q: %v", id, err)
               }

           default:
               continue
           }
       }
       return nil
   }
   ```

   These functions usually test only for the resource directly under test.
