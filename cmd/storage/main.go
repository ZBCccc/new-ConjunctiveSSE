package main

import (
	"ConjunctiveSSE/pkg/utils"
	"flag"
	"fmt"
)

func main() {
	// Database parameters
	dbName := flag.String("db", "Crime_USENIX_REV", "Database name")
	numDocs := flag.Int("docs", 0, "Number of documents (0 = use defaults)")
	avgKw := flag.Int("kw", 0, "Average keywords per document (0 = use defaults)")
	avgID := flag.Int("idlen", 0, "Average ID length (0 = use defaults)")
	flag.Parse()

	// Database-specific defaults
	var numDocuments, avgKeywordsPerDoc, avgIDLength int

	switch *dbName {
	case "Crime_USENIX_REV":
		// Crime dataset: ~500 documents, ~25 keywords/doc, ID ~10 chars
		numDocuments = 500
		avgKeywordsPerDoc = 25
		avgIDLength = 10
	case "Enron_USENIX":
		// Enron dataset: ~2000 documents, ~20 keywords/doc, ID ~10 chars
		numDocuments = 2000
		avgKeywordsPerDoc = 20
		avgIDLength = 10
	case "Wiki_USENIX":
		// Wiki dataset: ~10000 documents, ~15 keywords/doc, ID ~10 chars
		numDocuments = 10000
		avgKeywordsPerDoc = 15
		avgIDLength = 10
	default:
		numDocuments = 1000
		avgKeywordsPerDoc = 15
		avgIDLength = 10
	}

	// Override with command line args if provided
	if *numDocs > 0 {
		numDocuments = *numDocs
	}
	if *avgKw > 0 {
		avgKeywordsPerDoc = *avgKw
	}
	if *avgID > 0 {
		avgIDLength = *avgID
	}

	// Print header
	fmt.Println("================================================================================")
	fmt.Println("                    CDB_T Storage Overhead Comparison")
	fmt.Println("================================================================================")
	fmt.Printf("Database: %s\n", *dbName)
	fmt.Printf("Documents: %d\n", numDocuments)
	fmt.Printf("Avg keywords/document: %d\n", avgKeywordsPerDoc)
	fmt.Printf("Avg ID length: %d bytes\n", avgIDLength)
	fmt.Printf("Total keyword-document pairs: %d\n", numDocuments*avgKeywordsPerDoc)
	fmt.Println()

	// Calculate storage for all schemes
	results := utils.CalculateStorageForDatabase(*dbName, numDocuments, avgKeywordsPerDoc, avgIDLength)

	// Print detailed breakdown
	utils.PrintStorageComparison(numDocuments, avgKeywordsPerDoc, avgIDLength)

	// Print CSV format for easy import
	fmt.Println()
	fmt.Println("================================================================================")
	fmt.Println("CSV Format Output:")
	fmt.Println("================================================================================")
	fmt.Println("Scheme,Total_Bytes,Total_KB,Total_MB")
	for scheme, result := range results {
		fmt.Printf("%s,%d,%.2f,%.4f\n", scheme, result.TotalBytes, result.TotalKB, result.TotalMB)
	}

	// Print breakdown CSV
	fmt.Println()
	fmt.Println("Breakdown:")
	fmt.Println("Component,FDXT,ODXT,SDSSE-CQ,Mitra,Bestie")

	// Get all breakdown keys
	breakdowns := make(map[string]bool)
	for _, r := range results {
		for k := range r.Breakdown {
			breakdowns[k] = true
		}
	}

	for component := range breakdowns {
		fmt.Printf("%s,", component)
		for _, scheme := range []string{"FDXT", "ODXT", "SDSSE-CQ", "Mitra", "Bestie"} {
			if r, ok := results[scheme]; ok {
				if v, ok := r.Breakdown[component]; ok {
					fmt.Printf("%d", v)
				}
			}
			if scheme != "Bestie" {
				fmt.Printf(",")
			}
		}
		fmt.Println()
	}
}
