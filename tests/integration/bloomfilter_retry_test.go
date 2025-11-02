package bloomfilter_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	bloomfilter "github.com/shaia/BloomFilter"
)

// TestAtomicRetryMechanism validates that the CAS retry loop in setBitsAtomic
// successfully handles contention and ensures all bits are set correctly.
//
// This test creates extreme contention by having many goroutines simultaneously
// write to the same small filter, forcing CAS retries. It then validates that
// all bits were successfully set (no false negatives).
func TestAtomicRetryMechanism(t *testing.T) {
	// Use a small filter to increase collision probability
	// 1000 expected elements with 0.01 FPR = ~9728 bits
	// With 6 hash functions, each insert touches 6 bits
	bf := bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.01)

	const (
		numGoroutines = 100
		insertsPerGoroutine = 100
	)

	// Track successful insertions
	var insertCount atomic.Int64

	// All goroutines insert into the same key space to maximize contention
	var wg sync.WaitGroup
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < insertsPerGoroutine; i++ {
				// Use a small key space to force contention
				// Only 100 unique keys, but 10,000 total insertions
				key := i % 100
				bf.AddUint64(uint64(key))
				insertCount.Add(1)
			}
		}(g)
	}

	wg.Wait()

	totalInserts := insertCount.Load()
	t.Logf("Completed %d concurrent insertions", totalInserts)

	// CRITICAL TEST: Verify no false negatives
	// All 100 unique keys must be found (Bloom filter correctness)
	notFound := 0
	for key := 0; key < 100; key++ {
		if !bf.ContainsUint64(uint64(key)) {
			notFound++
			t.Errorf("Key %d not found after concurrent insertions (FALSE NEGATIVE)", key)
		}
	}

	if notFound > 0 {
		t.Fatalf("CRITICAL: Found %d false negatives - CAS retry mechanism failed!", notFound)
	}

	t.Logf("SUCCESS: All %d unique keys found (no false negatives)", 100)

	// Verify statistics
	stats := bf.GetCacheStats()
	t.Logf("Filter stats: bits_set=%d, load_factor=%.4f, estimated_fpp=%.6f",
		stats.BitsSet, stats.LoadFactor, stats.EstimatedFPP)
}

// TestExtremeContentionSameWord validates that multiple threads writing to
// the exact same bit positions (worst-case contention) still succeed.
func TestExtremeContentionSameWord(t *testing.T) {
	bf := bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.01)

	const (
		numGoroutines = 50
		iterations = 1000
	)

	var wg sync.WaitGroup

	// All goroutines insert the EXACT same key to maximize contention on same bits
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < iterations; i++ {
				// Every goroutine inserts the same key "contention_test"
				bf.AddString("contention_test")
			}
		}()
	}

	wg.Wait()

	// CRITICAL: The key must be found (no false negative)
	if !bf.ContainsString("contention_test") {
		t.Fatal("CRITICAL: Key 'contention_test' not found after extreme contention - CAS retry failed!")
	}

	t.Logf("SUCCESS: Key found after %d concurrent writes to same bit positions", numGoroutines*iterations)
}

// TestCASRetriesEventualSuccess validates that even under artificial contention,
// the retry loop eventually succeeds without hanging.
func TestCASRetriesEventualSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create multiple filters to test retry behavior across different scenarios
	testCases := []struct {
		name              string
		expectedElements  uint64
		fpr               float64
		numGoroutines     int
		keysPerGoroutine  int
	}{
		{"Small filter, high contention", 100, 0.01, 100, 50},
		{"Medium filter, medium contention", 10000, 0.01, 50, 100},
		{"Large filter, low contention", 100000, 0.01, 100, 1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bf := bloomfilter.NewCacheOptimizedBloomFilter(tc.expectedElements, tc.fpr)

			var wg sync.WaitGroup
			uniqueKeys := make(map[uint64]bool)
			var keyMutex sync.Mutex

			// Track all unique keys for verification
			for g := 0; g < tc.numGoroutines; g++ {
				wg.Add(1)
				go func(goroutineID int) {
					defer wg.Done()

					for i := 0; i < tc.keysPerGoroutine; i++ {
						key := uint64(goroutineID*tc.keysPerGoroutine + i)

						keyMutex.Lock()
						uniqueKeys[key] = true
						keyMutex.Unlock()

						bf.AddUint64(key)
					}
				}(g)
			}

			wg.Wait()

			// Verify no false negatives
			notFound := 0
			for key := range uniqueKeys {
				if !bf.ContainsUint64(key) {
					notFound++
					if notFound <= 5 {
						t.Errorf("Key %d not found (FALSE NEGATIVE)", key)
					}
				}
			}

			if notFound > 0 {
				t.Fatalf("Found %d false negatives out of %d keys", notFound, len(uniqueKeys))
			}

			t.Logf("SUCCESS: All %d unique keys found", len(uniqueKeys))
		})
	}
}

// TestNoHangUnderContention validates that the retry loop doesn't hang
// by using a timeout-based approach.
func TestNoHangUnderContention(t *testing.T) {
	done := make(chan bool, 1)

	go func() {
		bf := bloomfilter.NewCacheOptimizedBloomFilter(1000, 0.01)

		const numGoroutines = 100
		var wg sync.WaitGroup

		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for i := 0; i < 100; i++ {
					bf.AddUint64(uint64(i))
				}
			}(g)
		}

		wg.Wait()
		done <- true
	}()

	// Wait for completion or timeout
	// 10 seconds should be more than enough for 10,000 insertions
	// If it takes longer, the retry mechanism has hung
	select {
	case <-done:
		t.Log("SUCCESS: Retry mechanism completed without hanging")
	case <-time.After(10 * time.Second):
		t.Fatal("CRITICAL: Retry mechanism appears to have hung (timeout exceeded)")
	}
}
