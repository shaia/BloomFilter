package simd

import (
	"runtime"
	"unsafe"
)

// Operations defines the interface for SIMD operations
// This allows us to support different SIMD instruction sets (NEON, AVX2, AVX512)
type Operations interface {
	PopCount(data unsafe.Pointer, length int) int
	VectorOr(dst, src unsafe.Pointer, length int)
	VectorAnd(dst, src unsafe.Pointer, length int)
	VectorClear(data unsafe.Pointer, length int)
}

// Get returns the best available SIMD implementation
func Get() Operations {
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

// HasAVX2 returns true if AVX2 instructions are available
func HasAVX2() bool {
	return hasAVX2
}

// HasAVX512 returns true if AVX512 instructions are available
func HasAVX512() bool {
	return hasAVX512
}

// HasNEON returns true if NEON instructions are available
func HasNEON() bool {
	return hasNEON
}

// HasAny returns true if any SIMD instructions are available
func HasAny() bool {
	return hasAVX2 || hasAVX512 || hasNEON
}

// SIMD capabilities detection
var (
	hasAVX2   bool
	hasAVX512 bool
	hasNEON   bool
)

func init() {
	detectCapabilities()
}

// detectCapabilities detects available SIMD instruction sets
func detectCapabilities() {
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
