package main

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"testing"
)

func TestEncode(t *testing.T) {
	keyword := "secret"
	message := "hello world"

	// Encode the message
	h := hmac.New(crypto.SHA256.New, []byte(keyword))
	_, err := h.Write([]byte(message))
	if err != nil {
		t.Fatal(err)
	}
	cipherText := h.Sum(nil)

	// Encode the cipherText
	encodeCipherText := base64.StdEncoding.EncodeToString(cipherText)

	// Decode the cipherText
	decodeCipherText, err := base64.StdEncoding.DecodeString(encodeCipherText)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(cipherText, decodeCipherText) {
		t.Fatal("Decode the cipherText failed")
	}
}
