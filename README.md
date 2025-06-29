# SIMD-Optimized Bloom Filter

A high-performance, cache-line optimized bloom filter implementation in Go with SIMD acceleration.

## Features

- **SIMD Optimizations**: Automatic detection and usage of AVX2, AVX512, and ARM NEON instructions
- **Cache-Line Optimized**: 64-byte aligned memory structures for optimal cache performance
- **Cross-Platform**: Supports x86_64 (Intel/AMD) and ARM64 architectures
- **High Performance**: Vectorized operations with unrolled loops
- **Memory Efficient**: Cache-line aware memory allocation and bulk operations

## Package Structure

``` bash
├── bloomfilter.go               # Core implementation 
├── bloomfilter_test.go          # Comprehensive benchmarks
├── simd_*.go                    # Platform-specific SIMD optimizations
├── docs/examples/               # Usage examples
│   └── basic/                   # Basic usage example
│       └── main.go              # Simple demonstration
└── go.mod                       # Module definition
```

## Installation

```bash
go get github.com/shaia/go-simd-bloomfilter
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    bf "github.com/shaia/go-simd-bloomfilter"
)

func main() {
    // Create a bloom filter for 1M elements with 1% false positive rate
    filter := bf.NewCacheOptimizedBloomFilter(1000000, 0.01)
    
    // Add elements
    filter.AddString("example")
    filter.AddUint64(42)
    
    // Check membership
    fmt.Println(filter.ContainsString("example"))  // true
    fmt.Println(filter.ContainsString("missing"))  // false (probably)
    
    // Get statistics
    stats := filter.GetCacheStats()
    fmt.Printf("SIMD enabled: %t\n", stats.SIMDEnabled)
    fmt.Printf("Memory usage: %d bytes\n", stats.MemoryUsage)
}
```

### SIMD Capabilities

```go
// Check SIMD support
fmt.Printf("AVX2: %t\n", bf.HasAVX2())
fmt.Printf("AVX512: %t\n", bf.HasAVX512()) 
fmt.Printf("NEON: %t\n", bf.HasNEON())
fmt.Printf("Any SIMD: %t\n", bf.HasSIMD())
```

### Bulk Operations

```go
filter1 := bf.NewCacheOptimizedBloomFilter(100000, 0.01)
filter2 := bf.NewCacheOptimizedBloomFilter(100000, 0.01)

// Add some data to filters...

// Union (SIMD optimized)
filter1.Union(filter2)

// Intersection (SIMD optimized) 
filter1.Intersection(filter2)

// Population count (SIMD optimized)
bitsSet := filter1.PopCount()
```

## Performance

### Benchmarks

Run comprehensive benchmarks:

```bash
go test -bench=. -v
```

### Results (Apple M3 Pro with NEON)

- **Insertions**: ~2.6M operations/second
- **Lookups**: ~2.6M operations/second  
- **Memory**: Perfect cache-line alignment (0 offset)
- **False Positive Rate**: ~1.05% (target: 1.0%)

## SIMD Optimizations

### Automatic Detection
- **x86_64**: AVX2 and AVX512 support detection
- **ARM64**: NEON support (enabled by default)
- **Fallback**: Optimized scalar implementations

### Vectorized Operations

- **Hash Functions**: 32-byte chunk processing (4x uint64 simultaneously)
- **Population Count**: Unrolled cache-line processing
- **Bulk Operations**: Vectorized Union, Intersection, Clear
- **Memory Access**: Cache-line grouped operations

### Cache Optimization

- **Alignment**: 64-byte cache-line aligned memory
- **Prefetching**: Memory access hints for better cache utilization
- **Bulk Processing**: Entire cache-lines (512 bits) processed together

## API Reference

### Types

```go
type CacheOptimizedBloomFilter struct { ... }

type CacheStats struct {
    BitCount       uint64
    HashCount      uint32
    BitsSet        uint64
    LoadFactor     float64
    EstimatedFPP   float64
    CacheLineCount uint64
    CacheLineSize  int
    MemoryUsage    uint64
    Alignment      uintptr
    HasAVX2        bool
    HasAVX512      bool
    HasNEON        bool
    SIMDEnabled    bool
}
```

### Functions

```go
// Constructor
func NewCacheOptimizedBloomFilter(expectedElements uint64, falsePositiveRate float64) *CacheOptimizedBloomFilter

// Core operations
func (bf *CacheOptimizedBloomFilter) Add(data []byte)
func (bf *CacheOptimizedBloomFilter) Contains(data []byte) bool
func (bf *CacheOptimizedBloomFilter) AddString(s string)
func (bf *CacheOptimizedBloomFilter) ContainsString(s string) bool
func (bf *CacheOptimizedBloomFilter) AddUint64(n uint64)
func (bf *CacheOptimizedBloomFilter) ContainsUint64(n uint64) bool

// Bulk operations
func (bf *CacheOptimizedBloomFilter) Union(other *CacheOptimizedBloomFilter) error
func (bf *CacheOptimizedBloomFilter) Intersection(other *CacheOptimizedBloomFilter) error
func (bf *CacheOptimizedBloomFilter) Clear()
func (bf *CacheOptimizedBloomFilter) PopCount() uint64

// Statistics
func (bf *CacheOptimizedBloomFilter) GetCacheStats() CacheStats
func (bf *CacheOptimizedBloomFilter) EstimatedFPP() float64

// SIMD capabilities
func HasAVX2() bool
func HasAVX512() bool
func HasNEON() bool
func HasSIMD() bool
```

## Architecture Support

- **x86_64 (amd64)**: Intel and AMD processors with AVX2/AVX512
- **ARM64**: Apple Silicon (M-series) and other ARM64 with NEON
- **Other**: Optimized scalar fallback implementations

## License

MIT License - see LICENSE file for details.
