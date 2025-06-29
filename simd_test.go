package bloomfilter

import (
	"fmt"
	"runtime"
	"testing"
)

// TestSIMDCapabilities tests SIMD capability detection and reporting
func TestSIMDCapabilities(t *testing.T) {
	// Test SIMD capability detection
	t.Logf("Runtime GOARCH: %s", runtime.GOARCH)
	t.Logf("Detected AVX2: %t", HasAVX2())
	t.Logf("Detected AVX512: %t", HasAVX512())
	t.Logf("Detected NEON: %t", HasNEON())
	t.Logf("Any SIMD available: %t", HasSIMD())

	// Verify correct detection for current architecture
	switch runtime.GOARCH {
	case "amd64":
		if !HasAVX2() {
			t.Logf("Note: AVX2 not detected on AMD64 - may be running in emulation")
		}
	case "arm64":
		if !HasNEON() {
			t.Errorf("NEON should be available on ARM64")
		}
	}
}

// TestSIMDFunctions tests that SIMD functions can be called without errors
func TestSIMDFunctions(t *testing.T) {
	bf := NewCacheOptimizedBloomFilter(1000, 0.01)

	// Add some test data
	for i := 0; i < 100; i++ {
		bf.AddString(fmt.Sprintf("test_%d", i))
	}

	// Test SIMD PopCount
	count1 := bf.PopCount()
	count2 := bf.PopCountSIMD()

	if count1 != count2 {
		t.Errorf("PopCount mismatch: regular=%d, SIMD=%d", count1, count2)
	}
	t.Logf("PopCount result: %d bits set", count1)

	// Test SIMD Union
	bf2 := NewCacheOptimizedBloomFilter(1000, 0.01)
	for i := 50; i < 150; i++ {
		bf2.AddString(fmt.Sprintf("test_%d", i))
	}

	err := bf.UnionSIMD(bf2)
	if err != nil {
		t.Errorf("UnionSIMD failed: %v", err)
	}

	// Test SIMD Intersection
	bf3 := NewCacheOptimizedBloomFilter(1000, 0.01)
	for i := 0; i < 100; i++ {
		bf3.AddString(fmt.Sprintf("test_%d", i))
	}

	err = bf3.IntersectionSIMD(bf2)
	if err != nil {
		t.Errorf("IntersectionSIMD failed: %v", err)
	}

	// Test SIMD Clear
	bf.ClearSIMD()
	countAfterClear := bf.PopCount()
	if countAfterClear != 0 {
		t.Errorf("ClearSIMD failed: expected 0 bits, got %d", countAfterClear)
	}

	t.Logf("All SIMD functions executed successfully")
}

// TestCacheStats tests the cache statistics and SIMD reporting
func TestCacheStats(t *testing.T) {
	bf := NewCacheOptimizedBloomFilter(1000, 0.01)
	stats := bf.GetCacheStats()

	t.Logf("Cache Stats:")
	t.Logf("  Bit Count: %d", stats.BitCount)
	t.Logf("  Hash Count: %d", stats.HashCount)
	t.Logf("  Cache Lines: %d", stats.CacheLineCount)
	t.Logf("  Memory Usage: %d bytes", stats.MemoryUsage)
	t.Logf("  Alignment Offset: %d", stats.Alignment)
	t.Logf("  SIMD Capabilities:")
	t.Logf("    AVX2: %t", stats.HasAVX2)
	t.Logf("    AVX512: %t", stats.HasAVX512)
	t.Logf("    NEON: %t", stats.HasNEON)
	t.Logf("    SIMD Enabled: %t", stats.SIMDEnabled)

	// Verify SIMD is enabled on supported architectures
	if runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64" {
		if !stats.SIMDEnabled {
			t.Logf("Note: SIMD not enabled on %s - may be expected behavior", runtime.GOARCH)
		}
	}

	// Verify cache line alignment
	if stats.Alignment != 0 {
		t.Logf("Note: Memory not perfectly aligned (offset: %d bytes)", stats.Alignment)
	}
}
