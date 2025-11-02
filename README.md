# SIMD-Optimized Bloom Filter

A high-performance, cache-line optimized bloom filter implementation in Go with hardware-accelerated SIMD operations and lock-free atomic operations.

**Optimized for performance and simplicity** with zero-allocation operations and built-in thread-safety.

## Features

- **Thread-Safe**: Lock-free concurrent operations using atomic CAS operations (no external locks required)
- **Zero Allocations**: Stack-based buffers for typical use cases (hashCount ≤ 16, covering 99% of scenarios)
- **SIMD Acceleration**: Automatic detection and usage of AVX2, AVX512, and ARM NEON instructions
- **Cache-Optimized**: 64-byte aligned memory structures for optimal CPU cache performance
- **Cross-Platform**: Supports x86_64 (Intel/AMD) and ARM64 architectures
- **High Performance**: 3-4x faster than popular alternatives (willf/bloom)
- **Simple Architecture**: ~400 lines of clean, maintainable code
- **Production Ready**: Comprehensive test suite with race detection and correctness validation

## Performance

### Real-World Performance (Intel i9-13980HX)

| Operation | Simplified Atomic | Willf/Bloom | Thread-Safe Pool | Winner |
|-----------|------------------|-------------|------------------|---------|
| **Add** | 26.02 ns/op | 85.64 ns/op | ~400 ns/op | **3-15x faster** |
| **Contains** | 23.41 ns/op | 90.34 ns/op | ~600 ns/op | **4-26x faster** |
| **AddUint64** | 20.16 ns/op | N/A | ~350 ns/op | **17x faster** |
| **Allocations** | **0 B/op** | 97 B/op | 17 B/op | **100% saved** |

### Throughput (1M elements, 0.01 FPR)

- **Insertions**: 18.6 million operations/second
- **Lookups**: 35.8 million operations/second
- **Memory**: Zero allocations on hot path
- **False Positive Rate**: 1.02% (target: 1.0%)

### SIMD Speedup

| Operation | Size | SIMD | Fallback | Speedup |
|-----------|------|------|----------|---------|
| PopCount | 1KB | 40 ns | 125 ns | **3.1x** |
| PopCount | 64KB | 2.6 µs | 7.7 µs | **3.0x** |
| VectorOr | 1KB | 10 ns | 37 ns | **3.7x** |
| VectorOr | 64KB | 1.1 µs | 2.4 µs | **2.2x** |
| VectorAnd | 16KB | 162 ns | 502 ns | **3.1x** |

*Benchmarked on Intel i9-13980HX with AVX2*

## Installation

```bash
go get github.com/shaia/BloomFilter
```

## Best Use Cases

This library is **optimized for high-performance applications** where speed and memory efficiency are critical:

**Ideal For:**
- **High-frequency operations**: Millions of operations per second with zero allocations
- **Multi-threaded applications**: Built-in thread-safety without external locks
- **Microservices**: Per-request or per-session filtering with minimal overhead
- **Rate limiting**: Token buckets, request deduplication
- **Cache systems**: Bloom filter for cache existence checks
- **Real-time streaming**: Per-connection or per-stream filters
- **API gateways**: Request deduplication, idempotency checks
- **Data processing**: Any size from small (10K) to large (100M+) elements

**Performance Characteristics:**
- Small filters (10K-100K): 26 ns/op, zero allocations
- Large filters (1M-10M+): 26 ns/op, zero allocations
- SIMD operations: 2-4x faster for bulk operations (Union, Intersection, PopCount)
- Thread-safe: No lock contention, scales with CPU cores

## Quick Start

```go
package main

import (
    "fmt"
    bf "github.com/shaia/BloomFilter"
)

func main() {
    // Create a bloom filter for 1M elements with 1% false positive rate
    filter := bf.NewCacheOptimizedBloomFilter(1000000, 0.01)

    // Add elements (thread-safe, zero allocations)
    filter.AddString("example")
    filter.AddUint64(42)
    filter.Add([]byte("custom data"))

    // Check membership (thread-safe, zero allocations)
    fmt.Println(filter.ContainsString("example"))  // true
    fmt.Println(filter.ContainsString("missing"))  // false (probably)
    fmt.Println(filter.ContainsUint64(42))         // true

    // Get statistics
    stats := filter.GetCacheStats()
    fmt.Printf("SIMD enabled: %t\n", stats.SIMDEnabled)
    fmt.Printf("Memory usage: %d bytes\n", stats.MemoryUsage)
    fmt.Printf("Load factor: %.2f%%\n", stats.LoadFactor * 100)
    fmt.Printf("Estimated FPP: %.4f%%\n", stats.EstimatedFPP * 100)
}
```

## Package Structure

```
BloomFilter/
├── bloomfilter.go              # Core bloom filter API (public interface)
├── *_test.go                   # Comprehensive test suite
├── internal/                   # Internal implementation (not importable by users)
│   ├── hash/                   # Hash function implementations
│   │   └── hash.go            # FNV-1a and variant hash functions
│   └── simd/                   # SIMD package (architecture-specific)
│       ├── simd.go            # Interface & runtime detection
│       ├── fallback.go        # Optimized scalar implementation
│       ├── amd64/             # x86-64 SIMD (AVX2)
│       │   ├── avx2.go       # Assembly declarations
│       │   └── avx2.s        # AVX2 assembly code
│       └── arm64/             # ARM64 SIMD (NEON)
│           ├── neon_asm.go   # Assembly declarations
│           └── neon.s        # NEON assembly code
├── docs/examples/             # Usage examples
│   └── basic/example.go      # Complete example
└── tests/                     # Test suite
    ├── benchmark/            # Performance benchmarks
    └── integration/          # Integration tests
```

**Note:** The `internal/` package follows Go conventions - it cannot be imported by external packages, ensuring a clean public API while allowing internal refactoring without breaking changes.

## Usage Examples

### Thread-Safe Concurrent Usage

```go
// No locks needed - built-in thread safety!
filter := bf.NewCacheOptimizedBloomFilter(1000000, 0.01)

// Safe to use from multiple goroutines
go func() {
    for i := 0; i < 1000; i++ {
        filter.AddUint64(uint64(i))
    }
}()

go func() {
    for i := 0; i < 1000; i++ {
        exists := filter.ContainsUint64(uint64(i))
        fmt.Println(exists)
    }
}()
```

### SIMD Capabilities Detection

```go
// Check what SIMD instructions are available
fmt.Printf("AVX2: %t\n", bf.HasAVX2())       // Intel/AMD x86-64
fmt.Printf("AVX512: %t\n", bf.HasAVX512())   // High-end Intel
fmt.Printf("NEON: %t\n", bf.HasNEON())       // ARM64 (Apple Silicon, etc.)
fmt.Printf("Any SIMD: %t\n", bf.HasSIMD())   // Any acceleration available
```

### Bulk Operations (SIMD Optimized)

```go
filter1 := bf.NewCacheOptimizedBloomFilter(100000, 0.01)
filter2 := bf.NewCacheOptimizedBloomFilter(100000, 0.01)

// Add data to filters...
filter1.AddString("shared")
filter2.AddString("shared")

// Union - combine two filters (SIMD accelerated)
filter1.Union(filter2)

// Intersection - keep only common elements (SIMD accelerated)
filter1.Intersection(filter2)

// Population count - count set bits (SIMD accelerated)
bitsSet := filter1.PopCount()

// Clear all bits (SIMD accelerated)
filter1.Clear()
```

### Statistics and Monitoring

```go
stats := filter.GetCacheStats()

fmt.Printf("Bit count: %d\n", stats.BitCount)
fmt.Printf("Hash functions: %d\n", stats.HashCount)
fmt.Printf("Bits set: %d\n", stats.BitsSet)
fmt.Printf("Load factor: %.4f\n", stats.LoadFactor)
fmt.Printf("Estimated FPP: %.6f\n", stats.EstimatedFPP)
fmt.Printf("Cache lines: %d\n", stats.CacheLineCount)
fmt.Printf("Memory usage: %d bytes\n", stats.MemoryUsage)
fmt.Printf("Memory aligned: %t\n", stats.Alignment == 0)

// SIMD capabilities
fmt.Printf("AVX2: %t, AVX512: %t, NEON: %t\n",
    stats.HasAVX2, stats.HasAVX512, stats.HasNEON)
fmt.Printf("SIMD enabled: %t\n", stats.SIMDEnabled)
```

## Building and Testing

### Build

```bash
# Build the library
go build

# Run example
go run docs/examples/basic/example.go
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run benchmarks
go test -bench=. -benchmem ./tests/benchmark/...

# Run with race detector
go test -race -v ./...

# Run integration tests
go test -v ./tests/integration/...
```

## SIMD Implementation Details

### Automatic Platform Detection

The library automatically detects and uses the best available SIMD instructions:

1. **x86_64 (amd64)**:
   - AVX2 (256-bit vectors, 32 bytes at a time)
   - AVX512 (512-bit vectors, 64 bytes at a time) - placeholder
   - Fallback to optimized scalar

2. **ARM64**:
   - NEON (128-bit vectors, 16 bytes at a time)
   - Fallback to optimized scalar

3. **Other architectures**:
   - Optimized scalar implementation using bit manipulation

### Vectorized Operations

All critical operations are SIMD-accelerated:

- **PopCount**: Count set bits using vector instructions
- **VectorOr**: Bitwise OR for Union operations
- **VectorAnd**: Bitwise AND for Intersection operations
- **VectorClear**: Fast memory zeroing

### Assembly Implementation

- **AMD64**: Hand-written AVX2 assembly in [internal/simd/amd64/avx2.s](internal/simd/amd64/avx2.s)
- **ARM64**: Hand-written NEON assembly in [internal/simd/arm64/neon.s](internal/simd/arm64/neon.s)
- Clean separation between Go and assembly code
- Platform-specific build tags ensure correct compilation

## API Reference

### Core Types

```go
type CacheOptimizedBloomFilter struct {
    // Internal fields (cache-line aligned)
}

type CacheStats struct {
    BitCount       uint64   // Total bits in filter
    HashCount      uint32   // Number of hash functions
    BitsSet        uint64   // Current bits set
    LoadFactor     float64  // Ratio of bits set
    EstimatedFPP   float64  // Estimated false positive probability
    CacheLineCount uint64   // Number of cache lines
    CacheLineSize  int      // Size of cache line (64 bytes)
    MemoryUsage    uint64   // Total memory used
    Alignment      uintptr  // Memory alignment offset (0 = perfect)
    HasAVX2        bool     // AVX2 available
    HasAVX512      bool     // AVX512 available
    HasNEON        bool     // NEON available
    SIMDEnabled    bool     // Any SIMD enabled
}
```

### Constructor

```go
// Creates a new bloom filter optimized for cache performance
// Uses SIMD-accelerated operations and lock-free atomic operations for thread-safety
// Achieves zero allocations for typical use cases (hashCount ≤ 16, covering 99% of scenarios)
func NewCacheOptimizedBloomFilter(
    expectedElements uint64,    // Expected number of elements
    falsePositiveRate float64,  // Target false positive rate (0.0-1.0)
) *CacheOptimizedBloomFilter
```

### Core Methods

```go
// Add operations (thread-safe, lock-free, zero allocations)
func (bf *CacheOptimizedBloomFilter) Add(data []byte)
func (bf *CacheOptimizedBloomFilter) AddString(s string)
func (bf *CacheOptimizedBloomFilter) AddUint64(n uint64)

// Contains operations (thread-safe, lock-free, zero allocations)
func (bf *CacheOptimizedBloomFilter) Contains(data []byte) bool
func (bf *CacheOptimizedBloomFilter) ContainsString(s string) bool
func (bf *CacheOptimizedBloomFilter) ContainsUint64(n uint64) bool

// Bulk operations (SIMD accelerated, thread-safe)
func (bf *CacheOptimizedBloomFilter) Union(other *CacheOptimizedBloomFilter) error
func (bf *CacheOptimizedBloomFilter) Intersection(other *CacheOptimizedBloomFilter) error
func (bf *CacheOptimizedBloomFilter) Clear()
func (bf *CacheOptimizedBloomFilter) PopCount() uint64

// Statistics
func (bf *CacheOptimizedBloomFilter) GetCacheStats() CacheStats
func (bf *CacheOptimizedBloomFilter) EstimatedFPP() float64
```

### Global Functions

```go
// SIMD capability detection
func HasAVX2() bool    // Check for AVX2 support
func HasAVX512() bool  // Check for AVX512 support
func HasNEON() bool    // Check for NEON support
func HasSIMD() bool    // Check for any SIMD support
```

## Architecture Support

| Architecture | SIMD Support | Status |
|--------------|--------------|--------|
| x86_64 (Intel/AMD) | AVX2 | Implemented & Tested |
| x86_64 (Intel/AMD) | AVX512 | Placeholder |
| ARM64 (Apple Silicon) | NEON | Implemented |
| ARM64 (Other) | NEON | Implemented |
| Other | Scalar | Optimized Fallback |

## Performance Comparison

### Quick Summary

| Implementation | Add (ns/op) | Contains (ns/op) | Memory (B/op) | Speedup |
|----------------|-------------|------------------|---------------|---------|
| **Simplified Atomic** | 26.02 | 23.41 | 0 | Baseline |
| Thread-Safe Pool | ~400 | ~600 | 17 | 15-26x slower |
| willf/bloom | 85.64 | 90.34 | 97 | 3-4x slower |

### Detailed Comparisons

**Note:** Complete benchmark results and detailed comparisons are available in a separate benchmarking repository. The performance numbers shown above are from comprehensive testing on Intel i9-13980HX.

**Key Findings:**
- **vs willf/bloom**: 3-4x faster with zero allocations (vs 97 B/op)
- **vs Thread-Safe Pool (v0.3.0)**: 15-26x faster with 99.93% less memory usage
- **Throughput**: 18.6M insertions/sec, 35.8M lookups/sec
- **SIMD Acceleration**: 2-4x speedup for bulk operations (PopCount, Union, Intersection)

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass: `go test -v ./...`
2. Benchmarks show improvement: `go test -bench=. -benchmem`
3. Code is formatted: `go fmt ./...`
4. Race detector passes: `go test -race ./...`
5. SIMD correctness is validated: `go test -run=TestSIMDCorrectness`

## License

MIT License - see LICENSE file for details.

## Credits

- SIMD optimizations inspired by modern CPU architectures
- Cache-line optimization techniques from high-performance computing
- Bloom filter algorithm by Burton Howard Bloom (1970)
- Simplified atomic approach for maximum performance and simplicity
