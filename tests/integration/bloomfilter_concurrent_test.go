package bloomfilter_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	bloomfilter "github.com/shaia/BloomFilter"
)

// TestConcurrentReads tests thread-safe concurrent read operations
func TestConcurrentReads(t *testing.T) {
	// Thread-safety fixed with sync.Pool solution

	bf := bloomfilter.NewCacheOptimizedBloomFilter(100_000, 0.01)

	// Pre-populate the filter
	// Scale down for race detector to avoid timeout during setup
	numElements := 10000
	numGoroutines := 100
	numReadsPerGoroutine := 1000
	if testing.Short() {
		// When running with -race, use -short flag to reduce workload
		numElements = 1000
		numGoroutines = 10
		numReadsPerGoroutine = 100
	}

	t.Logf("Pre-populating with %d elements...", numElements)
	for i := 0; i < numElements; i++ {
		bf.AddString(fmt.Sprintf("key_%d", i))
	}

	t.Logf("Testing concurrent reads: %d goroutines × %d reads each", numGoroutines, numReadsPerGoroutine)

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)
	startTime := time.Now()

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < numReadsPerGoroutine; i++ {
				key := fmt.Sprintf("key_%d", i%numElements)
				if !bf.ContainsString(key) {
					errors <- fmt.Errorf("goroutine %d: key not found: %s", goroutineID, key)
					return
				}
			}
		}(g)
	}

	wg.Wait()
	close(errors)

	totalTime := time.Since(startTime)
	totalReads := numGoroutines * numReadsPerGoroutine

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
		if errorCount >= 10 {
			t.Error("Too many errors, stopping error reporting")
			break
		}
	}

	if errorCount == 0 {
		t.Logf("Concurrent reads successful:")
		t.Logf("  Total reads: %d", totalReads)
		t.Logf("  Time: %v", totalTime)
		t.Logf("  Rate: %.0f reads/sec", float64(totalReads)/totalTime.Seconds())
	}
}

// TestConcurrentWrites tests thread-safe concurrent write operations
func TestConcurrentWrites(t *testing.T) {
	// Thread-safety fixed with sync.Pool solution

	bf := bloomfilter.NewCacheOptimizedBloomFilter(100_000, 0.01)

	// Scale down workload for race detector (has 5-10x overhead on sync operations)
	numGoroutines := 50
	numWritesPerGoroutine := 1000
	if testing.Short() {
		// When running with -race, use -short flag to reduce workload
		numGoroutines = 10
		numWritesPerGoroutine = 100
	}

	t.Logf("Testing concurrent writes: %d goroutines × %d writes each", numGoroutines, numWritesPerGoroutine)

	var wg sync.WaitGroup
	startTime := time.Now()

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < numWritesPerGoroutine; i++ {
				key := fmt.Sprintf("g%d_key_%d", goroutineID, i)
				bf.AddString(key)
			}
		}(g)
	}

	wg.Wait()
	totalTime := time.Since(startTime)
	totalWrites := numGoroutines * numWritesPerGoroutine

	t.Logf("Concurrent writes completed:")
	t.Logf("  Total writes: %d", totalWrites)
	t.Logf("  Time: %v", totalTime)
	t.Logf("  Rate: %.0f writes/sec", float64(totalWrites)/totalTime.Seconds())

	// Verify a sample of written keys
	t.Logf("Verifying written keys...")
	sampleSize := 1000
	notFound := 0

	for g := 0; g < numGoroutines && notFound < 10; g++ {
		for i := 0; i < sampleSize/numGoroutines; i++ {
			key := fmt.Sprintf("g%d_key_%d", g, i)
			if !bf.ContainsString(key) {
				notFound++
				if notFound <= 5 {
					t.Errorf("Key not found after concurrent write: %s", key)
				}
			}
		}
	}

	if notFound > 0 {
		t.Errorf("Failed to find %d keys after concurrent writes", notFound)
	} else {
		t.Logf("All sampled keys found successfully")
	}
}

// TestMixedConcurrentOperations tests concurrent reads and writes
func TestMixedConcurrentOperations(t *testing.T) {
	// Thread-safety fixed with sync.Pool solution

	bf := bloomfilter.NewCacheOptimizedBloomFilter(100_000, 0.01)

	// Pre-populate
	// Scale down for race detector to avoid timeout during setup
	numInitialElements := 5000
	numReaders := 25
	numWriters := 25
	opsPerGoroutine := 500
	if testing.Short() {
		// When running with -race, use -short flag to reduce workload
		numInitialElements = 500
		numReaders = 10
		numWriters = 10
		opsPerGoroutine = 50
	}

	t.Logf("Pre-populating with %d elements...", numInitialElements)
	for i := 0; i < numInitialElements; i++ {
		bf.AddString(fmt.Sprintf("initial_%d", i))
	}

	t.Logf("Testing mixed operations: %d readers + %d writers × %d ops each",
		numReaders, numWriters, opsPerGoroutine)

	var wg sync.WaitGroup
	errors := make(chan error, numReaders+numWriters)
	startTime := time.Now()

	// Start readers
	for r := 0; r < numReaders; r++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()

			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("initial_%d", i%numInitialElements)
				if !bf.ContainsString(key) {
					errors <- fmt.Errorf("reader %d: key not found: %s", readerID, key)
					return
				}
			}
		}(r)
	}

	// Start writers
	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()

			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("writer_%d_key_%d", writerID, i)
				bf.AddString(key)
			}
		}(w)
	}

	wg.Wait()
	close(errors)

	totalTime := time.Since(startTime)
	totalOps := (numReaders + numWriters) * opsPerGoroutine

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
		if errorCount >= 10 {
			t.Error("Too many errors, stopping error reporting")
			break
		}
	}

	if errorCount == 0 {
		t.Logf("Mixed operations successful:")
		t.Logf("  Total operations: %d", totalOps)
		t.Logf("  Time: %v", totalTime)
		t.Logf("  Rate: %.0f ops/sec", float64(totalOps)/totalTime.Seconds())
	}
}
