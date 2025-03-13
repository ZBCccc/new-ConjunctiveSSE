package pbc

import (
	"crypto/rand"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/Nik-U/pbc"
)

func TestPrfToZr(t *testing.T) {
	key := []byte("testkey")
	message := []byte("testmessage")

	zr, err := PrfToZr(key, message)
	if err != nil {
		t.Fatalf("PrfToZr failed: %v", err)
	}

	if zr.Is0() {
		t.Error("PrfToZr returned zero element")
	}
}

func TestGToPower(t *testing.T) {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	zr := pairing.NewZr().SetBig(n)

	gPowN := GToPower(zr)
	if gPowN.Is0() {
		t.Error("GToPower returned zero element")
	}

	// 验证 gPowN 是否等于 g1^n
	expected := pairing.NewG1().PowZn(g1, zr)
	if !gPowN.Equals(expected) {
		t.Error("GToPower result is incorrect")
	}
}

func TestGToPower2(t *testing.T) {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	m, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	zrN := pairing.NewZr().SetBig(n)
	zrM := pairing.NewZr().SetBig(m)

	gPowNM := GToPower2(zrN, zrM)
	if gPowNM.Is0() {
		t.Error("GToPower2 returned zero element")
	}

	// 验证 gPowNM 是否等于 g1^(n*m)
	expected := pairing.NewG1().PowZn(g1, pairing.NewZr().Mul(zrN, zrM))
	if !gPowNM.Equals(expected) {
		t.Error("GToPower2 result is incorrect")
	}
}

func TestZrDiv(t *testing.T) {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	m, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	zrN := pairing.NewZr().SetBig(n)
	zrM := pairing.NewZr().SetBig(m)

	zrDivNM := ZrDiv(zrN, zrM)

	// 验证 zrDivNM 是否等于 n/m
	expected := pairing.NewZr().Div(zrN, zrM)
	if !zrDivNM.Equals(expected) {
		t.Error("ZrDiv result is incorrect")
	}
}

// func TestPow(t *testing.T) {
// 	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
// 	m, _ := rand.Int(rand.Reader, big.NewInt(1000000))
// 	zrN := pairing.NewZr().SetBig(n)
// 	zrM := pairing.NewZr().SetBig(m)

// 	zrPowNM := Pow(zrN, zrM)

// 	// 验证 zrPowNM 是否等于 n^m
// 	expected := pairing.NewZr().PowZn(zrN, zrM)
// 	if !zrPowNM.Equals(expected) {
// 		t.Error("Pow result is incorrect")
// 	}
// }

func TestBytesToElement(t *testing.T) {
	key := []byte("testkey")
	message := []byte("testmessage")

	zr, _ := PrfToZr(key, message)
	g := GToPower(zr)
	bytesN := g.Bytes()
	expected := BytesToG1(bytesN)
	if !g.Equals(expected) {
		t.Error("BytesToElement result is incorrect")
	}
}

func ComputeAlpha(Ky, Kz, id []byte, op int, wWc []byte) (*pbc.Element, *pbc.Element, error) {
	// 计算 PRF_p(Ky, id||op)
	idOp := append(id, byte(op))

	alpha1, err := PrfToZr(Ky, idOp)
	if err != nil {
		return nil, nil, err
	}

	// 计算 PRF_p(Kz, w||wc)
	alpha2, err := PrfToZr(Kz, wWc)
	if err != nil {
		return nil, nil, err
	}

	alpha := ZrDiv(alpha1, alpha2)

	return alpha, alpha1, nil
}

func TestTimeCost(t *testing.T) {
	// kt := []byte("0123456789123456")
	kx := []byte("0123456789123456")
	ky := []byte("0123456789123456")
	kz := []byte("0123456789123456")
	w1 := "F101"
	id := "0"
	start := time.Now()
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = i
			alpha, _, _ := ComputeAlpha(ky, kz, []byte(id), 0, append([]byte(w1), byte(5)))

			zr1, _ := PrfToZr(kx, []byte(w1))
			zr2, _ := PrfToZr(kz, append([]byte(w1), byte(5)))
			g := GToPower2(zr1, zr2)
			g = Pow(g, alpha)
			_ = g
		}()
	}
	wg.Wait()

	t.Log("time cost:", time.Since(start))
}
