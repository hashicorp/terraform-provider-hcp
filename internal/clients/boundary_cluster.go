// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-boundary-service/stable/2021-12-21/client/boundary_service"
	boundarymodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-boundary-service/stable/2021-12-21/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// GetBoundaryClusterByID gets a Boundary cluster by its ID.
func GetBoundaryClusterByID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	boundaryClusterID string) (*boundarymodels.HashicorpCloudBoundary20211221Cluster, error) {

	getParams := boundary_service.NewBoundaryServiceGetParams()
	getParams.Context = ctx
	getParams.ClusterID = boundaryClusterID
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResp, err := client.Boundary.BoundaryServiceGet(getParams, nil)
	if err != nil {
		return nil, err
	}

	return getResp.Payload.Cluster, nil
}

// CreateBoundaryCluster will make a call to the Boundary service to initiate the create Boundary
// cluster workflow.
func CreateBoundaryCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	boundaryCreateRequest *boundarymodels.HashicorpCloudBoundary20211221CreateRequest) (*boundarymodels.HashicorpCloudBoundary20211221CreateResponse, error) {

	p := boundary_service.NewBoundaryServiceCreateParams()
	p.Context = ctx
	p.Body = boundaryCreateRequest

	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	resp, err := client.Boundary.BoundaryServiceCreate(p, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// DeleteBoundaryCluster will make a call to the Boundary service to initiate the delete Boundary
// cluster workflow.
func DeleteBoundaryCluster(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	boundaryClusterID string) (*boundarymodels.HashicorpCloudBoundary20211221DeleteResponse, error) {

	p := boundary_service.NewBoundaryServiceDeleteParams()
	p.Context = ctx
	p.ClusterID = boundaryClusterID
	p.LocationOrganizationID = loc.OrganizationID
	p.LocationProjectID = loc.ProjectID

	deleteResp, err := client.Boundary.BoundaryServiceDelete(p, nil)
	if err != nil {
		return nil, err
	}

	return deleteResp.Payload, nil
}

// SetBoundaryClusterMaintenanceWindow updates the maintenance window configuration for a Boundary cluster.
func SetBoundaryClusterMaintenanceWindow(
	ctx context.Context,
	client *Client,
	loc *sharedmodels.HashicorpCloudLocationLocation,
	boundaryClusterID string,
	mwUpdateRequest *boundarymodels.HashicorpCloudBoundary20211221MaintenanceWindowUpdateRequest,
) (*boundarymodels.HashicorpCloudBoundary20211221MaintenanceWindowUpdateResponse, error) {

	params := boundary_service.NewBoundaryServiceMaintenanceWindowUpdateParams()
	params.Context = ctx
	params.Body = mwUpdateRequest

	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.ClusterID = boundaryClusterID

	resp, err := client.Boundary.BoundaryServiceMaintenanceWindowUpdate(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// GetBoundaryClusterMaintenanceWindow gets the maintenance window configuration for a Boundary cluster.
func GetBoundaryClusterMaintenanceWindow(
	ctx context.Context,
	client *Client,
	loc *sharedmodels.HashicorpCloudLocationLocation,
	boundaryClusterID string,
) (*boundarymodels.HashicorpCloudBoundary20211221UpgradeType, *boundarymodels.HashicorpCloudBoundary20211221MaintenanceWindow, error) {

	params := boundary_service.NewBoundaryServiceMaintenanceWindowGetParams()
	params.Context = ctx
	params.ClusterID = boundaryClusterID
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID

	resp, err := client.Boundary.BoundaryServiceMaintenanceWindowGet(params, nil)
	if err != nil {
		return nil, nil, err
	}

	return resp.Payload.UpgradeType, resp.Payload.MaintenanceWindow, nil
}
