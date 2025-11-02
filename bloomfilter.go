package bloomfilter

import (
	"fmt"
	"math"
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

// Add adds an element with cache line optimization
func (bf *CacheOptimizedBloomFilter) Add(data []byte) {
	h1 := hash.Optimized1(data)
	h2 := hash.Optimized2(data)

	// Stack buffer for typical filters
	var stackBuf [16]uint64
	var positions []uint64
	if bf.hashCount <= 16 {
		positions = stackBuf[:bf.hashCount]
	} else {
		positions = make([]uint64, bf.hashCount)
	}

	// Generate positions
	for i := uint32(0); i < bf.hashCount; i++ {
		positions[i] = (h1 + uint64(i)*h2) % bf.bitCount
	}

	// Set bits atomically
	bf.setBitsAtomic(positions)
}

// Contains checks membership with cache line optimization
func (bf *CacheOptimizedBloomFilter) Contains(data []byte) bool {
	h1 := hash.Optimized1(data)
	h2 := hash.Optimized2(data)

	var stackBuf [16]uint64
	var positions []uint64
	if bf.hashCount <= 16 {
		positions = stackBuf[:bf.hashCount]
	} else {
		positions = make([]uint64, bf.hashCount)
	}

	for i := uint32(0); i < bf.hashCount; i++ {
		positions[i] = (h1 + uint64(i)*h2) % bf.bitCount
	}

	return bf.checkBitsAtomic(positions)
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
