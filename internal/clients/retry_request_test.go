// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import "testing"

func TestShouldRetryErrorCode(t *testing.T) {
	errorCodesToRetry := []int{502, 503, 504}

	shouldFail := shouldRetryErrorCode(200, errorCodesToRetry)
	if shouldFail != false {
		t.Errorf("shouldRetryErrorCode(200, []int{502, 503, 504}[:]) = %v; want false", shouldFail)
	}

	shouldSucceed := shouldRetryErrorCode(503, errorCodesToRetry)
	if shouldSucceed != true {
		t.Errorf("shouldRetryErrorCode(503, []int{502, 503, 504}[:]) = %v; want true", shouldSucceed)
	}
}
