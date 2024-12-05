package utils

import (
	"testing"
)

func TestTexFileRead(t *testing.T) {
	// test file path
	filePath := "../../config/id_keywords/Wiki_USENIX_id_counts.txt"

	// read file
	idCounts, err := texFileRead(filePath)
	if err != nil {
		t.Fatal(err)
	}

	// export [:3] results
	idCounts = idCounts[:3]
	t.Log(idCounts)
}

func TestGenDeletePairs(t *testing.T) {
	// test file path
	filePath := "../../config/id_keywords/Wiki_USENIX_id_counts.txt"

	// read file
	deletePairs := GenDeletePairs(filePath, 100)
	t.Log(deletePairs)
}
