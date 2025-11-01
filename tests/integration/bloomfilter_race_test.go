// +build race

package bloomfilter_test

import (
	"fmt"
	"sync"
	"testing"

	bloomfilter "github.com/shaia/BloomFilter"
)

// This file contains tests specifically designed to detect race conditions
// Run with: go test -race ./tests/integration

// TestRaceConcurrentAdds tests for races during concurrent additions
func TestRaceConcurrentAdds(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(100_000, 0.01)

	numGoroutines := 100
	addsPerGoroutine := 100

	var wg sync.WaitGroup
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < addsPerGoroutine; i++ {
				bf.AddString(fmt.Sprintf("g%d_k%d", id, i))
			}
		}(g)
	}

	wg.Wait()
	t.Logf("Completed %d concurrent adds", numGoroutines*addsPerGoroutine)
}

// TestRaceConcurrentReads tests for races during concurrent reads
func TestRaceConcurrentReads(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(10_000, 0.01)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		bf.AddString(fmt.Sprintf("key_%d", i))
	}

	numGoroutines := 100
	readsPerGoroutine := 100

	var wg sync.WaitGroup
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < readsPerGoroutine; i++ {
				_ = bf.ContainsString(fmt.Sprintf("key_%d", i%1000))
			}
		}(g)
	}

	wg.Wait()
	t.Logf("Completed %d concurrent reads", numGoroutines*readsPerGoroutine)
}

// TestRaceMixedReadWrite tests for races during mixed read/write operations
func TestRaceMixedReadWrite(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(50_000, 0.01)

	// Pre-populate
	for i := 0; i < 500; i++ {
		bf.AddString(fmt.Sprintf("initial_%d", i))
	}

	numReaders := 50
	numWriters := 50
	opsPerGoroutine := 100

	var wg sync.WaitGroup

	// Start readers
	for r := 0; r < numReaders; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				_ = bf.ContainsString(fmt.Sprintf("initial_%d", i%500))
			}
		}(r)
	}

	// Start writers
	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				bf.AddString(fmt.Sprintf("writer_%d_%d", id, i))
			}
		}(w)
	}

	wg.Wait()
	t.Logf("Completed mixed read/write operations")
}

// TestRacePopCount tests for races during PopCount operations
func TestRacePopCount(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(10_000, 0.01)

	// Pre-populate
	for i := 0; i < 100; i++ {
		bf.AddString(fmt.Sprintf("key_%d", i))
	}

	numGoroutines := 50
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = bf.PopCount()
			}
		}()
	}

	wg.Wait()
	t.Log("Completed concurrent PopCount operations")
}

// TestRaceClear tests for races during Clear operations
func TestRaceClear(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(10_000, 0.01)

	var wg sync.WaitGroup
	numCycles := 50

	for cycle := 0; cycle < numCycles; cycle++ {
		// Add data
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				bf.AddString(fmt.Sprintf("cycle_%d_key_%d", c, i))
			}
		}(cycle)

		// Clear occasionally
		if cycle%10 == 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				bf.Clear()
			}()
		}
	}

	wg.Wait()
	t.Log("Completed concurrent add/clear operations")
}

// TestRaceUnion tests for races during Union operations
func TestRaceUnion(t *testing.T) {
	bf1 := bloomfilter.NewCacheOptimizedBloomFilter(10_000, 0.01)
	bf2 := bloomfilter.NewCacheOptimizedBloomFilter(10_000, 0.01)

	// Pre-populate both filters
	for i := 0; i < 100; i++ {
		bf1.AddString(fmt.Sprintf("bf1_%d", i))
		bf2.AddString(fmt.Sprintf("bf2_%d", i))
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent reads from bf1 while doing union
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = bf1.ContainsString(fmt.Sprintf("bf1_%d", i%100))
			}
		}()
	}

	// Perform union
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = bf1.Union(bf2)
	}()

	wg.Wait()
	t.Log("Completed concurrent union operations")
}

// TestRaceIntersection tests for races during Intersection operations
func TestRaceIntersection(t *testing.T) {
	bf1 := bloomfilter.NewCacheOptimizedBloomFilter(10_000, 0.01)
	bf2 := bloomfilter.NewCacheOptimizedBloomFilter(10_000, 0.01)

	// Pre-populate both filters with overlapping data
	for i := 0; i < 100; i++ {
		bf1.AddString(fmt.Sprintf("key_%d", i))
		if i%2 == 0 {
			bf2.AddString(fmt.Sprintf("key_%d", i))
		}
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent reads from bf1 while doing intersection
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = bf1.ContainsString(fmt.Sprintf("key_%d", i%100))
			}
		}()
	}

	// Perform intersection
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = bf1.Intersection(bf2)
	}()

	wg.Wait()
	t.Log("Completed concurrent intersection operations")
}

// TestRaceGetCacheStats tests for races when reading stats
func TestRaceGetCacheStats(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(10_000, 0.01)

	var wg sync.WaitGroup
	numReaders := 50
	numWriters := 10

	// Concurrent stats readers
	for r := 0; r < numReaders; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = bf.GetCacheStats()
			}
		}()
	}

	// Concurrent writers (modifying the filter)
	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				bf.AddString(fmt.Sprintf("writer_%d_%d", id, i))
			}
		}(w)
	}

	wg.Wait()
	t.Log("Completed concurrent GetCacheStats operations")
}

// TestRaceMultipleOperations tests various operations happening concurrently
func TestRaceMultipleOperations(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(100_000, 0.01)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		bf.AddString(fmt.Sprintf("init_%d", i))
	}

	var wg sync.WaitGroup
	duration := 100 // operations per goroutine

	// Adders
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < duration; j++ {
				bf.AddString(fmt.Sprintf("add_%d_%d", id, j))
			}
		}(i)
	}

	// Readers
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < duration; j++ {
				_ = bf.ContainsString(fmt.Sprintf("init_%d", j%1000))
			}
		}(i)
	}

	// PopCount callers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < duration; j++ {
				_ = bf.PopCount()
			}
		}()
	}

	// Stats readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < duration; j++ {
				_ = bf.GetCacheStats()
			}
		}()
	}

	wg.Wait()
	t.Log("Completed multiple concurrent operations")
}

// TestRaceArrayVsMapMode tests races in both storage modes
func TestRaceArrayVsMapMode(t *testing.T) {
	tests := []struct {
		name     string
		elements uint64
		mode     string
	}{
		{"Array mode", 10_000, "ARRAY"},
		{"Map mode", 1_000_000, "MAP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := bloomfilter.NewCacheOptimizedBloomFilter(tt.elements, 0.01)

			var wg sync.WaitGroup
			numGoroutines := 50
			opsPerGoroutine := 100

			for g := 0; g < numGoroutines; g++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for i := 0; i < opsPerGoroutine; i++ {
						key := fmt.Sprintf("g%d_k%d", id, i)
						bf.AddString(key)
						_ = bf.ContainsString(key)
					}
				}(g)
			}

			wg.Wait()
			t.Logf("Completed race test for %s", tt.mode)
		})
	}
}
