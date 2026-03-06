package utils

import (
	"fmt"
)

// StorageSizeCalculator calculates the storage overhead for different SSE schemes
// Based on the following cryptographic primitives:
// - AES-256 in CBC mode (32-byte key, IV = 16 bytes)
// - HMAC-SHA256 for PRFs (32-byte output)
// - SHA-256 for hash functions (32-byte output)
// - PBC type A (standard): G1 element = 128 bytes, Zr element = 20 bytes
// Note: PBC uses standard Type A curve (q ~ 160 bits)

const (
	// AES-256
	AESKeySize   = 32 // bytes
	AESBlockSize = 16 // bytes (IV size)

	// HMAC-SHA256
	HMACOutputSize = 32 // bytes

	// SHA-256
	SHA256OutputSize = 32 // bytes

	// PBC type A (standard) - ACTUAL MEASURED SIZES
	// Measured via TestPBCG1Size with standard Type A: G1 = 128 bytes, Zr = 20 bytes
	PBCG1Size = 128 // bytes (actual measured with standard Type A)
	PBCZrSize = 20  // bytes (actual measured with standard Type A)
)

// CDBTSizeResult stores the storage size breakdown for CDB_T
type CDBTSizeResult struct {
	Scheme        string
	TotalBytes    int
	TotalKB       float64
	TotalMB       float64
	Breakdown     map[string]int
	Description   string
}

// CalculateFDXTODXTSDSSECQSize calculates the storage for FDXT, ODXT, and SDSSE-CQ
// These three schemes have the same theoretical CDB_T structure:
// - addr: 32 bytes (HMAC-SHA256 output)
// - val: 32 bytes (HMAC-SHA256 output)
// - alpha: ~128 bytes (PBC G1 element)
// Plus XTag/XSet storage: l (32B) + c (32B)
func CalculateFDXTODXTSDSSECQSize(numDocuments int, avgKeywordsPerDoc int) CDBTSizeResult {
	// Each document-keyword pair produces one entry in CDB_T
	totalEntries := numDocuments * avgKeywordsPerDoc

	breakdown := make(map[string]int)

	// CDBTSet: addr (key) + val + alpha (value)
	addrSize := totalEntries * HMACOutputSize // address keys
	valSize := totalEntries * HMACOutputSize  // val values
	alphaSize := totalEntries * PBCG1Size     // alpha (PBC G1 element)

	breakdown["addr (CDBTSet key)"] = addrSize
	breakdown["val (CDBTSet value)"] = valSize
	breakdown["alpha (CDBTSet value)"] = alphaSize
	cdbtsetSize := addrSize + valSize + alphaSize

	// CDBXtag: l + c (for each keyword)
	// Each keyword appears in multiple documents
	totalKeywordDocPairs := numDocuments * avgKeywordsPerDoc
	lSize := totalKeywordDocPairs * HMACOutputSize // l keys
	cSize := totalKeywordDocPairs * HMACOutputSize // c values (xtag XOR t)

	breakdown["l (CDBXtag key)"] = lSize
	breakdown["c (CDBXtag value)"] = cSize
	cdbxtagSize := lSize + cSize

	totalBytes := cdbtsetSize + cdbxtagSize

	return CDBTSizeResult{
		Scheme:      "FDXT/ODXT/SDSSE-CQ",
		TotalBytes:  totalBytes,
		TotalKB:     float64(totalBytes) / 1024,
		TotalMB:     float64(totalBytes) / (1024 * 1024),
		Breakdown:   breakdown,
		Description: "addr (32B) + val (32B) + alpha (128B) + l (32B) + c (32B) per keyword-document pair",
	}
}

// CalculateMitraSize calculates the storage for Mitra scheme
// Mitra CDB_T: addr + val (no alpha)
// Equivalent to FDXT without alpha
func CalculateMitraSize(numDocuments int, avgKeywordsPerDoc int) CDBTSizeResult {
	totalEntries := numDocuments * avgKeywordsPerDoc

	breakdown := make(map[string]int)

	// CDB_T (MitraCipherList): addr + val
	addrSize := totalEntries * HMACOutputSize // address keys
	valSize := totalEntries * HMACOutputSize  // val values

	breakdown["addr"] = addrSize
	breakdown["val"] = valSize

	totalBytes := addrSize + valSize

	return CDBTSizeResult{
		Scheme:      "Mitra",
		TotalBytes:  totalBytes,
		TotalKB:     float64(totalBytes) / 1024,
		TotalMB:     float64(totalBytes) / (1024 * 1024),
		Breakdown:   breakdown,
		Description: "addr (32B) + val (32B) per keyword-document pair (no alpha)",
	}
}

// CalculateBestieSize calculates the storage for Bestie scheme
// Bestie CDB_T: hash(L||D) + symmetric encryption ciphertext C
// - hash(L||D): λ' + λ + 1 bits (consider as 2*λ + 1 bits)
// - For SHA-256: λ = 256 bits = 32 bytes, so hash = 2*32 = 64 bytes (approximately)
// - C: AES-256-CBC encrypted data = original_size + 16 bytes (IV)
func CalculateBestieSize(numDocuments int, avgKeywordsPerDoc int, avgIDLength int) CDBTSizeResult {
	totalEntries := numDocuments * avgKeywordsPerDoc

	breakdown := make(map[string]int)

	// hash(L||D): 2 * SHA256 = 64 bytes (considering λ' + λ + 1)
	hashSize := totalEntries * 2 * SHA256OutputSize

	breakdown["hash(L||D)"] = hashSize

	// C: AES-256-CBC encrypted ID = ID length + 16 bytes (IV)
	// Assuming average ID length
	encryptedIDSize := totalEntries * (avgIDLength + AESBlockSize)

	breakdown["ciphertext C (encrypted ID + IV)"] = encryptedIDSize

	totalBytes := hashSize + encryptedIDSize

	return CDBTSizeResult{
		Scheme:      "Bestie",
		TotalBytes:  totalBytes,
		TotalKB:     float64(totalBytes) / 1024,
		TotalMB:     float64(totalBytes) / (1024 * 1024),
		Breakdown:   breakdown,
		Description: "hash(L||D) (~64B) + AES-256-CBC ciphertext (~ID_length + 16B) per keyword-document pair",
	}
}

// PrintStorageComparison prints a comparison of storage sizes for all schemes
func PrintStorageComparison(numDocuments int, avgKeywordsPerDoc int, avgIDLength int) {
	fdxtSize := CalculateFDXTODXTSDSSECQSize(numDocuments, avgKeywordsPerDoc)
	mitraSize := CalculateMitraSize(numDocuments, avgKeywordsPerDoc)
	bestieSize := CalculateBestieSize(numDocuments, avgKeywordsPerDoc, avgIDLength)

	schemes := []CDBTSizeResult{fdxtSize, mitraSize, bestieSize}

	fmt.Println("========================================")
	fmt.Println("CDB_T Storage Overhead Comparison")
	fmt.Println("========================================")
	fmt.Printf("Database: %d documents, ~%d keywords/doc\n", numDocuments, avgKeywordsPerDoc)
	fmt.Printf("Total keyword-document pairs: %d\n", numDocuments*avgKeywordsPerDoc)
	fmt.Println("----------------------------------------")

	for _, s := range schemes {
		fmt.Printf("\n%s:\n", s.Scheme)
		fmt.Printf("  Total: %d bytes (%.2f KB, %.2f MB)\n", s.TotalBytes, s.TotalKB, s.TotalMB)
		fmt.Printf("  Breakdown:\n")
		for k, v := range s.Breakdown {
			fmt.Printf("    - %s: %d bytes (%.2f KB)\n", k, v, float64(v)/1024)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("Storage Ratio (relative to FDXT/ODXT/SDSSE-CQ):")
	fmt.Println("----------------------------------------")
	fmt.Printf("FDXT/ODXT/SDSSE-CQ: 1.00x (baseline)\n")
	fmt.Printf("Mitra: %.2fx\n", float64(mitraSize.TotalBytes)/float64(fdxtSize.TotalBytes))
	fmt.Printf("Bestie: %.2fx\n", float64(bestieSize.TotalBytes)/float64(fdxtSize.TotalBytes))
}

// CalculateStorageForDatabase calculates storage for a specific database
// This function takes actual database parameters
func CalculateStorageForDatabase(dbName string, numDocuments int, avgKeywordsPerDoc int, avgIDLength int) map[string]CDBTSizeResult {
	results := make(map[string]CDBTSizeResult)

	results["FDXT"] = CalculateFDXTODXTSDSSECQSize(numDocuments, avgKeywordsPerDoc)
	results["ODXT"] = CalculateFDXTODXTSDSSECQSize(numDocuments, avgKeywordsPerDoc)
	results["SDSSE-CQ"] = CalculateFDXTODXTSDSSECQSize(numDocuments, avgKeywordsPerDoc)
	results["Mitra"] = CalculateMitraSize(numDocuments, avgKeywordsPerDoc)
	results["Bestie"] = CalculateBestieSize(numDocuments, avgKeywordsPerDoc, avgIDLength)

	return results
}
