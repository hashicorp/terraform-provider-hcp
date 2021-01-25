package helper

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// ToError will convert the passed in diag.Diagnostics d
// to an error if it is non-nil and contains an error.
func ToError(d diag.Diagnostics) *multierror.Error {
	if d == nil {
		return nil
	}

	if !d.HasError() {
		return nil
	}

	var err *multierror.Error
	for _, dia := range d {
		summary := dia.Summary
		if dia.Detail != "" {
			summary = fmt.Sprintf("%s; %s", summary, dia.Detail)
		}

		err = multierror.Append(err, errors.New(summary))
	}

	return err
}
