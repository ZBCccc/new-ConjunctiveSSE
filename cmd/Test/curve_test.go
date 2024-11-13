package main

import (
	"ConjunctiveSSE/pkg/utils"
	"golang.org/x/crypto/curve25519"
	"math/big"
	"testing"
)

func BenchmarkCurve(b *testing.B) {
	// 定义 a 和 p
	kx := "kx"
	kz := "kz"
	w1 := "F0"
	j := 1
	p := big.NewInt(1)
	p = p.Lsh(p, 255)            // 2^255
	p = p.Sub(p, big.NewInt(19)) // 2^255 - 19
	g := big.NewInt(9)
	xtoken1, _ := utils.PrfFp([]byte(kx), []byte(w1), p, g)
	xtoken2, _ := utils.PrfFp([]byte(kz), append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), p, g)
	a := new(big.Int).Mul(xtoken1, xtoken2) // 这里假设 a 是已知的

	// 计算 a * p
	result := new([32]byte)
	a = a.Mod(a, p)
	var aBytes [32]byte
	copy(aBytes[:], a.Bytes())

	for i := 0; i < b.N; i++ {
		curve25519.ScalarBaseMult(result, &aBytes)
	}
}

func BenchmarkHMACSHA256(b *testing.B) {
	keyword := "kt"
	w1 := "F0"
	j := 1
	message := append(append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), big.NewInt(int64(0)).Bytes()...)

	// 测试 HMAC-SHA256

	for i := 0; i < b.N; i++ {
		_, err := utils.PrfF([]byte(keyword), message)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkXtoken(b *testing.B) {
	kx := "kx"
	kz := "kz"
	w1 := "F0"
	j := 1
	p := big.NewInt(1)
	p = p.Lsh(p, 255) // 2^255
	p = p.Sub(p, big.NewInt(19))
	g := big.NewInt(9)
	xByte, w1Byte, zByte := []byte(kx), []byte(w1), []byte(kz)
	for i := 0; i < b.N; i++ {
		xtoken1, _ := utils.PrfFp(xByte, w1Byte, p, g)
		xtoken2, _ := utils.PrfFp(zByte, append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), p, g)
		_ = xtoken1
		_ = xtoken2
	}
}

func BenchmarkGAB(b *testing.B) {
	kx := "kx"
	kz := "kz"
	w1 := "F0"
	j := 1
	p := big.NewInt(1)
	p = p.Lsh(p, 255) // 2^255-19
	p = p.Sub(p, big.NewInt(19))
	g := big.NewInt(9)

	// 测试g^ab
	xtoken1, _ := utils.PrfFp([]byte(kx), []byte(w1), p, g)
	xtoken2, _ := utils.PrfFp([]byte(kz), append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), p, g)
	a := new(big.Int).Mul(xtoken1, xtoken2) // 这里假设 a 是已知的

	for i := 0; i < b.N; i++ {
		_ = new(big.Int).Exp(g, a, p)
	}
}
