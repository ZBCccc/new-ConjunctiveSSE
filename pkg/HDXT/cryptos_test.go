package HDXT

import (
	"fmt"
	"testing"
	"time"
)


func TestPrfF(t *testing.T) {
	key := []byte("1234567890123456")
	timeStart := time.Now()
	for i := range 60 {
		message := []byte(fmt.Sprintf("hello%d", i))
		result, _ := PrfF(key, message)
		_ = result
	}
	t.Logf("PrfF time: %v", time.Since(timeStart))
}
