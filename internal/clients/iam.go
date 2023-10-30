package clients

import (
	"context"
	"fmt"

	iam "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	rmModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
)

const (
	// maxBatchGetPrincipalsSize is the maximum number of principals that
	// can be retrieved in a given batch.
	maxBatchGetPrincipalsSize = 1000
)

// BatchGetPrincipals retrieves the requested principals in a batch. If the
// number of principals exceeds the batch limit, multiple requests will be made.
func BatchGetPrincipals(ctx context.Context, client *Client, principals []string, view *models.HashicorpCloudIamPrincipalView) ([]*models.HashicorpCloudIamPrincipal, error) {
	var allPrincipals []*models.HashicorpCloudIamPrincipal

	n := len(principals)
	for i := 0; i < n; i += maxBatchGetPrincipalsSize {
		params := iam.NewIamServiceBatchGetPrincipalsParams()
		params.OrganizationID = client.Config.OrganizationID
		params.View = (*string)(models.HashicorpCloudIamPrincipalViewPRINCIPALVIEWBASIC.Pointer())
		params.PrincipalIds = principals[i:min(i+maxBatchGetPrincipalsSize, n)]

		resp, err := client.IAM.IamServiceBatchGetPrincipals(params, nil)
		if err != nil {
			return nil, err
		}

		allPrincipals = append(allPrincipals, resp.Payload.Principals...)
	}

	return allPrincipals, nil
}

// IamPrincipalTypeToBindingType converts an IAM principal type to a resource
// manager binding member type.
func IamPrincipalTypeToBindingType(p *models.HashicorpCloudIamPrincipal) (*rmModels.HashicorpCloudResourcemanagerPolicyBindingMemberType, error) {
	if p == nil || p.Type == nil {
		return nil, fmt.Errorf("nil principal type")
	}

	switch *p.Type {
	case models.HashicorpCloudIamPrincipalTypePRINCIPALTYPEUSER:
		return rmModels.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(), nil
	case models.HashicorpCloudIamPrincipalTypePRINCIPALTYPEGROUP:
		return rmModels.HashicorpCloudResourcemanagerPolicyBindingMemberTypeGROUP.Pointer(), nil
	case models.HashicorpCloudIamPrincipalTypePRINCIPALTYPESERVICE:
		return rmModels.HashicorpCloudResourcemanagerPolicyBindingMemberTypeSERVICEPRINCIPAL.Pointer(), nil
	default:
		return nil, fmt.Errorf("Unsupported principal type (%s) for IAM Policy", *p.Type)
	}
}
