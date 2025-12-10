package bloomfilter

import (
	"fmt"
	"math/rand"
	"testing"
	"unsafe"
)

// Helper: Sequential Union implementation (bypassing the Parallel check)
func (bf *CacheOptimizedBloomFilter) unionSequential(other *CacheOptimizedBloomFilter) {
	totalBytes := int(bf.cacheLineCount * CacheLineSize)
	bf.simdOps.VectorOr(
		unsafe.Pointer(&bf.cacheLines[0]),
		unsafe.Pointer(&other.cacheLines[0]),
		totalBytes,
	)
}

// Helper: Sequential Add implementation
func (bf *CacheOptimizedBloomFilter) addBatchSequential(data [][]byte) {
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
		bf := NewCacheOptimizedBloomFilter(size, 0.01)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bf.addBatchSequential(data)
		}
	})

	b.Run("Parallel_AddBatch", func(b *testing.B) {
		bf := NewCacheOptimizedBloomFilter(size, 0.01)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bf.AddBatch(data)
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
		// Clone to prevent dirtying state if union modified anything (Union is idempotent-ish for bitwise OR)
		// But repeating Union on same filter is fine, it just stays set.
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bf1.unionSequential(bf2)
		}
	})

	b.Run("Parallel_Union", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bf1.Union(bf2)
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
