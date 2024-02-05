// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testclient

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

type ErrorWithCode interface {
	error
	Code() int
}

func isAlreadyExistsError(e ErrorWithCode) bool {
	switch e.Code() {
	case int(codes.AlreadyExists), http.StatusConflict:
		return true
	default:
		return false
	}
}
