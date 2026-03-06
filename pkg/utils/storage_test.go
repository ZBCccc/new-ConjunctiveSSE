package utils

import (
	"fmt"
	"testing"
)

func TestStorageCalculation(t *testing.T) {
	// Test with sample database parameters
	// Crime_USENIX_REV: ~1000 documents, ~15 keywords/doc
	numDocuments := 1000
	avgKeywordsPerDoc := 15
	avgIDLength := 10

	fmt.Println("Testing storage calculation with sample data:")
	fmt.Printf("Documents: %d, Avg keywords/doc: %d, Avg ID length: %d\n",
		numDocuments, avgKeywordsPerDoc, avgIDLength)
	fmt.Println()

	// Test FDXT/ODXT/SDSSE-CQ
	fdxtSize := CalculateFDXTODXTSDSSECQSize(numDocuments, avgKeywordsPerDoc)
	fmt.Printf("FDXT/ODXT/SDSSE-CQ:\n")
	fmt.Printf("  Total: %d bytes (%.2f KB, %.2f MB)\n",
		fdxtSize.TotalBytes, fdxtSize.TotalKB, fdxtSize.TotalMB)

	// Verify expected size with ACTUAL measured PBC size
	// Per entry: addr(32) + val(32) + alpha(16) + l(32) + c(32) = 144 bytes
	// Note: alpha is now 16 bytes (measured) instead of 128 bytes (theoretical)
	expectedPerEntry := 32 + 32 + 16 + 32 + 32
	expectedTotal := numDocuments * avgKeywordsPerDoc * expectedPerEntry
	if fdxtSize.TotalBytes != expectedTotal {
		t.Errorf("FDXT size mismatch: got %d, expected %d", fdxtSize.TotalBytes, expectedTotal)
	}

	// Test Mitra
	mitraSize := CalculateMitraSize(numDocuments, avgKeywordsPerDoc)
	fmt.Printf("Mitra:\n")
	fmt.Printf("  Total: %d bytes (%.2f KB, %.2f MB)\n",
		mitraSize.TotalBytes, mitraSize.TotalKB, mitraSize.TotalMB)

	// Per entry: addr(32) + val(32) = 64 bytes
	mitraExpectedPerEntry := 32 + 32
	mitraExpectedTotal := numDocuments * avgKeywordsPerDoc * mitraExpectedPerEntry
	if mitraSize.TotalBytes != mitraExpectedTotal {
		t.Errorf("Mitra size mismatch: got %d, expected %d", mitraSize.TotalBytes, mitraExpectedTotal)
	}

	// Test Bestie
	bestieSize := CalculateBestieSize(numDocuments, avgKeywordsPerDoc, avgIDLength)
	fmt.Printf("Bestie:\n")
	fmt.Printf("  Total: %d bytes (%.2f KB, %.2f MB)\n",
		bestieSize.TotalBytes, bestieSize.TotalKB, bestieSize.TotalMB)

	// Verify ratios
	mitraRatio := float64(mitraSize.TotalBytes) / float64(fdxtSize.TotalBytes)
	bestieRatio := float64(bestieSize.TotalBytes) / float64(fdxtSize.TotalBytes)

	fmt.Printf("\nRatios (based on actual PBC G1 size = 16 bytes):\n")
	fmt.Printf("  Mitra/FDXT: %.2f\n", mitraRatio)
	fmt.Printf("  Bestie/FDXT: %.2f\n", bestieRatio)

	// Expected ratios with actual PBC size (16 bytes)
	// FDXT: 144 bytes per entry (32+32+16+32+32)
	// Mitra: 64 bytes per entry = 0.44x
	// Bestie: 64 + 10 + 16 = 90 bytes per entry = 0.62x
	expectedMitraRatio := 64.0 / 144.0
	expectedBestieRatio := float64(64+avgIDLength+16) / 144.0

	if abs(mitraRatio-expectedMitraRatio) > 0.01 {
		t.Errorf("Mitra ratio unexpected: got %.2f, expected %.2f", mitraRatio, expectedMitraRatio)
	}
	if abs(bestieRatio-expectedBestieRatio) > 0.01 {
		t.Errorf("Bestie ratio unexpected: got %.2f, expected %.2f", bestieRatio, expectedBestieRatio)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestPrintStorageComparison(t *testing.T) {
	// Test print function
	PrintStorageComparison(1000, 15, 10)
}
