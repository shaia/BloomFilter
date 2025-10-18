package arm64

import "unsafe"

// PopCount performs SIMD population count using NEON
func PopCount(data unsafe.Pointer, length int) int {
	return neonPopCount(data, length)
}

// VectorOr performs SIMD OR operation using NEON
func VectorOr(dst, src unsafe.Pointer, length int) {
	neonVectorOr(dst, src, length)
}

// VectorAnd performs SIMD AND operation using NEON
func VectorAnd(dst, src unsafe.Pointer, length int) {
	neonVectorAnd(dst, src, length)
}

// VectorClear performs SIMD clear operation using NEON
func VectorClear(data unsafe.Pointer, length int) {
	neonVectorClear(data, length)
}
