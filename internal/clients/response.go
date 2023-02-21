// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/runtime"
)

// IsResponseCodeNotFound takes an error returned from a client service
// request, and returns true if the response code was 404 not found
func IsResponseCodeNotFound(err error) bool {
	var apiErr *runtime.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == http.StatusNotFound
	} else {
		return strings.Contains(err.Error(), fmt.Sprintf("[%d]", http.StatusNotFound))
	}
}
