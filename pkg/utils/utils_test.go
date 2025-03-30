package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestHmacDigest(t *testing.T) {
	plaintext := []byte("F1234")
	key := []byte("0123456789123456")
	token := HmacDigest(plaintext, key)
	start := time.Now()
	for counter := range 1000 {
		_ = HmacDigest([]byte(fmt.Sprintf("%d", counter)), token)
	}
	elapsed := time.Since(start)
	t.Log(elapsed)
}
