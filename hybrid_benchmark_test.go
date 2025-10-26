package bloomfilter

import (
	"fmt"
	"testing"
	"unsafe"
)

// Benchmark array mode vs map mode for different filter sizes
func BenchmarkHybridModes(b *testing.B) {
	benchmarks := []struct {
		name     string
		elements uint64
		fpr      float64
	}{
		{"Small_1K_Array", 1_000, 0.01},
		{"Small_10K_Array", 10_000, 0.01},
		{"Medium_100K_Array", 100_000, 0.01},
		{"Large_1M_Map", 1_000_000, 0.01},
		{"Large_10M_Map", 10_000_000, 0.01},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name+"_Add", func(b *testing.B) {
			bf := NewCacheOptimizedBloomFilter(bm.elements, bm.fpr)

			data := make([]byte, 8)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				*(*uint64)(unsafe.Pointer(&data[0])) = uint64(i)
				bf.Add(data)
			}

			b.ReportMetric(float64(bf.cacheLineCount), "cache_lines")
			b.ReportMetric(float64(bf.bitCount)/8/1024/1024, "MB")
			b.SetBytes(8)
		})

		b.Run(bm.name+"_Contains", func(b *testing.B) {
			bf := NewCacheOptimizedBloomFilter(bm.elements, bm.fpr)

			// Pre-populate with some data
			data := make([]byte, 8)
			for i := 0; i < 10000; i++ {
				*(*uint64)(unsafe.Pointer(&data[0])) = uint64(i)
				bf.Add(data)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				*(*uint64)(unsafe.Pointer(&data[0])) = uint64(i % 20000)
				_ = bf.Contains(data)
			}

			b.ReportMetric(float64(bf.cacheLineCount), "cache_lines")
			b.SetBytes(8)

			})
	}
}

// Benchmark memory allocation patterns
func BenchmarkHybridMemoryAllocation(b *testing.B) {
	sizes := []struct {
		name     string
		elements uint64
		fpr      float64
	}{
		{"Array_Small", 1_000, 0.01},
		{"Array_Medium", 100_000, 0.01},
		{"Map_Large", 1_000_000, 0.01},
		{"Map_Huge", 10_000_000, 0.01},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				bf := NewCacheOptimizedBloomFilter(size.elements, size.fpr)
				_ = bf
			}
		})
	}
}

// Benchmark throughput comparison
func BenchmarkHybridThroughput(b *testing.B) {
	configs := []struct {
		name     string
		elements uint64
		fpr      float64
		ops      int
	}{
		{"Array_10K_ops1K", 10_000, 0.01, 1_000},
		{"Array_100K_ops10K", 100_000, 0.01, 10_000},
		{"Map_1M_ops100K", 1_000_000, 0.01, 100_000},
	}

	for _, cfg := range configs {
		b.Run(cfg.name, func(b *testing.B) {
			bf := NewCacheOptimizedBloomFilter(cfg.elements, cfg.fpr)


			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Perform a mix of operations
				for j := 0; j < cfg.ops; j++ {
					if j%2 == 0 {
						bf.AddUint64(uint64(j))
					} else {
						_ = bf.ContainsUint64(uint64(j))
					}
				}
			}

			opsPerSec := float64(b.N*cfg.ops) / b.Elapsed().Seconds()
			b.ReportMetric(opsPerSec/1000000, "Mops/sec")

			})
	}
}

// Benchmark to show the crossover point between array and map efficiency
func BenchmarkHybridCrossoverPoint(b *testing.B) {
	// Test around the threshold (10K cache lines)
	sizes := []uint64{
		1_000_000,   // ~1,873 cache lines - array mode
		3_000_000,   // ~5,619 cache lines - array mode
		5_000_000,   // ~9,365 cache lines - array mode
		5_500_000,   // ~10,302 cache lines - map mode (just over threshold)
		10_000_000,  // ~18,721 cache lines - map mode
		50_000_000,  // ~93,607 cache lines - map mode
	}

	for _, size := range sizes {
		name := fmt.Sprintf("Elements_%dM", size/1_000_000)
		if size < 1_000_000 {
			name = fmt.Sprintf("Elements_%dK", size/1_000)
		}

		b.Run(name, func(b *testing.B) {
			bf := NewCacheOptimizedBloomFilter(size, 0.01)

			mode := "ARRAY"
			if !bf.useArrayMode {
				mode = "MAP"
			}
			b.Logf("Mode: %s, Cache lines: %d, Threshold: %d",
				mode, bf.cacheLineCount, ArrayModeThreshold)

			data := make([]byte, 8)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				*(*uint64)(unsafe.Pointer(&data[0])) = uint64(i)
				bf.Add(data)
				_ = bf.Contains(data)
			}

			b.ReportMetric(float64(bf.cacheLineCount), "cache_lines")
			b.SetBytes(16) // 8 bytes for Add + 8 for Contains
		})
	}
}
