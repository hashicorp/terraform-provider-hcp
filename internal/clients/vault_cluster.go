// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/client/vault_service"
	vaultmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/models"
)

// GetVaultClusterByID gets an Vault cluster by its ID.
func GetVaultClusterByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	vaultClusterID string) (*vaultmodels.HashicorpCloudVault20201125Cluster, error) {

	getParams := vault_service.NewGetParams()
	getParams.Context = ctx
	getParams.ClusterID = vaultClusterID
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResp, err := client.Vault.Get(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp.Payload.Cluster, nil
}

// CreateVaultCluster will make a call to the Vault service to initiate the create Vault
// cluster workflow.
func CreateVaultCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	vaultCluster *vaultmodels.HashicorpCloudVault20201125InputCluster) (*vaultmodels.HashicorpCloudVault20201125CreateResponse, error) {

	p := vault_service.NewCreateParams()
	p.Context = ctx
	p.Body = &vaultmodels.HashicorpCloudVault20201125CreateRequest{Cluster: vaultCluster}

	p.ClusterLocationOrganizationID = loc.OrganizationID
	p.ClusterLocationProjectID = loc.ProjectID

	resp, err := client.Vault.Create(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// DeleteVaultCluster will make a call to the Vault service to initiate the delete Vault
// cluster workflow.
func DeleteVaultCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string) (*vaultmodels.HashicorpCloudVault20201125DeleteResponse, error) {

	p := vault_service.NewDeleteParams()
	p.Context = ctx
	p.ClusterID = clusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	deleteResp, err := client.Vault.Delete(p, nil)
	if err != nil {
		return nil, err
	}

	return deleteResp.Payload, nil
}

// CreateVaultClusterAdminToken will make a call to the Vault service to generate an admin token for the Vault cluster
// that expires after 6 hours.
func CreateVaultClusterAdminToken(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	vaultClusterID string) (*vaultmodels.HashicorpCloudVault20201125GetAdminTokenResponse, error) {

	p := vault_service.NewGetAdminTokenParams()
	p.Context = ctx
	p.ClusterID = vaultClusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID
	p.LocationRegionProvider = &loc.Region.Provider
	p.LocationRegionRegion = &loc.Region.Region

	resp, err := client.Vault.GetAdminToken(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// UpdateVaultClusterPublicIps will make a call to the Vault service to enable or disable public IPs for the Vault cluster.
func UpdateVaultClusterPublicIps(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string, enablePublicIps bool) (*vaultmodels.HashicorpCloudVault20201125UpdatePublicIpsResponse, error) {

	updateParams := vault_service.NewUpdatePublicIpsParams()
	updateParams.Context = ctx
	updateParams.ClusterID = clusterID
	updateParams.LocationProjectID = loc.ProjectID
	updateParams.LocationOrganizationID = loc.OrganizationID
	updateParams.Body = &vaultmodels.HashicorpCloudVault20201125UpdatePublicIpsRequest{
		// ClusterID and Location are repeated because the values above are required to populate the URL,
		// and the values below are required in the API request body
		ClusterID:       clusterID,
		Location:        loc,
		EnablePublicIps: enablePublicIps,
	}

	updateResp, err := client.Vault.UpdatePublicIps(updateParams, nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}

// UpdateVaultMajorVersionUpgradeConfig will make a call to the Vault service to update the major version upgrade config for the Vault cluster.
func UpdateVaultMajorVersionUpgradeConfig(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, clusterID string,
	config *vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfig) (vaultmodels.HashicorpCloudVault20201125UpdateMajorVersionUpgradeConfigResponse, error) {

	request := &vaultmodels.HashicorpCloudVault20201125UpdateMajorVersionUpgradeConfigRequest{
		// ClusterID and Location are repeated because the values above are required to populate the URL,
		// and the values below are required in the API request body
		ClusterID:   clusterID,
		Location:    loc,
		UpgradeType: config.UpgradeType,
	}
	if config.MaintenanceWindow != nil {
		request.MaintenanceWindow = &vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfigMaintenanceWindow{
			DayOfWeek:     config.MaintenanceWindow.DayOfWeek,
			TimeWindowUtc: config.MaintenanceWindow.TimeWindowUtc,
		}
	}
	updateParams := vault_service.NewUpdateMajorVersionUpgradeConfigParams()
	updateParams.Context = ctx
	updateParams.ClusterID = clusterID
	updateParams.LocationProjectID = loc.ProjectID
	updateParams.LocationOrganizationID = loc.OrganizationID
	updateParams.Body = request

	updateResp, err := client.Vault.UpdateMajorVersionUpgradeConfig(updateParams, nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}

// UpdateVaultCluster will make a call to the Vault service to update the Vault cluster configuration.
func UpdateVaultClusterConfig(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string, tier *string, metrics *vaultmodels.HashicorpCloudVault20201125ObservabilityConfig,
	auditLog *vaultmodels.HashicorpCloudVault20201125ObservabilityConfig) (*vaultmodels.HashicorpCloudVault20201125UpdateResponse, error) {

	config := &vaultmodels.HashicorpCloudVault20201125InputClusterConfig{}
	updateMaskPaths := []string{}

	if tier != nil {
		tier := vaultmodels.HashicorpCloudVault20201125Tier(*tier)
		config.Tier = &tier
		updateMaskPaths = append(updateMaskPaths, "config.tier")
	}
	if metrics != nil {
		config.MetricsConfig = metrics
		updateMaskPaths = append(updateMaskPaths, "config.metrics_config")
	}
	if auditLog != nil {
		config.AuditLogExportConfig = auditLog
		updateMaskPaths = append(updateMaskPaths, "config.audit_log_export_config")
	}
	updateParams := vault_service.NewUpdateParams()
	updateParams.Context = ctx
	updateParams.ClusterID = clusterID
	updateParams.ClusterLocationProjectID = loc.ProjectID
	updateParams.ClusterLocationOrganizationID = loc.OrganizationID
	updateParams.UpdateMaskPaths = updateMaskPaths
	updateParams.Body = &vaultmodels.HashicorpCloudVault20201125InputCluster{
		// ClusterID and Location are repeated because the values above are required to populate the URL,
		// and the values below are required in the API request body
		ID:       clusterID,
		Location: loc,
		// NOTE: if this function is ever modified to update more than just the tier,
		// the tier must ALWAYS be specified, since the 0-value is valid and will not
		// be ignored.
		Config: config,
	}

	updateResp, err := client.Vault.Update(updateParams, nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}

// UpdateVaultPathsFilter will make a call to the Vault service to update the paths filter for a secondary cluster
func UpdateVaultPathsFilter(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string, params vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationPathsFilter) (*vaultmodels.HashicorpCloudVault20201125UpdatePathsFilterResponse, error) {

	updateParams := vault_service.NewUpdatePathsFilterParams()
	updateParams.Context = ctx
	updateParams.ClusterID = clusterID
	updateParams.LocationProjectID = loc.ProjectID
	updateParams.LocationOrganizationID = loc.OrganizationID
	updateParams.Body = &vaultmodels.HashicorpCloudVault20201125UpdatePathsFilterRequest{
		// ClusterID and Location are repeated because the values above are required to populate the URL,
		// and the values below are required in the API request body
		ClusterID: clusterID,
		Location:  loc,
		Mode:      params.Mode,
		Paths:     params.Paths,
	}

	updateResp, err := client.Vault.UpdatePathsFilter(updateParams, nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}

// DeleteVaultPathsFilter will make a call to the Vault service to delete the paths filter for a secondary cluster
func DeleteVaultPathsFilter(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	clusterID string) (*vaultmodels.HashicorpCloudVault20201125DeletePathsFilterResponse, error) {

	deleteParams := vault_service.NewDeletePathsFilterParams()
	deleteParams.Context = ctx
	deleteParams.ClusterID = clusterID
	deleteParams.LocationProjectID = loc.ProjectID
	deleteParams.LocationOrganizationID = loc.OrganizationID
	deleteParams.LocationRegionProvider = &loc.Region.Provider
	deleteParams.LocationRegionRegion = &loc.Region.Region

	deleteResp, err := client.Vault.DeletePathsFilter(deleteParams, nil)
	if err != nil {
		return nil, err
	}

	return deleteResp.Payload, nil
}
