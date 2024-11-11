package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/big"
)

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

// PrffAes256Ctr 基于 AES-256 in counter mode 实现的 PRF 函数
func PrffAes256Ctr(key, message []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key length must be 32 bytes for AES-256")
	}

	// 创建 AES-256 块加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 初始化计数器（IV）
	iv := make([]byte, aes.BlockSize)
	binary.BigEndian.PutUint64(iv[aes.BlockSize-8:], 1)

	// 创建 CTR 模式的流加密器
	stream := cipher.NewCTR(block, iv)

	// 加密消息
	ciphertext := make([]byte, len(message))
	stream.XORKeyStream(ciphertext, message)

	return ciphertext, nil
}

var one = big.NewInt(1)
var zero = big.NewInt(0)

func PrfFp(key, message []byte, p, g *big.Int) (*big.Int, error) {
	// // 生成一个HMAC对象
	// h := hmac.New(sha256.New, key)
	// // 写入消息
	// if _, err := h.Write(message); err != nil {
	// 	return nil, err
	// }
	// // 计算消息的MAC
	// mac := h.Sum(nil)

	// // Convert mac result to big.Int
	// res := new(big.Int).SetBytes(mac)
	// res.Mod(res, p)

	// // Check if res % p == 0 and add 1 if true
	// if res.Cmp(zero) == 0 {
	// 	res.Add(res, one)
	// }

	// return res, nil
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
	if new(big.Int).Mod(res, p).Cmp(zero) == 0 {
		res.Add(res, one)
	}

	// Calculate ex = res % p
	ex := new(big.Int).Mod(res, p)

	// Calculate pow(g, ex, p-1)
	pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	result := new(big.Int).Exp(g, ex, pMinus1)

	return result, nil
}

func ComputeAlpha(Ky, Kz, id []byte, op int, wWc []byte, p, g *big.Int) (*big.Int, *big.Int, error) {
	// // 计算 PRF_p(Ky, id||op)
	// idOp := append(id, byte(op))
	// alpha1, err := PrfFp(Ky, idOp, p, g)
	// if err != nil {
	// 	log.Println(err)
	// 	return nil, nil, err
	// }

	// // 计算 PRF_p(Kz, w||wc)
	// alpha2, err := PrfFp(Kz, wWc, p, g)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil, nil, err
	// }

	// // Calculate alpha = alpha1 * alpha2
	// //pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	// alpha2 = new(big.Int).ModInverse(alpha2, p)

	// alpha := new(big.Int).Mod(new(big.Int).Mul(alpha1, alpha2), p)

	// return alpha, alpha1, nil
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
