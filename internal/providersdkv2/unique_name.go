package providersdkv2

import (
	"math/rand"
	"strconv"
	"time"
)

// uniqueName will generate a unique name that is <= 36 characters long.
// E.g. hcp-provider-test-20060102150405-123
func uniqueName() string {
	return "hcp-provider-test-" + time.Now().Format("20060102150405") + "-" + strconv.Itoa(rand.Intn(1000))
}
