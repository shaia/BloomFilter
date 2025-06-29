//go:build arm64 && !purego

package bloomfilter

import (
	"fmt"
	"math/bits"
	"unsafe"
)

// popCountNEON performs vectorized population count using ARM NEON-style operations
//
//nolint:unused // Called conditionally via SIMD dispatch
func (bf *CacheOptimizedBloomFilter) popCountNEON() uint64 {
	count := uint64(0)

	// Process cache lines with unrolled loops optimized for NEON
	// This simulates 128-bit vector processing by processing 2 uint64 at a time
	for i := uint64(0); i < bf.cacheLineCount; i++ {
		cacheLine := &bf.cacheLines[i]

		// Process 2 uint64 values at once (simulating 128-bit NEON vector)
		// NEON processes in 128-bit chunks, so we do 4 iterations per cache line
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

// unionNEON performs vectorized union using ARM NEON-style operations
//
//nolint:unused // Called conditionally via SIMD dispatch
func (bf *CacheOptimizedBloomFilter) unionNEON(other *CacheOptimizedBloomFilter) error {
	if bf.cacheLineCount != other.cacheLineCount {
		return fmt.Errorf("bloom filters must have same size for union")
	}

	// Process entire cache lines with vectorized operations
	for i := uint64(0); i < bf.cacheLineCount; i++ {
		dst := &bf.cacheLines[i]
		src := &other.cacheLines[i]

		// Vectorized OR operations (simulating NEON 128-bit vectors)
		// Process four 16-byte chunks per cache line
		for j := 0; j < 8; j += 2 {
			dstPtr := (*[2]uint64)(unsafe.Pointer(&dst.words[j]))
			srcPtr := (*[2]uint64)(unsafe.Pointer(&src.words[j]))

			// 16-byte chunk (2 uint64)
			dstPtr[0] |= srcPtr[0]
			dstPtr[1] |= srcPtr[1]
		}
	}

	return nil
}

// intersectionNEON performs vectorized intersection using ARM NEON-style operations
//
//nolint:unused // Called conditionally via SIMD dispatch
func (bf *CacheOptimizedBloomFilter) intersectionNEON(other *CacheOptimizedBloomFilter) error {
	if bf.cacheLineCount != other.cacheLineCount {
		return fmt.Errorf("bloom filters must have same size for intersection")
	}

	// Process entire cache lines with vectorized operations
	for i := uint64(0); i < bf.cacheLineCount; i++ {
		dst := &bf.cacheLines[i]
		src := &other.cacheLines[i]

		// Vectorized AND operations (simulating NEON 128-bit vectors)
		// Process four 16-byte chunks per cache line
		for j := 0; j < 8; j += 2 {
			dstPtr := (*[2]uint64)(unsafe.Pointer(&dst.words[j]))
			srcPtr := (*[2]uint64)(unsafe.Pointer(&src.words[j]))

			// 16-byte chunk (2 uint64)
			dstPtr[0] &= srcPtr[0]
			dstPtr[1] &= srcPtr[1]
		}
	}

	return nil
}

// clearNEON performs vectorized clear using ARM NEON-style operations
//
//nolint:unused // Called conditionally via SIMD dispatch
func (bf *CacheOptimizedBloomFilter) clearNEON() {
	// Process entire cache lines with vectorized operations
	for i := range bf.cacheLines {
		cacheLine := &bf.cacheLines[i]

		// Vectorized clear operations (simulating NEON 128-bit vectors)
		// Clear four 16-byte chunks per cache line
		for j := 0; j < 8; j += 2 {
			ptr := (*[2]uint64)(unsafe.Pointer(&cacheLine.words[j]))
			ptr[0] = 0
			ptr[1] = 0
		}
	}
}
