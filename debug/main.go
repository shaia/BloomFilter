package main

import (
	"fmt"
	"unsafe"

	bf "github.com/shaia/go-simd-bloomfilter"
)

func main() {
	fmt.Println("🎯 Assembly Debugging Demo")

	// Create a small bloom filter for easier debugging
	filter := bf.NewCacheOptimizedBloomFilter(100, 0.01)

	// Add some test data
	testData := []string{"apple", "banana", "cherry"}
	for _, item := range testData {
		fmt.Printf("Adding: %s\n", item)
		filter.AddString(item)
	}

	// Get SIMD stats
	stats := filter.GetCacheStats()
	fmt.Printf("\n📊 SIMD Stats:\n")
	fmt.Printf("  NEON Available: %v\n", stats.HasNEON)
	fmt.Printf("  AVX2 Available: %v\n", stats.HasAVX2)
	fmt.Printf("  SIMD Enabled: %v\n", stats.SIMDEnabled)

	// This will call the SIMD PopCount function - good breakpoint location
	bitsSet := filter.PopCount()
	fmt.Printf("  Bits Set: %d\n", bitsSet)

	// Test SIMD Union operation - another good breakpoint location
	filter2 := bf.NewCacheOptimizedBloomFilter(100, 0.01)
	filter2.AddString("date")
	filter2.AddString("elderberry")

	fmt.Printf("\n🔄 Performing SIMD Union...\n")
	err := filter.Union(filter2)
	if err != nil {
		fmt.Printf("Union error: %v\n", err)
	}

	finalBitsSet := filter.PopCount()
	fmt.Printf("  Final Bits Set: %d\n", finalBitsSet)

	// Test the core SIMD functions directly for debugging
	fmt.Printf("\n🔧 Direct SIMD Function Test:\n")
	testDirectSIMD()
}

func testDirectSIMD() {
	// Create test data that we can easily track in assembly
	testData := make([]uint64, 8)    // 64 bytes = 1 cache line
	testData[0] = 0xAAAAAAAAAAAAAAAA // Alternating bits
	testData[1] = 0x5555555555555555 // Opposite pattern
	testData[2] = 0xFFFFFFFFFFFFFFFF // All bits set
	testData[3] = 0x0000000000000000 // No bits set
	testData[4] = 0x1111111111111111 // Every 4th bit
	testData[5] = 0x8888888888888888 // Every 4th bit shifted
	testData[6] = 0xF0F0F0F0F0F0F0F0 // Alternating nibbles
	testData[7] = 0x0F0F0F0F0F0F0F0F // Opposite nibbles

	// Get SIMD operations
	simdOps := bf.GetSIMDOperations()

	// This will call neonPopCount - perfect for setting breakpoints
	fmt.Printf("  Testing PopCount on test data...\n")
	count := simdOps.PopCount(unsafe.Pointer(&testData[0]), len(testData)*8)
	fmt.Printf("  PopCount result: %d bits\n", count)

	// Test VectorOr operation
	testData2 := make([]uint64, 8)
	testData2[0] = 0x1111111111111111
	testData2[1] = 0x2222222222222222
	testData2[2] = 0x4444444444444444
	testData2[3] = 0x8888888888888888

	fmt.Printf("  Testing VectorOr operation...\n")
	simdOps.VectorOr(
		unsafe.Pointer(&testData[0]),
		unsafe.Pointer(&testData2[0]),
		len(testData)*8,
	)

	// Check result
	newCount := simdOps.PopCount(unsafe.Pointer(&testData[0]), len(testData)*8)
	fmt.Printf("  PopCount after OR: %d bits\n", newCount)

	fmt.Printf("  Direct SIMD test completed!\n")
}
