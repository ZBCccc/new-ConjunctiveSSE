package util

/*
#cgo CFLAGS: -I/usr/local/opt/openssl/include
#cgo LDFLAGS: -L/usr/local/opt/openssl/lib -lssl -lcrypto
#include "crypto.h"
#include <stdlib.h>
*/
import "C"
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

type Operation int

const (
	Del Operation = iota // 0
	Add                  // 1
)

type SEOp struct {
	J    int
	Sval string
	Cnt  int
}

func PrfF_C(key, message []byte) ([]byte, error) {
	output := make([]byte, 32)
	outputLen := C.int(len(output))
	C.hmac_sha256(C.CString(key), C.int(len(key)), C.CString(message), C.int(len(message)), (*C.uchar)(unsafe.Pointer(&output[0])), (*C.int)(unsafe.Pointer(&outputLen)))
	return output, nil
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

// PrfF_AES256_CTR 基于 AES-256 in counter mode 实现的 PRF 函数
func PrfF_AES256_CTR(key, message []byte) ([]byte, error) {
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

func PrfFp(key, message []byte, p, g *big.Int) (*big.Int, error) {
	// 生成一个HMAC对象
	h := hmac.New(sha256.New, key)
	// 写入消息
	if _, err := h.Write(message); err != nil {
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

func Base64ToBigInt(base64Str string) (*big.Int, error) {
	// Base64解码
	decodedBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, err
	}

	// 将[]byte转换为big.Int
	bigIntValue := new(big.Int).SetBytes(decodedBytes)
	return bigIntValue, nil
}

// RemoveElement 删除列表中的特定元素，返回新的列表
func RemoveElement(slice []string, target string) []string {
	newSlice := []string{}
	for _, v := range slice {
		if v != target {
			newSlice = append(newSlice, v)
		}
	}
	return newSlice
}

func RemoveElementFromSlice(slice []string, target string) []string {
	for i, v := range slice {
		if v == target {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// WriteResult 将结果写入CSV文件
func WriteResultToCSV(filePath string, headers []string, data [][]string) error {
	// 创建文件，如果文件所在的目录不存在则创建
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	if err := writer.Write(headers); err != nil {
		fmt.Println("Error writing headers:", err)
		return err
	}

	// 写入数据
	for _, record := range data {
		if err := writer.Write(record); err != nil {
			fmt.Println("Error writing data:", err)
			return err
		}
	}

	return nil
}

// WriteResultToFile 将结果写入文件
func WriteResultToFile(filePath string, data [][]string) error {
	// 创建文件，如果文件所在的目录不存在则创建
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, record := range data {
		_, err = file.WriteString(strings.Join(record, "#") + "\n")
		if err != nil {
			fmt.Println("Error writing data:", err)
			return err
		}
	}

	fmt.Println("Data written to file:", filePath)
	return nil
}
