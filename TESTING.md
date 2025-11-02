# Testing Guide

This document describes the testing structure and best practices for the BloomFilter project.

## Test Organization

The project follows Go best practices for test organization:

```
BloomFilter/
├── bloomfilter_test.go              # Core functionality tests
├── bloomfilter_simd_test.go         # SIMD capability detection tests
├── bloomfilter_validation_test.go   # Input validation tests (32 sub-tests)
└── tests/
    ├── TEST_COVERAGE_SUMMARY.md     # Comprehensive test coverage summary
    ├── benchmark/
    │   └── bloomfilter_benchmark_test.go  # Performance benchmarks
    └── integration/
        ├── bloomfilter_concurrent_test.go       # Thread-safety tests
        ├── bloomfilter_edge_cases_test.go       # Edge cases and boundary conditions
        ├── bloomfilter_race_test.go             # Race detector tests (build tag: race)
        └── bloomfilter_simd_comparison_test.go  # SIMD comparison (build tag: simd_comparison)
```

## Test Categories

### 1. Unit Tests (Root Package)

Located in the root package directory, these test individual components and functions.

**Files:**
- `bloomfilter_test.go` - Core Bloom filter operations (Add, Contains, Union, Intersection, etc.)
- `bloomfilter_simd_test.go` - SIMD capability detection and runtime functions
- `bloomfilter_validation_test.go` - Input validation with 32 sub-tests covering all validation paths

**Running:**
```bash
# All unit tests
go test -v .

# Specific test
go test -v -run=TestBasicFunctionality .

# Validation tests only
go test -v -run=TestInputValidation .
```

### 2. Benchmarks (tests/benchmark/)

Performance benchmarks for insertion, lookup, false positive rates, and cache performance.

**Files:**
- `bloomfilter_benchmark_test.go` - Comprehensive performance benchmarks

**Running:**
```bash
# All benchmarks
go test -bench=. -benchmem ./tests/benchmark/...

# Specific benchmark
go test -bench=BenchmarkInsertion -benchmem ./tests/benchmark

# Quick benchmarks (using Makefile)
make bench-short

# Full benchmark comparison (SIMD vs Pure Go)
make bench-all

# With CPU profiling
go test -bench=BenchmarkInsertion -cpuprofile=cpu.prof ./tests/benchmark
```

### 3. Integration Tests (tests/integration/)

Tests that verify thread-safety, edge cases, and cross-component interactions.

**Files:**
- `bloomfilter_concurrent_test.go` - Thread-safety tests with 100+ concurrent goroutines
- `bloomfilter_edge_cases_test.go` - Boundary conditions, invalid inputs, extreme sizes
- `bloomfilter_race_test.go` - Race detector tests (build tag: `race`)
- `bloomfilter_simd_comparison_test.go` - SIMD vs fallback validation (build tag: `simd_comparison`)

**Running:**
```bash
# All integration tests
go test -v ./tests/integration

# Thread-safety tests
go test -v ./tests/integration -run=TestConcurrent

# With race detector
go test -race -v ./tests/integration

# Edge cases and validation
go test -v ./tests/integration -run=TestZeroAndNegativeInputs

# SIMD comparison tests (requires build tag)
go test -tags=simd_comparison -v ./tests/integration -run=TestSIMDPerformanceImprovement

# SIMD comparison benchmarks
go test -tags=simd_comparison -bench=BenchmarkSIMDvsScalar ./tests/integration
```

## Running Tests

### Quick Test Commands (Using Makefile)

```bash
# Quick sanity check (skips long-running tests)
make test-short

# Run all tests
make test

# Run with race detector
make test-race

# Run integration tests only
make test-integration

# Run benchmarks
make bench

# Run quick benchmarks
make bench-short

# Full validation (tests + race + pure Go)
make test-all
```

### Standard Test Suite

```bash
# Run all tests
go test -v ./...

# Quick iteration (skip long-running tests)
go test -short -v ./...

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
- `bloomfilter_*_test.go` - Test files with descriptive names (e.g., `bloomfilter_simd_test.go`)
- `bloomfilter_*_benchmark_test.go` - Dedicated benchmark files (e.g., `bloomfilter_storage_mode_benchmark_test.go`)
- Use `bloomfilter_` prefix for consistency across test files

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

### Tests Workflow (on push/PR)
- Standard unit tests (`go test -v ./...`)
- Race detector tests (`go test -race -short -timeout=10m -v ./...`)
  - Uses `-short` flag to reduce workload for race detector (5-10x overhead)
  - 10-minute timeout for comprehensive race detection
  - Uploads race logs on failure
- Build validation for all platforms (Ubuntu, Windows, macOS)
- Build with race detector enabled
- Benchmark dry run

### Key Features
- **Race Detection**: Automated data race detection on every push/PR
- **Cross-Platform**: Tests on Ubuntu, Windows, and macOS
- **Comprehensive Coverage**: Unit, integration, and stress tests
- **Performance Validation**: Benchmark tests ensure no regressions

See [.github/workflows/test.yml](.github/workflows/test.yml) for full workflow definition.

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
- [tests/TEST_COVERAGE_SUMMARY.md](tests/TEST_COVERAGE_SUMMARY.md) - Comprehensive test coverage summary
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Go Benchmark Guidelines](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
