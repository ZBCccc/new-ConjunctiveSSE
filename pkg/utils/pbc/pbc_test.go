package pbc

import (
	"crypto/rand"
	"math/big"
	"testing"
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

func TestPow(t *testing.T) {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	m, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	zrN := pairing.NewZr().SetBig(n)
	zrM := pairing.NewZr().SetBig(m)

	zrPowNM := Pow(zrN, zrM)

	// 验证 zrPowNM 是否等于 n^m
	expected := pairing.NewZr().PowZn(zrN, zrM)
	if !zrPowNM.Equals(expected) {
		t.Error("Pow result is incorrect")
	}
}
