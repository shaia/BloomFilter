# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Thread-Safety**: Full concurrent support with lock-free atomic operations
  - Lock-free bit operations using atomic Compare-And-Swap (CAS)
  - Bounded retry limits with exponential backoff under contention
  - sync.Pool optimization for zero-allocation temporary storage reuse
- **Batch Operations**: High-throughput batch Add functions
  - `AddBatch(items [][]byte)` - Batch byte slice operations
  - `AddBatchString(items []string)` - Batch string operations with zero-copy conversion
  - `AddBatchUint64(items []uint64)` - Batch uint64 operations
  - Pooled resource reuse across batch items for optimal performance
- **Comprehensive Test Suite**: Thread-safety and performance validation
  - Race detector integration with GitHub Actions CI/CD
  - Concurrent read/write tests (100+ goroutines)
  - Stress tests with millions of operations
  - Edge case and boundary condition tests

### Changed

- Refactored codebase for better maintainability and thread-safety
- Split monolithic `bloomfilter.go` (660 lines) into focused modules:
  - `bloomfilter.go` (394 lines): Core API and public interface
  - `internal/hash/hash.go` (108 lines): Hash function implementations
  - `internal/storage/storage.go` (186 lines): Hybrid storage with sync.Pool
- Modernized unsafe string-to-byte conversion using Go 1.20+ stdlib (`unsafe.StringData`/`unsafe.Slice`)
- Optimized CI/CD workflow to use `-short` flag with race detector to prevent timeouts
- Fixed stack allocation comments based on escape analysis verification
- Eliminated 150+ lines of duplicate code between array and map modes
- Simplified complex functions by 59-65% with proper resource pooling
- Added `IsArrayMode()` accessor method for better encapsulation
- Updated all documentation to reflect thread-safety improvements

### Performance

- **Concurrent Writes**: 18-23M operations/second (50 goroutines)
- **Concurrent Reads**: 10M+ operations/second (100 goroutines)
- **Lock-Free Operations**: Zero mutex contention with atomic CAS
- **Resource Pooling**: Eliminates allocations in hot paths with sync.Pool
- **Race Detector Compatible**: Tests pass with race detector in <1 second (reduced workload)

### Fixed

- Critical nested pool operation bug in batch functions causing race detector timeouts
- Empty spin loop backoff properly documented (compiler optimization acceptable)
- Defer-in-loop bug that caused pool exhaustion under high concurrency
- Pool storage slice return bug that could cause data corruption
- CAS retry limit prevents indefinite spinning under extreme contention
- Defensive copying of pooled storage slices in AddBatch functions for consistency and future-proofing

### Quality Improvements

- Zero performance regression - improved concurrency performance
- All tests pass including race detector validation
- GitHub Actions CI/CD with automated race detection
- Better separation of concerns with clear module boundaries
- Internal packages cannot be imported by users, ensuring API stability
- Production-ready thread-safety with comprehensive testing

## [0.2.0] - 2025-10-26

### Added

- Hybrid array/map architecture for automatic optimization based on filter size
- Automatic mode selection between array mode (≤10K cache lines) and map mode (>10K cache lines)
- Zero-allocation operations for small filters in array mode
- Unlimited scalability for large filters in map mode
- Competitive analysis documentation vs willf/bloom
- Comprehensive profiling documentation (CPU and memory)
- SIMD comparison tests with build tags for optional performance validation
- GitHub Actions workflow integration for SIMD tests before releases

### Changed

- **BREAKING**: Removed external dependencies (willf/bloom) for cleaner package
- Optimized map clearing using Go 1.21+ `clear()` built-in (41% CPU improvement in hot path)
- Eliminated double map lookups by using length-based existence checks (16.7% map overhead reduction)
- Updated README with hybrid architecture performance metrics and use case guidance
- Comprehensive documentation updates across all files
- Modernized build tags from `// +build` to `//go:build` directive (Go 1.17+)

### Removed

- Removed unused `MaxCacheLines` legacy constant (replaced by `ArrayModeThreshold`)
- Removed competitive benchmark code from main package (moved to separate project)
- Removed all external dependencies - package is now zero-dependency

### Performance

- **Small filters (10K-100K)**: 55-65 ns/op with zero allocations (1.5x faster than alternatives)
- **Large filters (1M+)**: 450-520 ns/op with unlimited scalability
- **Memory reduction**: 95% reduction for small filters (14.4 MB → 720 KB)
- **Map operations**: 41% faster with `clear()` optimization
- **Map lookups**: 16.7% overhead reduction from single lookup pattern
- **SIMD operations**: 2-4x speedup verified across PopCount, VectorOr, VectorAnd

### Fixed

- Flaky SIMD performance tests now use 0.7x tolerance threshold for system load variance
- SIMD comparison tests are now optional (use `-tags=simd_comparison` to run)
- Code smell: Removed unused mode variables from benchmarks
- Code smell: Removed duplicate constant declarations in test files
- Map access pattern: Eliminated redundant lookup operations
- Documentation: Updated all docs to reflect current hybrid implementation

## [0.1.0] - 2025-10-25

### Added

- Initial release of SIMD-optimized Bloom Filter
- Cache-optimized Bloom Filter implementation with cache line alignment
- AVX2 SIMD support for AMD64 architecture
- NEON SIMD support for ARM64 architecture
- CPUID-based runtime feature detection for optimal SIMD selection
- Cross-platform support (AMD64/ARM64)
- Comprehensive test suite including SIMD correctness tests
- Performance benchmarks for SIMD vs scalar operations
- PR-based versioning and release workflow system
- Automated release preparation via GitHub Actions
- Branch protection documentation and guidelines

### Performance

- Cache-aligned memory layout for optimal CPU cache utilization
- Pre-allocated arrays to minimize memory allocations in hot paths
- SIMD-optimized bit operations for Set/Test operations
- Efficient hash distribution across cache lines

### Documentation

- Complete API documentation
- Performance benchmarking results and analysis
- Example usage code in `docs/examples/`
- Versioning and release process documentation
- Branch protection setup guide

[0.2.0]: https://github.com/shaia/BloomFilter/releases/tag/v0.2.0
[0.1.0]: https://github.com/shaia/BloomFilter/releases/tag/v0.1.0
