package pbc

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"log"

	"github.com/Nik-U/pbc"
)

var (
	pairing *pbc.Pairing
	g1      *pbc.Element
)

func init() {
	params := `
		type a
		q 8780710799663312522437781984754049815806883199414208211028653399266475630880222957078625179422662221423155858769582317459277713367317481324925129998224791
		h 12016012264891146079388821366740534204802954401251311822919615131047207289359704531102844802183906537786776
		r 730750818665451621361119245571504901405976559617
		exp2 159
		exp1 107
		sign1 1
		sign0 1
		`
	var err error
	pairing, err = pbc.NewPairingFromString(params)
	if err != nil {
		log.Fatal("Pairing initialization failed: ", err)
	}

	g1 = pairing.NewG1().Rand()
}

// GetPairing 返回全局的 pairing 实例
func GetPairing() *pbc.Pairing {
	return pairing
}

// PrfToZr converts a hash of key and message to a pbc.Element
func PrfToZr(key, message []byte) (*pbc.Element, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(message)
	if err != nil {
		return nil, errors.New("HMAC write error")
	}
	mac := h.Sum(nil)

	zr := pairing.NewZr().SetFromHash(mac)
	return zr, nil
}

// GToPower returns g1^n
func GToPower(n *pbc.Element) *pbc.Element {
	return pairing.NewG1().PowZn(g1, n)
}

// GToPower2 returns g1^(n*m)
func GToPower2(n, m *pbc.Element) *pbc.Element {
	nm := pairing.NewZr().Mul(n, m)
	return pairing.NewG1().PowZn(g1, nm)
}

// ZrDiv returns n / m
func ZrDiv(n, m *pbc.Element) *pbc.Element {
	return pairing.NewZr().Div(n, m)
}

func Pow(n, m *pbc.Element) *pbc.Element {
	return pairing.NewG1().PowZn(n, m)
}
