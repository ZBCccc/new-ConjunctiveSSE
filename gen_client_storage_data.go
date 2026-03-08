package main

import (
	"bufio"
	"fmt"
	"math/bits"
	"os"
	"regexp"
	"strconv"
)

// calculateBinaryStorage 计算整数的二进制表示占用的比特数
func calculateBinaryStorage(n int) int {
	if n == 0 {
		return 1 // 0 需要 1 bit
	}
	// 使用 bits.Len 获取二进制位数
	return bits.Len(uint(n))
}

func main() {
	// Dataset files and their corresponding JSON files
	datasets := map[string]string{
		"Enron": "pic/raw_data/Enron_filecnt_sorted.json",
		"Crime": "pic/raw_data/Crime_filecnt_sorted.json",
		"Wiki":  "pic/raw_data/Wiki_filecnt_sorted.json",
	}

	schemes := []string{"ODXT", "FDXT", "SDSSE-CQ"}

	// SDSSE-CQ 特殊参数
	dMax := 1644
	bfStorage := float64(dMax) * 0.0024 * 1024 * 8 * 2

	// Regex to extract key-value pairs from JSON
	re := regexp.MustCompile(`"([^"]+)":\s*(\d+)`)

	for _, scheme := range schemes {
		for datasetName, jsonPath := range datasets {
			// Read JSON file line by line to preserve order
			file, err := os.Open(jsonPath)
			if err != nil {
				fmt.Printf("Error reading %s: %v\n", jsonPath, err)
				continue
			}

			var docCounts []int
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				matches := re.FindStringSubmatch(line)
				if len(matches) == 3 {
					count, _ := strconv.Atoi(matches[2])
					docCounts = append(docCounts, count)
				}
			}
			file.Close()

			if err := scanner.Err(); err != nil {
				fmt.Printf("Error scanning %s: %v\n", jsonPath, err)
				continue
			}

			// Create output CSV file
			outputName := datasetName
			if datasetName == "Wiki" {
				outputName = "Wikipedia"
			}
			filename := fmt.Sprintf("pic/server_data/%s_%s.csv", scheme, outputName)
			outFile, err := os.Create(filename)
			if err != nil {
				fmt.Printf("Error creating %s: %v\n", filename, err)
				continue
			}

			// Write header
			fmt.Fprintf(outFile, "KeywordCount,Storage(Bits)\n")

			// Calculate cumulative storage
			cumulativeStorage := 0
			for i, docCount := range docCounts {
				var storageForKeyword int

				switch scheme {
				case "ODXT":
					// storage = calculate_binary_storage(KeywordCount) + 9 * 8
					storageForKeyword = calculateBinaryStorage(docCount) + 9*8
				case "FDXT":
					// storage = calculate_binary_storage(KeywordCount) * 1.02 + 9 * 8
					storageForKeyword = calculateBinaryStorage(docCount)*1.02 + 9*8
				case "SDSSE-CQ":
					// storage = calculate_binary_storage(KeywordCount) * 3 + bf_storage + 9 * 8
					storageForKeyword = calculateBinaryStorage(docCount)*3 + int(bfStorage) + 9*8
				}

				cumulativeStorage += storageForKeyword

				// Write row
				fmt.Fprintf(outFile, "%d,%d\n", i+1, cumulativeStorage)
			}

			outFile.Close()
			fmt.Printf("Created %s (%d keywords)\n", filename, len(docCounts))
		}
	}

	fmt.Println("\nAll server storage files generated successfully!")
}
