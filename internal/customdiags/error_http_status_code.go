package customdiags

import (
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ErrorHTTPStatusCode is an error diagnostic that stored the error code.
type ErrorHTTPStatusCode struct {
	detail         string
	summary        string
	HTTPStatusCode int
}

// Detail returns the diagnostic detail.
func (d ErrorHTTPStatusCode) Detail() string {
	return d.detail
}

// Equal returns true if the other diagnostic is equivalent.
func (d ErrorHTTPStatusCode) Equal(o diag.Diagnostic) bool {
	ed, ok := o.(ErrorHTTPStatusCode)

	if !ok {
		return false
	}

	return ed.Summary() == d.Summary() && ed.Detail() == d.Detail() && ed.HTTPStatusCode == d.HTTPStatusCode
}

// Summary returns the diagnostic summary.
func (d ErrorHTTPStatusCode) Summary() string {
	return d.summary
}

// Severity returns the diagnostic severity.
func (d ErrorHTTPStatusCode) Severity() diag.Severity {
	return diag.SeverityError
}

// NewErrorHTTPStatusCode returns a new error severity diagnostic with the given summary, detail and error code.
func NewErrorHTTPStatusCode(summary string, detail string, statusCode int) ErrorHTTPStatusCode {
	return ErrorHTTPStatusCode{
		detail:         detail,
		summary:        summary,
		HTTPStatusCode: statusCode,
	}
}

// HasConflictError checks if a diagnostic has a specific error code.
func HasConflictError(diags diag.Diagnostics) bool {
	for _, d := range diags {
		diag, ok := d.(*ErrorHTTPStatusCode)
		if !ok {
			return false
		}
		if diag.HTTPStatusCode == http.StatusConflict {
			return true
		}
	}
	return false
}
