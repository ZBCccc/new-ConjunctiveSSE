package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"math/big"

	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"

	"github.com/Nik-U/pbc"
)

func PrfF(key, message []byte) ([]byte, error) {
	// Generate an HMAC object
	h := hmac.New(sha256.New, key)
	// Write message
	_, err := h.Write(message)
	if err != nil {
		return nil, err
	}
	// Calculate MAC of message
	return h.Sum(nil), nil
}

func PrfFp(key, message []byte, p, g *big.Int) (*big.Int, error) {
	// Generate an HMAC object
	h := hmac.New(sha256.New, key)
	// Write message
	_, err := h.Write(message)
	if err != nil {
		return nil, err
	}
	// Calculate MAC of message
	mac := h.Sum(nil)

	// Convert mac result to big.Int
	res := new(big.Int).SetBytes(mac)

	// Check if res % p == 0 and add 1 if true
	if new(big.Int).Mod(res, p).Cmp(big.NewInt(0)) == 0 {
		res.Add(res, big.NewInt(1))
	}

	// Calculate ex = res % p
	ex := new(big.Int).Mod(res, p)

	// Calculate pow(g, ex, p-1)
	pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	result := new(big.Int).Exp(g, ex, pMinus1)

	return result, nil
}

func ComputeAlpha(Ky, Kz, id []byte, op int, wWc []byte) (*pbc.Element, *pbc.Element, error) {
	// Calculate PRF_p(Ky, id||op)
	idOp := append(id, byte(op))

	alpha1, err := pbcUtil.PrfToZr(Ky, idOp)
	if err != nil {
		return nil, nil, err
	}

	// Calculate PRF_p(Kz, w||wc)
	alpha2, err := pbcUtil.PrfToZr(Kz, wWc)
	if err != nil {
		return nil, nil, err
	}

	alpha := pbcUtil.ZrDiv(alpha1, alpha2)

	return alpha, alpha1, nil
}

func HmacDigest(plaintext, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(plaintext)
	return h.Sum(nil)
}
