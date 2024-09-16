package main

import (
	"fmt"
	"math/big"
	"time"

	"ConjunctiveSSE/util"
)

func main() {
	keyword := "kt"
	kx := "kx"
	kz := "kz"
	w1 := "F0"
	j := 1
	p, _ := new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)
	g := big.NewInt(65537)
	message := append(append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), big.NewInt(int64(0)).Bytes()...)

	// 测试 HMAC-SHA256
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := util.PrfF([]byte(keyword), message)
		if err != nil {
			fmt.Println(err)
		}
	}
	t2 := time.Since(start)
	fmt.Println("PrfF time:", t2)

	// 测试 AES-256-CTR
	key := make([]byte, 32)
	copy(key, kx)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		_, err := util.PrfF_AES256_CTR(key, message)
		if err != nil {
			fmt.Println(err)
		}
	}
	t2 = time.Since(start)
	fmt.Println("PrfF_AES256_CTR time:", t2)

	// 测试 xtoken
	
	start = time.Now()
	for i := 0; i < 1000; i++ {
		xtoken1, _ := util.PrfFp([]byte(kx), []byte(w1), p, g)
		xtoken2, _ := util.PrfFp([]byte(kz), append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), p, g)
		_ = xtoken1
		_ = xtoken2
		// _ = new(big.Int).Exp(g, new(big.Int).Mul(xtoken1, xtoken2), p)
	}
	t2 = time.Since(start)
	fmt.Println("xtoken time:", t2)

	// 测试g^ab
	xtoken1, _ := util.PrfFp([]byte(kx), []byte(w1), p, g)
	xtoken2, _ := util.PrfFp([]byte(kz), append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), p, g)
	start = time.Now()
	for i := 0; i < 1000; i++ {
		_ = new(big.Int).Exp(g, new(big.Int).Mul(xtoken1, xtoken2), p)
	}
	t2 = time.Since(start)
	fmt.Println("g^ab time:", t2)

	// 测试ab
	start = time.Now()
	for i := 0; i < 1000; i++ {
		ab := new(big.Int).Mul(xtoken1, xtoken2)
		_ = ab
	}
	t2 = time.Since(start)
	fmt.Println("ab time:", t2)
}
