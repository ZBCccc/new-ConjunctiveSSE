package pbc

import (
	"fmt"
	"testing"
)

// TestPBCG1Size measures the actual output size of PBC G1 elements
// This test generates multiple G1 elements and calculates the average size
func TestPBCG1Size(t *testing.T) {
	testCases := []int{100, 1000, 10000}

	for _, n := range testCases {
		totalSize := 0
		sizes := make([]int, n)

		for i := 0; i < n; i++ {
			// Generate a random G1 element
			g1Element := pairing.NewG1().Rand()
			bytes := g1Element.Bytes()
			sizes[i] = len(bytes)
			totalSize += len(bytes)
		}

		avgSize := float64(totalSize) / float64(n)

		// Calculate statistics
		minSize := sizes[0]
		maxSize := sizes[0]
		for _, s := range sizes {
			if s < minSize {
				minSize = s
			}
			if s > maxSize {
				maxSize = s
			}
		}

		fmt.Printf("\n=== PBC G1 Size Test (n=%d) ===\n", n)
		fmt.Printf("Total size: %d bytes\n", totalSize)
		fmt.Printf("Average size: %.2f bytes\n", avgSize)
		fmt.Printf("Min size: %d bytes\n", minSize)
		fmt.Printf("Max size: %d bytes\n", maxSize)
		fmt.Printf("Theoretical size: ~128 bytes (PBC Type A with 32-bit security)\n")
	}
}

// TestPBCZrSize measures the actual output size of PBC Zr elements
func TestPBCZrSize(t *testing.T) {
	testCases := []int{100, 1000, 10000}

	for _, n := range testCases {
		totalSize := 0
		sizes := make([]int, n)

		for i := 0; i < n; i++ {
			// Generate a random Zr element
			zrElement := pairing.NewZr().Rand()
			bytes := zrElement.Bytes()
			sizes[i] = len(bytes)
			totalSize += len(bytes)
		}

		avgSize := float64(totalSize) / float64(n)

		fmt.Printf("\n=== PBC Zr Size Test (n=%d) ===\n", n)
		fmt.Printf("Total size: %d bytes\n", totalSize)
		fmt.Printf("Average size: %.2f bytes\n", avgSize)
		fmt.Printf("Theoretical size: ~32 bytes (PBC Type A)\n")
	}
}

// TestPBCG1SizeConsistency tests if G1 elements have consistent sizes
func TestPBCG1SizeConsistency(t *testing.T) {
	n := 1000
	sizes := make(map[int]int)

	for i := 0; i < n; i++ {
		g1Element := pairing.NewG1().Rand()
		size := len(g1Element.Bytes())
		sizes[size]++
	}

	fmt.Printf("\n=== PBC G1 Size Distribution (n=%d) ===\n", n)
	for size, count := range sizes {
		fmt.Printf("Size %d bytes: %d occurrences (%.2f%%)\n", size, count, float64(count)/float64(n)*100)
	}
}

// BenchmarkPBCG1Size benchmarks G1 element generation and sizing
func BenchmarkPBCG1Size(b *testing.B) {
	b.Run("100", func(b *testing.B) {
		for i := 0; i < 100; i++ {
			g1Element := pairing.NewG1().Rand()
			_ = g1Element.Bytes()
		}
	})

	b.Run("1000", func(b *testing.B) {
		for i := 0; i < 1000; i++ {
			g1Element := pairing.NewG1().Rand()
			_ = g1Element.Bytes()
		}
	})

	b.Run("10000", func(b *testing.B) {
		for i := 0; i < 10000; i++ {
			g1Element := pairing.NewG1().Rand()
			_ = g1Element.Bytes()
		}
	})
}
