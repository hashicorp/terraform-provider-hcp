// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var validLogLevels = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

// logger implements the logging interface required by our openapi clients
type logger struct{}

// Printf prints an info log
func (l logger) Printf(format string, args ...interface{}) {
	log.Printf("[INFO] %s", fmt.Sprintf(format, args...))
}

// Debugf prints a debug log
func (l logger) Debugf(format string, args ...interface{}) {
	log.Printf("[DEBUG] %s", fmt.Sprintf(format, args...))
}

// ShouldLog determines if TF_LOG is set to a valid level.
func ShouldLog() bool {
	logLevel := strings.ToUpper(os.Getenv("TF_LOG"))
	if logLevel == "" {
		return false
	}

	for _, l := range validLogLevels {
		if l == logLevel {
			return true
		}
	}

	return false
}
