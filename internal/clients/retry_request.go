// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"fmt"
	"time"

	billing "github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/client/billing_account_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
)

const (
	retryCount   = 10
	retryDelay   = 10
	counterStart = 1
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
		serviceErr, ok := err.(*organization_service.OrganizationServiceListDefault)
		if !ok {
			return nil, err
		}
		counter := counterStart
		for shouldRetryErrorCode(serviceErr.Code(), errorCodesToRetry[:]) && counter < retryCount {
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
		serviceErr, ok := err.(*project_service.ProjectServiceListDefault)
		if !ok {
			return nil, err
		}

		counter := counterStart
		for shouldRetryErrorCode(serviceErr.Code(), errorCodesToRetry[:]) && counter < retryCount {
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

// Wraps the ProjectServiceGet function in a loop that supports retrying the GET request
func RetryProjectServiceGet(client *Client, params *project_service.ProjectServiceGetParams) (*project_service.ProjectServiceGetOK, error) {
	resp, err := client.Project.ProjectServiceGet(params, nil)

	if err != nil {
		serviceErr, ok := err.(*project_service.ProjectServiceGetDefault)
		if !ok {
			return nil, err
		}

		counter := counterStart
		for shouldRetryErrorCode(serviceErr.Code(), errorCodesToRetry[:]) && counter < retryCount {
			resp, err = client.Project.ProjectServiceGet(params, nil)
			if err == nil {
				break
			}
			// Avoid wasting time if we're not going to retry next loop cycle
			if (counter + 1) != retryCount {
				fmt.Printf("Error trying to get configured project. Retrying in %d seconds...", retryDelay*counter)
				time.Sleep(time.Duration(retryDelay*counter) * time.Second)
			}
			counter++
		}
	}
	return resp, err
}

// Wraps the BillingServiceUpdate function in a loop that supports retrying the PUT request
func RetryBillingServiceUpdate(client *Client, params *billing.BillingAccountServiceUpdateParams) (*billing.BillingAccountServiceUpdateOK, error) {
	resp, err := client.Billing.BillingAccountServiceUpdate(params, nil)

	if err != nil {
		serviceErr, ok := err.(*project_service.ProjectServiceGetDefault)
		if !ok {
			return nil, err
		}

		counter := counterStart
		for shouldRetryErrorCode(serviceErr.Code(), []int{503}) && counter < retryCount {
			resp, err = client.Billing.BillingAccountServiceUpdate(params, nil)
			if err == nil {
				break
			}
			// Avoid wasting time if we're not going to retry next loop cycle
			if (counter + 1) != retryCount {
				fmt.Printf("Error trying to update billing account. Retrying in %d seconds...", retryDelay*counter)
				time.Sleep(time.Duration(retryDelay*counter) * time.Second)
			}
			counter++
		}
	}
	return resp, err
}
