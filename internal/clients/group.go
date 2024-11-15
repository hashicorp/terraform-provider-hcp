// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"github.com/cenkalti/backoff/v4"

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
		return err
	}

	getCode := func(err error) (int, error) {
		serviceErr, ok := err.(*groups_service.GroupsServiceCreateGroupDefault)
		if !ok {
			return 0, err
		}
		return serviceErr.Code(), nil
	}

	err := backoff.Retry(newBackoffOp(op, getCode), newBackoff())

	return res, err
}

// UpdateGroupRetry wraps the groups client with an exponential backoff retry mechanism.
func UpdateGroupRetry(client *Client, params *groups_service.GroupsServiceUpdateGroup2Params) (*groups_service.GroupsServiceUpdateGroup2OK, error) {
	var res *groups_service.GroupsServiceUpdateGroup2OK
	op := func() error {
		var err error
		res, err = client.Groups.GroupsServiceUpdateGroup2(params, nil)
		return err
	}

	getCode := func(err error) (int, error) {
		serviceErr, ok := err.(*groups_service.GroupsServiceUpdateGroup2Default)
		if !ok {
			return 0, err
		}
		return serviceErr.Code(), nil
	}

	err := backoff.Retry(newBackoffOp(op, getCode), newBackoff())

	return res, err
}

// DeleteGroupRetry wraps the groups client with an exponential backoff retry mechanism.
func DeleteGroupRetry(client *Client, params *groups_service.GroupsServiceDeleteGroupParams) (*groups_service.GroupsServiceDeleteGroupOK, error) {
	var res *groups_service.GroupsServiceDeleteGroupOK
	op := func() error {
		var err error
		res, err = client.Groups.GroupsServiceDeleteGroup(params, nil)
		return err
	}

	getCode := func(err error) (int, error) {
		serviceErr, ok := err.(*groups_service.GroupsServiceDeleteGroupDefault)
		if !ok {
			return 0, err
		}
		return serviceErr.Code(), nil
	}

	err := backoff.Retry(newBackoffOp(op, getCode), newBackoff())

	return res, err
}

// Group Members

// UpdateGroupMembersRetry wraps the groups client with an exponential backoff retry mechanism.
func UpdateGroupMembersRetry(client *Client, params *groups_service.GroupsServiceUpdateGroupMembersParams) (*groups_service.GroupsServiceUpdateGroupMembersOK, error) {
	var res *groups_service.GroupsServiceUpdateGroupMembersOK
	op := func() error {
		var err error
		res, err = client.Groups.GroupsServiceUpdateGroupMembers(params, nil)
		return err
	}

	getCode := func(err error) (int, error) {
		serviceErr, ok := err.(*groups_service.GroupsServiceUpdateGroupMembersDefault)
		if !ok {
			return 0, err
		}
		return serviceErr.Code(), nil
	}

	err := backoff.Retry(newBackoffOp(op, getCode), newBackoff())

	return res, err
}

// newBackoff creates a new exponential backoff with default values.
func newBackoff() backoff.BackOff {
	// Create a new exponential backoff with explicit default values.
	return backoff.NewExponentialBackOff(
		backoff.WithInitialInterval(backoff.DefaultInitialInterval),
		backoff.WithRandomizationFactor(backoff.DefaultRandomizationFactor),
		backoff.WithMultiplier(backoff.DefaultMultiplier),
		backoff.WithMaxInterval(backoff.DefaultMaxInterval),
		backoff.WithMaxElapsedTime(backoff.DefaultMaxElapsedTime),
	)
}

func newBackoffOp(op func() error, getCode func(error) (int, error)) func() error {
	return func() error {
		err := op()

		if err == nil {
			return nil
		}

		code, codeErr := getCode(err)
		if codeErr != nil {
			return backoff.Permanent(codeErr)
		}

		if !shouldRetryErrorCode(code, groupErrorCodesToRetry[:]) {
			return backoff.Permanent(err)
		}

		return err
	}
}
