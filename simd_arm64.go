//go:build arm64 && !purego

package bloomfilter

import (
	"unsafe"
)

// NEON SIMD intrinsics for ARM64
// These functions use actual ARM NEON vector instructions and are implemented in assembly

//go:noescape
func neonPopCount(data unsafe.Pointer, length int) int

//go:noescape
func neonVectorOr(dst, src unsafe.Pointer, length int)

//go:noescape
func neonVectorAnd(dst, src unsafe.Pointer, length int)

//go:noescape
func neonVectorClear(data unsafe.Pointer, length int)
