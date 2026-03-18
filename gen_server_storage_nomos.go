package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func main() {
	// Storage sizes per keyword-document pair in bytes
	storageSizes := map[string]int{
		"Nomos":   96 + 12 * 32,
		"MC-ODXT": 96 + 32,
		"VQNomos": 96 + 12 * 32 + 3248,
	}

	// Dataset files and their corresponding JSON files
	datasets := map[string]string{
		"Enron": "pic/raw_data/Enron_filecnt_sorted.json",
		"Crime": "pic/raw_data/Crime_filecnt_sorted.json",
		"Wiki":  "pic/raw_data/Wiki_filecnt_sorted.json",
	}

	schemes := []string{"Nomos", "MC-ODXT", "VQNomos"}

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
			filename := fmt.Sprintf("pic/nomos_server_data/%s_%s.csv", scheme, outputName)
			outFile, err := os.Create(filename)
			if err != nil {
				fmt.Printf("Error creating %s: %v\n", filename, err)
				continue
			}

			// Write header
			fmt.Fprintf(outFile, "KeywordCount,Storage(Bits)\n")

			bytesPerPair := storageSizes[scheme]
			bitsPerPair := bytesPerPair * 8

			// Calculate cumulative storage
			cumulativeStorage := 0
			if scheme == "VQNomos" {
				// For VQNomos, add the fixed overhead for each keyword
				cumulativeStorage += 2147483392 // 256 MB in bits
			}
			for i, docCount := range docCounts {
				// Calculate storage for this keyword's documents
				storageForKeyword := docCount * bitsPerPair
				cumulativeStorage += storageForKeyword

				// Write row
				fmt.Fprintf(outFile, "%d,%d\n", i+1, cumulativeStorage)
			}

			outFile.Close()
			fmt.Printf("Created %s (%d keywords)\n", filename, len(docCounts))
		}
	}

	fmt.Println("\nAll files generated successfully!")
}
