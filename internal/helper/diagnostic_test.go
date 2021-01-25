package helper

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/require"
)

func Test_toError(t *testing.T) {
	tcs := []struct {
		name string
		diag diag.Diagnostics
		err  *multierror.Error
	}{
		{
			name: "nil diagnostics",
			diag: nil,
			err:  nil,
		},
		{
			name: "diagnostics with no error",
			diag: []diag.Diagnostic{
				{
					Severity: diag.Warning,
				},
			},
			err: nil,
		},
		{
			name: "diagnostics with single error",
			diag: []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "there was an error",
				},
			},
			err: multierror.Append(nil, errors.New("there was an error")),
		},
		{
			name: "diagnostics with two errors",
			diag: []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "there was an error",
				},
				{
					Severity: diag.Error,
					Summary:  "there was a second error",
				},
			},
			err: multierror.Append(errors.New("there was an error"), errors.New("there was a second error")),
		},
		{
			name: "diagnostics with detailed error",
			diag: []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "there was an error",
					Detail:   "with detail",
				},
			},
			err: multierror.Append(nil, errors.New("there was an error; with detail")),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			err := ToError(tc.diag)
			r.Equal(tc.err, err)
		})
	}
}
