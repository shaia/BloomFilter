package bloomfilter

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// BenchmarkCachePerformance runs a benchmark for cache performance analysis
// Usage: go test -bench=BenchmarkCachePerformance
func BenchmarkCachePerformance(b *testing.B) {
	sizes := []uint64{10000, 100000, 1000000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			bf := NewCacheOptimizedBloomFilter(size, 0.01)
			testData := make([]string, 1000)
			for i := range testData {
				testData[i] = fmt.Sprintf("test_%d", i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Add
				for _, item := range testData {
					bf.AddString(item)
				}
				// Check
				for _, item := range testData {
					bf.ContainsString(item)
				}
				bf.Clear()
			}
			b.StopTimer()

			// Report stats as benchmark output
			stats := bf.GetCacheStats()
			b.ReportMetric(float64(stats.CacheLineCount), "cachelines")
			b.ReportMetric(float64(stats.MemoryUsage)/1024, "KB_mem")
		})
	}
}

// BenchmarkInsertion benchmarks insertion performance
func BenchmarkInsertion(b *testing.B) {
	const numElements = 1000000
	const fpp = 0.01

	bf := NewCacheOptimizedBloomFilter(numElements, fpp)

	// Generate test data
	testData := make([]string, numElements)
	for i := 0; i < numElements; i++ {
		testData[i] = fmt.Sprintf("item_%d_%d", i, rand.Int())
	}

	b.ResetTimer()
	b.Run("String_Insertion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, item := range testData {
				bf.AddString(item)
			}
			bf.Clear()
		}

		// Report performance metrics
		stats := bf.GetCacheStats()
		b.ReportMetric(float64(stats.CacheLineCount), "cachelines")
		b.ReportMetric(float64(stats.MemoryUsage)/(1024*1024), "MB_mem")
		b.ReportMetric(float64(stats.Alignment), "alignment_offset")
	})
}

// BenchmarkLookup benchmarks lookup performance
func BenchmarkLookup(b *testing.B) {
	const numElements = 1000000
	const fpp = 0.01

	bf := NewCacheOptimizedBloomFilter(numElements, fpp)

	// Generate and insert test data
	testData := make([]string, numElements)
	for i := 0; i < numElements; i++ {
		testData[i] = fmt.Sprintf("item_%d_%d", i, rand.Int())
		bf.AddString(testData[i])
	}

	b.ResetTimer()
	b.Run("String_Lookup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			found := 0
			for _, item := range testData {
				if bf.ContainsString(item) {
					found++
				}
			}
			// Verify all items were found
			if found != numElements {
				b.Fatalf("Expected %d items found, got %d", numElements, found)
			}
		}

		// Report final statistics
		finalStats := bf.GetCacheStats()
		b.ReportMetric(finalStats.LoadFactor, "load_factor")
		b.ReportMetric(finalStats.EstimatedFPP*100, "estimated_fpp_percent")
		b.ReportMetric(float64(finalStats.BitsSet), "bits_set")
	})
}

// BenchmarkFalsePositives benchmarks false positive rate testing
func BenchmarkFalsePositives(b *testing.B) {
	const numElements = 1000000
	const fpp = 0.01
	const testNegatives = 100000

	bf := NewCacheOptimizedBloomFilter(numElements, fpp)

	// Insert test data
	testData := make([]string, numElements)
	for i := 0; i < numElements; i++ {
		testData[i] = fmt.Sprintf("item_%d_%d", i, rand.Int())
		bf.AddString(testData[i])
	}

	b.ResetTimer()
	b.Run("False_Positive_Rate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			falsePositives := 0
			for j := 0; j < testNegatives; j++ {
				item := fmt.Sprintf("negative_item_%d", j)
				if bf.ContainsString(item) {
					falsePositives++
				}
			}

			actualFPP := float64(falsePositives) / testNegatives
			// Report false positive rate
			b.ReportMetric(actualFPP*100, "actual_fpp_percent")
			b.ReportMetric(fpp*100, "target_fpp_percent")
		}
	})
}

// BenchmarkComprehensive runs a comprehensive benchmark with detailed reporting
func BenchmarkComprehensive(b *testing.B) {
	const numElements = 1000000
	const fpp = 0.01

	b.Run("Comprehensive_Test", func(b *testing.B) {
		bf := NewCacheOptimizedBloomFilter(numElements, fpp)

		// Display cache line information (reported once)
		stats := bf.GetCacheStats()
		b.Logf("Cache line size: %d bytes", stats.CacheLineSize)
		b.Logf("Cache lines used: %d", stats.CacheLineCount)
		b.Logf("Memory alignment: %d bytes offset", stats.Alignment)
		b.Logf("Total memory: %d bytes (%.2f MB)", stats.MemoryUsage, float64(stats.MemoryUsage)/(1024*1024))

		// Generate test data
		testData := make([]string, numElements)
		for i := 0; i < numElements; i++ {
			testData[i] = fmt.Sprintf("item_%d_%d", i, rand.Int())
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Insertion phase
			insertStart := time.Now()
			for _, item := range testData {
				bf.AddString(item)
			}
			insertTime := time.Since(insertStart)

			// Lookup phase
			lookupStart := time.Now()
			found := 0
			for _, item := range testData {
				if bf.ContainsString(item) {
					found++
				}
			}
			lookupTime := time.Since(lookupStart)

			// False positive test
			falsePositives := 0
			const testNegatives = 10000 // Reduced for benchmark performance
			for j := 0; j < testNegatives; j++ {
				item := fmt.Sprintf("negative_item_%d", j)
				if bf.ContainsString(item) {
					falsePositives++
				}
			}

			// Report metrics
			actualFPP := float64(falsePositives) / testNegatives
			finalStats := bf.GetCacheStats()

			b.ReportMetric(float64(numElements)/insertTime.Seconds(), "insertions_per_sec")
			b.ReportMetric(float64(numElements)/lookupTime.Seconds(), "lookups_per_sec")
			b.ReportMetric(actualFPP*100, "actual_fpp_percent")
			b.ReportMetric(finalStats.LoadFactor, "load_factor")
			b.ReportMetric(finalStats.EstimatedFPP*100, "estimated_fpp_percent")

			bf.Clear() // Reset for next iteration
		}
	})
}
