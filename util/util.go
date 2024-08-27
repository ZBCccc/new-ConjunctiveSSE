package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"math/big"
)

const MAXBYTES = 64

type Operation int

const (
	del Operation = iota // 0
	Add                  // 1
)

type DataPacket struct {
	Address []byte
	Val     []byte
	Alpha   *big.Int
	Xtag    *big.Int
}

func PrfF(key, message []byte) ([]byte, error) {
	// 生成一个HMAC对象
	h := hmac.New(sha256.New, key)
	// 写入消息
	_, err := h.Write(message)
	if err != nil {
		return nil, err
	}
	// 计算消息的MAC
	return h.Sum(nil), nil
}

func PrfFp(key, message []byte, p, g *big.Int) (*big.Int, error) {
	// 生成一个HMAC对象
	h := hmac.New(sha256.New, key)
	// 写入消息
	_, err := h.Write(message)
	if err != nil {
		return nil, err
	}
	// 计算消息的MAC
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

func bytesXOR(a, b []byte) []byte {
	result := make([]byte, len(a))
	for i := range a {
		result[i] = a[i] ^ b[i]
	}
	return result
}

// MulInv 计算 a 在模 b 下的乘法逆元
func MulInv(a, b *big.Int) *big.Int {
	// 确保 a 为正数
	a = new(big.Int).Mod(a, b)

	// 使用扩展欧几里得算法
	x := new(big.Int)
	y := new(big.Int)
	gcd := new(big.Int).GCD(x, y, a, b)

	// 如果 gcd 不为 1，则不存在乘法逆元
	if gcd.Cmp(big.NewInt(1)) != 0 {
		return nil
	}

	// 确保结果为正数
	if x.Sign() < 0 {
		x.Add(x, b)
	}

	return x.Mod(x, b)
}
