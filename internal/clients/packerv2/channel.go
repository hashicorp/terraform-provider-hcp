package packerv2

import (
	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/models"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

type Channel = packermodels.HashicorpCloudPacker20230101Channel
type GetChannelParams = packerservice.PackerServiceGetChannelParams

func GetChannelByName(client *clients.Client, location location.BucketLocation, name string) (*Channel, error) {
	params := packerservice.NewPackerServiceGetChannelParams()
	params.SetLocationOrganizationID(location.GetOrganizationID())
	params.SetLocationProjectID(location.GetProjectID())
	params.SetBucketName(location.GetBucketName())
	params.SetChannelName(name)

	resp, err := getChannelRaw(client, params)
	if err != nil {
		return nil, err
	}

	return resp.GetPayload().Channel, nil
}

func getChannelRaw(client *clients.Client, params *GetChannelParams) (*packerservice.PackerServiceGetChannelOK, error) {
	resp, err := client.PackerV2.PackerServiceGetChannel(params, nil)
	if err != nil {
		return nil, formatGRPCError[*packerservice.PackerServiceGetChannelDefault](err)
	}

	return resp, nil
}
