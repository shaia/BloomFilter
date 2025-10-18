//go:build !amd64 || purego

package amd64

import "unsafe"

// Stub implementations for non-AMD64 platforms
// These will never be called due to runtime detection, but are needed for compilation

func avx2PopCount(data unsafe.Pointer, length int) int {
	// This should never be called on non-AMD64 platforms
	panic("avx2PopCount called on non-AMD64 platform")
}

func avx2VectorOr(dst, src unsafe.Pointer, length int) {
	// This should never be called on non-AMD64 platforms
	panic("avx2VectorOr called on non-AMD64 platform")
}

func avx2VectorAnd(dst, src unsafe.Pointer, length int) {
	// This should never be called on non-AMD64 platforms
	panic("avx2VectorAnd called on non-AMD64 platform")
}

func avx2VectorClear(data unsafe.Pointer, length int) {
	// This should never be called on non-AMD64 platforms
	panic("avx2VectorClear called on non-AMD64 platform")
}

func hasAVX2Support() bool {
	// AVX2 is only available on x86-64
	return false
}
