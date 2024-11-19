package utils

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/bits-and-blooms/bloom/v3"
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
	result := make([]byte, len(mac))
	copy(result, mac)
	if len(mac) != 32 {
		return nil, fmt.Errorf("MAC length must be 32 bytes")
	}

	// 确保id的长度为小于等于31字节
	if len(id) > 31 {
		return nil, fmt.Errorf("id length must be less than or equal to 31 bytes")
	}

	// 执行异或操作
	for i := 0; i < len(id); i++ {
		result[i] = mac[i] ^ id[i]
	}

	// 将MAC的最后一个字节与op异或
	if op != 0 && op != 1 {
		return nil, fmt.Errorf("op must be 0 or 1")
	}
	result[31] = mac[31] ^ byte(op)

	return result, nil
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

// RemoveElement 删除slice中的特定元素
func RemoveElement(slice []string, target string) []string {
	// 遍历slice，删除target元素
	result := make([]string, len(slice))
	copy(result, slice)
	for i, v := range result {
		if v == target {
			result = append(result[:i], result[i+1:]...)
			return result
		}
	}
	return result
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

func BytesXOR(b1, b2 []byte) []byte {
	// b1, b2的长度相等
	result := make([]byte, len(b1))
	copy(result, b1)

	for i := 0; i < len(b1); i++ {
		result[i] = b1[i] ^ b2[i]
	}

	return result
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
		// 将auhmeKeys[i]哈希为16字节
		hash := sha256.Sum256(auhmeKeys[i])
		auhmeKeys[i] = hash[:16]
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

// 保存 filecnt 到文件
func SaveFileCntToFile(fileCnt map[string]int, filename string) error {
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

	// 将 fileCnt 写入Json文件
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(fileCnt)
}
