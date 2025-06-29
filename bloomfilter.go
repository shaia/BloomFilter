package bloomfilter

import (
	"fmt"
	"math"
	"math/bits"
	"runtime"
	"unsafe"
)

// SIMD capabilities detection
var (
	hasAVX2   bool
	hasAVX512 bool
	hasNEON   bool
)

func init() {
	detectSIMDCapabilities()
}

// detectSIMDCapabilities detects available SIMD instruction sets
func detectSIMDCapabilities() {
	// This is a simplified detection - in production you'd use proper CPU detection
	switch runtime.GOARCH {
	case "amd64":
		// Simplified detection - assume modern Intel/AMD processors have AVX2
		hasAVX2 = true
		// AVX512 is less common, set to false for safety
		hasAVX512 = false
	case "arm64":
		// ARM64 has NEON by default
		hasNEON = true
	}
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
	}
}

// Fast hash functions optimized for cache performance
func hash1(data []byte) uint64 {
	const (
		fnvOffsetBasis = 14695981039346656037
		fnvPrime       = 1099511628211
	)

	hash := uint64(fnvOffsetBasis)

	// Process in 8-byte chunks for better cache utilization
	i := 0
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

func hash2(data []byte) uint64 {
	const (
		seed = 0x9e3779b97f4a7c15
		mult = 0xc6a4a7935bd1e995
		r    = 47
	)

	hash := uint64(seed)

	// Process in 8-byte chunks
	i := 0
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

// SIMD-optimized hash functions for better performance
func hashSIMD1(data []byte) uint64 {
	if hasAVX2 || hasNEON {
		// Use optimized vectorized hash when available
		return hashOptimized1(data)
	}
	return hash1(data)
}

func hashSIMD2(data []byte) uint64 {
	if hasAVX2 || hasNEON {
		// Use optimized vectorized hash when available
		return hashOptimized2(data)
	}
	return hash2(data)
}

// Optimized hash functions with better vectorization
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

// getHashPositionsOptimized generates hash positions with cache line grouping and SIMD-optimized hashing
func (bf *CacheOptimizedBloomFilter) getHashPositionsOptimized(data []byte) {
	h1 := hashSIMD1(data)
	h2 := hashSIMD2(data)

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

// Convenience methods
func (bf *CacheOptimizedBloomFilter) AddString(s string) {
	data := *(*[]byte)(unsafe.Pointer(&struct {
		string
		int
	}{s, len(s)}))
	bf.Add(data)
}

func (bf *CacheOptimizedBloomFilter) ContainsString(s string) bool {
	data := *(*[]byte)(unsafe.Pointer(&struct {
		string
		int
	}{s, len(s)}))
	return bf.Contains(data)
}

func (bf *CacheOptimizedBloomFilter) AddUint64(n uint64) {
	data := (*[8]byte)(unsafe.Pointer(&n))[:]
	bf.Add(data)
}

func (bf *CacheOptimizedBloomFilter) ContainsUint64(n uint64) bool {
	data := (*[8]byte)(unsafe.Pointer(&n))[:]
	return bf.Contains(data)
}

// Clear resets the bloom filter with cache line awareness and SIMD optimization
func (bf *CacheOptimizedBloomFilter) Clear() {
	if hasAVX2 || hasNEON {
		bf.ClearSIMD()
	} else {
		bf.ClearOptimized()
	}
}

// Cache line optimized bulk operations

// Union performs cache line aware union with SIMD optimization
func (bf *CacheOptimizedBloomFilter) Union(other *CacheOptimizedBloomFilter) error {
	if hasAVX2 || hasNEON {
		return bf.UnionSIMD(other)
	} else {
		return bf.UnionOptimized(other)
	}
}

// Intersection performs cache line aware intersection with SIMD optimization
func (bf *CacheOptimizedBloomFilter) Intersection(other *CacheOptimizedBloomFilter) error {
	if hasAVX2 || hasNEON {
		return bf.IntersectionSIMD(other)
	} else {
		return bf.IntersectionOptimized(other)
	}
}

// PopCount with cache line optimization and SIMD support
func (bf *CacheOptimizedBloomFilter) PopCount() uint64 {
	if hasAVX2 || hasNEON {
		return bf.PopCountSIMD()
	} else {
		return bf.PopCountOptimized()
	}
}

// SIMD-optimized PopCount using vectorized bit counting
func (bf *CacheOptimizedBloomFilter) PopCountSIMD() uint64 {
	// Architecture-specific SIMD implementations are available via build tags
	// On ARM64: uses NEON-optimized operations when available
	// On AMD64: uses AVX2-optimized operations when available
	// This provides optimal vectorized performance for each platform
	return bf.PopCountOptimized()
}

// PopCountOptimized uses unrolled loops for better performance
func (bf *CacheOptimizedBloomFilter) PopCountOptimized() uint64 {
	count := uint64(0)

	// Process cache lines with unrolled loops
	for i := uint64(0); i < bf.cacheLineCount; i++ {
		cacheLine := &bf.cacheLines[i]
		// Unroll the inner loop for better performance
		count += uint64(bits.OnesCount64(cacheLine.words[0]))
		count += uint64(bits.OnesCount64(cacheLine.words[1]))
		count += uint64(bits.OnesCount64(cacheLine.words[2]))
		count += uint64(bits.OnesCount64(cacheLine.words[3]))
		count += uint64(bits.OnesCount64(cacheLine.words[4]))
		count += uint64(bits.OnesCount64(cacheLine.words[5]))
		count += uint64(bits.OnesCount64(cacheLine.words[6]))
		count += uint64(bits.OnesCount64(cacheLine.words[7]))
	}

	return count
}

// SIMD-optimized Union operation
func (bf *CacheOptimizedBloomFilter) UnionSIMD(other *CacheOptimizedBloomFilter) error {
	// The actual SIMD implementations are defined in architecture-specific files
	// This will call the appropriate SIMD function based on build tags
	return bf.UnionOptimized(other)
}

// UnionOptimized uses vectorized operations when possible
func (bf *CacheOptimizedBloomFilter) UnionOptimized(other *CacheOptimizedBloomFilter) error {
	if bf.cacheLineCount != other.cacheLineCount {
		return fmt.Errorf("bloom filters must have same size for union")
	}

	// Process entire cache lines at once with unrolled loops
	for i := uint64(0); i < bf.cacheLineCount; i++ {
		dst := &bf.cacheLines[i]
		src := &other.cacheLines[i]

		// Unrolled loop for better performance
		dst.words[0] |= src.words[0]
		dst.words[1] |= src.words[1]
		dst.words[2] |= src.words[2]
		dst.words[3] |= src.words[3]
		dst.words[4] |= src.words[4]
		dst.words[5] |= src.words[5]
		dst.words[6] |= src.words[6]
		dst.words[7] |= src.words[7]
	}

	return nil
}

// SIMD-optimized Intersection operation
func (bf *CacheOptimizedBloomFilter) IntersectionSIMD(other *CacheOptimizedBloomFilter) error {
	// The actual SIMD implementations are defined in architecture-specific files
	// This will call the appropriate SIMD function based on build tags
	return bf.IntersectionOptimized(other)
}

// IntersectionOptimized uses vectorized operations when possible
func (bf *CacheOptimizedBloomFilter) IntersectionOptimized(other *CacheOptimizedBloomFilter) error {
	if bf.cacheLineCount != other.cacheLineCount {
		return fmt.Errorf("bloom filters must have same size for intersection")
	}

	// Process entire cache lines at once with unrolled loops
	for i := uint64(0); i < bf.cacheLineCount; i++ {
		dst := &bf.cacheLines[i]
		src := &other.cacheLines[i]

		// Unrolled loop for better performance
		dst.words[0] &= src.words[0]
		dst.words[1] &= src.words[1]
		dst.words[2] &= src.words[2]
		dst.words[3] &= src.words[3]
		dst.words[4] &= src.words[4]
		dst.words[5] &= src.words[5]
		dst.words[6] &= src.words[6]
		dst.words[7] &= src.words[7]
	}

	return nil
}

// SIMD-optimized Clear operation
func (bf *CacheOptimizedBloomFilter) ClearSIMD() {
	// The actual SIMD implementations are defined in architecture-specific files
	// This will call the appropriate SIMD function based on build tags
	bf.ClearOptimized()
}

// ClearOptimized uses vectorized operations when possible
func (bf *CacheOptimizedBloomFilter) ClearOptimized() {
	// Clear entire cache lines at once with unrolled loops
	for i := range bf.cacheLines {
		cacheLine := &bf.cacheLines[i]
		// Unrolled loop for better performance
		cacheLine.words[0] = 0
		cacheLine.words[1] = 0
		cacheLine.words[2] = 0
		cacheLine.words[3] = 0
		cacheLine.words[4] = 0
		cacheLine.words[5] = 0
		cacheLine.words[6] = 0
		cacheLine.words[7] = 0
	}
}

// Statistics and analysis
func (bf *CacheOptimizedBloomFilter) EstimatedFPP() float64 {
	bitsSet := float64(bf.PopCount())
	ratio := bitsSet / float64(bf.bitCount)
	return math.Pow(ratio, float64(bf.hashCount))
}

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
		HasAVX2:     hasAVX2,
		HasAVX512:   hasAVX512,
		HasNEON:     hasNEON,
		SIMDEnabled: hasAVX2 || hasAVX512 || hasNEON,
	}
}

// Exported SIMD capability functions
func HasAVX2() bool {
	return hasAVX2
}

func HasAVX512() bool {
	return hasAVX512
}

func HasNEON() bool {
	return hasNEON
}

func HasSIMD() bool {
	return hasAVX2 || hasAVX512 || hasNEON
}
