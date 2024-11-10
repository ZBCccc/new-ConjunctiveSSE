package utils

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/bits-and-blooms/bloom/v3"
	mapset "github.com/deckarep/golang-set/v2"
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
	res.Mod(res, p)

	// Check if res % p == 0 and add 1 if true
	if res.Cmp(zero) == 0 {
		res.Add(res, one)
	}

	return res, nil
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
	//pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	alpha2 = new(big.Int).ModInverse(alpha2, p)

	alpha := new(big.Int).Mod(new(big.Int).Mul(alpha1, alpha2), p)

	return alpha, alpha1, nil
}

// BytesXORWithOp 将MAC值的前31个字节与id异或，并将MAC的最后一个字节与op异或
func BytesXORWithOp(mac, id []byte, op int) ([]byte, error) {
	if len(mac) != 32 {
		return nil, fmt.Errorf("MAC length must be 32 bytes")
	}

	// 确保id的长度为小于等于31字节，截断只取前31个字节
	if len(id) > 31 {
		id = id[:31]
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

// WriteResultToCSV 将结果写入CSV文件
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
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
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

func HdxtReadKeys(filePath string) ([]byte, [3][]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, [3][]byte{}, err
	}
	defer file.Close()

	var mitraKey []byte
	var auhmeKeys [3][]byte

	scanner := bufio.NewScanner(file)

	line := scanner.Text()
	mitraKey, err = base64.StdEncoding.DecodeString(line)
	if err != nil {
		return nil, [3][]byte{}, err
	}

	for i := 0; i < 3; i++ {
		line := scanner.Text()
		auhmeKeys[i], err = base64.StdEncoding.DecodeString(line)
		if err != nil {
			return nil, [3][]byte{}, err
		}
	}

	return mitraKey, auhmeKeys, nil
}

// RemoveDuplicates 去除切片中的重复元素
func RemoveDuplicates(intSlice []string) []string {
	// 创建一个新的string集合
	stringSet := mapset.NewSet[string]()

	// 将切片中的元素添加到集合中
	for _, v := range intSlice {
		stringSet.Add(v)
	}

	// 转换为切片
	return stringSet.ToSlice()
}

// SaveBloomFilterToFile 保存 Bloom filter 到文件
func SaveBloomFilterToFile(filter *bloom.BloomFilter, filename string) error {
	// 创建文件，如果所在目录不存在，则先创建目录，再创建文件
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将 Bloom filter 写入文件
	_, err = filter.WriteTo(file)
	return err
}

// LoadBloomFilterFromFile 从文件加载 Bloom filter
func LoadBloomFilterFromFile(filename string) (*bloom.BloomFilter, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	filter := bloom.NewWithEstimates(1000000, 0.01) // 创建一个新的 Bloom filter，使用整数参数
	_, err = filter.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	return filter, nil
}

// SaveUpdateCntToFile 保存 UpdateCnt 到文件
func SaveUpdateCntToFile(updateCnt map[string]int, filename string) error {
	// 创建文件，如果所在目录不存在，则先创建目录，再创建文件
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将 UpdateCnt 写入Json文件
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(updateCnt)
}

// LoadUpdateCntFromFile 从文件加载 UpdateCnt
func LoadUpdateCntFromFile(filename string) (map[string]int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	// 从文件加载 UpdateCnt
	updateCnt := make(map[string]int)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&updateCnt)
	if err != nil {
		return nil, err
	}
	return updateCnt, nil
}

func QueryKeywordsFromFile(fileName string) [][]string {
	// 读取待搜索的连接关键词文件，文件格式为：
	// 每一行都是关键词的集合，关键词之间用#隔开
	// 例如：
	// 关键词1#关键词2#关键词3
	// 关键词4#关键词5
	// 关键词6
	// 读取待搜索的关键词文件
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("无法打开文件:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var keywordsList [][]string

	for scanner.Scan() {
		line := scanner.Text()
		keywords := strings.Split(line, "#")
		keywordsList = append(keywordsList, keywords)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("读取文件时出错:", err)
	}

	return keywordsList
}

func BytesXOR(b1, b2 []byte) []byte {
	// b1, b2的长度均为32字节

	var result []byte
	result = make([]byte, len(b1))
	copy(result, b1)

	// 对最小长度的部分进行异或操作
	for i := 0; i < len(b1); i++ {
		result[i] = b1[i] ^ b2[i]
	}

	return result
}
