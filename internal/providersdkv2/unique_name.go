package providersdkv2

import (
	"math/rand"
	"strconv"
	"time"
)

// uniqueName will generate a unique name that is <= 36 characters long.
// E.g. hcp-provider-20060102150405-1234567
func uniqueName() string {
	return "hcp-provider-" + time.Now().Format("20060102150405") + "-" + strconv.Itoa(rand.Intn(10000000))
}
