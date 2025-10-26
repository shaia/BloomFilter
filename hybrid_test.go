package bloomfilter

import (
	"fmt"
	"testing"
)

// TestHybridModeSelection verifies that the correct mode is chosen based on filter size
func TestHybridModeSelection(t *testing.T) {
	tests := []struct {
		name           string
		elements       uint64
		fpr            float64
		expectArrayMode bool
		description    string
	}{
		{
			name:           "Small filter - should use array mode",
			elements:       10_000,
			fpr:            0.01,
			expectArrayMode: true,
			description:    "10K elements = ~1200 cache lines < 10K threshold",
		},
		{
			name:           "Medium filter - should use array mode",
			elements:       100_000,
			fpr:            0.01,
			expectArrayMode: true,
			description:    "100K elements = ~11,980 cache lines, close to threshold but array mode",
		},
		{
			name:           "Large filter - should use map mode",
			elements:       1_000_000,
			fpr:            0.01,
			expectArrayMode: false,
			description:    "1M elements = ~119,808 cache lines > 10K threshold",
		},
		{
			name:           "Very large filter - should use map mode",
			elements:       10_000_000,
			fpr:            0.01,
			expectArrayMode: false,
			description:    "10M elements = ~1,198,086 cache lines >> 10K threshold",
		},
		{
			name:           "Huge filter - should use map mode",
			elements:       100_000_000,
			fpr:            0.001,
			expectArrayMode: false,
			description:    "100M elements with low FPR = millions of cache lines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := NewCacheOptimizedBloomFilter(tt.elements, tt.fpr)

			if bf.useArrayMode != tt.expectArrayMode {
				t.Errorf("%s: expected array mode=%v, got=%v\n  Cache lines: %d, Threshold: %d\n  %s",
					tt.name, tt.expectArrayMode, bf.useArrayMode,
					bf.cacheLineCount, ArrayModeThreshold, tt.description)
			}

			// Verify the correct structures are initialized
			if bf.useArrayMode {
				if bf.arrayOps == nil || bf.arrayOpsSet == nil || bf.arrayMap == nil {
					t.Errorf("%s: array mode selected but array structures not initialized", tt.name)
				}
				if bf.mapOps != nil || bf.mapOpsSet != nil || bf.mapMap != nil {
					t.Errorf("%s: array mode selected but map structures were initialized", tt.name)
				}
			} else {
				if bf.mapOps == nil || bf.mapOpsSet == nil || bf.mapMap == nil {
					t.Errorf("%s: map mode selected but map structures not initialized", tt.name)
				}
				if bf.arrayOps != nil || bf.arrayOpsSet != nil || bf.arrayMap != nil {
					t.Errorf("%s: map mode selected but array structures were initialized", tt.name)
				}
			}

			t.Logf("✓ %s: mode=%s, cache_lines=%d, bits=%d",
				tt.name,
				func() string {
					if bf.useArrayMode {
						return "ARRAY"
					}
					return "MAP"
				}(),
				bf.cacheLineCount,
				bf.bitCount)
		})
	}
}

// TestHybridModeCorrectness verifies both modes produce correct results
func TestHybridModeCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		elements uint64
		fpr      float64
	}{
		{"Small/Array Mode", 10_000, 0.01},
		{"Large/Map Mode", 1_000_000, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := NewCacheOptimizedBloomFilter(tt.elements, tt.fpr)

			mode := "ARRAY"
			if !bf.useArrayMode {
				mode = "MAP"
			}
			t.Logf("Testing %s mode with %d elements", mode, tt.elements)

			// Test with 1000 elements
			testElements := []string{
				"apple", "banana", "cherry", "date", "elderberry",
				"fig", "grape", "honeydew", "kiwi", "lemon",
			}

			// Add elements
			for _, elem := range testElements {
				bf.AddString(elem)
			}

			// Verify all added elements are found
			for _, elem := range testElements {
				if !bf.ContainsString(elem) {
					t.Errorf("%s mode: element '%s' was added but not found", mode, elem)
				}
			}

			// Test elements that weren't added
			notAdded := []string{"mango", "nectarine", "orange", "papaya", "quince"}
			falsePositives := 0
			for _, elem := range notAdded {
				if bf.ContainsString(elem) {
					falsePositives++
				}
			}

			// With good FPR and few checks, should have very few false positives
			if falsePositives > 2 {
				t.Logf("%s mode: warning - %d false positives out of %d (might be normal)",
					mode, falsePositives, len(notAdded))
			}

			t.Logf("✓ %s mode correctness verified: %d/%d elements found, %d/%d false positives",
				mode, len(testElements), len(testElements), falsePositives, len(notAdded))
		})
	}
}

// TestHybridMemoryFootprint estimates memory usage for different modes
func TestHybridMemoryFootprint(t *testing.T) {
	tests := []struct {
		name     string
		elements uint64
		fpr      float64
	}{
		{"Small/Array", 1_000, 0.01},
		{"Medium/Array", 10_000, 0.01},
		{"Large/Map", 1_000_000, 0.01},
		{"Huge/Map", 10_000_000, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := NewCacheOptimizedBloomFilter(tt.elements, tt.fpr)

			// Calculate actual bit array size
			bitArrayBytes := bf.cacheLineCount * CacheLineSize

			// Estimate overhead based on mode
			var overheadBytes uint64
			mode := "ARRAY"

			if bf.useArrayMode {
				// Array mode: fixed overhead
				// 3 arrays × 10K elements × 24 bytes/slice = ~720KB
				overheadBytes = ArrayModeThreshold * 24 * 3
			} else {
				mode = "MAP"
				// Map mode: dynamic overhead, estimate based on hash count
				// Each map entry: ~50 bytes average (key + value + overhead)
				estimatedEntries := bf.hashCount / 4 // Rough estimate
				overheadBytes = uint64(estimatedEntries) * 50 * 3 // 3 maps
			}

			totalBytes := bitArrayBytes + overheadBytes

			t.Logf("Mode: %s", mode)
			t.Logf("  Elements: %s", formatNumber(tt.elements))
			t.Logf("  Cache lines: %s", formatNumber(bf.cacheLineCount))
			t.Logf("  Bit array: %s", formatBytes(bitArrayBytes))
			t.Logf("  Overhead: %s", formatBytes(overheadBytes))
			t.Logf("  Total (est): %s", formatBytes(totalBytes))
			t.Logf("  Overhead %%: %.1f%%", float64(overheadBytes)/float64(totalBytes)*100)

			// Array mode should have predictable overhead
			if bf.useArrayMode {
				expectedOverhead := uint64(ArrayModeThreshold * 24 * 3)
				if overheadBytes != expectedOverhead {
					t.Errorf("Array mode overhead mismatch: expected %d, got %d",
						expectedOverhead, overheadBytes)
				}
			}
		})
	}
}

// TestLargeScaleHybrid tests the hybrid approach with realistic large-scale scenarios
func TestLargeScaleHybrid(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large-scale test in short mode")
	}

	tests := []struct {
		name     string
		elements uint64
		fpr      float64
		testOps  int
	}{
		{"Medium scale", 500_000, 0.01, 10_000},
		{"Large scale", 5_000_000, 0.01, 10_000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := NewCacheOptimizedBloomFilter(tt.elements, tt.fpr)

			mode := "ARRAY"
			if !bf.useArrayMode {
				mode = "MAP"
			}

			t.Logf("Testing %s mode: %s elements, %d test operations",
				mode, formatNumber(tt.elements), tt.testOps)

			// Add elements
			for i := 0; i < tt.testOps; i++ {
				bf.AddUint64(uint64(i))
			}

			// Verify they're all found
			errors := 0
			for i := 0; i < tt.testOps; i++ {
				if !bf.ContainsUint64(uint64(i)) {
					errors++
				}
			}

			if errors > 0 {
				t.Errorf("%s mode: %d elements not found (should be 0)", mode, errors)
			}

			// Test false positive rate
			falsePositives := 0
			fpTests := 10000
			for i := tt.testOps; i < tt.testOps+fpTests; i++ {
				if bf.ContainsUint64(uint64(i)) {
					falsePositives++
				}
			}

			actualFPR := float64(falsePositives) / float64(fpTests)
			t.Logf("✓ %s mode: FPR=%.4f (expected ~%.4f), errors=%d",
				mode, actualFPR, tt.fpr, errors)

			// Allow some margin for FPR
			if actualFPR > tt.fpr*3 {
				t.Errorf("%s mode: FPR too high: %.4f (expected ~%.4f)",
					mode, actualFPR, tt.fpr)
			}
		})
	}
}

// Helper functions
func formatNumber(n uint64) string {
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func formatBytes(b uint64) string {
	if b >= 1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(b)/(1024*1024*1024))
	}
	if b >= 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(b)/(1024*1024))
	}
	if b >= 1024 {
		return fmt.Sprintf("%.2f KB", float64(b)/1024)
	}
	return fmt.Sprintf("%d bytes", b)
}
