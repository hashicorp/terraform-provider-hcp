package acctest

import (
	"math/rand"
	"time"
)

// RandString generates a random string with the given length.
func RandString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}
