// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testclient

import (
	"net/http"

	"google.golang.org/grpc/codes"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func isAlreadyExistsError(e clients.ErrorWithCode) bool {
	switch e.Code() {
	case int(codes.AlreadyExists), http.StatusConflict:
		return true
	default:
		return false
	}
}
