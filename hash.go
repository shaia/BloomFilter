package bloomfilter

import "unsafe"

// hashOptimized1 implements FNV-1a hash with optimized chunking for cache efficiency.
// Processes data in 32-byte chunks (AVX2-friendly) for better performance.
func hashOptimized1(data []byte) uint64 {
	const (
		fnvOffsetBasis = 14695981039346656037
		fnvPrime       = 1099511628211
	)

	hash := uint64(fnvOffsetBasis)
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

// hashOptimized2 implements a variant hash function with different constants.
// Using two independent hash functions provides better distribution.
func hashOptimized2(data []byte) uint64 {
	const (
		seed = 0x9e3779b97f4a7c15
		mult = 0xc6a4a7935bd1e995
		r    = 47
	)

	hash := uint64(seed)
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
