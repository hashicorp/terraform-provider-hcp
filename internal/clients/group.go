// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
)

// Groups

// CreateGroupRetry wraps the groups client with an exponential backoff retry mechanism.
func CreateGroupRetry(client *Client, params *groups_service.GroupsServiceCreateGroupParams) (*groups_service.GroupsServiceCreateGroupOK, error) {
	res, err := client.Groups.GroupsServiceCreateGroup(params, nil)

	return res, err
}

// UpdateGroupRetry wraps the groups client with an exponential backoff retry mechanism.
func UpdateGroupRetry(client *Client, params *groups_service.GroupsServiceUpdateGroup2Params) (*groups_service.GroupsServiceUpdateGroup2OK, error) {
	res, err := client.Groups.GroupsServiceUpdateGroup2(params, nil)

	return res, err
}

// DeleteGroupRetry wraps the groups client with an exponential backoff retry mechanism.
func DeleteGroupRetry(client *Client, params *groups_service.GroupsServiceDeleteGroupParams) (*groups_service.GroupsServiceDeleteGroupOK, error) {
	res, err := client.Groups.GroupsServiceDeleteGroup(params, nil)

	return res, err
}

// Group Members

// UpdateGroupMembersRetry wraps the groups client with an exponential backoff retry mechanism.
func UpdateGroupMembersRetry(client *Client, params *groups_service.GroupsServiceUpdateGroupMembersParams) (*groups_service.GroupsServiceUpdateGroupMembersOK, error) {
	res, err := client.Groups.GroupsServiceUpdateGroupMembers(params, nil)

	return res, err
}
