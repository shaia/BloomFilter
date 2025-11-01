package bloomfilter_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	bloomfilter "github.com/shaia/BloomFilter"
)

// TestLargeDatasetInsertion tests adding millions of keys
func TestLargeDatasetInsertion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	tests := []struct {
		name     string
		elements uint64
		fpr      float64
		addCount int
	}{
		{
			name:     "1 Million elements",
			elements: 1_000_000,
			fpr:      0.01,
			addCount: 1_000_000,
		},
		{
			name:     "5 Million elements",
			elements: 5_000_000,
			fpr:      0.01,
			addCount: 5_000_000,
		},
		{
			name:     "10 Million elements",
			elements: 10_000_000,
			fpr:      0.01,
			addCount: 10_000_000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := bloomfilter.NewCacheOptimizedBloomFilter(tt.elements, tt.fpr)

			startTime := time.Now()
			startMem := getMemStats()

			// Add elements
			t.Logf("Adding %d elements...", tt.addCount)
			for i := 0; i < tt.addCount; i++ {
				key := fmt.Sprintf("key_%d", i)
				bf.AddString(key)

				// Progress indicator for very large datasets
				if i > 0 && i%(tt.addCount/10) == 0 {
					t.Logf("  Progress: %d%% (%d elements)", (i*100)/tt.addCount, i)
				}
			}

			insertTime := time.Since(startTime)
			endMem := getMemStats()

			t.Logf("Insertion complete:")
			t.Logf("  Time: %v", insertTime)
			t.Logf("  Rate: %.0f ops/sec", float64(tt.addCount)/insertTime.Seconds())
			t.Logf("  Memory used: %.2f MB", float64(endMem-startMem)/(1024*1024))

			// Verify a sample of elements
			t.Logf("Verifying sample of elements...")
			sampleSize := 10000
			if sampleSize > tt.addCount {
				sampleSize = tt.addCount
			}

			notFound := 0
			verifyStart := time.Now()

			for i := 0; i < sampleSize; i++ {
				// Sample evenly across the range
				idx := (i * tt.addCount) / sampleSize
				key := fmt.Sprintf("key_%d", idx)
				if !bf.ContainsString(key) {
					notFound++
				}
			}

			verifyTime := time.Since(verifyStart)

			if notFound > 0 {
				t.Errorf("Failed to find %d out of %d sampled elements (%.2f%%)",
					notFound, sampleSize, float64(notFound)*100/float64(sampleSize))
			}

			t.Logf("Verification complete:")
			t.Logf("  Time: %v", verifyTime)
			t.Logf("  Rate: %.0f lookups/sec", float64(sampleSize)/verifyTime.Seconds())
			t.Logf("  Sample size: %d", sampleSize)
			t.Logf("  All samples found: %v", notFound == 0)

			// Check stats
			stats := bf.GetCacheStats()
			t.Logf("Filter stats:")
			t.Logf("  Mode: %s", func() string {
				if bf.IsArrayMode() {
					return "ARRAY"
				}
				return "MAP"
			}())
			t.Logf("  Bits set: %d / %d (%.2f%%)", stats.BitsSet, stats.BitCount, stats.LoadFactor*100)
			t.Logf("  Estimated FPP: %.4f%%", stats.EstimatedFPP*100)
		})
	}
}

// TestLongRunningStability tests filter behavior over extended use
func TestLongRunningStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	bf := bloomfilter.NewCacheOptimizedBloomFilter(100_000, 0.01)

	numCycles := 10
	elementsPerCycle := 10000

	t.Logf("Testing stability over %d cycles of %d elements each", numCycles, elementsPerCycle)

	initialMem := getMemStats()

	for cycle := 0; cycle < numCycles; cycle++ {
		startMem := getMemStats()

		// Add elements
		for i := 0; i < elementsPerCycle; i++ {
			key := fmt.Sprintf("cycle_%d_key_%d", cycle, i)
			bf.AddString(key)
		}

		// Verify elements from this cycle
		notFound := 0
		for i := 0; i < elementsPerCycle; i++ {
			key := fmt.Sprintf("cycle_%d_key_%d", cycle, i)
			if !bf.ContainsString(key) {
				notFound++
			}
		}

		endMem := getMemStats()
		cycleMem := endMem - startMem

		if notFound > 0 {
			t.Errorf("Cycle %d: failed to find %d elements", cycle, notFound)
		}

		t.Logf("Cycle %d: added %d elements, memory delta: %.2f MB",
			cycle, elementsPerCycle, float64(cycleMem)/(1024*1024))
	}

	finalMem := getMemStats()
	totalMemGrowth := finalMem - initialMem

	stats := bf.GetCacheStats()
	t.Logf("Stability test complete:")
	t.Logf("  Total cycles: %d", numCycles)
	t.Logf("  Total elements added: %d", numCycles*elementsPerCycle)
	t.Logf("  Total memory growth: %.2f MB", float64(totalMemGrowth)/(1024*1024))
	t.Logf("  Load factor: %.2f%%", stats.LoadFactor*100)
	t.Logf("  Estimated FPP: %.4f%%", stats.EstimatedFPP*100)
}

// TestExtremeEdgeCases tests unusual input conditions
func TestExtremeEdgeCases(t *testing.T) {
	t.Run("Very small filter", func(t *testing.T) {
		bf := bloomfilter.NewCacheOptimizedBloomFilter(10, 0.01)

		// Add more elements than expected capacity
		for i := 0; i < 100; i++ {
			bf.AddString(fmt.Sprintf("key_%d", i))
		}

		// Verify all elements are found
		notFound := 0
		for i := 0; i < 100; i++ {
			if !bf.ContainsString(fmt.Sprintf("key_%d", i)) {
				notFound++
			}
		}

		if notFound > 0 {
			t.Errorf("Failed to find %d elements in overloaded small filter", notFound)
		}

		stats := bf.GetCacheStats()
		t.Logf("Small filter stats: load=%.2f%%, estimated_fpp=%.4f%%",
			stats.LoadFactor*100, stats.EstimatedFPP*100)
	})

	t.Run("Very long strings", func(t *testing.T) {
		bf := bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.01)

		longStrings := []string{
			string(make([]byte, 1024)),      // 1 KB
			string(make([]byte, 10*1024)),   // 10 KB
			string(make([]byte, 100*1024)),  // 100 KB
		}

		// Fill with unique data
		for i, s := range longStrings {
			data := []byte(s)
			for j := range data {
				data[j] = byte(i + j)
			}
			longStrings[i] = string(data)
		}

		// Add and verify
		for i, s := range longStrings {
			bf.AddString(s)
			if !bf.ContainsString(s) {
				t.Errorf("Failed to find long string %d (len=%d)", i, len(s))
			}
		}

		t.Logf("Successfully handled %d long strings", len(longStrings))
	})

	t.Run("Empty and nil inputs", func(t *testing.T) {
		bf := bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.01)

		// Empty string
		bf.AddString("")
		if !bf.ContainsString("") {
			t.Error("Failed to find empty string")
		}

		// Empty byte slice
		bf.Add([]byte{})
		if !bf.Contains([]byte{}) {
			t.Error("Failed to find empty byte slice")
		}

		// Zero value
		bf.AddUint64(0)
		if !bf.ContainsUint64(0) {
			t.Error("Failed to find uint64 zero value")
		}

		t.Log("Empty and zero value inputs handled correctly")
	})

	t.Run("Extreme FPR values", func(t *testing.T) {
		// Very low FPR
		bf1 := bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.0001)
		stats1 := bf1.GetCacheStats()
		t.Logf("Low FPR filter: bits=%d, hash_count=%d", stats1.BitCount, stats1.HashCount)

		// High FPR (not recommended but should work)
		bf2 := bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.5)
		stats2 := bf2.GetCacheStats()
		t.Logf("High FPR filter: bits=%d, hash_count=%d", stats2.BitCount, stats2.HashCount)

		// Both should work
		bf1.AddString("test")
		bf2.AddString("test")

		if !bf1.ContainsString("test") || !bf2.ContainsString("test") {
			t.Error("Extreme FPR filters failed basic operations")
		}
	})
}

// TestHighThroughputSequential tests sequential high-throughput operations
func TestHighThroughputSequential(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high throughput test in short mode")
	}

	bf := bloomfilter.NewCacheOptimizedBloomFilter(1_000_000, 0.01)

	numOperations := 1_000_000

	// Test insert throughput
	t.Logf("Testing insert throughput...")
	startTime := time.Now()
	for i := 0; i < numOperations; i++ {
		bf.AddUint64(uint64(i))
	}
	insertDuration := time.Since(startTime)
	insertRate := float64(numOperations) / insertDuration.Seconds()

	t.Logf("Insert performance:")
	t.Logf("  Operations: %d", numOperations)
	t.Logf("  Time: %v", insertDuration)
	t.Logf("  Rate: %.0f ops/sec", insertRate)

	// Test lookup throughput
	t.Logf("Testing lookup throughput...")
	startTime = time.Now()
	for i := 0; i < numOperations; i++ {
		_ = bf.ContainsUint64(uint64(i))
	}
	lookupDuration := time.Since(startTime)
	lookupRate := float64(numOperations) / lookupDuration.Seconds()

	t.Logf("Lookup performance:")
	t.Logf("  Operations: %d", numOperations)
	t.Logf("  Time: %v", lookupDuration)
	t.Logf("  Rate: %.0f ops/sec", lookupRate)

	// Lookup should be faster than or similar to insert
	if lookupRate < insertRate*0.5 {
		t.Logf("Warning: Lookup rate (%.0f) is significantly slower than insert rate (%.0f)",
			lookupRate, insertRate)
	}
}

// TestMemoryFootprintGrowth tests memory usage patterns
func TestMemoryFootprintGrowth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory footprint test in short mode")
	}

	tests := []struct {
		name     string
		elements uint64
		fpr      float64
	}{
		{"Small (10K)", 10_000, 0.01},
		{"Medium (100K)", 100_000, 0.01},
		{"Large (1M)", 1_000_000, 0.01},
		{"Very large (10M)", 10_000_000, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime.GC()
			beforeMem := getMemStats()

			bf := bloomfilter.NewCacheOptimizedBloomFilter(tt.elements, tt.fpr)
			stats := bf.GetCacheStats()

			runtime.GC()
			afterMem := getMemStats()

			actualMem := afterMem - beforeMem
			expectedMem := stats.MemoryUsage

			t.Logf("Memory footprint for %d elements:", tt.elements)
			t.Logf("  Expected (from stats): %.2f MB", float64(expectedMem)/(1024*1024))
			t.Logf("  Actual (measured): %.2f MB", float64(actualMem)/(1024*1024))
			t.Logf("  Mode: %s", func() string {
				if bf.IsArrayMode() {
					return "ARRAY"
				}
				return "MAP"
			}())
			t.Logf("  Cache lines: %d", stats.CacheLineCount)

			// Measured memory may include Go runtime overhead
			// Allow some deviation
			if actualMem > expectedMem*2 {
				t.Logf("Warning: Actual memory (%.2f MB) significantly exceeds expected (%.2f MB)",
					float64(actualMem)/(1024*1024), float64(expectedMem)/(1024*1024))
			}
		})
	}
}

// getMemStats returns current memory allocation in bytes
func getMemStats() uint64 {
	runtime.GC() // Force GC to get more accurate reading
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}
