package packerv2

import (
	"fmt"

	packerservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2023-01-01/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

type Version = packermodels.HashicorpCloudPacker20230101Version
type GetVersionParams = packerservice.PackerServiceGetVersionParams

func GetVersionByFingerprint(client *clients.Client, location location.BucketLocation, fingerprint string) (*Version, error) {
	params := packerservice.NewPackerServiceGetVersionParams()
	params.SetLocationOrganizationID(location.GetOrganizationID())
	params.SetLocationProjectID(location.GetProjectID())
	params.SetBucketName(location.GetBucketName())
	params.SetFingerprint(fingerprint)

	resp, err := getVersionRaw(client, params)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload().Version, nil
}

func GetVersionByFingerprintDiags(client *clients.Client, location location.BucketLocation, fingerprint string) (*Version, diag.Diagnostics) {
	var diags diag.Diagnostics
	version, err := GetVersionByFingerprint(client, location, fingerprint)

	if err != nil {
		diags.AddError(
			"failed to get Version by Fingerprint, received an error from the HCP Packer API",
			err.Error(),
		)
	} else if version == nil || version.Fingerprint == "" {
		diags.AddError(
			"failed to get Version by Fingerprint, received an invalid Version response from the HCP Packer API",
			"this is an internal error, please report this issue to the provider developers",
		)
	}
	if diags.HasError() {
		return nil, diags
	}

	return version, diags
}

func GetVersionByChannelName(client *clients.Client, location location.BucketLocation, channelName string) (*Version, error) {
	channel, err := GetChannelByName(client, location, channelName)
	if err != nil {
		return nil, err
	}

	if channel == nil || channel.Name == "" {
		return nil, fmt.Errorf(
			"received an invalid channel response." +
				"This is an internal error. Report this issue to the provider developers",
		)
	}

	return channel.Version, nil
}

func GetVersionByChannelNameDiags(client *clients.Client, location location.BucketLocation, channelName string) (*Version, diag.Diagnostics) {
	var diags diag.Diagnostics
	version, err := GetVersionByChannelName(client, location, channelName)

	if err != nil {
		diags.AddError(
			"failed to get Version by Channel Name, received an error from the HCP Packer API",
			err.Error(),
		)
		return nil, diags
	}

	if version == nil {
		diags.AddError(
			"provided Channel does not have an assigned Version",
			"the Channel was found, but did not have an assigned Version. ",
		)
		return nil, diags
	}

	return version, nil
}

func getVersionRaw(client *clients.Client, params *GetVersionParams) (*packerservice.PackerServiceGetVersionOK, error) {
	resp, err := client.PackerV2.PackerServiceGetVersion(params, nil)
	if err != nil {
		return nil, formatGRPCError[*packerservice.PackerServiceGetVersionDefault](err)
	}

	return resp, nil
}
