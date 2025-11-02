package main

import (
	"fmt"
	"runtime"

	bf "github.com/shaia/BloomFilter"
)

// Build information, set via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
	BuildUser = "unknown"
)

func main() {
	fmt.Println("Cache Line Optimized Bloom Filter")
	fmt.Println("=================================")

	// Show version information
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Commit: %s\n", Commit)
	fmt.Printf("Build Date: %s\n", BuildDate)
	fmt.Printf("Build User: %s\n", BuildUser)

	// Show system information
	fmt.Printf("System: GOMAXPROCS=%d, NumCPU=%d\n", runtime.GOMAXPROCS(0), runtime.NumCPU())
	fmt.Printf("Cache line size: %d bytes\n", bf.CacheLineSize)
	fmt.Printf("Words per cache line: %d\n", bf.WordsPerCacheLine)
	fmt.Printf("Bits per cache line: %d\n", bf.BitsPerCacheLine)

	// Show SIMD capabilities
	fmt.Printf("\nSIMD Capabilities:\n")
	fmt.Printf("AVX2: %t\n", bf.HasAVX2())
	fmt.Printf("AVX512: %t\n", bf.HasAVX512())
	fmt.Printf("NEON: %t\n", bf.HasNEON())
	fmt.Printf("SIMD Enabled: %t\n\n", bf.HasAVX2() || bf.HasAVX512() || bf.HasNEON())

	// Example 1: Basic usage
	fmt.Println("\nExample 1: Basic Usage")
	fmt.Println("----------------------")

	filter := bf.NewCacheOptimizedBloomFilter(10000, 0.001)

	filter.AddString("cache_optimized")
	filter.AddString("bloom_filter")
	filter.AddUint64(42)

	fmt.Printf("Contains 'cache_optimized': %t\n", filter.ContainsString("cache_optimized"))
	fmt.Printf("Contains 'not_present': %t\n", filter.ContainsString("not_present"))
	fmt.Printf("Contains 42: %t\n", filter.ContainsUint64(42))

	stats := filter.GetCacheStats()
	fmt.Printf("Memory aligned: %t\n", stats.Alignment == 0)
	fmt.Printf("Cache lines used: %d\n", stats.CacheLineCount)
	fmt.Printf("SIMD optimized: %t\n", stats.SIMDEnabled)

	// Example 2: Multiple operations (zero allocations)
	fmt.Println("\nExample 2: Multiple Operations")
	fmt.Println("-------------------------------")

	filter2 := bf.NewCacheOptimizedBloomFilter(100000, 0.01)

	// Add multiple strings (zero allocations per operation)
	urls := []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/page3",
	}
	for _, url := range urls {
		filter2.AddString(url)
	}

	// Add multiple uint64s (zero allocations per operation)
	userIDs := []uint64{1001, 1002, 1003, 1004, 1005}
	for _, id := range userIDs {
		filter2.AddUint64(id)
	}

	fmt.Printf("Contains 'https://example.com/page2': %t\n", filter2.ContainsString("https://example.com/page2"))
	fmt.Printf("Contains user ID 1003: %t\n", filter2.ContainsUint64(1003))
	fmt.Printf("Contains user ID 9999: %t\n", filter2.ContainsUint64(9999))

	// Example 3: Thread-safe concurrent operations
	fmt.Println("\nExample 3: Thread-Safe Concurrent Operations")
	fmt.Println("--------------------------------------------")

	filter3 := bf.NewCacheOptimizedBloomFilter(100000, 0.01)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		filter3.AddUint64(uint64(i))
	}

	// Concurrent writes (safe with atomic operations)
	done := make(chan bool, 10)
	for g := 0; g < 10; g++ {
		go func(goroutineID int) {
			for i := 0; i < 100; i++ {
				filter3.AddString(fmt.Sprintf("goroutine_%d_key_%d", goroutineID, i))
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	fmt.Printf("Thread-safe writes completed successfully\n")
	fmt.Printf("Contains 'goroutine_5_key_50': %t\n", filter3.ContainsString("goroutine_5_key_50"))

	finalStats := filter3.GetCacheStats()
	fmt.Printf("\nFinal Statistics:\n")
	fmt.Printf("  Total bits: %d\n", finalStats.BitCount)
	fmt.Printf("  Bits set: %d\n", finalStats.BitsSet)
	fmt.Printf("  Load factor: %.2f%%\n", finalStats.LoadFactor*100)
	fmt.Printf("  Estimated FPP: %.6f\n", finalStats.EstimatedFPP)
}
