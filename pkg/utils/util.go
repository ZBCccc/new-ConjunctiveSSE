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

// BytesXORWithOp XORs the first 31 bytes of the MAC with id, and XORs the last byte of the MAC with op
func BytesXORWithOp(mac, id []byte, op int) ([]byte, error) {
	result := make([]byte, len(mac))
	copy(result, mac)
	if len(mac) != 32 {
		return nil, fmt.Errorf("MAC length must be 32 bytes")
	}

	// Ensure id length is at most 31 bytes
	if len(id) > 31 {
		return nil, fmt.Errorf("id length must be less than or equal to 31 bytes")
	}

	// Perform XOR operation
	for i := 0; i < len(id); i++ {
		result[i] = mac[i] ^ id[i]
	}

	// XOR the last byte of the MAC with op
	if op != 0 && op != 1 {
		return nil, fmt.Errorf("op must be 0 or 1")
	}
	result[31] = mac[31] ^ byte(op)

	return result, nil
}

// RemoveElement removes a specific element from a slice
func RemoveElement(slice []string, target string) []string {
	// Iterate over the slice and remove the target element
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

// WriteResultToCSV writes results to a CSV file
func WriteResultToCSV(filePath string, headers []string, data [][]string) error {
	// Create the file; if the directory does not exist, create it first
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(headers); err != nil {
		fmt.Println("Error writing headers:", err)
		return err
	}

	// Write data
	for _, record := range data {
		if err := writer.Write(record); err != nil {
			fmt.Println("Error writing data:", err)
			return err
		}
	}

	return nil
}

// WriteResultToFile writes results to a file
func WriteResultToFile(filePath string, data [][]string) error {
	// Create the file; if the directory does not exist, create it first
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// Create the file
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
	// Determine the longer slice
	var longer, shorter []byte
	if len(b1) > len(b2) {
		longer, shorter = b1, b2
	} else {
		longer, shorter = b2, b1
	}

	result := make([]byte, len(longer))
	copy(result, longer)

	// XOR the shorter slice into the longer slice
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
		// Hash auhmeKeys[i] to 16 bytes
		hash := sha256.Sum256(auhmeKeys[i])
		auhmeKeys[i] = hash[:16]
		if err != nil {
			return nil, [3][]byte{}, err
		}
	}

	return mitraKey, auhmeKeys, nil
}

// RemoveDuplicates removes duplicate elements from a slice
func RemoveDuplicates(intSlice []string) []string {
	return slice.Unique(intSlice)
}

// QueryKeywordsFromFile reads a conjunctive query file where each line is a
// set of keywords separated by '#'. Example line: "keyword1#keyword2#keyword3"
func QueryKeywordsFromFile(fileName string) [][]string {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("cannot open query file: ", err)
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
		log.Fatal("error reading query file: ", err)
	}

	return keywordsList
}

// SaveFileCntToFile saves filecnt to a file
func SaveFileCntToFile(fileCnt map[string]int, filename string) error {
	// Create the file; if the directory does not exist, create it first
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write fileCnt to a JSON file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(fileCnt)
}
