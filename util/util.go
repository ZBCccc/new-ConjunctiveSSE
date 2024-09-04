package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"sync"
)

const MAXBYTES = 64

type Operation int

const (
	Del Operation = iota // 0
	Add                  // 1
)

type DataPacket struct {
	Address []byte
	Val     []byte
	Alpha   *big.Int
	Xtag    *big.Int
}

type SEOp struct {
	J    int
	Sval []byte
	Cnt  int
}

// RequestType 请求类型
type RequestType int

// 请求类型
const (
	Update RequestType = iota
	Search
)

// Request 请求
type Request struct {
	Type    RequestType
	Payload interface{}
}

type UpdatePayload struct {
	Address []byte
	Val     []byte
	Alpha   *big.Int
	Xtag    *big.Int
}

type SearchPayload struct {
	StokenList [][]byte
	XtokenList [][]*big.Int
}

type Response struct {
	SEOpList []SEOp
}

var registerOnce sync.Once

func RegisterTypes() {
	registerOnce.Do(func() {
		gob.Register(UpdatePayload{})
		gob.Register(SearchPayload{})
		gob.Register(Response{})
		gob.Register(Request{})
	})
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

func ComputeAlpha(Ky, Kz, id []byte, op int, wWc []byte, p, g *big.Int) (*big.Int, *big.Int, error) {
	// 计算 PRF_p(Ky, id||op)
	idOp := append(id, byte(op))
	alpha1, err := PrfFp(Ky, idOp, p, g)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	// 计算 PRF_p(Kz, w||wc)
	alpha2, err := PrfFp(Kz, wWc, p, g)
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	// Calculate alpha = alpha1 * alpha2
	pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	alpha2 = new(big.Int).ModInverse(alpha2, pMinus1)

	alpha := new(big.Int).Mul(alpha1, alpha2)

	return alpha, alpha1, nil
}

// BytesXORWithOp 将MAC值的前31个字节与id异或，并将MAC的最后一个字节与op异或
func BytesXORWithOp(mac, id []byte, op int) ([]byte, error) {
	if len(mac) != 32 {
		return nil, fmt.Errorf("MAC length must be 32 bytes")
	}

	// 确保id的长度为小于等于31字节
	if len(id) > 31 {
		return nil, fmt.Errorf("id length must be less than or equal to 31 bytes")
	}

	// 执行异或操作
	for i := 0; i < len(id); i++ {
		mac[i] = mac[i] ^ id[i]
	}

	// 将MAC的最后一个字节与op异或
	if op != 0 && op != 1 {
		return nil, fmt.Errorf("op must be 0 or 1")
	}
	mac[31] = mac[31] ^ byte(op)

	return mac, nil
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
