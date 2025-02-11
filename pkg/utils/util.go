package utils

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Operation int

const (
	Add Operation = iota // 0
	Del                  // 1
)

type SEOp struct {
	J    int
	Sval string
	Cnt  int
}

func (seop *SEOp) computeSize() int {
	return 4 + len(seop.Sval) + 4
}

func CalculatePayloadSize(seopList []SEOp) int {
	size := 0
	for _, seop := range seopList {
		size += seop.computeSize()
	}
	return size
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
	// 判断最长的切片
	var longer, shorter []byte
	if len(b1) > len(b2) {
		longer, shorter = b1, b2
	} else {
		longer, shorter = b2, b1
	}

	result := make([]byte, len(longer))
	copy(result, longer)

	// 将较短的切片异或到较长切片中
	for i := 0; i < len(shorter); i++ {
		result[i] ^= shorter[i]
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
	return slice.Unique(intSlice)
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

// SaveFileCntToFile 保存 filecnt 到文件
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
