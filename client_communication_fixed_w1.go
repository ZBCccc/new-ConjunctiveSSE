package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func main() {
	// Communication sizes per keyword-document pair in bytes
	storageSizes := map[string]int{
		"Nomos":   10 * (1 + 3) * 32,
		"MC-ODXT": 10 * 1 * 32,
		"VQNomos": 10*(1+3)*32 + 16,
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
			seenCounts := make(map[int]struct{})
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				matches := re.FindStringSubmatch(line)
				if len(matches) == 3 {
					count, _ := strconv.Atoi(matches[2])
					if _, seen := seenCounts[count]; seen {
						continue
					}
					seenCounts[count] = struct{}{}
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
			filename := fmt.Sprintf("pic/nomos_client_communication_data_w1/%s_%s.csv", scheme, outputName)
			outFile, err := os.Create(filename)
			if err != nil {
				fmt.Printf("Error creating %s: %v\n", filename, err)
				continue
			}

			// Write header
			fmt.Fprintf(outFile, "KeywordCount,Storage(Bytes)\n")

			bytesPerPair := storageSizes[scheme]

			// Calculate cumulative storage
			for _, docCount := range docCounts {
				// Calculate storage for this keyword's documents
				storageForKeyword := bytesPerPair

				// Write row
				fmt.Fprintf(outFile, "%d,%d\n", docCount, storageForKeyword)
			}

			outFile.Close()
			fmt.Printf("Created %s (%d keywords)\n", filename, len(docCounts))
		}
	}

	fmt.Println("\nAll files generated successfully!")
}
