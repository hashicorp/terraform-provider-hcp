package clients

import (
	"errors"
	"net/http"

	"github.com/go-openapi/runtime"
)

// IsResponseCodeNotFound takes an error returned from a client service
// request, and returns true if the response code was 404 not found
func IsResponseCodeNotFound(err error) bool {
	var apiErr *runtime.APIError
	return errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound
}
