// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packerv2

import (
	"errors"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

type GRPCError interface {
	error
	GetPayload() *sharedmodels.GoogleRPCStatus
}

func formatGRPCError[E GRPCError](err error) error {
	if err == nil {
		return nil
	}
	grpcErr, ok := err.(E)
	if !ok {
		return fmt.Errorf("unexpected error format. Got: %v", err)
	}

	return errors.New(grpcErr.GetPayload().Message)
}
