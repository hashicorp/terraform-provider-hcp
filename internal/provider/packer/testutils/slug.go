package testutils

import (
	"fmt"
	"time"
)

// CreateTestSlug creates a slug for testing purposes.
// The slug is prefixed by the test name, and suffixed by the current time
// The test name will be truncated to 25 characters so that the
// time suffix
func CreateTestSlug(testName string) string {
	suffix := fmt.Sprintf("-%s", time.Now().Format("0601021504"))
	return fmt.Sprintf("%.*s%s", 36-len(suffix), testName, suffix)
}
