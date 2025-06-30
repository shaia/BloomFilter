package bloomfilter

import (
	"fmt"
	"testing"
)

// TestBasicFunctionality tests core bloom filter operations
func TestBasicFunctionality(t *testing.T) {
	// Create a small bloom filter for testing
	bf := NewCacheOptimizedBloomFilter(1000, 0.01)

	// Test string operations
	testStrings := []string{"apple", "banana", "cherry", "date", "elderberry"}

	// Add strings
	for _, str := range testStrings {
		bf.AddString(str)
	}

	// Test that all added strings are found
	for _, str := range testStrings {
		if !bf.ContainsString(str) {
			t.Errorf("Expected to find string '%s' but it was not found", str)
		}
	}

	// Test that non-added strings might not be found (though false positives are possible)
	nonAddedStrings := []string{"grape", "honeydew", "kiwi"}
	falsePositives := 0
	for _, str := range nonAddedStrings {
		if bf.ContainsString(str) {
			falsePositives++
		}
	}

	// With a 1% FPP and only a few test strings, we shouldn't have many false positives
	if falsePositives > 1 {
		t.Logf("Warning: Higher than expected false positives (%d out of %d)", falsePositives, len(nonAddedStrings))
	}
}

// TestUint64Operations tests uint64 specific operations
func TestUint64Operations(t *testing.T) {
	bf := NewCacheOptimizedBloomFilter(1000, 0.01)

	testNumbers := []uint64{1, 42, 100, 999, 12345, 0xDEADBEEF}

	// Add numbers
	for _, num := range testNumbers {
		bf.AddUint64(num)
	}

	// Test that all added numbers are found
	for _, num := range testNumbers {
		if !bf.ContainsUint64(num) {
			t.Errorf("Expected to find number %d but it was not found", num)
		}
	}

	// Test some numbers that weren't added
	nonAddedNumbers := []uint64{2, 43, 101, 1000}
	falsePositives := 0
	for _, num := range nonAddedNumbers {
		if bf.ContainsUint64(num) {
			falsePositives++
		}
	}

	t.Logf("False positives for uint64: %d out of %d", falsePositives, len(nonAddedNumbers))
}

// TestByteOperations tests raw byte slice operations
func TestByteOperations(t *testing.T) {
	bf := NewCacheOptimizedBloomFilter(1000, 0.01)

	testData := [][]byte{
		[]byte("hello"),
		[]byte("world"),
		[]byte{0x01, 0x02, 0x03},
		[]byte{0xFF, 0xFE, 0xFD},
		[]byte(""), // empty byte slice
	}

	// Add byte slices
	for _, data := range testData {
		bf.Add(data)
	}

	// Test that all added data is found
	for i, data := range testData {
		if !bf.Contains(data) {
			t.Errorf("Expected to find data at index %d but it was not found", i)
		}
	}
}

// TestClearOperation tests the clear functionality
func TestClearOperation(t *testing.T) {
	bf := NewCacheOptimizedBloomFilter(1000, 0.01)

	// Add some data
	testStrings := []string{"test1", "test2", "test3"}
	for _, str := range testStrings {
		bf.AddString(str)
	}

	// Verify data is there
	for _, str := range testStrings {
		if !bf.ContainsString(str) {
			t.Errorf("Data should be present before clear: %s", str)
		}
	}

	// Get initial bit count
	initialBits := bf.PopCount()
	if initialBits == 0 {
		t.Error("Expected some bits to be set before clear")
	}

	// Clear the filter
	bf.Clear()

	// Verify data is gone
	for _, str := range testStrings {
		if bf.ContainsString(str) {
			t.Errorf("Data should not be present after clear: %s", str)
		}
	}

	// Verify bit count is zero
	clearedBits := bf.PopCount()
	if clearedBits != 0 {
		t.Errorf("Expected 0 bits after clear, got %d", clearedBits)
	}
}

// TestPopCount tests population count functionality
func TestPopCount(t *testing.T) {
	bf := NewCacheOptimizedBloomFilter(1000, 0.01)

	// Initially should have 0 bits set
	if count := bf.PopCount(); count != 0 {
		t.Errorf("Expected 0 bits initially, got %d", count)
	}

	// Add some elements
	bf.AddString("test1")
	count1 := bf.PopCount()
	if count1 == 0 {
		t.Error("Expected some bits to be set after adding element")
	}

	// Add more elements
	bf.AddString("test2")
	bf.AddString("test3")
	count2 := bf.PopCount()

	// Should have at least as many bits set (possibly more due to hash collisions)
	if count2 < count1 {
		t.Errorf("Expected bit count to increase or stay same, got %d then %d", count1, count2)
	}

	t.Logf("Bit count progression: 0 -> %d -> %d", count1, count2)
}

// TestUnionOperation tests union of two bloom filters
func TestUnionOperation(t *testing.T) {
	bf1 := NewCacheOptimizedBloomFilter(1000, 0.01)
	bf2 := NewCacheOptimizedBloomFilter(1000, 0.01)

	// Add different data to each filter
	set1 := []string{"apple", "banana", "cherry"}
	set2 := []string{"date", "elderberry", "fig"}

	for _, str := range set1 {
		bf1.AddString(str)
	}

	for _, str := range set2 {
		bf2.AddString(str)
	}

	// Union bf2 into bf1
	err := bf1.Union(bf2)
	if err != nil {
		t.Fatalf("Union operation failed: %v", err)
	}

	// bf1 should now contain elements from both sets
	allElements := append(set1, set2...)
	for _, str := range allElements {
		if !bf1.ContainsString(str) {
			t.Errorf("Expected to find '%s' in union result", str)
		}
	}
}

// TestIntersectionOperation tests intersection of two bloom filters
func TestIntersectionOperation(t *testing.T) {
	bf1 := NewCacheOptimizedBloomFilter(1000, 0.01)
	bf2 := NewCacheOptimizedBloomFilter(1000, 0.01)

	// Add overlapping data to both filters
	set1 := []string{"apple", "banana", "cherry", "shared1", "shared2"}
	set2 := []string{"date", "elderberry", "fig", "shared1", "shared2"}

	for _, str := range set1 {
		bf1.AddString(str)
	}

	for _, str := range set2 {
		bf2.AddString(str)
	}

	// Intersect bf2 with bf1
	err := bf1.Intersection(bf2)
	if err != nil {
		t.Fatalf("Intersection operation failed: %v", err)
	}

	// bf1 should contain shared elements
	sharedElements := []string{"shared1", "shared2"}
	for _, str := range sharedElements {
		if !bf1.ContainsString(str) {
			t.Errorf("Expected to find shared element '%s' in intersection result", str)
		}
	}

	// Note: Due to false positives, non-shared elements might still be found
	// so we don't test for their absence
}

// TestMismatchedSizeOperations tests error handling for mismatched filter sizes
func TestMismatchedSizeOperations(t *testing.T) {
	bf1 := NewCacheOptimizedBloomFilter(1000, 0.01)
	bf2 := NewCacheOptimizedBloomFilter(2000, 0.01) // Different size

	// Union should fail
	err := bf1.Union(bf2)
	if err == nil {
		t.Error("Expected error when unioning filters of different sizes")
	}

	// Intersection should fail
	err = bf1.Intersection(bf2)
	if err == nil {
		t.Error("Expected error when intersecting filters of different sizes")
	}
}

// TestCacheStatistics tests the statistics functionality
func TestCacheStatistics(t *testing.T) {
	bf := NewCacheOptimizedBloomFilter(1000, 0.01)

	// Add some data
	for i := 0; i < 100; i++ {
		bf.AddString(fmt.Sprintf("test_%d", i))
	}

	stats := bf.GetCacheStats()

	// Basic sanity checks
	if stats.BitCount == 0 {
		t.Error("Expected non-zero bit count")
	}

	if stats.HashCount == 0 {
		t.Error("Expected non-zero hash count")
	}

	if stats.CacheLineCount == 0 {
		t.Error("Expected non-zero cache line count")
	}

	if stats.MemoryUsage == 0 {
		t.Error("Expected non-zero memory usage")
	}

	if stats.BitsSet == 0 {
		t.Error("Expected some bits to be set")
	}

	if stats.LoadFactor < 0 || stats.LoadFactor > 1 {
		t.Errorf("Load factor should be between 0 and 1, got %f", stats.LoadFactor)
	}

	if stats.EstimatedFPP < 0 || stats.EstimatedFPP > 1 {
		t.Errorf("Estimated FPP should be between 0 and 1, got %f", stats.EstimatedFPP)
	}

	t.Logf("Stats: BitCount=%d, HashCount=%d, BitsSet=%d, LoadFactor=%.4f, EstimatedFPP=%.4f",
		stats.BitCount, stats.HashCount, stats.BitsSet, stats.LoadFactor, stats.EstimatedFPP)
}

// TestSIMDFunctionality tests SIMD-specific operations
func TestSIMDFunctionality(t *testing.T) {
	bf := NewCacheOptimizedBloomFilter(1000, 0.01)

	// Add some data
	testData := []string{"simd1", "simd2", "simd3", "simd4", "simd5"}
	for _, str := range testData {
		bf.AddString(str)
	}

	// Test PopCount (now uses SIMD automatically)
	count1 := bf.PopCount()
	count2 := bf.PopCount()

	if count1 != count2 {
		t.Errorf("PopCount inconsistent: first=%d, second=%d", count1, count2)
	}

	// Test SIMD operations with another filter
	bf2 := NewCacheOptimizedBloomFilter(1000, 0.01)
	bf2.AddString("simd6")
	bf2.AddString("simd7")

	// Test Union (now uses SIMD automatically)
	err := bf.Union(bf2)
	if err != nil {
		t.Errorf("Union failed: %v", err)
	}

	// Verify union worked
	if !bf.ContainsString("simd6") || !bf.ContainsString("simd7") {
		t.Error("Union did not properly merge filters")
	}

	// Test Clear (now uses SIMD automatically)
	bf.Clear()
	if bf.PopCount() != 0 {
		t.Error("Clear did not properly clear the filter")
	}
}

// TestFalsePositiveRate tests that the false positive rate is reasonable
func TestFalsePositiveRate(t *testing.T) {
	targetFPP := 0.01 // 1%
	bf := NewCacheOptimizedBloomFilter(10000, targetFPP)

	// Add known elements
	numElements := 5000
	for i := 0; i < numElements; i++ {
		bf.AddString(fmt.Sprintf("element_%d", i))
	}

	// Test false positive rate with non-added elements
	numTests := 10000
	falsePositives := 0

	for i := numElements; i < numElements+numTests; i++ {
		if bf.ContainsString(fmt.Sprintf("element_%d", i)) {
			falsePositives++
		}
	}

	actualFPP := float64(falsePositives) / float64(numTests)

	// Allow some tolerance (2x the target rate should be reasonable)
	maxAllowedFPP := targetFPP * 2

	if actualFPP > maxAllowedFPP {
		t.Errorf("False positive rate too high: actual=%.4f, target=%.4f, max_allowed=%.4f",
			actualFPP, targetFPP, maxAllowedFPP)
	}

	t.Logf("False positive rate test: actual=%.4f%%, target=%.4f%%, elements=%d, tests=%d",
		actualFPP*100, targetFPP*100, numElements, numTests)
}
