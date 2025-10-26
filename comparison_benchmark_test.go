package bloomfilter_test

import (
	"fmt"
	"testing"
	"unsafe" // For your library's uint64 conversion

	// Your library
	shaia_bf "github.com/shaia/go-simd-bloomfilter"

	// Competitor library
	willf_bf "github.com/willf/bloom"
)

// --- Configuration for Comparison Benchmarks ---
var comparisonBenchmarks = []struct {
	name     string
	elements uint64  // Number of elements to insert
	fpr      float64 // Target False Positive Rate
	ops      int     // Number of operations per b.N iteration (for Add/Contains)
}{
	// Small Filter Test (Likely Array Mode for shaia_bf)
	{"Size_10K_FPR_1%", 10_000, 0.01, 1000},
	// Medium Filter Test (Likely Array Mode for shaia_bf)
	{"Size_100K_FPR_1%", 100_000, 0.01, 1000},
	// Large Filter Test (Likely Map Mode for shaia_bf)
	{"Size_1M_FPR_1%", 1_000_000, 0.01, 1000},
	// Very Large Filter Test (Map Mode for shaia_bf)
	{"Size_10M_FPR_1%", 10_000_000, 0.01, 1000},
	// High Precision Test
	{"Size_1M_FPR_0.1%", 1_000_000, 0.001, 1000},
}

// --- Benchmark Functions ---

func BenchmarkComparisonAdd(b *testing.B) {
	for _, cfg := range comparisonBenchmarks {
		// --- Your Library (shaia_bf) ---
		b.Run(fmt.Sprintf("%s/shaia_bf", cfg.name), func(b *testing.B) {
			bf := shaia_bf.NewCacheOptimizedBloomFilter(cfg.elements, cfg.fpr)
			data := make([]byte, 8) // Use byte slice for Add method
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate adding cfg.ops unique elements
				for j := 0; j < cfg.ops; j++ {
					// Convert counter to byte slice for Add method
					// Ensure unique data for each inner loop iteration
					val := uint64(i*cfg.ops + j)
					*(*uint64)(unsafe.Pointer(&data[0])) = val
					bf.Add(data)
				}
			}
		})

		// --- Competitor Library (willf_bf) ---
		b.Run(fmt.Sprintf("%s/willf_bf", cfg.name), func(b *testing.B) {
			// willf/bloom calculates m and k
			m, k := willf_bf.EstimateParameters(uint(cfg.elements), cfg.fpr)
			bf := willf_bf.New(m, k)
			data := make([]byte, 8) // willf also uses byte slices
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate adding cfg.ops unique elements
				for j := 0; j < cfg.ops; j++ {
					// Ensure unique data for each inner loop iteration
					val := uint64(i*cfg.ops + j)
					*(*uint64)(unsafe.Pointer(&data[0])) = val
					bf.Add(data)
				}
			}
		})
	}
}

func BenchmarkComparisonContains(b *testing.B) {
	for _, cfg := range comparisonBenchmarks {
		// --- Setup Data (Common for both) ---
		var testData [][]byte
		for i := 0; i < cfg.ops; i++ {
			data := make([]byte, 8)
			val := uint64(i) // Use simple sequence for test data
			*(*uint64)(unsafe.Pointer(&data[0])) = val
			testData = append(testData, data)
		}

		// --- Your Library (shaia_bf) ---
		b.Run(fmt.Sprintf("%s/shaia_bf", cfg.name), func(b *testing.B) {
			bf := shaia_bf.NewCacheOptimizedBloomFilter(cfg.elements, cfg.fpr)
			// Pre-fill the filter outside the timer
			for _, data := range testData {
				bf.Add(data)
			}

			b.ReportAllocs()
			b.ResetTimer() // Start timing Contains checks

			for i := 0; i < b.N; i++ {
				// Check cfg.ops elements
				for j := 0; j < cfg.ops; j++ {
					_ = bf.Contains(testData[j])
				}
			}
		})

		// --- Competitor Library (willf_bf) ---
		b.Run(fmt.Sprintf("%s/willf_bf", cfg.name), func(b *testing.B) {
			m, k := willf_bf.EstimateParameters(uint(cfg.elements), cfg.fpr)
			bf := willf_bf.New(m, k)
			// Pre-fill the filter outside the timer
			for _, data := range testData {
				bf.Add(data)
			}

			b.ReportAllocs()
			b.ResetTimer() // Start timing Contains checks

			for i := 0; i < b.N; i++ {
				// Check cfg.ops elements
				for j := 0; j < cfg.ops; j++ {
					_ = bf.Test(testData[j]) // willf uses 'Test' instead of 'Contains'
				}
			}
		})
	}
}