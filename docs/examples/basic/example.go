package main

import (
	"fmt"
	"runtime"

	bf "github.com/shaia/go-simd-bloomfilter"
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

	// Example usage
	fmt.Println("\nExample Usage:")
	fmt.Println("--------------")

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
	fmt.Printf("SIMD capabilities: AVX2=%t, AVX512=%t, NEON=%t\n", stats.HasAVX2, stats.HasAVX512, stats.HasNEON)
}
