//go:build amd64 && !purego

package amd64

import "unsafe"

// AVX2 SIMD intrinsics for AMD64/x86-64
// These functions use actual AVX2 vector instructions and are implemented in assembly

//go:noescape
func avx2PopCount(data unsafe.Pointer, length int) int

//go:noescape
func avx2VectorOr(dst, src unsafe.Pointer, length int)

//go:noescape
func avx2VectorAnd(dst, src unsafe.Pointer, length int)

//go:noescape
func avx2VectorClear(data unsafe.Pointer, length int)

//go:noescape
func hasAVX2Support() bool
