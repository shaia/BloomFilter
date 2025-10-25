## Release v0.1.0

**Type:** patch

### Changes

- Prepare release v0.1.0 (b3824da)
- Add PR-based versioning and release workflows (ed2e579)
- Implement proper CPUID-based AVX2 detection (e0e3ed0)
- restructure of the code (acd0e2a)
- fixing gh workflow (c98cfba)
- restructure of the code (b22dc16)
- changed structure of bloom filter (fc7489d)
- streamline the code change to structure of the code to be simpler (06def17)
- fixing go version (229f996)
- fix small lint error (0897603)
- Create go.yml (9e6542e)
- first commit (7a98ac5)

---

**Release Checklist:**
- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG updated (if applicable)
- [ ] Breaking changes documented (if any)
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.0]: https://github.com/shaia/go-simd-bloomfilter/releases/tag/v0.1.0
