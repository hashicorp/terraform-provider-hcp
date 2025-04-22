// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"github.com/cenkalti/backoff/v4"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
)

var groupErrorCodesToRetry = [...]int{502, 503, 504}

// Groups

// CreateGroupRetry wraps the groups client with an exponential backoff retry mechanism.
func CreateGroupRetry(client *Client, params *groups_service.GroupsServiceCreateGroupParams) (*groups_service.GroupsServiceCreateGroupOK, error) {
	var res *groups_service.GroupsServiceCreateGroupOK
	op := func() error {
		var err error
		res, err = client.Groups.GroupsServiceCreateGroup(params, nil)
		if err != nil {
			return err
		}
		if res.Payload.OperationID != "" {
			loc := &sharedmodels.HashicorpCloudLocationLocation{OrganizationID: client.Config.OrganizationID}
			return WaitForOperation(params.Context, client, "create group", loc, res.Payload.OperationID)
		}
		return nil
	}

	serviceErr := &groups_service.GroupsServiceCreateGroupDefault{}
	err := backoff.Retry(newBackoffOp(op, serviceErr), newBackoff())

	return res, err
}

// UpdateGroupRetry wraps the groups client with an exponential backoff retry mechanism.
func UpdateGroupRetry(client *Client, params *groups_service.GroupsServiceUpdateGroup2Params) (*groups_service.GroupsServiceUpdateGroup2OK, error) {
	var res *groups_service.GroupsServiceUpdateGroup2OK
	op := func() error {
		var err error
		res, err = client.Groups.GroupsServiceUpdateGroup2(params, nil)
		if err != nil {
			return err
		}
		if res.Payload.OperationID != "" {
			loc := &sharedmodels.HashicorpCloudLocationLocation{OrganizationID: client.Config.OrganizationID}
			return WaitForOperation(params.Context, client, "update group", loc, res.Payload.OperationID)
		}
		return nil
	}

	serviceErr := &groups_service.GroupsServiceUpdateGroup2Default{}
	err := backoff.Retry(newBackoffOp(op, serviceErr), newBackoff())

	return res, err
}

// DeleteGroupRetry wraps the groups client with an exponential backoff retry mechanism.
func DeleteGroupRetry(client *Client, params *groups_service.GroupsServiceDeleteGroupParams) (*groups_service.GroupsServiceDeleteGroupOK, error) {
	var res *groups_service.GroupsServiceDeleteGroupOK
	op := func() error {
		var err error
		res, err = client.Groups.GroupsServiceDeleteGroup(params, nil)
		if err != nil {
			return err
		}
		if res.Payload.OperationID != "" {
			loc := &sharedmodels.HashicorpCloudLocationLocation{OrganizationID: client.Config.OrganizationID}
			return WaitForOperation(params.Context, client, "delete group", loc, res.Payload.OperationID)
		}
		return nil
	}

	serviceErr := &groups_service.GroupsServiceDeleteGroupDefault{}
	err := backoff.Retry(newBackoffOp(op, serviceErr), newBackoff())

	return res, err
}

// Group Members

// UpdateGroupMembersRetry wraps the groups client with an exponential backoff retry mechanism.
func UpdateGroupMembersRetry(client *Client, params *groups_service.GroupsServiceUpdateGroupMembersParams) (*groups_service.GroupsServiceUpdateGroupMembersOK, error) {
	var res *groups_service.GroupsServiceUpdateGroupMembersOK
	op := func() error {
		var err error
		res, err = client.Groups.GroupsServiceUpdateGroupMembers(params, nil)
		if err != nil {
			return err
		}
		if res.Payload.OperationID != "" {
			loc := &sharedmodels.HashicorpCloudLocationLocation{OrganizationID: client.Config.OrganizationID}
			return WaitForOperation(params.Context, client, "update group members", loc, res.Payload.OperationID)
		}
		return nil
	}

	serviceErr := &groups_service.GroupsServiceUpdateGroupMembersDefault{}
	err := backoff.Retry(newBackoffOp(op, serviceErr), newBackoff())

	return res, err
}
