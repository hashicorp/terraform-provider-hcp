package customdiags

import (
	"google.golang.org/grpc/codes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ErrorDiagnosticWithErrorCode is an error diagnostic that stored the error code.
type ErrorDiagnosticWithErrorCode struct {
	detail    string
	summary   string
	errorCode codes.Code
}

// Detail returns the diagnostic detail.
func (d ErrorDiagnosticWithErrorCode) Detail() string {
	return d.detail
}

// Equal returns true if the other diagnostic is equivalent.
func (d ErrorDiagnosticWithErrorCode) Equal(o diag.Diagnostic) bool {
	ed, ok := o.(ErrorDiagnosticWithErrorCode)

	if !ok {
		return false
	}

	return ed.Summary() == d.Summary() && ed.Detail() == d.Detail() && ed.errorCode == d.errorCode
}

// Summary returns the diagnostic summary.
func (d ErrorDiagnosticWithErrorCode) Summary() string {
	return d.summary
}

// Severity returns the diagnostic severity.
func (d ErrorDiagnosticWithErrorCode) Severity() diag.Severity {
	return diag.SeverityError
}

// NewErrorDiagnosticWithErrorCode returns a new error severity diagnostic with the given summary, detail and error code.
func NewErrorDiagnosticWithErrorCode(summary string, detail string, errorCode codes.Code) ErrorDiagnosticWithErrorCode {
	return ErrorDiagnosticWithErrorCode{
		detail:    detail,
		summary:   summary,
		errorCode: errorCode,
	}
}

// HasConflictError checks if a diagnostic has a specific error code.
func HasConflictError(diags diag.Diagnostics) bool {
	for _, d := range diags {
		diag, ok := d.(*ErrorDiagnosticWithErrorCode)
		if !ok {
			return false
		}
		if diag.errorCode == codes.Aborted {
			return true
		}
	}
	return false
}
