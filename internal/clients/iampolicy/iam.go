package iampolicy

import (
	"context"
	"slices"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"golang.org/x/exp/maps"
)

// The ResourceIamUpdater interface is implemented for each HCP resource
// supporting IAM policy (Organization/Project/Resource).
//
// Implementations should be created per resource and should keep track of the
// resource identifier.
type ResourceIamUpdater interface {
	// Fetch the existing IAM policy attached to a resource.
	GetResourceIamPolicy(context.Context) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics)

	// Replaces the existing IAM Policy attached to a resource.
	SetResourceIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, diag.Diagnostics)

	// A mutex guards against concurrent to call to the SetResourceIamPolicy method.
	// The mutex key should be globally unique.
	GetMutexKey() string
}

type TerraformResourceData interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
	SetAttribute(ctx context.Context, path path.Path, val interface{}) diag.Diagnostics
}

// Factory for generating ResourceIamUpdater for given ResourceData resource
type NewResourceIamUpdaterFunc func(ctx context.Context, d TerraformResourceData, clients *clients.Client) (ResourceIamUpdater, diag.Diagnostics)

// Equal returns if the passed Policies are equal.
func Equal(p1, p2 *models.HashicorpCloudResourcemanagerPolicy) bool {
	if p1 == nil && p2 == nil {
		return true
	} else if p1 == nil || p2 == nil {
		return false
	}

	// Check if the Etags are equal
	if p1.Etag != p2.Etag {
		return false
	}

	// Convert the policies to maps so they are easier to compare
	m1, m2 := ToMap(p1), ToMap(p2)

	// Check if they have the same number of roles
	roles1, roles2 := maps.Keys(m1), maps.Keys(m2)
	if len(roles1) != len(roles2) {
		return false
	}

	// Sort and compare the roles
	slices.Sort(roles1)
	slices.Sort(roles2)
	if !slices.Equal(roles1, roles2) {
		return false
	}

	for r1, members1 := range m1 {
		// Ensure both policies contain the given role
		members2, ok := m2[r1]
		if !ok {
			return false
		}

		for m1, mtype1 := range members1 {
			// Ensure both roles have the same member ID
			mtype2, ok := members2[m1]
			if !ok {
				return false
			}

			// Ensure both consider the member as having the same type
			if *mtype1 != *mtype2 {
				return false
			}
		}

	}

	return true
}

// ToMap to map converts an IAM policy to a set of maps. The first map is keyed
// by Role ID, and the second map is keyed by PrincipalID.
func ToMap(p *models.HashicorpCloudResourcemanagerPolicy) map[string]map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType {
	bindings := make(map[string]map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType, len(p.Bindings))
	for _, b := range p.Bindings {
		bindings[b.RoleID] = make(map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType, len(b.Members))
		for _, m := range b.Members {
			bindings[b.RoleID][m.MemberID] = m.MemberType
		}
	}

	return bindings
}

// FromMap converts the map generated by ToMap to an IAM Policy object.
func FromMap(etag string, bindings map[string]map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType) *models.HashicorpCloudResourcemanagerPolicy {
	up := &models.HashicorpCloudResourcemanagerPolicy{
		Bindings: []*models.HashicorpCloudResourcemanagerPolicyBinding{},
		Etag:     etag,
	}

	for role, members := range bindings {
		b := &models.HashicorpCloudResourcemanagerPolicyBinding{
			Members: []*models.HashicorpCloudResourcemanagerPolicyBindingMember{},
			RoleID:  role,
		}

		for id, mtype := range members {
			m := &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   id,
				MemberType: mtype,
			}

			b.Members = append(b.Members, m)
		}

		up.Bindings = append(up.Bindings, b)
	}

	return up
}
