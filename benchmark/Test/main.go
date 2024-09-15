package main

import (
	"fmt"
	"math/big"
	"sync"
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

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := util.PrfF([]byte(keyword), append(append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), big.NewInt(int64(0)).Bytes()...))
		if err != nil {
			fmt.Println(err)
		}
	}
	t2 := time.Since(start)
	fmt.Println("PrfF time:", t2)

	start = time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			xtoken1, _ := util.PrfFp([]byte(kx), []byte(w1), p, g)
			xtoken2, _ := util.PrfFp([]byte(kz), append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), p, g)
			// _ = xtoken1 // 使用xtoken1避免报错
			// _ = xtoken2 // 使用xtoken2避免报错
			_ = new(big.Int).Exp(g, new(big.Int).Mul(xtoken1, xtoken2), p)
		}()
	}
	wg.Wait()
	t2 = time.Since(start)
	fmt.Println("xtoken time:", t2)
}
