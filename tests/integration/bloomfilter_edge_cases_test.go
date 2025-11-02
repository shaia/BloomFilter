package bloomfilter_test

import (
	"math"
	"testing"

	bloomfilter "github.com/shaia/BloomFilter"
)

// TestBoundaryConditions tests exact boundary conditions
func TestBoundaryConditions(t *testing.T) {
	t.Run("Small Filter", func(t *testing.T) {
		// Test small filter (10K elements)
		bf1 := bloomfilter.NewCacheOptimizedBloomFilter(10_000, 0.01)
		stats1 := bf1.GetCacheStats()
		t.Logf("Small filter: elements=10K, cache_lines=%d, memory=%d bytes",
			stats1.CacheLineCount, stats1.MemoryUsage)

		// Verify works correctly
		for i := 0; i < 1000; i++ {
			key := string([]byte{byte(i >> 8), byte(i)})
			bf1.Add([]byte(key))
		}

		notFound := 0
		for i := 0; i < 1000; i++ {
			key := string([]byte{byte(i >> 8), byte(i)})
			if !bf1.Contains([]byte(key)) {
				notFound++
			}
		}

		if notFound > 0 {
			t.Errorf("Small filter: Expected all 1000 items to be found, got %d missing", notFound)
		}
	})

	t.Run("Large Filter", func(t *testing.T) {
		// Test large filter (1M elements)
		bf2 := bloomfilter.NewCacheOptimizedBloomFilter(1_000_000, 0.01)
		stats2 := bf2.GetCacheStats()
		t.Logf("Large filter: elements=1M, cache_lines=%d, memory=%d bytes",
			stats2.CacheLineCount, stats2.MemoryUsage)

		// Verify works correctly
		for i := 0; i < 1000; i++ {
			key := string([]byte{byte(i >> 8), byte(i)})
			bf2.Add([]byte(key))
		}

		notFound := 0
		for i := 0; i < 1000; i++ {
			key := string([]byte{byte(i >> 8), byte(i)})
			if !bf2.Contains([]byte(key)) {
				notFound++
			}
		}

		if notFound > 0 {
			t.Errorf("Large filter: Expected all 1000 items to be found, got %d missing", notFound)
		}
	})
}

// TestExtremelySmallFilter tests filters with very few expected elements
func TestExtremelySmallFilter(t *testing.T) {
	testCases := []struct {
		name             string
		expectedElements uint64
		falsePositiveRate float64
	}{
		{"Single Element", 1, 0.01},
		{"Ten Elements", 10, 0.01},
		{"Hundred Elements", 100, 0.001},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bf := bloomfilter.NewCacheOptimizedBloomFilter(tc.expectedElements, tc.falsePositiveRate)

			// Add expected number of elements
			for i := uint64(0); i < tc.expectedElements; i++ {
				bf.AddUint64(i)
			}

			// Verify all elements are found
			for i := uint64(0); i < tc.expectedElements; i++ {
				if !bf.ContainsUint64(i) {
					t.Errorf("Element %d should be found but wasn't", i)
				}
			}

			stats := bf.GetCacheStats()
			t.Logf("Filter stats: bits=%d, hash_count=%d, cache_lines=%d",
				stats.BitCount, stats.HashCount, stats.CacheLineCount)
		})
	}
}

// TestExtremeFalsePositiveRates tests very low and very high FPRs
func TestExtremeFalsePositiveRates(t *testing.T) {
	testCases := []struct {
		name string
		fpr  float64
	}{
		{"Very Low FPR", 0.000001},  // 0.0001%
		{"Low FPR", 0.001},          // 0.1%
		{"Medium FPR", 0.01},        // 1%
		{"High FPR", 0.1},           // 10%
	}

	const elements = 10000

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bf := bloomfilter.NewCacheOptimizedBloomFilter(elements, tc.fpr)

			// Add elements
			for i := 0; i < elements/10; i++ {
				bf.AddUint64(uint64(i))
			}

			// Check all added elements are found
			notFound := 0
			for i := 0; i < elements/10; i++ {
				if !bf.ContainsUint64(uint64(i)) {
					notFound++
				}
			}

			if notFound > 0 {
				t.Errorf("FPR %f: Expected all %d items to be found, got %d missing",
					tc.fpr, elements/10, notFound)
			}

			stats := bf.GetCacheStats()
			t.Logf("FPR %f: bits=%d, hash_count=%d, load_factor=%.4f, estimated_fpp=%.6f",
				tc.fpr, stats.BitCount, stats.HashCount, stats.LoadFactor, stats.EstimatedFPP)
		})
	}
}

// TestMaximumHashCount tests filters with very low FPR (high hash count)
func TestMaximumHashCount(t *testing.T) {
	// Very low FPR will result in high hash count
	bf := bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.0000001)

	stats := bf.GetCacheStats()
	t.Logf("Hash count for FPR 0.0000001: %d", stats.HashCount)

	// Verify it still works correctly
	for i := 0; i < 100; i++ {
		bf.AddUint64(uint64(i))
	}

	notFound := 0
	for i := 0; i < 100; i++ {
		if !bf.ContainsUint64(uint64(i)) {
			notFound++
		}
	}

	if notFound > 0 {
		t.Errorf("Expected all 100 items to be found, got %d missing", notFound)
	}
}

// TestZeroAndNegativeInputs tests that invalid inputs panic with clear error messages
func TestZeroAndNegativeInputs(t *testing.T) {
	t.Run("Zero Expected Elements", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for zero expected elements, but didn't panic")
			} else {
				t.Logf("Correctly panicked for zero expected elements: %v", r)
				// Verify panic message is informative
				if msg, ok := r.(string); ok {
					if msg != "bloomfilter: expectedElements must be greater than 0" {
						t.Errorf("Unexpected panic message: %s", msg)
					}
				}
			}
		}()

		bloomfilter.NewCacheOptimizedBloomFilter(0, 0.01)
		t.Error("Should not reach here - expected panic")
	})

	t.Run("Invalid FPR - Too High", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for FPR > 1.0, but didn't panic")
			} else {
				t.Logf("Correctly panicked for FPR > 1.0: %v", r)
			}
		}()

		bloomfilter.NewCacheOptimizedBloomFilter(1000, 1.5)
		t.Error("Should not reach here - expected panic")
	})

	t.Run("Invalid FPR - Negative", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for negative FPR, but didn't panic")
			} else {
				t.Logf("Correctly panicked for negative FPR: %v", r)
			}
		}()

		bloomfilter.NewCacheOptimizedBloomFilter(1000, -0.01)
		t.Error("Should not reach here - expected panic")
	})

	t.Run("Invalid FPR - Zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for zero FPR, but didn't panic")
			} else {
				t.Logf("Correctly panicked for zero FPR: %v", r)
			}
		}()

		bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.0)
		t.Error("Should not reach here - expected panic")
	})

	t.Run("Invalid FPR - Exactly 1.0", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for FPR = 1.0, but didn't panic")
			} else {
				t.Logf("Correctly panicked for FPR = 1.0: %v", r)
			}
		}()

		bloomfilter.NewCacheOptimizedBloomFilter(1000, 1.0)
		t.Error("Should not reach here - expected panic")
	})

	t.Run("NaN FPR", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for NaN FPR, but didn't panic")
			} else {
				t.Logf("Correctly panicked for NaN FPR: %v", r)
				// Verify panic message is informative
				if msg, ok := r.(string); ok {
					if msg != "bloomfilter: falsePositiveRate cannot be NaN" {
						t.Errorf("Unexpected panic message: %s", msg)
					}
				}
			}
		}()

		bloomfilter.NewCacheOptimizedBloomFilter(1000, math.NaN())
		t.Error("Should not reach here - expected panic")
	})
}

// TestEmptyData tests adding and checking empty/nil data
func TestEmptyData(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.01)

	t.Run("Empty Byte Slice", func(t *testing.T) {
		bf.Add([]byte{})
		if !bf.Contains([]byte{}) {
			t.Error("Empty byte slice should be found")
		}
	})

	t.Run("Empty String", func(t *testing.T) {
		bf.AddString("")
		if !bf.ContainsString("") {
			t.Error("Empty string should be found")
		}
	})

	t.Run("Zero Uint64", func(t *testing.T) {
		bf.AddUint64(0)
		if !bf.ContainsUint64(0) {
			t.Error("Zero uint64 should be found")
		}
	})
}

// TestVeryLargeElements tests filters with billions of expected elements
func TestVeryLargeElements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large element test in short mode")
	}

	// Create filter for 10M elements
	bf := bloomfilter.NewCacheOptimizedBloomFilter(10_000_000, 0.01)

	stats := bf.GetCacheStats()
	t.Logf("Large filter stats: bits=%d, cache_lines=%d, memory=%d MB",
		stats.BitCount, stats.CacheLineCount, stats.MemoryUsage/(1024*1024))

	// Add and verify a sample
	const sampleSize = 10000
	for i := 0; i < sampleSize; i++ {
		bf.AddUint64(uint64(i))
	}

	notFound := 0
	for i := 0; i < sampleSize; i++ {
		if !bf.ContainsUint64(uint64(i)) {
			notFound++
		}
	}

	if notFound > 0 {
		t.Errorf("Expected all %d items to be found, got %d missing", sampleSize, notFound)
	}
}
