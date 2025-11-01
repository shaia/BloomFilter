package bloomfilter_test

import (
	"math"
	"testing"

	bloomfilter "github.com/shaia/BloomFilter"
)

// TestBoundaryConditions tests exact boundary conditions
func TestBoundaryConditions(t *testing.T) {
	t.Run("Exact ArrayModeThreshold", func(t *testing.T) {
		// Calculate elements that will produce exactly ArrayModeThreshold cache lines
		fpr := 0.01

		// Formula: cacheLines = (bitCount + 511) / 512
		// bitCount = elements * ln(fpr) / (ln(2)^2)
		// We need to find elements such that cacheLines ≈ threshold
		// ArrayModeThreshold is the dividing line between array and map mode

		// Test just below threshold
		bf1 := bloomfilter.NewCacheOptimizedBloomFilter(800_000, fpr)
		stats1 := bf1.GetCacheStats()
		t.Logf("Below threshold: elements=800K, cache_lines=%d, mode=%s",
			stats1.CacheLineCount, func() string {
				if bf1.IsArrayMode() {
					return "ARRAY"
				}
				return "MAP"
			}())

		// Test just above threshold
		bf2 := bloomfilter.NewCacheOptimizedBloomFilter(900_000, fpr)
		stats2 := bf2.GetCacheStats()
		t.Logf("Above threshold: elements=900K, cache_lines=%d, mode=%s",
			stats2.CacheLineCount, func() string {
				if bf2.IsArrayMode() {
					return "ARRAY"
				}
				return "MAP"
			}())

		// Verify both work correctly
		for i := 0; i < 1000; i++ {
			key := string([]byte{byte(i >> 8), byte(i)})
			bf1.Add([]byte(key))
			bf2.Add([]byte(key))
		}

		notFound1, notFound2 := 0, 0
		for i := 0; i < 1000; i++ {
			key := string([]byte{byte(i >> 8), byte(i)})
			if !bf1.Contains([]byte(key)) {
				notFound1++
			}
			if !bf2.Contains([]byte(key)) {
				notFound2++
			}
		}

		if notFound1 > 0 || notFound2 > 0 {
			t.Errorf("Boundary filters failed: below_threshold_missing=%d, above_threshold_missing=%d",
				notFound1, notFound2)
		}
	})

	t.Run("Cache line alignment boundaries", func(t *testing.T) {
		// Test filters that create different cache line counts
		testCases := []uint64{
			1,      // Minimal
			63,     // Just under 1 cache line of data
			64,     // Exactly 1 cache line
			65,     // Just over 1 cache line
			511,    // Just under many cache lines
			512,    // Exactly fills cache lines
			513,    // Just over cache line boundary
			1023,   // Near power of 2
			1024,   // Power of 2
			1025,   // Just over power of 2
		}

		for _, elements := range testCases {
			bf := bloomfilter.NewCacheOptimizedBloomFilter(elements, 0.01)
			stats := bf.GetCacheStats()

			// Add elements
			for i := uint64(0); i < elements; i++ {
				bf.AddUint64(i)
			}

			// Verify all elements
			notFound := 0
			for i := uint64(0); i < elements; i++ {
				if !bf.ContainsUint64(i) {
					notFound++
				}
			}

			if notFound > 0 {
				t.Errorf("Elements=%d: failed to find %d elements, cache_lines=%d",
					elements, notFound, stats.CacheLineCount)
			} else {
				t.Logf("Elements=%d: OK, cache_lines=%d, bits=%d",
					elements, stats.CacheLineCount, stats.BitCount)
			}
		}
	})

	t.Run("Bit and byte alignment", func(t *testing.T) {
		// Test with data sizes that exercise different alignment paths
		dataSizes := []int{
			1, 2, 3, 4, 5, 6, 7, 8,      // Single bytes to uint64
			9, 15, 16, 17,                // Around 16-byte boundary
			31, 32, 33,                   // Around 32-byte boundary (AVX2)
			63, 64, 65,                   // Around 64-byte boundary (cache line)
			127, 128, 129,                // Power of 2 boundaries
		}

		bf := bloomfilter.NewCacheOptimizedBloomFilter(10000, 0.01)

		for _, size := range dataSizes {
			data := make([]byte, size)
			for i := 0; i < size; i++ {
				data[i] = byte(i)
			}

			bf.Add(data)
			if !bf.Contains(data) {
				t.Errorf("Failed to find data of size %d bytes", size)
			}
		}

		t.Logf("Successfully tested %d different data sizes", len(dataSizes))
	})
}

// TestHashDistribution tests quality of hash distribution
func TestHashDistribution(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(10000, 0.01)
	stats := bf.GetCacheStats()
	totalBits := stats.BitCount

	// Add elements and track bit positions
	numElements := 1000
	initialBitsSet := bf.PopCount()

	for i := 0; i < numElements; i++ {
		bf.AddUint64(uint64(i))
	}

	finalBitsSet := bf.PopCount()
	bitsSetByElements := finalBitsSet - initialBitsSet

	// Calculate expected bits set
	// Formula: m * (1 - (1 - 1/m)^(k*n))
	// where m = total bits, k = hash count, n = elements
	m := float64(totalBits)
	k := float64(stats.HashCount)
	n := float64(numElements)

	expectedBitsSet := m * (1 - math.Pow(1-1/m, k*n))
	actualBitsSet := float64(bitsSetByElements)

	// Allow 10% deviation from expected
	deviation := math.Abs(actualBitsSet-expectedBitsSet) / expectedBitsSet

	if deviation > 0.10 {
		t.Errorf("Hash distribution deviation too high: expected=%.0f, actual=%.0f, deviation=%.2f%%",
			expectedBitsSet, actualBitsSet, deviation*100)
	}

	t.Logf("Hash distribution test:")
	t.Logf("  Elements added: %d", numElements)
	t.Logf("  Hash count: %d", stats.HashCount)
	t.Logf("  Expected bits set: %.0f", expectedBitsSet)
	t.Logf("  Actual bits set: %.0f", actualBitsSet)
	t.Logf("  Deviation: %.2f%%", deviation*100)
}

// TestCollisionResistance tests resistance to hash collisions
func TestCollisionResistance(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(10000, 0.01)

	// Test patterns known to cause collisions in poor hash functions
	collisionPronePatterns := [][]byte{
		// Sequential patterns
		{0, 1, 2, 3, 4, 5, 6, 7},
		{1, 2, 3, 4, 5, 6, 7, 8},
		{2, 3, 4, 5, 6, 7, 8, 9},

		// Repeating patterns
		{0xAA, 0xAA, 0xAA, 0xAA},
		{0x55, 0x55, 0x55, 0x55},
		{0xFF, 0xFF, 0xFF, 0xFF},

		// Shifted patterns
		{1, 0, 0, 0, 0, 0, 0, 0},
		{0, 1, 0, 0, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 0, 0, 0},

		// Palindromes
		{1, 2, 3, 4, 4, 3, 2, 1},
		{0xDE, 0xAD, 0xBE, 0xEF, 0xEF, 0xBE, 0xAD, 0xDE},
	}

	// Add all patterns
	for i, pattern := range collisionPronePatterns {
		bf.Add(pattern)
		if !bf.Contains(pattern) {
			t.Errorf("Failed to add collision-prone pattern %d: %v", i, pattern)
		}
	}

	// Verify all patterns are still found
	notFound := 0
	for i, pattern := range collisionPronePatterns {
		if !bf.Contains(pattern) {
			t.Errorf("Failed to find collision-prone pattern %d: %v", i, pattern)
			notFound++
		}
	}

	if notFound == 0 {
		t.Logf("All %d collision-prone patterns handled correctly", len(collisionPronePatterns))
	}
}

// TestExtremeFalsePositiveRates tests filters with extreme FPR settings
func TestExtremeFalsePositiveRates(t *testing.T) {
	tests := []struct {
		name     string
		elements uint64
		fpr      float64
	}{
		{"Very low FPR", 1000, 0.00001},
		{"Low FPR", 1000, 0.0001},
		{"Normal FPR", 1000, 0.01},
		{"High FPR", 1000, 0.1},
		{"Very high FPR", 1000, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := bloomfilter.NewCacheOptimizedBloomFilter(tt.elements, tt.fpr)
			stats := bf.GetCacheStats()

			t.Logf("%s configuration:", tt.name)
			t.Logf("  Target FPR: %.6f", tt.fpr)
			t.Logf("  Bit count: %d", stats.BitCount)
			t.Logf("  Hash count: %d", stats.HashCount)
			t.Logf("  Bits per element: %.2f", float64(stats.BitCount)/float64(tt.elements))

			// Add elements
			for i := uint64(0); i < tt.elements; i++ {
				bf.AddUint64(i)
			}

			// Verify elements
			notFound := 0
			for i := uint64(0); i < tt.elements; i++ {
				if !bf.ContainsUint64(i) {
					notFound++
				}
			}

			if notFound > 0 {
				t.Errorf("Failed to find %d/%d elements", notFound, tt.elements)
			}

			// Measure actual false positive rate
			numTests := 10000
			falsePositives := 0
			for i := tt.elements; i < tt.elements+uint64(numTests); i++ {
				if bf.ContainsUint64(i) {
					falsePositives++
				}
			}

			actualFPR := float64(falsePositives) / float64(numTests)
			t.Logf("  Measured FPR: %.6f", actualFPR)

			// For very low FPR, allow up to 2x target
			// For normal/high FPR, allow up to 3x target
			maxMultiplier := 3.0
			if tt.fpr < 0.001 {
				maxMultiplier = 2.0
			}

			if actualFPR > tt.fpr*maxMultiplier {
				t.Errorf("Actual FPR (%.6f) exceeds %.1fx target (%.6f)",
					actualFPR, maxMultiplier, tt.fpr*maxMultiplier)
			}
		})
	}
}

// TestZeroAndMinimalCases tests edge cases around zero values
func TestZeroAndMinimalCases(t *testing.T) {
	t.Run("Zero uint64", func(t *testing.T) {
		bf := bloomfilter.NewCacheOptimizedBloomFilter(100, 0.01)
		bf.AddUint64(0)
		if !bf.ContainsUint64(0) {
			t.Error("Failed to find uint64(0)")
		}

		// Verify it's different from uint64(1)
		if bf.ContainsUint64(1) {
			t.Log("Note: uint64(1) also found (possible false positive)")
		}
	})

	t.Run("Empty string vs nil slice", func(t *testing.T) {
		bf := bloomfilter.NewCacheOptimizedBloomFilter(100, 0.01)

		bf.AddString("")
		bf.Add([]byte{})
		bf.Add(nil)

		// All should be found (they're all empty)
		if !bf.ContainsString("") {
			t.Error("Failed to find empty string")
		}
		if !bf.Contains([]byte{}) {
			t.Error("Failed to find empty byte slice")
		}
		if !bf.Contains(nil) {
			t.Error("Failed to find nil slice")
		}
	})

	t.Run("Single bit patterns", func(t *testing.T) {
		bf := bloomfilter.NewCacheOptimizedBloomFilter(100, 0.01)

		// Test all single-bit patterns in a byte
		for i := 0; i < 8; i++ {
			pattern := []byte{1 << i}
			bf.Add(pattern)
			if !bf.Contains(pattern) {
				t.Errorf("Failed to find single-bit pattern: 0x%02X", pattern[0])
			}
		}

		t.Log("All single-bit patterns handled correctly")
	})
}

// TestMemoryBehavior tests memory-related edge cases
func TestMemoryBehavior(t *testing.T) {
	t.Run("Multiple clear cycles", func(t *testing.T) {
		bf := bloomfilter.NewCacheOptimizedBloomFilter(10000, 0.01)

		// Run multiple add/clear cycles
		numCycles := 100
		elementsPerCycle := 100

		for cycle := 0; cycle < numCycles; cycle++ {
			// Add elements
			for i := 0; i < elementsPerCycle; i++ {
				bf.AddUint64(uint64(cycle*elementsPerCycle + i))
			}

			// Verify some elements
			if !bf.ContainsUint64(uint64(cycle * elementsPerCycle)) {
				t.Errorf("Cycle %d: element not found", cycle)
			}

			// Clear
			bf.Clear()

			// Verify cleared
			if bf.PopCount() != 0 {
				t.Errorf("Cycle %d: filter not properly cleared, %d bits still set",
					cycle, bf.PopCount())
			}
		}

		t.Logf("Completed %d add/clear cycles successfully", numCycles)
	})

	t.Run("Overload beyond capacity", func(t *testing.T) {
		// Create small filter
		bf := bloomfilter.NewCacheOptimizedBloomFilter(100, 0.01)

		// Add 10x the expected capacity
		numElements := 1000
		for i := 0; i < numElements; i++ {
			bf.AddUint64(uint64(i))
		}

		// All elements should still be found (but FPR will be high)
		notFound := 0
		for i := 0; i < numElements; i++ {
			if !bf.ContainsUint64(uint64(i)) {
				notFound++
			}
		}

		finalStats := bf.GetCacheStats()

		if notFound > 0 {
			t.Errorf("Overloaded filter failed to find %d/%d elements", notFound, numElements)
		}

		t.Logf("Overloaded filter stats:")
		t.Logf("  Capacity: 100 elements")
		t.Logf("  Actual: %d elements", numElements)
		t.Logf("  Load factor: %.2f%%", finalStats.LoadFactor*100)
		t.Logf("  Estimated FPP: %.4f%%", finalStats.EstimatedFPP*100)
		t.Logf("  All elements found: %v", notFound == 0)

		// FPR should be very high
		if finalStats.EstimatedFPP < 0.5 {
			t.Logf("Note: Overloaded filter has lower FPR than expected (%.4f%%)",
				finalStats.EstimatedFPP*100)
		}
	})
}

// TestUnicodeAndSpecialCharacters tests handling of special strings
func TestUnicodeAndSpecialCharacters(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.01)

	specialStrings := []string{
		"Hello, 世界",                    // Chinese
		"Привет, мир",                  // Russian
		"مرحبا بالعالم",                // Arabic
		"שלום עולם",                    // Hebrew
		"こんにちは世界",                    // Japanese
		"🚀🌟💻🔥",                      // Emojis
		"\x00\x01\x02\x03",            // Control characters
		"\n\r\t",                      // Whitespace
		"a\u0000b",                    // Null byte in middle
		string([]byte{0xFF, 0xFE}),   // Invalid UTF-8
	}

	// Add all special strings
	for _, s := range specialStrings {
		bf.AddString(s)
	}

	// Verify all are found
	notFound := 0
	for i, s := range specialStrings {
		if !bf.ContainsString(s) {
			t.Errorf("Failed to find special string %d: %q (bytes: %v)", i, s, []byte(s))
			notFound++
		}
	}

	if notFound == 0 {
		t.Logf("All %d special/unicode strings handled correctly", len(specialStrings))
	}
}
