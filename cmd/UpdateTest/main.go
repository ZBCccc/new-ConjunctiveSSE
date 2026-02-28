package main

import (
	"ConjunctiveSSE/pkg/utils"
	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

func main() {
	keyword := "F1438"
	op := 1
	Max := 12
	srch := 1
	updt := 12

	// Fixed keys for reproducible benchmarking as used in the paper.
	kw := []byte("0123456789123456")
	kt := []byte("0123456789123456")
	kx := []byte("0123456789123456")
	ky := []byte("0123456789123456")
	kz := []byte("0123456789123456")

	start := time.Now()

	for i := range 100 {
		msgLen := len(keyword) + len(big.NewInt(int64(Max)).Bytes()) + 1
		msg := make([]byte, 0, msgLen)
		msg = append(msg, []byte(keyword)...)
		msg = append(msg, big.NewInt(int64(Max)).Bytes()...)
		id := strconv.Itoa(i)

		addr, _ := utils.PrfF(kw, append(msg, byte(0)))
		val, _ := utils.PrfF(kw, append(msg, byte(1)))

		val, _ = utils.BytesXORWithOp(val, []byte(id), int(op))

		msgLen = len(keyword) + len(big.NewInt(int64(srch)).Bytes()) + len(big.NewInt(int64(updt)).Bytes()) + 1
		msg = make([]byte, 0, msgLen)
		msg = append(msg, []byte(keyword)...)
		msg = append(msg, big.NewInt(int64(srch)).Bytes()...)
		msg = append(msg, big.NewInt(int64(updt)).Bytes()...)
		l, _ := utils.PrfF(kt, append(msg, byte(0)))
		tt, _ := utils.PrfF(kt, append(msg, byte(1)))
		xtag1, _ := pbcUtil.PrfToZr(kx, []byte(keyword))
		xtag2, _ := pbcUtil.PrfToZr(ky, append([]byte(id), byte(op)))
		xtag := pbcUtil.GToPower2(xtag1, xtag2)
		c := utils.BytesXOR(xtag.Bytes(), tt)
		alpha, _, _ := utils.ComputeAlpha(ky, kz, []byte(id), int(op), append([]byte(keyword), big.NewInt(int64(Max)).Bytes()...))

		_ = addr
		_ = val
		_ = l
		_ = c
		_ = alpha
	}

	elapsed := time.Since(start)
	fmt.Printf("Update test completed in %s\n", elapsed)
}
