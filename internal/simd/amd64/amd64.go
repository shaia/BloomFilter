package amd64

import "unsafe"

// PopCount performs SIMD population count using AVX2
func PopCount(data unsafe.Pointer, length int) int {
	return avx2PopCount(data, length)
}

// VectorOr performs SIMD OR operation using AVX2
func VectorOr(dst, src unsafe.Pointer, length int) {
	avx2VectorOr(dst, src, length)
}

// VectorAnd performs SIMD AND operation using AVX2
func VectorAnd(dst, src unsafe.Pointer, length int) {
	avx2VectorAnd(dst, src, length)
}

// VectorClear performs SIMD clear operation using AVX2
func VectorClear(data unsafe.Pointer, length int) {
	avx2VectorClear(data, length)
}

// HasAVX2 returns true if AVX2 is supported
func HasAVX2() bool {
	return hasAVX2Support()
}
