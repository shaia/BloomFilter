# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Refactored codebase for better maintainability and readability
- Split monolithic `bloomfilter.go` (660 lines) into focused modules:
  - `bloomfilter.go` (394 lines): Core API and public interface
  - `internal/hash/hash.go` (108 lines): Hash function implementations
  - `internal/storage/storage.go` (186 lines): Hybrid storage abstraction
- Moved implementation details to `internal/` package following Go conventions
- Eliminated 150+ lines of duplicate code between array and map modes
- Simplified complex functions by 59-65% (getHashPositionsOptimized, setBitCacheOptimized, getBitCacheOptimized)
- Added `IsArrayMode()` accessor method for better encapsulation
- Updated package structure documentation in README

### Quality Improvements

- Zero performance regression - all benchmarks unchanged
- All tests pass (18/18)
- Better separation of concerns with clear module boundaries
- Internal packages cannot be imported by users, ensuring API stability
- Easier to maintain and extend codebase

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
