package main

import (
	"ConjunctiveSSE/Database"
	"ConjunctiveSSE/utils"
	"bytes"
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"fmt"
	"math/big"
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

func TestGenerateKeywords(t *testing.T) {
	err := Database.GenQuerydataFromDB("Crime_USENIX_REV", "id_keywords", 10000)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearch(t *testing.T) {
	//g := big.NewInt(9)
	p := big.NewInt(1)
	p = p.Lsh(p, 255)
	p = p.Sub(p, big.NewInt(19))
	//fmt.Println("p:", p, "g:", g)

	alpha := new(big.Int).Mul(new(big.Int).Mod(big.NewInt(4), p), new(big.Int).ModInverse(big.NewInt(5), p))
	alpha = alpha.Mod(alpha, p)

	xtagHead := new(big.Int).Mul(new(big.Int).Mod(big.NewInt(3), p), new(big.Int).Mod(big.NewInt(4), p))
	xtag := xtagHead.Mod(xtagHead, p)
	//xtag := new(big.Int).Exp(g, xtagHead, p)

	xtoken := new(big.Int).Mul(new(big.Int).Mod(big.NewInt(3), p), new(big.Int).Mod(big.NewInt(5), p))
	//xtokenHead = xtagHead.Mod(xtokenHead, p)
	//xtoken := new(big.Int).Exp(g, xtokenHead, p)

	fmt.Println("alpha:", alpha)
	fmt.Println("xtag:", xtag)
	fmt.Println("xtoken:", xtoken)

	xtag2 := new(big.Int).Mul(alpha, xtoken)
	xtag2 = xtag2.Mod(xtag2, p)
	fmt.Println("xtag2:", xtag2)

	fmt.Println("xtag == xtag2:", xtag.Cmp(xtag2) == 0)
}

func TestPrfp(t *testing.T) {
	p := big.NewInt(1)
	p = p.Lsh(p, 255)
	p = p.Sub(p, big.NewInt(19))

	key := []byte("key")
	message := []byte("message")
	hello := []byte("hello")
	test := []byte("test")
	g := big.NewInt(9)

	a, err := utils.PrfFp(key, message, p, g)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := utils.PrfFp(key, hello, p, g)
	c, _ := utils.PrfFp(key, test, p, g)

	xtagHead := new(big.Int).Mod(new(big.Int).Mul(b, c), p)
	xtag := new(big.Int).Exp(g, xtagHead, p)

	inverse := new(big.Int).ModInverse(a, p)
	fmt.Println("a^inverse:", new(big.Int).Mod(new(big.Int).Mul(a, inverse), p))

	xtag2Head1 := new(big.Int).Mod(new(big.Int).Mul(a, b), p)
	xtag2Head2 := new(big.Int).Mod(new(big.Int).Mul(c, inverse), p)
	xtag2Head := new(big.Int).Mod(new(big.Int).Mul(xtag2Head1, xtag2Head2), p)
	xtag2 := new(big.Int).Exp(g, xtag2Head, p)

	fmt.Println("xtagHead == xtag2Head:", xtagHead.Cmp(xtag2Head) == 0)
	fmt.Println("xtag == xtag2:", xtag.Cmp(xtag2) == 0)
}
