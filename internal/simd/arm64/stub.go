//go:build !arm64 || purego

package arm64

import "unsafe"

// Stub implementations for non-ARM64 platforms
// These will never be called due to runtime detection, but are needed for compilation

func neonPopCount(data unsafe.Pointer, length int) int {
	// This should never be called on non-ARM64 platforms
	panic("neonPopCount called on non-ARM64 platform")
}

func neonVectorOr(dst, src unsafe.Pointer, length int) {
	// This should never be called on non-ARM64 platforms
	panic("neonVectorOr called on non-ARM64 platform")
}

func neonVectorAnd(dst, src unsafe.Pointer, length int) {
	// This should never be called on non-ARM64 platforms
	panic("neonVectorAnd called on non-ARM64 platform")
}

func neonVectorClear(data unsafe.Pointer, length int) {
	// This should never be called on non-ARM64 platforms
	panic("neonVectorClear called on non-ARM64 platform")
}
