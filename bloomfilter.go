package bloomfilter

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/shaia/BloomFilter/internal/hash"
	"github.com/shaia/BloomFilter/internal/simd"
)

// CacheOptimizedBloomFilter uses cache line aligned storage with SIMD optimization and atomic operations for thread-safety.
type CacheOptimizedBloomFilter struct {
	// Cache line aligned bitset
	cacheLines     []CacheLine
	bitCount       uint64
	hashCount      uint32
	cacheLineCount uint64

	// SIMD operations instance (initialized once for performance)
	simdOps simd.Operations
}

// CacheStats provides detailed statistics about the bloom filter
type CacheStats struct {
	BitCount       uint64
	HashCount      uint32
	BitsSet        uint64
	LoadFactor     float64
	EstimatedFPP   float64
	CacheLineCount uint64
	CacheLineSize  int
	MemoryUsage    uint64
	Alignment      uintptr
	// SIMD capability information
	HasAVX2     bool
	HasAVX512   bool
	HasNEON     bool
	SIMDEnabled bool
}

// NewCacheOptimizedBloomFilter creates a cache line optimized bloom filter.
// Uses SIMD-accelerated operations and lock-free atomic operations for thread-safety.
// Achieves zero allocations for typical use cases (hashCount ≤ 16, which covers 99% of scenarios).
//
// Panics if:
//   - expectedElements is 0
//   - falsePositiveRate is <= 0, >= 1.0, or NaN
func NewCacheOptimizedBloomFilter(expectedElements uint64, falsePositiveRate float64) *CacheOptimizedBloomFilter {
	// Validate inputs
	if expectedElements == 0 {
		panic("bloomfilter: expectedElements must be greater than 0")
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1.0 {
		panic(fmt.Sprintf("bloomfilter: falsePositiveRate must be in range (0, 1), got %f", falsePositiveRate))
	}
	if math.IsNaN(falsePositiveRate) {
		panic("bloomfilter: falsePositiveRate cannot be NaN")
	}

	// Calculate optimal parameters
	ln2 := math.Ln2
	bitCount := uint64(-float64(expectedElements) * math.Log(falsePositiveRate) / (ln2 * ln2))
	hashCount := uint32(float64(bitCount) * ln2 / float64(expectedElements))

	// Validate calculated parameters
	if bitCount == 0 {
		panic(fmt.Sprintf("bloomfilter: falsePositiveRate too high (%f) for %d elements, results in zero bits", falsePositiveRate, expectedElements))
	}

	if hashCount < 1 {
		hashCount = 1
	}

	// Align to cache line boundaries (512 bits per cache line)
	cacheLineCount := (bitCount + BitsPerCacheLine - 1) / BitsPerCacheLine
	if cacheLineCount == 0 {
		cacheLineCount = 1 // Ensure at least one cache line
	}
	bitCount = cacheLineCount * BitsPerCacheLine

	// Allocate cache line aligned memory
	cacheLines := make([]CacheLine, cacheLineCount)

	// Verify alignment
	if uintptr(unsafe.Pointer(&cacheLines[0]))%CacheLineSize != 0 {
		// Force alignment by creating a larger slice and finding aligned offset
		oversized := make([]byte, int(cacheLineCount)*CacheLineSize+CacheLineSize)
		alignedPtr := (uintptr(unsafe.Pointer(&oversized[0])) + CacheLineSize - 1) &^ (CacheLineSize - 1)
		cacheLines = *(*[]CacheLine)(unsafe.Pointer(&struct {
			ptr uintptr
			len int
			cap int
		}{alignedPtr, int(cacheLineCount), int(cacheLineCount)}))
	}

	bf := &CacheOptimizedBloomFilter{
		cacheLines:     cacheLines,
		bitCount:       bitCount,
		hashCount:      hashCount,
		cacheLineCount: cacheLineCount,
		simdOps:        simd.Get(), // Initialize SIMD operations once
	}

	return bf
}

// Clone creates a deep copy of the bloom filter
func (bf *CacheOptimizedBloomFilter) Clone() *CacheOptimizedBloomFilter {
	newBf := *bf
	newBf.cacheLines = make([]CacheLine, len(bf.cacheLines))
	copy(newBf.cacheLines, bf.cacheLines)
	return &newBf
}

// Add adds an element with cache line optimization
func (bf *CacheOptimizedBloomFilter) Add(data []byte) {
	// Stack buffer for typical filters
	var stackBuf [16]uint64
	var positions []uint64
	if bf.hashCount <= 16 {
		positions = stackBuf[:bf.hashCount]
	} else {
		positions = make([]uint64, bf.hashCount)
	}

	// Generate positions
	bf.calculatePositions(data, positions)

	// Set bits atomically
	bf.setBitsAtomic(positions)
}

// Contains checks membership with cache line optimization
func (bf *CacheOptimizedBloomFilter) Contains(data []byte) bool {
	var stackBuf [16]uint64
	var positions []uint64
	if bf.hashCount <= 16 {
		positions = stackBuf[:bf.hashCount]
	} else {
		positions = make([]uint64, bf.hashCount)
	}

	bf.calculatePositions(data, positions)

	return bf.checkBitsAtomic(positions)
}

// AddBatch adds multiple elements concurrently using available CPU cores.
//
// Thread Safety: This method is thread-safe and uses atomic operations (CAS) to set bits.
// It can be safely called concurrently from multiple goroutines, although the internal
// Parallelization itself spawns goroutines.
//
// Parallel Execution: Parallel processing is triggered only if the batch size exceeds
// (runtime.NumCPU() * MinBatchSizePerCPU). For smaller batches, it falls back to
// sequential addition to avoid the overhead of spawning goroutines.
//
// Performance: checking benchmarks, for large batches (e.g., 50k items), this method
// can be 2x or more faster than a sequential loop, depending on the number of cores.
func (bf *CacheOptimizedBloomFilter) AddBatch(data [][]byte) {
	if len(data) == 0 {
		return
	}

	numCPU := runtime.NumCPU()
	if len(data) < numCPU*MinBatchSizePerCPU { // fallback for small batches
		for _, item := range data {
			bf.Add(item)
		}
		return
	}

	bf.executeParallel(uint64(len(data)), func(start, end uint64) {
		items := data[start:end]

		// Thread-local buffers to avoid allocation in loop
		// We can't reuse bf.Add because it allocates internally
		// So we duplicate the add logic here for performance

		hCount := bf.hashCount
		var stackBuf [16]uint64
		var positions []uint64

		if hCount <= 16 {
			positions = stackBuf[:hCount]
		} else {
			positions = make([]uint64, hCount)
		}

		for _, item := range items {
			bf.calculatePositions(item, positions)
			bf.setBitsAtomic(positions)
		}
	})
}

// ContainsBatch checks multiple elements concurrently.
//
// Returns: A slice of booleans where each boolean corresponds to the element at the same index
// in the input slice. true indicates the element might be in the set, false indicates it definitely is not.
//
// Thread Safety: This method is thread-safe (read-only operations on the bitset).
//
// Parallel Execution: Parallel processing is triggered only if the batch size exceeds
// (runtime.NumCPU() * MinBatchSizePerCPU). For smaller batches, it falls back to
// sequential checks.
//
// Performance: Optimizes throughput for large batches by utilizing multiple cores.
func (bf *CacheOptimizedBloomFilter) ContainsBatch(data [][]byte) []bool {
	if len(data) == 0 {
		return nil
	}

	results := make([]bool, len(data))
	numCPU := runtime.NumCPU()

	if len(data) < numCPU*MinBatchSizePerCPU { // fallback for small batches
		for i, item := range data {
			results[i] = bf.Contains(item)
		}
		return results
	}

	bf.executeParallel(uint64(len(data)), func(start, end uint64) {
		chunkIdx := int(start)
		items := data[start:end]

		hCount := bf.hashCount
		var stackBuf [16]uint64
		var positions []uint64

		if hCount <= 16 {
			positions = stackBuf[:hCount]
		} else {
			positions = make([]uint64, hCount)
		}

		for j, item := range items {
			bf.calculatePositions(item, positions)

			// Calculate global index for results
			results[chunkIdx+j] = bf.checkBitsAtomic(positions)
		}
	})

	return results
}

// AddString adds a string element to the bloom filter
func (bf *CacheOptimizedBloomFilter) AddString(s string) {
	data := *(*[]byte)(unsafe.Pointer(&struct {
		string
		int
	}{s, len(s)}))
	bf.Add(data)
}

// ContainsString checks if a string element exists in the bloom filter
func (bf *CacheOptimizedBloomFilter) ContainsString(s string) bool {
	data := *(*[]byte)(unsafe.Pointer(&struct {
		string
		int
	}{s, len(s)}))
	return bf.Contains(data)
}

// AddUint64 adds a uint64 element to the bloom filter
func (bf *CacheOptimizedBloomFilter) AddUint64(n uint64) {
	data := (*[8]byte)(unsafe.Pointer(&n))[:]
	bf.Add(data)
}

// ContainsUint64 checks if a uint64 element exists in the bloom filter
func (bf *CacheOptimizedBloomFilter) ContainsUint64(n uint64) bool {
	data := (*[8]byte)(unsafe.Pointer(&n))[:]
	return bf.Contains(data)
}

// Clear resets the bloom filter using vectorized operations with automatic fallback
func (bf *CacheOptimizedBloomFilter) Clear() {
	if bf.cacheLineCount == 0 {
		return
	}

	// Calculate total data size in bytes
	totalBytes := int(bf.cacheLineCount * CacheLineSize)

	// Use the pre-initialized SIMD operations for vectorized clear operation
	bf.simdOps.VectorClear(unsafe.Pointer(&bf.cacheLines[0]), totalBytes)
}

// Union performs vectorized union operation with automatic fallback to optimized scalar
func (bf *CacheOptimizedBloomFilter) Union(other *CacheOptimizedBloomFilter) error {
	if bf.cacheLineCount != other.cacheLineCount {
		return fmt.Errorf("bloom filters must have same size for union")
	}

	if bf.cacheLineCount == 0 {
		return nil
	}

	// Parallel execution for large filters
	if bf.cacheLineCount >= ParallelThreshold {
		bf.executeParallel(bf.cacheLineCount, func(start, end uint64) {
			startIdx := int(start)
			endIdx := int(end)
			chunkBytes := (endIdx - startIdx) * CacheLineSize
			bf.simdOps.VectorOr(
				unsafe.Pointer(&bf.cacheLines[startIdx]),
				unsafe.Pointer(&other.cacheLines[startIdx]),
				chunkBytes,
			)
		})
		return nil
	}

	// Calculate total data size in bytes
	totalBytes := int(bf.cacheLineCount * CacheLineSize)

	// Use the pre-initialized SIMD operations for vectorized OR operation
	bf.simdOps.VectorOr(
		unsafe.Pointer(&bf.cacheLines[0]),
		unsafe.Pointer(&other.cacheLines[0]),
		totalBytes,
	)

	return nil
}

// Intersection performs vectorized intersection operation with automatic fallback to optimized scalar
func (bf *CacheOptimizedBloomFilter) Intersection(other *CacheOptimizedBloomFilter) error {
	if bf.cacheLineCount != other.cacheLineCount {
		return fmt.Errorf("bloom filters must have same size for intersection")
	}

	if bf.cacheLineCount == 0 {
		return nil
	}

	// Parallel execution for large filters
	if bf.cacheLineCount >= ParallelThreshold {
		bf.executeParallel(bf.cacheLineCount, func(start, end uint64) {
			startIdx := int(start)
			endIdx := int(end)
			chunkBytes := (endIdx - startIdx) * CacheLineSize
			bf.simdOps.VectorAnd(
				unsafe.Pointer(&bf.cacheLines[startIdx]),
				unsafe.Pointer(&other.cacheLines[startIdx]),
				chunkBytes,
			)
		})
		return nil
	}

	// Calculate total data size in bytes
	totalBytes := int(bf.cacheLineCount * CacheLineSize)

	// Use the pre-initialized SIMD operations for vectorized AND operation
	bf.simdOps.VectorAnd(
		unsafe.Pointer(&bf.cacheLines[0]),
		unsafe.Pointer(&other.cacheLines[0]),
		totalBytes,
	)

	return nil
}

// PopCount uses vectorized bit counting with automatic fallback to optimized scalar
func (bf *CacheOptimizedBloomFilter) PopCount() uint64 {
	if bf.cacheLineCount == 0 {
		return 0
	}

	// Parallel execution for large filters
	if bf.cacheLineCount >= ParallelThreshold {
		var totalCount uint64
		bf.executeParallel(bf.cacheLineCount, func(start, end uint64) {
			startIdx := int(start)
			endIdx := int(end)
			chunkBytes := (endIdx - startIdx) * CacheLineSize
			c := bf.simdOps.PopCount(unsafe.Pointer(&bf.cacheLines[startIdx]), chunkBytes)
			atomic.AddUint64(&totalCount, uint64(c))
		})
		return totalCount
	}

	// Calculate total data size in bytes
	totalBytes := int(bf.cacheLineCount * CacheLineSize)

	// Use the pre-initialized SIMD operations for vectorized population count
	count := bf.simdOps.PopCount(unsafe.Pointer(&bf.cacheLines[0]), totalBytes)

	return uint64(count)
}

// EstimatedFPP calculates the estimated false positive probability
func (bf *CacheOptimizedBloomFilter) EstimatedFPP() float64 {
	bitsSet := float64(bf.PopCount())
	ratio := bitsSet / float64(bf.bitCount)
	return math.Pow(ratio, float64(bf.hashCount))
}

// GetCacheStats returns detailed statistics about the bloom filter
func (bf *CacheOptimizedBloomFilter) GetCacheStats() CacheStats {
	bitsSet := bf.PopCount()
	alignment := uintptr(unsafe.Pointer(&bf.cacheLines[0])) % CacheLineSize

	return CacheStats{
		BitCount:       bf.bitCount,
		HashCount:      bf.hashCount,
		BitsSet:        bitsSet,
		LoadFactor:     float64(bitsSet) / float64(bf.bitCount),
		EstimatedFPP:   bf.EstimatedFPP(),
		CacheLineCount: bf.cacheLineCount,
		CacheLineSize:  CacheLineSize,
		MemoryUsage:    bf.cacheLineCount * CacheLineSize,
		Alignment:      alignment,
		// SIMD capability information
		HasAVX2:     simd.HasAVX2(),
		HasAVX512:   simd.HasAVX512(),
		HasNEON:     simd.HasNEON(),
		SIMDEnabled: simd.HasAny(),
	}
}

// HasAVX2 returns true if AVX2 SIMD instructions are available
func HasAVX2() bool {
	return simd.HasAVX2()
}

// HasAVX512 returns true if AVX512 SIMD instructions are available
func HasAVX512() bool {
	return simd.HasAVX512()
}

// HasNEON returns true if NEON SIMD instructions are available
func HasNEON() bool {
	return simd.HasNEON()
}

// HasSIMD returns true if any SIMD instructions are available
func HasSIMD() bool {
	return simd.HasAny()
}

const (
	// Cache line size for most modern CPUs (Intel, AMD, ARM)
	CacheLineSize = 64
	// Number of uint64 words per cache line
	WordsPerCacheLine = CacheLineSize / 8 // 8 words per 64-byte cache line
	// Bits per cache line
	BitsPerCacheLine = CacheLineSize * 8 // 512 bits per cache line

	// SIMD vector sizes
	AVX2VectorSize   = 32 // 256-bit vectors = 32 bytes = 4 uint64
	AVX512VectorSize = 64 // 512-bit vectors = 64 bytes = 8 uint64
	NEONVectorSize   = 16 // 128-bit vectors = 16 bytes = 2 uint64

	// Threshold for choosing between array and map mode
	// Arrays: ≤10K cache lines = ~5MB bloom filter = efficient for small/medium filters
	// Maps: >10K cache lines = scalable for large filters (up to billions of elements)
	// Memory overhead: Array mode ~240KB fixed, Map mode grows dynamically
	ArrayModeThreshold = 10000

	// ParallelThreshold defines the minimum number of cache lines to trigger parallel processing
	// for bulk operations.
	// 4096 cache lines * 64 bytes = 262,144 bytes = 256 KiB
	ParallelThreshold = 4096

	// MinBatchSizePerCPU defines the minimum number of items per CPU core
	// required to trigger parallel processing in batch operations.
	// This avoids overhead of starting goroutines for small batches.
	MinBatchSizePerCPU = 100
)

// CacheLine represents a single 64-byte cache line containing 8 uint64 words
type CacheLine struct {
	words [WordsPerCacheLine]uint64
}

// setBitsAtomic sets multiple bits atomically using lock-free CAS operations.
//
// CORRECTNESS GUARANTEE: This function MUST successfully set all bits to maintain
// Bloom filter correctness. Bloom filters can have false positives but NEVER false
// negatives. Failing to set a bit would introduce false negatives, breaking the
// data structure's mathematical guarantees.
//
// RETRY STRATEGY: Uses unlimited retries with CAS. Under extreme contention (hundreds
// of concurrent writers targeting the same word), this could theoretically spin for
// a while, but:
//   - Each CAS operation is extremely fast (~1-10ns)
//   - The probability of 100+ consecutive failures is astronomically low
//   - The alternative (giving up) would corrupt the Bloom filter
//
// CONTENTION ANALYSIS: With 512 bits per cache line and typical hash distributions,
// the probability of multiple threads colliding on the same 64-bit word is very low.
// Even with 100 concurrent writers, most CAS operations succeed on the first try.
//
// PERFORMANCE: Benchmarks show this approach achieves 14M+ writes/sec with 50
// concurrent goroutines without any backoff mechanism, indicating that contention
// is naturally low due to the large bit array size.
func (bf *CacheOptimizedBloomFilter) setBitsAtomic(positions []uint64) {
	for _, bitPos := range positions {
		cacheLineIdx := bitPos / BitsPerCacheLine
		wordIdx := (bitPos % BitsPerCacheLine) / 64
		bitOffset := bitPos % 64

		mask := uint64(1 << bitOffset)
		wordPtr := &bf.cacheLines[cacheLineIdx].words[wordIdx]

		// Retry indefinitely until successful. This is safe because:
		// 1. CAS is lock-free and will eventually succeed
		// 2. If the bit is already set (old == new), we exit immediately
		// 3. Bloom filter correctness requires all bits to be set
		for {
			old := atomic.LoadUint64(wordPtr)
			new := old | mask

			// Fast path: bit already set, no need to CAS
			if old == new {
				break
			}

			// Attempt to set the bit
			if atomic.CompareAndSwapUint64(wordPtr, old, new) {
				break
			}

			// CAS failed, retry (another thread modified the word)
			// No backoff needed - natural hash distribution provides low contention
		}
	}
}

func (bf *CacheOptimizedBloomFilter) checkBitsAtomic(positions []uint64) bool {
	for _, bitPos := range positions {
		cacheLineIdx := bitPos / BitsPerCacheLine
		wordIdx := (bitPos % BitsPerCacheLine) / 64
		bitOffset := bitPos % 64

		word := atomic.LoadUint64(&bf.cacheLines[cacheLineIdx].words[wordIdx])
		if (word & (1 << bitOffset)) == 0 {
			return false
		}
	}
	return true
}

// executeParallel executes a task concurrently across multiple goroutines.
// It divides the work into chunks based on the number of available CPUs.
func (bf *CacheOptimizedBloomFilter) executeParallel(totalItems uint64, task func(start, end uint64)) {
	numCPU := runtime.NumCPU()
	var wg sync.WaitGroup
	chunkSize := (totalItems + uint64(numCPU) - 1) / uint64(numCPU)

	for i := 0; i < numCPU; i++ {
		start := uint64(i) * chunkSize
		end := start + chunkSize
		if start >= totalItems {
			break
		}
		if end > totalItems {
			end = totalItems
		}

		wg.Add(1)
		go func(s, e uint64) {
			defer wg.Done()
			task(s, e)
		}(start, end)
	}
	wg.Wait()
}

// calculatePositions computes the bit positions for a given item
func (bf *CacheOptimizedBloomFilter) calculatePositions(data []byte, positions []uint64) {
	h1 := hash.Optimized1(data)
	h2 := hash.Optimized2(data)
	bitCount := bf.bitCount
	hCount := bf.hashCount

	for i := uint32(0); i < hCount; i++ {
		positions[i] = (h1 + uint64(i)*h2) % bitCount
	}
}
