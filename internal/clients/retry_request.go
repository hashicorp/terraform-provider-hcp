// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
)

const (
	retryCount          = 10
	retryDelay          = 10
	counterStart        = 1
	unknownErrorMessage = "could not complete request: please ensure your HCP_API_HOST, HCP_CLIENT_ID, and HCP_CLIENT_SECRET are correct"
)

var errorCodesToRetry = [...]int{502, 503, 504}

// Helper to check what requests to retry based on the response HTTP code
func shouldRetryErrorCode(errorCode int, errorCodesToRetry []int) bool {
	for i := range errorCodesToRetry {
		if errorCodesToRetry[i] == errorCode {
			return true
		}
	}
	return false
}

// Wraps the OrganizationServiceList function in a loop that supports retrying the GET request
func RetryOrganizationServiceList(client *Client, params *organization_service.OrganizationServiceListParams) (*organization_service.OrganizationServiceListOK, error) {
	resp, err := client.Organization.OrganizationServiceList(params, nil)

	if err != nil {
		if reflect.TypeOf(err.Error()).Name() != "organization_service.OrganizationServiceListDefault" {
			return nil, errors.New(unknownErrorMessage)
		}
		counter := counterStart
		for shouldRetryErrorCode(err.(*organization_service.OrganizationServiceListDefault).Code(), errorCodesToRetry[:]) && counter < retryCount {
			resp, err = client.Organization.OrganizationServiceList(params, nil)
			if err == nil {
				break
			}
			// Avoid wasting time if we're not going to retry next loop cycle
			if (counter + 1) != retryCount {
				fmt.Printf("Error trying to get list of organizations. Retrying in %d seconds...", retryDelay*counter)
				time.Sleep(time.Duration(retryDelay*counter) * time.Second)
			}
			counter++
		}
	}
	return resp, err
}

// Wraps the ProjectServiceList function in a loop that supports retrying the GET request
func RetryProjectServiceList(client *Client, params *project_service.ProjectServiceListParams) (*project_service.ProjectServiceListOK, error) {
	resp, err := client.Project.ProjectServiceList(params, nil)

	if err != nil {
		if reflect.TypeOf(err.Error()).Name() != "project_service.ProjectServiceListDefault" {
			return nil, errors.New(unknownErrorMessage)
		}

		counter := counterStart
		for shouldRetryErrorCode(err.(*project_service.ProjectServiceListDefault).Code(), errorCodesToRetry[:]) && counter < retryCount {
			resp, err = client.Project.ProjectServiceList(params, nil)
			if err == nil {
				break
			}
			// Avoid wasting time if we're not going to retry next loop cycle
			if (counter + 1) != retryCount {
				fmt.Printf("Error trying to get list of projects. Retrying in %d seconds...", retryDelay*counter)
				time.Sleep(time.Duration(retryDelay*counter) * time.Second)
			}
			counter++
		}
	}
	return resp, err
}
