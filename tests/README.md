# Tests Directory

This directory contains special test files that require specific build tags or conditions.

## Directory Structure

```
tests/
└── integration/
    └── simd_comparison_test.go  # SIMD performance comparison tests
```

## Integration Tests

### SIMD Comparison Tests

**File:** `integration/simd_comparison_test.go`

**Purpose:** Comprehensive SIMD performance and correctness validation tests.

**Build Tag:** `simd_comparison`

These tests compare SIMD implementations against scalar fallback implementations across various data sizes and operations. They are separated because:

1. **Resource Intensive:** Run extensive performance benchmarks
2. **Special Build Tag:** Require `-tags=simd_comparison` to run
3. **Release Validation:** Used to verify SIMD performance before releases
4. **Not Run by Default:** Excluded from normal `go test ./...` runs

**Running these tests:**

```bash
# Run SIMD performance comparison tests
go test -tags=simd_comparison -v ./tests/integration

# Run specific test
go test -tags=simd_comparison -v ./tests/integration -run=TestSIMDPerformanceImprovement

# Run benchmarks
go test -tags=simd_comparison -bench=BenchmarkSIMDvsScalar ./tests/integration
```

**Test Coverage:**
- `BenchmarkSIMDvsScalar` - Compares SIMD vs fallback for PopCount, VectorOr, VectorAnd, VectorClear
- `TestSIMDPerformanceImprovement` - Validates SIMD speedup meets minimum thresholds
- `TestSIMDCorrectness` - Ensures SIMD produces identical results to fallback
- `BenchmarkBloomFilterWithSIMD` - Full Bloom filter benchmarks with SIMD

## Regular Unit Tests

Regular unit tests remain in the root package directory:

```
bloomfilter_test.go              # Core Bloom filter functionality tests
benchmark_test.go                # Performance benchmarks for main package
storage_mode_test.go             # Array vs Map storage mode selection tests
storage_mode_benchmark_test.go   # Storage mode performance benchmarks
simd_test.go                     # SIMD capability detection tests
```

These follow Go conventions and are run with standard `go test ./...`.

## Running Tests

**All standard tests:**
```bash
go test -v ./...
```

**With benchmarks:**
```bash
go test -bench=. -benchmem ./...
```

**Include integration tests:**
```bash
# Run all tests including integration
go test -tags=simd_comparison -v ./...

# Run only integration tests
go test -tags=simd_comparison -v ./tests/integration
```

**Automated benchmark suite:**
```bash
bash scripts/benchmark.sh
```

This creates a timestamped results folder with all benchmark outputs, CPU profiles, and analysis files.

## CI/CD Integration

The integration tests are automatically run in CI/CD workflows:

- **Pull Requests:** Regular tests only
- **Pre-Release:** Includes SIMD comparison tests
- **Release:** Full validation including performance thresholds

See `.github/workflows/` for workflow configurations.

## Best Practices

1. **Use build tags** for tests that are resource-intensive or require special conditions
2. **Keep unit tests** in the same directory as the code they test
3. **Separate integration tests** that test cross-package functionality or performance
4. **Document requirements** for running special tests (build tags, environment, etc.)
5. **Automate** with scripts and CI/CD for consistency

## Adding New Tests

**Regular unit tests:** Add to appropriate `*_test.go` file in root package

**Integration tests:** Add to `tests/integration/` with appropriate build tags

**Benchmarks:** Add to `*_benchmark_test.go` files, use consistent naming

For questions or issues, see [BENCHMARK_WORKFLOW.md](../scripts/BENCHMARK_WORKFLOW.md).
