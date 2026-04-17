// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

type ServiceError interface {
	IsSuccess() bool
	IsRedirect() bool
	IsClientError() bool
	IsServerError() bool
	IsCode(code int) bool
	Code() int
	Error() string
	String() string
	GetPayload() *models.GoogleRPCStatus
}
