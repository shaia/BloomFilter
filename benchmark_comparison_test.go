package bloomfilter

import (
	"fmt"
	"math/rand"
	"testing"
	"unsafe"
)

// Helper: Sequential Union implementation (bypassing the Parallel check)
// Helper: Sequential Union implementation (bypassing the Parallel check)
func unionSequential(bf, other *CacheOptimizedBloomFilter) {
	totalBytes := int(bf.cacheLineCount * CacheLineSize)
	bf.simdOps.VectorOr(
		unsafe.Pointer(&bf.cacheLines[0]),
		unsafe.Pointer(&other.cacheLines[0]),
		totalBytes,
	)
}

// Helper: Sequential Add implementation
func addBatchSequential(bf *CacheOptimizedBloomFilter, data [][]byte) {
	for _, item := range data {
		bf.Add(item)
	}
}

// -----------------------------------------------------------------------------
// Benchmarks: Add Batch
// -----------------------------------------------------------------------------

func BenchmarkAddBatch_Comparison(b *testing.B) {
	// Size large enough to benefit from parallelism
	size := uint64(1000000)
	batchSize := 50000

	// Pre-generate data
	data := make([][]byte, batchSize)
	for i := 0; i < batchSize; i++ {
		data[i] = []byte(fmt.Sprintf("bench-item-%d", rand.Int()))
	}

	b.Run("Sequential_Loop", func(b *testing.B) {
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			bf := NewCacheOptimizedBloomFilter(size, 0.01)
			b.StartTimer()
			addBatchSequential(bf, data)
			b.StopTimer()
		}
	})

	b.Run("Parallel_AddBatch", func(b *testing.B) {
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			bf := NewCacheOptimizedBloomFilter(size, 0.01)
			b.StartTimer()
			bf.AddBatch(data)
			b.StopTimer()
		}
	})
}

// -----------------------------------------------------------------------------
// Benchmarks: Contains Batch
// -----------------------------------------------------------------------------

func BenchmarkContainsBatch_Comparison(b *testing.B) {
	size := uint64(1000000)
	batchSize := 50000

	data := make([][]byte, batchSize)
	for i := 0; i < batchSize; i++ {
		data[i] = []byte(fmt.Sprintf("bench-item-%d", rand.Int()))
	}

	bf := NewCacheOptimizedBloomFilter(size, 0.01)
	bf.AddBatch(data)

	b.Run("Sequential_Loop", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, item := range data {
				bf.Contains(item)
			}
		}
	})

	b.Run("Parallel_ContainsBatch", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bf.ContainsBatch(data)
		}
	})
}

// -----------------------------------------------------------------------------
// Benchmarks: Union
// -----------------------------------------------------------------------------

func BenchmarkUnion_Comparison(b *testing.B) {
	// Must be larger than ParallelThreshold (4096 lines)
	// 500k elements -> ~10k lines
	size := uint64(1000000)

	bf1 := NewCacheOptimizedBloomFilter(size, 0.01)
	bf2 := NewCacheOptimizedBloomFilter(size, 0.01)

	// Fill them up a bit
	batchSize := 50000
	data1 := make([][]byte, batchSize)
	data2 := make([][]byte, batchSize)
	for i := 0; i < batchSize; i++ {
		data1[i] = []byte(fmt.Sprintf("set1-%d", i))
		data2[i] = []byte(fmt.Sprintf("set2-%d", i))
	}
	bf1.AddBatch(data1)
	bf2.AddBatch(data2)

	b.Run("Sequential_Union", func(b *testing.B) {
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			// Create a fresh copy of bf1 to avoid state accumulation
			// We can't easily clone, so we just recreate it or copy the buffer.
			// Recreating with AddBatch is slow. Copying memory is faster.
			// Let's just create a new one and copy the cacheLines from the prepared bf1.
			bfDest := NewCacheOptimizedBloomFilter(size, 0.01)
			copy(bfDest.cacheLines, bf1.cacheLines)

			b.StartTimer()
			unionSequential(bfDest, bf2)
			b.StopTimer()
		}
	})

	b.Run("Parallel_Union", func(b *testing.B) {
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			bfDest := NewCacheOptimizedBloomFilter(size, 0.01)
			copy(bfDest.cacheLines, bf1.cacheLines)

			b.StartTimer()
			bfDest.Union(bf2)
			b.StopTimer()
		}
	})
}

// -----------------------------------------------------------------------------
// Benchmarks: PopCount
// -----------------------------------------------------------------------------

func BenchmarkPopCount_Comparison(b *testing.B) {
	size := uint64(1000000)
	bf := NewCacheOptimizedBloomFilter(size, 0.01)

	// Fill
	data := make([][]byte, 50000)
	for i := 0; i < 50000; i++ {
		data[i] = []byte(fmt.Sprintf("item-%d", i))
	}
	bf.AddBatch(data)

	b.Run("Sequential_PopCount", func(b *testing.B) {
		totalBytes := int(bf.cacheLineCount * CacheLineSize)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bf.simdOps.PopCount(unsafe.Pointer(&bf.cacheLines[0]), totalBytes)
		}
	})

	b.Run("Parallel_PopCount", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bf.PopCount()
		}
	})
}
