package main

import (
	"ConjunctiveSSE/pkg/utils"
	"encoding/base64"
	"testing"
)

func TestBytes(t *testing.T) {
	// 将base64编码的字符串解码为字节数组
	h := "F3222"
	t.Log([]byte(h))
	e1 := "mvzxB5BuNNFYR22AABjyJObYpj0ZohXGkeSWYcMLZ4Q="
	d1, err := base64.StdEncoding.DecodeString(e1)
	if err != nil {
		t.Fatalf("Error decoding base64 string: %v", err)
	}
	t.Log(d1)

	e2 := "mv71DZNcNNFYR22AABjyJObYpj0ZohXGkeSWYcMLZ4Q="
	d2, err := base64.StdEncoding.DecodeString(e2)
	if err != nil {
		t.Fatalf("Error decoding base64 string: %v", err)
	}
	t.Log(d2)
	t.Log(string(utils.BytesXOR(d1, d2)))
	t.Log(1 ^ 3)
}
