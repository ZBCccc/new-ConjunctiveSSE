package main

import (
	"ConjunctiveSSE/util"
	"golang.org/x/crypto/curve25519"
	"math/big"
	"sync"
	"testing"
	"time"
)

func TestCurve(t *testing.T) {
	// 定义 a 和 p
	kx := "kx"
	kz := "kz"
	w1 := "F0"
	j := 1
	p := big.NewInt(1)
	p = p.Lsh(p, 255)            // 2^255
	p = p.Sub(p, big.NewInt(19)) // 2^255 - 19
	g := big.NewInt(9)
	xtoken1, _ := util.PrfFp([]byte(kx), []byte(w1), p, g)
	xtoken2, _ := util.PrfFp([]byte(kz), append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), p, g)
	a := new(big.Int).Mul(xtoken1, xtoken2) // 这里假设 a 是已知的

	// 计算 a * p
	result := new([32]byte)
	a = a.Mod(a, p)
	var aBytes [32]byte
	copy(aBytes[:], a.Bytes())

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			curve25519.ScalarBaseMult(result, &aBytes)
		}()
	}
	wg.Wait()
	t2 := time.Since(start)
	t.Log("curve25519 time:", t2)
}

func TestHMACSHA256(t *testing.T) {
	keyword := "kt"
	w1 := "F0"
	j := 1
	message := append(append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), big.NewInt(int64(0)).Bytes()...)

	// 测试 HMAC-SHA256
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := util.PrfF([]byte(keyword), message)
		if err != nil {
			t.Fatal(err)
		}
	}
	t2 := time.Since(start)
	t.Log("PrfF time:", t2)
}

func TestAES256CTR(t *testing.T) {
	keyword := "kt"
	w1 := "F0"
	j := 1
	message := append(append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), big.NewInt(int64(0)).Bytes()...)

	// 测试 AES-256-CTR
	key := make([]byte, 32)
	copy(key, keyword)

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := util.PrfF_AES256_CTR(key, message)
		if err != nil {
			t.Fatal(err)
		}
	}
	t2 := time.Since(start)
	t.Log("PrfF_AES256_CTR time:", t2)
}

func TestXtoken(t *testing.T) {
	kx := "kx"
	kz := "kz"
	w1 := "F0"
	j := 1
	p := big.NewInt(1)
	p = p.Lsh(p, 255) // 2^255
	p = p.Sub(p, big.NewInt(19))
	g := big.NewInt(9)

	start := time.Now()
	for i := 0; i < 1000; i++ {
		xtoken1, _ := util.PrfFp([]byte(kx), []byte(w1), p, g)
		xtoken2, _ := util.PrfFp([]byte(kz), append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), p, g)
		_ = xtoken1
		_ = xtoken2
	}
	t2 := time.Since(start)
	t.Log("xtoken time:", t2)
}

func TestGAB(t *testing.T) {
	kx := "kx"
	kz := "kz"
	w1 := "F0"
	j := 1
	p := big.NewInt(1)
	p = p.Lsh(p, 255) // 2^255-19
	p = p.Sub(p, big.NewInt(19))
	g := big.NewInt(9)

	// 测试g^ab
	xtoken1, _ := util.PrfFp([]byte(kx), []byte(w1), p, g)
	xtoken2, _ := util.PrfFp([]byte(kz), append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), p, g)
	a := new(big.Int).Mul(xtoken1, xtoken2) // 这里假设 a 是已知的

	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = new(big.Int).Exp(g, a, p)
		}()
	}
	wg.Wait()
	t2 := time.Since(start)
	t.Log("g^ab time:", t2)
}
