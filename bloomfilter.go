package bloomfilter

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/shaia/go-simd-bloomfilter/internal/simd"
)

// CacheOptimizedBloomFilter uses cache line aligned storage
type CacheOptimizedBloomFilter struct {
	// Cache line aligned bitset
	cacheLines     []CacheLine
	bitCount       uint64
	hashCount      uint32
	cacheLineCount uint64

	// Pre-allocated arrays to avoid allocations in hot paths
	positions        []uint64
	cacheLineIndices []uint64

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

// NewCacheOptimizedBloomFilter creates a cache line optimized bloom filter
func NewCacheOptimizedBloomFilter(expectedElements uint64, falsePositiveRate float64) *CacheOptimizedBloomFilter {
	// Calculate optimal parameters
	ln2 := math.Ln2
	bitCount := uint64(-float64(expectedElements) * math.Log(falsePositiveRate) / (ln2 * ln2))
	hashCount := uint32(float64(bitCount) * ln2 / float64(expectedElements))

	if hashCount < 1 {
		hashCount = 1
	}

	// Align to cache line boundaries (512 bits per cache line)
	cacheLineCount := (bitCount + BitsPerCacheLine - 1) / BitsPerCacheLine
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

	return &CacheOptimizedBloomFilter{
		cacheLines:       cacheLines,
		bitCount:         bitCount,
		hashCount:        hashCount,
		cacheLineCount:   cacheLineCount,
		positions:        make([]uint64, hashCount),
		cacheLineIndices: make([]uint64, hashCount),
		simdOps:          simd.Get(), // Initialize SIMD operations once
	}
}

// Add adds an element with cache line optimization
func (bf *CacheOptimizedBloomFilter) Add(data []byte) {
	bf.getHashPositionsOptimized(data)
	bf.prefetchCacheLines()
	bf.setBitCacheOptimized(bf.positions[:bf.hashCount])
}

// Contains checks membership with cache line optimization
func (bf *CacheOptimizedBloomFilter) Contains(data []byte) bool {
	bf.getHashPositionsOptimized(data)
	bf.prefetchCacheLines()
	return bf.getBitCacheOptimized(bf.positions[:bf.hashCount])
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
)

// CacheLine represents a single 64-byte cache line containing 8 uint64 words
type CacheLine struct {
	words [WordsPerCacheLine]uint64
}

// Optimized hash functions with better vectorization and cache utilization
func hashOptimized1(data []byte) uint64 {
	const (
		fnvOffsetBasis = 14695981039346656037
		fnvPrime       = 1099511628211
	)

	hash := uint64(fnvOffsetBasis)

	// Process in larger chunks for better cache utilization
	i := 0

	// Process 32-byte chunks when possible (AVX2 friendly)
	for i+32 <= len(data) {
		// Unroll the loop for 4 uint64 values
		chunk1 := *(*uint64)(unsafe.Pointer(&data[i]))
		chunk2 := *(*uint64)(unsafe.Pointer(&data[i+8]))
		chunk3 := *(*uint64)(unsafe.Pointer(&data[i+16]))
		chunk4 := *(*uint64)(unsafe.Pointer(&data[i+24]))

		hash ^= chunk1
		hash *= fnvPrime
		hash ^= chunk2
		hash *= fnvPrime
		hash ^= chunk3
		hash *= fnvPrime
		hash ^= chunk4
		hash *= fnvPrime

		i += 32
	}

	// Process remaining 8-byte chunks
	for i+8 <= len(data) {
		chunk := *(*uint64)(unsafe.Pointer(&data[i]))
		hash ^= chunk
		hash *= fnvPrime
		i += 8
	}

	// Handle remaining bytes
	for i < len(data) {
		hash ^= uint64(data[i])
		hash *= fnvPrime
		i++
	}

	return hash
}

func hashOptimized2(data []byte) uint64 {
	const (
		seed = 0x9e3779b97f4a7c15
		mult = 0xc6a4a7935bd1e995
		r    = 47
	)

	hash := uint64(seed)

	// Process in larger chunks for better cache utilization
	i := 0

	// Process 32-byte chunks when possible (AVX2 friendly)
	for i+32 <= len(data) {
		// Unroll the loop for 4 uint64 values
		chunk1 := *(*uint64)(unsafe.Pointer(&data[i]))
		chunk2 := *(*uint64)(unsafe.Pointer(&data[i+8]))
		chunk3 := *(*uint64)(unsafe.Pointer(&data[i+16]))
		chunk4 := *(*uint64)(unsafe.Pointer(&data[i+24]))

		hash ^= chunk1
		hash *= mult
		hash ^= hash >> r
		hash ^= chunk2
		hash *= mult
		hash ^= hash >> r
		hash ^= chunk3
		hash *= mult
		hash ^= hash >> r
		hash ^= chunk4
		hash *= mult
		hash ^= hash >> r

		i += 32
	}

	// Process remaining 8-byte chunks
	for i+8 <= len(data) {
		chunk := *(*uint64)(unsafe.Pointer(&data[i]))
		hash ^= chunk
		hash *= mult
		hash ^= hash >> r
		i += 8
	}

	// Handle remaining bytes
	for i < len(data) {
		hash ^= uint64(data[i])
		hash *= mult
		hash ^= hash >> r
		i++
	}

	return hash
}

// getHashPositionsOptimized generates hash positions with cache line grouping and vectorized hashing
func (bf *CacheOptimizedBloomFilter) getHashPositionsOptimized(data []byte) {
	h1 := hashOptimized1(data)
	h2 := hashOptimized2(data)

	// Generate positions and group by cache line to improve locality
	cacheLineMap := make(map[uint64][]uint64)

	for i := uint32(0); i < bf.hashCount; i++ {
		hash := h1 + uint64(i)*h2
		bitPos := hash % bf.bitCount
		cacheLineIdx := bitPos / BitsPerCacheLine

		bf.positions[i] = bitPos
		cacheLineMap[cacheLineIdx] = append(cacheLineMap[cacheLineIdx], bitPos)
	}

	// Store unique cache line indices for prefetching
	bf.cacheLineIndices = bf.cacheLineIndices[:0]
	for cacheLineIdx := range cacheLineMap {
		bf.cacheLineIndices = append(bf.cacheLineIndices, cacheLineIdx)
	}
}

// prefetchCacheLines provides hints to prefetch cache lines
func (bf *CacheOptimizedBloomFilter) prefetchCacheLines() {
	// In Go, we can't directly issue prefetch instructions,
	// but we can hint to the runtime by touching memory
	for _, idx := range bf.cacheLineIndices {
		if idx < bf.cacheLineCount {
			// Touch the cache line to bring it into cache
			_ = bf.cacheLines[idx].words[0]
		}
	}
}

// setBitCacheOptimized sets multiple bits with cache line awareness
func (bf *CacheOptimizedBloomFilter) setBitCacheOptimized(positions []uint64) {
	// Group operations by cache line to minimize cache misses
	cacheLineOps := make(map[uint64][]struct{ wordIdx, bitOffset uint64 })

	for _, bitPos := range positions {
		cacheLineIdx := bitPos / BitsPerCacheLine
		wordInCacheLine := (bitPos % BitsPerCacheLine) / 64
		bitOffset := bitPos % 64

		cacheLineOps[cacheLineIdx] = append(cacheLineOps[cacheLineIdx], struct{ wordIdx, bitOffset uint64 }{
			wordIdx: wordInCacheLine, bitOffset: bitOffset,
		})
	}

	// Process each cache line's operations together
	for cacheLineIdx, ops := range cacheLineOps {
		if cacheLineIdx < bf.cacheLineCount {
			cacheLine := &bf.cacheLines[cacheLineIdx]
			for _, op := range ops {
				cacheLine.words[op.wordIdx] |= 1 << op.bitOffset
			}
		}
	}
}

// getBitCacheOptimized checks multiple bits with cache line awareness
func (bf *CacheOptimizedBloomFilter) getBitCacheOptimized(positions []uint64) bool {
	// Group operations by cache line
	cacheLineOps := make(map[uint64][]struct{ wordIdx, bitOffset uint64 })

	for _, bitPos := range positions {
		cacheLineIdx := bitPos / BitsPerCacheLine
		wordInCacheLine := (bitPos % BitsPerCacheLine) / 64
		bitOffset := bitPos % 64

		cacheLineOps[cacheLineIdx] = append(cacheLineOps[cacheLineIdx], struct{ wordIdx, bitOffset uint64 }{
			wordIdx: wordInCacheLine, bitOffset: bitOffset,
		})
	}

	// Check each cache line's bits together
	for cacheLineIdx, ops := range cacheLineOps {
		if cacheLineIdx >= bf.cacheLineCount {
			return false
		}

		cacheLine := &bf.cacheLines[cacheLineIdx]
		for _, op := range ops {
			if (cacheLine.words[op.wordIdx] & (1 << op.bitOffset)) == 0 {
				return false
			}
		}
	}

	return true
}
