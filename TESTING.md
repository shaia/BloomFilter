# Testing Guide

This document describes the testing structure and best practices for the BloomFilter project.

## Test Organization

The project follows Go best practices for test organization:

```
BloomFilter/
├── bloomfilter_test.go              # Core functionality tests
├── benchmark_test.go                # Performance benchmarks
├── simd_test.go                     # SIMD capability detection tests
├── storage_mode_test.go             # Storage mode selection tests (array vs map)
├── storage_mode_benchmark_test.go   # Storage mode performance benchmarks
└── tests/
    ├── README.md                    # Tests directory documentation
    └── integration/
        └── simd_comparison_test.go  # SIMD comparison integration tests (build tag: simd_comparison)
```

## Test Categories

### 1. Unit Tests (Root Package)

Located in the root package directory, these test individual components and functions.

**Files:**
- `bloomfilter_test.go` - Core Bloom filter operations
- `simd_test.go` - SIMD capability detection
- `storage_mode_test.go` - Storage mode selection logic

**Running:**
```bash
go test -v ./...
```

### 2. Benchmarks (Root Package)

Performance benchmarks for the main package functionality.

**Files:**
- `benchmark_test.go` - Comprehensive performance benchmarks
- `storage_mode_benchmark_test.go` - Array vs Map storage mode benchmarks

**Running:**
```bash
# All benchmarks
go test -bench=. -benchmem ./...

# Specific benchmark
go test -bench=BenchmarkInsertion -benchmem

# With CPU profiling
go test -bench=BenchmarkInsertion -cpuprofile=cpu.prof
```

### 3. Integration Tests (tests/integration/)

Special tests that require build tags or test cross-package functionality.

**Files:**
- `simd_comparison_test.go` - SIMD vs fallback performance validation

**Build Tag:** `simd_comparison`

**Running:**
```bash
# All integration tests
go test -tags=simd_comparison -v ./tests/integration

# Specific test
go test -tags=simd_comparison -v ./tests/integration -run=TestSIMDPerformanceImprovement

# Integration benchmarks
go test -tags=simd_comparison -bench=. ./tests/integration
```

## Running Tests

### Standard Test Suite

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Generate coverage report
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v -run=TestBasicFunctionality
```

### Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark pattern
go test -bench=BenchmarkSIMD -benchmem

# Save benchmark results
go test -bench=. -benchmem > results/benchmark_$(date +%Y%m%d_%H%M%S).txt

# Compare benchmarks (using benchstat)
go test -bench=. -count=10 > old.txt
# Make changes...
go test -bench=. -count=10 > new.txt
benchstat old.txt new.txt
```

### Integration Tests

```bash
# Run SIMD comparison tests
go test -tags=simd_comparison -v ./tests/integration

# Run with benchmarks
go test -tags=simd_comparison -bench=. -benchmem ./tests/integration
```

### Automated Benchmark Suite

For comprehensive benchmarking with profiling:

```bash
bash scripts/benchmark.sh
```

This creates a timestamped folder with:
- All benchmark results
- CPU profiles
- Profile analysis
- Call trees

See [BENCHMARK_WORKFLOW.md](BENCHMARK_WORKFLOW.md) for details.

## Test Naming Conventions

### Test Functions
- `Test<Functionality>` - Unit tests (e.g., `TestBasicFunctionality`)
- `Test<Feature><Aspect>` - Specific aspect tests (e.g., `TestHybridModeSelection`)
- Use descriptive names that explain what is being tested

### Benchmark Functions
- `Benchmark<Operation>` - Basic benchmarks (e.g., `BenchmarkInsertion`)
- `Benchmark<Feature><Operation>` - Feature-specific benchmarks (e.g., `BenchmarkHybridModes`)
- Include size/scale in sub-benchmarks (e.g., `Size_1K`, `Size_1M`)

### File Names
- `*_test.go` - Test files (e.g., `bloomfilter_test.go`)
- `*_benchmark_test.go` - Dedicated benchmark files (e.g., `storage_mode_benchmark_test.go`)
- Match the file being tested when possible

## Writing Tests

### Unit Test Example

```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected interface{}
    }{
        {"description", input1, expected1},
        {"description", input2, expected2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FunctionUnderTest(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Benchmark Example

```go
func BenchmarkOperation(b *testing.B) {
    // Setup (excluded from timing)
    bf := NewCacheOptimizedBloomFilter(10000, 0.01)

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        bf.AddUint64(uint64(i))
    }

    // Optional: Report custom metrics
    b.ReportMetric(float64(value), "custom_metric")
}
```

### Integration Test Example

```go
//go:build integration_tag

package integration

import (
    "testing"
    bloomfilter "github.com/shaia/BloomFilter"
)

func TestIntegrationScenario(t *testing.T) {
    // Test cross-package or special scenarios
    bf := bloomfilter.NewCacheOptimizedBloomFilter(10000, 0.01)
    // ... test logic ...
}
```

## CI/CD Integration

Tests are automatically run in GitHub Actions workflows:

### Pull Request Workflow
- Standard unit tests (`go test ./...`)
- Basic SIMD correctness tests
- Build validation

### Pre-Release Workflow
- All unit tests
- SIMD comparison tests (`-tags=simd_comparison`)
- Build validation
- Version validation

### Release Workflow
- Full test suite including integration tests
- SIMD performance validation
- Build for all platforms

See `.github/workflows/` for full workflow definitions.

## Test Coverage Goals

- **Unit Tests:** Aim for 80%+ coverage of core functionality
- **Integration Tests:** Cover critical paths and performance requirements
- **Benchmarks:** Cover all major operations and size ranges

Check coverage:
```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Performance Testing Best Practices

1. **Isolate benchmarks** - Test one operation at a time
2. **Use realistic data** - Test with representative data sizes
3. **Report allocations** - Always use `b.ReportAllocs()`
4. **Run multiple times** - Use `-count=10` for statistical significance
5. **Compare carefully** - Use `benchstat` for comparison
6. **Profile when needed** - Use `-cpuprofile` and `-memprofile`
7. **Document thresholds** - Set and document acceptable performance ranges

## Troubleshooting

### Tests fail with "undefined: HasSIMD"
- Check package declarations match file locations
- Ensure imports are correct for integration tests

### Integration tests not running
- Verify build tag is specified: `-tags=simd_comparison`
- Check file is in correct directory: `tests/integration/`

### Benchmarks show high variance
- Run with `-count=10` for multiple samples
- Check for background processes affecting results
- Use `benchstat` to analyze variance

### Coverage report missing integration tests
- Integration tests with build tags are excluded from standard coverage
- Run with tags: `go test -tags=simd_comparison -cover ./...`

## Additional Resources

- [scripts/BENCHMARK_WORKFLOW.md](scripts/BENCHMARK_WORKFLOW.md) - Automated benchmarking guide
- [tests/README.md](tests/README.md) - Tests directory documentation
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Go Benchmark Guidelines](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
