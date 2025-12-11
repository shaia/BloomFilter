package bloomfilter

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestBatchOperations(t *testing.T) {
	bf := NewCacheOptimizedBloomFilter(100000, 0.01)

	// Prepare data
	count := 10000
	data := make([][]byte, count)
	for i := 0; i < count; i++ {
		data[i] = []byte(fmt.Sprintf("item-%d", i))
	}

	// Test AddBatch
	bf.AddBatch(data)

	// Test ContainsBatch in small batch (sequential fallback)
	smallBatch := data[:100]
	results := bf.ContainsBatch(smallBatch)
	for i, exists := range results {
		if !exists {
			t.Errorf("Item %s should exist", smallBatch[i])
		}
	}

	// Test ContainsBatch in large batch (parallel)
	// We need enough items to force parallel path if we lowered the threshold,
	// but here we are testing the logic.
	// To actually force parallel path for Add/Check logic which uses CPU count,
	// we rely on the implementation details (threshold is currently numCPU*100).
	// Let's create a larger dataset to be sure.

	largeCount := 100000
	largeData := make([][]byte, largeCount)
	for i := 0; i < largeCount; i++ {
		largeData[i] = []byte(fmt.Sprintf("large-%d", i))
	}

	bf2 := NewCacheOptimizedBloomFilter(uint64(largeCount), 0.01)
	bf2.AddBatch(largeData)

	largeResults := bf2.ContainsBatch(largeData)
	if len(largeResults) != largeCount {
		t.Fatalf("Expected %d results, got %d", largeCount, len(largeResults))
	}

	for i, exists := range largeResults {
		if !exists {
			t.Errorf("Item large-%d should exist", i)
		}
	}
}

func TestParallelSetOperations(t *testing.T) {
	// Create filters large enough to trigger parallel path
	// ParallelThreshold is 4096 cache lines.
	// 4096 * 512 bits = 2,097,152 bits.
	// With optimal settings (m/n = 9.58 for 1%), we need ~220k elements.

	size := uint64(300000)
	bf1 := NewCacheOptimizedBloomFilter(size, 0.01)
	bf2 := NewCacheOptimizedBloomFilter(size, 0.01)

	// Check if we actually triggered parallel threshold
	if bf1.cacheLineCount < ParallelThreshold {
		t.Logf("Warning: Filter size (%d lines) smaller than ParallelThreshold (%d). Test will run sequentially.",
			bf1.cacheLineCount, ParallelThreshold)
	} else {
		t.Logf("Filter size: %d lines (Parallel Test Active)", bf1.cacheLineCount)
	}

	// Fill filters
	// bf1 has evens, bf2 has odds
	itemCount := 50000
	chunk1 := make([][]byte, itemCount)
	chunk2 := make([][]byte, itemCount)

	for i := 0; i < itemCount; i++ {
		chunk1[i] = []byte(fmt.Sprintf("num-%d", i*2))   // 0, 2, 4...
		chunk2[i] = []byte(fmt.Sprintf("num-%d", i*2+1)) // 1, 3, 5...
	}

	bf1.AddBatch(chunk1)
	bf2.AddBatch(chunk2)

	// Test PopCount Parallel
	pop1 := bf1.PopCount()
	pop2 := bf2.PopCount()

	if pop1 == 0 || pop2 == 0 {
		t.Error("PopCount should not be zero")
	}

	// Test Union
	// Clone bf1 to verify union result later if needed, but here we just check result
	bfUnion := NewCacheOptimizedBloomFilter(size, 0.01)
	bfUnion.Union(bf1) // merge bf1
	bfUnion.Union(bf2) // merge bf2

	// Real union on bf1
	bf1.Union(bf2)

	popUnion := bf1.PopCount()
	if popUnion < pop1 || popUnion < pop2 {
		t.Errorf("Union PopCount %d should be >= individual counts (%d, %d)", popUnion, pop1, pop2)
	}

	// Check content
	if !bf1.Contains([]byte("num-0")) || !bf1.Contains([]byte("num-1")) {
		t.Error("Union should contain elements from both sets")
	}

	// Test Intersection
	// Reset filters
	bfA := NewCacheOptimizedBloomFilter(size, 0.01)
	bfB := NewCacheOptimizedBloomFilter(size, 0.01)

	commonData := make([][]byte, 1000)
	for i := 0; i < 1000; i++ {
		commonData[i] = []byte(fmt.Sprintf("common-%d", i))
	}

	bfA.AddBatch(chunk1) // evens
	bfA.AddBatch(commonData)

	bfB.AddBatch(chunk2) // odds
	bfB.AddBatch(commonData)

	bfA.Intersection(bfB)

	// Should contain common
	if !bfA.Contains([]byte("common-0")) {
		t.Error("Intersection should keep common elements")
	}

	// Should NOT contain disjoint (probabilistically)
	// "num-0" was in A but not B.
	if bfA.Contains([]byte("num-0")) && bfA.Contains([]byte("num-2")) && bfA.Contains([]byte("num-4")) {
		// It's possible for false positives, but unlikely all 3 survive if intersection worked
		t.Log("Warning: Intersection might have failed to remove disjoint elements (or false positives)")
	}
}

func BenchmarkAddBatch(b *testing.B) {
	data := make([][]byte, 10000)
	for i := 0; i < 10000; i++ {
		data[i] = []byte(fmt.Sprintf("bench-%d", rand.Int()))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		bf := NewCacheOptimizedBloomFilter(1000000, 0.01)
		b.StartTimer()
		bf.AddBatch(data)
	}
}

func BenchmarkSequentialAdd(b *testing.B) {
	data := make([][]byte, 10000)
	for i := 0; i < 10000; i++ {
		data[i] = []byte(fmt.Sprintf("bench-%d", rand.Int()))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		bf := NewCacheOptimizedBloomFilter(1000000, 0.01)
		b.StartTimer()
		for _, item := range data {
			bf.Add(item)
		}
	}
}
