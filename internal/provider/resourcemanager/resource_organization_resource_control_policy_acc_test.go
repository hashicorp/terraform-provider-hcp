// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccOrganizationResourceControlPolicyResource(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", resource.EnvTfAcc)
	}

	resourceName := "hcp_resource_control_policy.test"
	constraintIDs := testAccOrganizationResourceControlPolicyConstraintIDs(t, 2)
	oneConstraint := constraintIDs[:1]
	twoConstraints := constraintIDs[:2]

	fixture := newOrganizationResourceControlPolicyFixture(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationResourceControlPolicyDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationResourceControlPolicyConfig(oneConstraint),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "organization_id", fixture.orgID),
					resource.TestCheckResourceAttr(resourceName, "id", fixture.orgID),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					testAccCheckOrganizationResourceControlPolicyBackend(t, oneConstraint),
					testAccCheckOrganizationResourceControlPolicyConstraintSet(resourceName, oneConstraint),
				),
			},
			{
				Config:   testAccOrganizationResourceControlPolicyConfig(oneConstraint),
				PlanOnly: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccOrganizationResourceControlPolicyImportID,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationResourceControlPolicyConfig(twoConstraints),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "organization_id", fixture.orgID),
					resource.TestCheckResourceAttr(resourceName, "id", fixture.orgID),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					testAccCheckOrganizationResourceControlPolicyBackend(t, twoConstraints),
					testAccCheckOrganizationResourceControlPolicyConstraintSet(resourceName, twoConstraints),
				),
			},
			{
				Config:   testAccOrganizationResourceControlPolicyConfig(twoConstraints),
				PlanOnly: true,
			},
		},
	})
}

func TestAccOrganizationResourceControlPolicyResource_InvalidConstraint(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", resource.EnvTfAcc)
	}

	constraintIDs := testAccOrganizationResourceControlPolicyConstraintIDs(t, 1)
	config := testAccOrganizationResourceControlPolicyConfig([]string{constraintIDs[0], "constraints/does-not-exist"})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`Unrecognized constraint ID`),
			},
		},
	})
}

type organizationResourceControlPolicyFixture struct {
	orgID               string
	originalConstraints []string
}

func newOrganizationResourceControlPolicyFixture(t *testing.T) organizationResourceControlPolicyFixture {
	t.Helper()

	client := acctest.HCPClients(t)
	orgID := client.Config.OrganizationID
	if orgID == "" {
		t.Fatal("test provider client is missing organization ID")
	}

	policy := testAccGetOrganizationResourceControlPolicy(t, orgID)
	originalConstraints := append([]string{}, policy.EnabledConstraints...)
	sort.Strings(originalConstraints)

	t.Cleanup(func() {
		testAccSetOrganizationResourceControlPolicy(t, orgID, originalConstraints)
	})

	return organizationResourceControlPolicyFixture{
		orgID:               orgID,
		originalConstraints: originalConstraints,
	}
}

func testAccOrganizationResourceControlPolicyConstraintIDs(t *testing.T, minCount int) []string {
	t.Helper()

	client := acctest.HCPClients(t)
	orgID := client.Config.OrganizationID
	if orgID == "" {
		t.Fatal("test provider client is missing organization ID")
	}

	params := organization_service.NewOrganizationServiceListConstraintsParamsWithContext(context.Background())
	params.ID = orgID

	res, err := client.Organization.OrganizationServiceListConstraints(params, nil)
	if err != nil {
		t.Fatalf("list constraints failed: %v", err)
	}

	available := map[string]struct{}{}
	for _, constraint := range res.GetPayload().Constraints {
		if constraint != nil && constraint.ID != "" {
			available[constraint.ID] = struct{}{}
		}
	}

	preferred := []string{
		"constraints/packer.deny-creation",
		"constraints/vault.deny-creation",
		"constraints/vagrant.deny-boxes-creation",
	}

	selected := make([]string, 0, minCount)
	seen := map[string]struct{}{}
	for _, candidate := range preferred {
		if _, ok := available[candidate]; ok {
			selected = append(selected, candidate)
			seen[candidate] = struct{}{}
			if len(selected) == minCount {
				return selected
			}
		}
	}

	remaining := make([]string, 0, len(available))
	for id := range available {
		if _, ok := seen[id]; !ok {
			remaining = append(remaining, id)
		}
	}
	sort.Strings(remaining)

	for _, id := range remaining {
		selected = append(selected, id)
		if len(selected) == minCount {
			return selected
		}
	}

	t.Skipf("acceptance test requires at least %d available resource control policy constraints, found %d", minCount, len(selected))
	return nil
}

func testAccOrganizationResourceControlPolicyConfig(constraintIDs []string) string {
	quoted := make([]string, 0, len(constraintIDs))
	for _, id := range constraintIDs {
		quoted = append(quoted, fmt.Sprintf("%q", id))
	}

	return fmt.Sprintf(`
data "hcp_organization" "test" {}

resource "hcp_resource_control_policy" "test" {
	organization_id     = data.hcp_organization.test.resource_id
	enabled_constraints = [%s]
}
`, strings.Join(quoted, ", "))
}

func testAccOrganizationResourceControlPolicyImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["hcp_resource_control_policy.test"]
	if !ok {
		return "", fmt.Errorf("resource not found")
	}

	id, ok := rs.Primary.Attributes["organization_id"]
	if !ok || id == "" {
		return "", fmt.Errorf("organization_id not set")
	}

	return id, nil
}

func testAccCheckOrganizationResourceControlPolicyConstraintSet(resourceName string, expected []string) resource.TestCheckFunc {
	checks := make([]resource.TestCheckFunc, 0, len(expected))
	for _, id := range expected {
		checks = append(checks, resource.TestCheckTypeSetElemAttr(resourceName, "enabled_constraints.*", id))
	}

	return func(s *terraform.State) error {
		for _, check := range checks {
			if err := check(s); err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccCheckOrganizationResourceControlPolicyBackend(t *testing.T, expected []string) resource.TestCheckFunc {
	t.Helper()

	return func(_ *terraform.State) error {
		policy := testAccGetOrganizationResourceControlPolicy(t, acctest.HCPClients(t).Config.OrganizationID)

		got := append([]string{}, policy.EnabledConstraints...)
		exp := append([]string{}, expected...)
		sort.Strings(got)
		sort.Strings(exp)

		if len(got) != len(exp) {
			return fmt.Errorf("expected %d enabled constraints, got %d (%v)", len(exp), len(got), got)
		}

		for i := range exp {
			if got[i] != exp[i] {
				return fmt.Errorf("expected enabled constraints %v, got %v", exp, got)
			}
		}

		return nil
	}
}

func testAccCheckOrganizationResourceControlPolicyDestroy(t *testing.T) func(*terraform.State) error {
	t.Helper()

	return func(_ *terraform.State) error {
		policy := testAccGetOrganizationResourceControlPolicy(t, acctest.HCPClients(t).Config.OrganizationID)
		if len(policy.EnabledConstraints) != 0 {
			return fmt.Errorf("expected delete to clear all enabled constraints, got %v", policy.EnabledConstraints)
		}
		return nil
	}
}

func testAccGetOrganizationResourceControlPolicy(t *testing.T, orgID string) *models.HashicorpCloudResourcemanagerOrganizationGetResourceControlPolicyResponse {
	t.Helper()

	client := acctest.HCPClients(t)
	params := organization_service.NewOrganizationServiceGetResourceControlPolicyParamsWithContext(context.Background())
	params.ID = orgID

	res, err := client.Organization.OrganizationServiceGetResourceControlPolicy(params, nil)
	if err != nil {
		t.Fatalf("get resource control policy failed: %v", err)
	}

	return res.GetPayload()
}

func testAccSetOrganizationResourceControlPolicy(t *testing.T, orgID string, constraintIDs []string) {
	t.Helper()

	client := acctest.HCPClients(t)
	current := testAccGetOrganizationResourceControlPolicy(t, orgID)

	params := organization_service.NewOrganizationServiceSetResourceControlPolicyParamsWithContext(context.Background())
	params.ID = orgID
	params.Body = &models.HashicorpCloudResourcemanagerOrganizationServiceSetResourceControlPolicyBody{
		ConstraintIds: constraintIDs,
		Etag:          current.Etag,
	}

	if _, err := client.Organization.OrganizationServiceSetResourceControlPolicy(params, nil); err != nil {
		t.Fatalf("restore resource control policy failed: %v", err)
	}
}