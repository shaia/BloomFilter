package bloomfilter

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

/*
Comprehensive Benchmark Suite for SIMD-Optimized Cache-Line Aligned Bloom Filter

This benchmark suite provides detailed performance analysis across multiple dimensions:

1. BenchmarkCachePerformance: Tests cache efficiency and memory scaling across different sizes
2. BenchmarkInsertion: Measures insertion throughput with memory layout analysis
3. BenchmarkLookup: Measures lookup throughput with load factor and accuracy metrics
4. BenchmarkFalsePositives: Tests statistical accuracy of false positive rates
5. BenchmarkComprehensive: Complete performance profile with throughput and accuracy analysis

Key metrics reported:
- Performance: insertions_per_sec, lookups_per_sec
- Memory: MB_mem, KB_mem, cachelines, alignment_offset
- Accuracy: actual_fpp_percent, estimated_fpp_percent, target_fpp_percent
- Utilization: load_factor, bits_set

Usage: go test -bench=. -benchmem
*/

// BenchmarkCachePerformance runs a benchmark for cache performance analysis across different sizes
// Tests cache line efficiency, memory usage, and performance scaling
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
				// Add operations
				for _, item := range testData {
					bf.AddString(item)
				}
				// Contains operations
				for _, item := range testData {
					bf.ContainsString(item)
				}
				bf.Clear()
			}
			b.StopTimer()

			// Report cache and memory metrics
			stats := bf.GetCacheStats()
			b.ReportMetric(float64(stats.MemoryUsage)/1024, "KB_mem")
			b.ReportMetric(float64(stats.CacheLineCount), "cachelines")
		})
	}
}

// BenchmarkInsertion benchmarks insertion performance with SIMD optimization
// Measures throughput of adding elements to the bloom filter
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

		// Report memory layout and alignment metrics
		stats := bf.GetCacheStats()
		b.ReportMetric(float64(stats.MemoryUsage)/(1024*1024), "MB_mem")
		b.ReportMetric(float64(stats.Alignment), "alignment_offset")
		b.ReportMetric(float64(stats.CacheLineCount), "cachelines")
	})
}

// BenchmarkLookup benchmarks lookup performance with SIMD optimization
// Measures throughput of querying elements in the bloom filter
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

		// Report bloom filter statistics and accuracy metrics
		finalStats := bf.GetCacheStats()
		b.ReportMetric(float64(finalStats.BitsSet), "bits_set")
		b.ReportMetric(finalStats.EstimatedFPP*100, "estimated_fpp_percent")
		b.ReportMetric(finalStats.LoadFactor, "load_factor")
	})
}

// BenchmarkFalsePositives benchmarks false positive rate accuracy
// Tests the statistical accuracy of the bloom filter's false positive rate
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
			// Report false positive rate accuracy metrics
			b.ReportMetric(actualFPP*100, "actual_fpp_percent")
			b.ReportMetric(fpp*100, "target_fpp_percent")
		}
	})
}

// BenchmarkComprehensive runs a comprehensive benchmark with detailed performance analysis
// Measures insertion speed, lookup speed, false positive accuracy, and memory efficiency
// Provides complete performance profile of the SIMD-optimized bloom filter
func BenchmarkComprehensive(b *testing.B) {
	const numElements = 1000000
	const fpp = 0.01

	b.Run("Comprehensive_Test", func(b *testing.B) {
		bf := NewCacheOptimizedBloomFilter(numElements, fpp)

		// Display cache line and memory layout information (reported once)
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
			// Insertion performance measurement
			insertStart := time.Now()
			for _, item := range testData {
				bf.AddString(item)
			}
			insertTime := time.Since(insertStart)

			// Lookup performance measurement
			lookupStart := time.Now()
			found := 0
			for _, item := range testData {
				if bf.ContainsString(item) {
					found++
				}
			}
			lookupTime := time.Since(lookupStart)

			// False positive accuracy test
			falsePositives := 0
			const testNegatives = 10000 // Reduced for benchmark performance
			for j := 0; j < testNegatives; j++ {
				item := fmt.Sprintf("negative_item_%d", j)
				if bf.ContainsString(item) {
					falsePositives++
				}
			}

			// Calculate and report comprehensive performance metrics
			actualFPP := float64(falsePositives) / testNegatives
			finalStats := bf.GetCacheStats()

			// Performance throughput metrics
			b.ReportMetric(actualFPP*100, "actual_fpp_percent")
			b.ReportMetric(finalStats.EstimatedFPP*100, "estimated_fpp_percent")
			b.ReportMetric(float64(numElements)/insertTime.Seconds(), "insertions_per_sec")
			b.ReportMetric(finalStats.LoadFactor, "load_factor")
			b.ReportMetric(float64(numElements)/lookupTime.Seconds(), "lookups_per_sec")

			bf.Clear() // Reset for next iteration
		}
	})
}
