package bloomfilter

import "unsafe"

// SIMDOperations defines the interface for SIMD operations
// This allows us to support different SIMD instruction sets (NEON, AVX2, AVX512)
type SIMDOperations interface {
	PopCount(data unsafe.Pointer, length int) int
	VectorOr(dst, src unsafe.Pointer, length int)
	VectorAnd(dst, src unsafe.Pointer, length int)
	VectorClear(data unsafe.Pointer, length int)
}

// GetSIMDOperations returns the best available SIMD implementation
func GetSIMDOperations() SIMDOperations {
	// Priority order: AVX512 > AVX2 > NEON > Fallback
	if hasAVX512 {
		return &AVX512Operations{}
	} else if hasAVX2 {
		return &AVX2Operations{}
	} else if hasNEON {
		return &NEONOperations{}
	}
	return &FallbackOperations{}
}
