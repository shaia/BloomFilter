# Benchmark and Profiling Results

This folder contains all benchmark results, CPU profiles, and performance analysis documents.

## 📊 Latest Results Summary

**Last Updated:** October 26, 2025

### Current Performance (After Hybrid + clear() Optimization)

#### Small Filters (Array Mode - 10K-100K elements)
| Metric | Value |
|--------|-------|
| Add Operation | 55-65 ns/op |
| Contains Operation | 55-65 ns/op |
| Allocations | **0 B/op (zero!)** |
| Memory Overhead | ~720 KB fixed |
| Throughput | **17M ops/sec** |
| vs willf/bloom | **1.5x faster** |

#### Large Filters (Map Mode - 1M+ elements)
| Metric | Value |
|--------|-------|
| Add Operation | 450-520 ns/op |
| Contains Operation | 445-485 ns/op |
| Allocations | 144 B/op |
| Memory Overhead | Dynamic |
| Throughput | **2.2M ops/sec** |
| Scalability | Unlimited |

### SIMD Speedups

| Operation | SIMD Time | Scalar Time | Speedup |
|-----------|-----------|-------------|---------|
| PopCount (65KB) | 2.67µs | 8.05µs | **3.0x** |
| VectorOr (65KB) | 1.15µs | 2.41µs | **2.1x** |
| VectorAnd (65KB) | 1.08µs | 2.49µs | **2.3x** |
| VectorClear (65KB) | 862ns | 1.67µs | **1.9x** |

### Recent Optimizations Impact

| Optimization | CPU Improvement | Impact |
|-------------|-----------------|--------|
| clear() built-in | 41.4% in hot path | High |
| Double lookup fix | 16.7% map overhead | Medium |
| Hybrid array/map | 95% memory reduction | Critical |

---

## 📁 File Organization

### Latest Analysis Documents (October 2025)

#### Core Optimization Reports
1. **CLEAR_OPTIMIZATION_PROFILING.md** - Go 1.21+ clear() optimization analysis
2. **DOUBLE_LOOKUP_FIX_ANALYSIS.md** - Map lookup optimization (16.7% improvement)
3. **COMPETITIVE_ANALYSIS.md** - Performance vs willf/bloom (market leader)
4. **HYBRID_FINAL_COMPARISON.md** - Hybrid vs pure array comparison
5. **HYBRID_VS_ARRAY_PROFILING.md** - Profiling comparison of architectures

#### Historical Documents (Pre-Hybrid)
- **ARRAY_OPTIMIZATION_RESULTS.md** - Pure array implementation (legacy)
- **OPTIMIZATION_RESULTS.md** - Slice pre-allocation optimization
- **FLAMEGRAPH_ANALYSIS.md** - CPU profiling and flamegraph analysis
- **PROFILING_COMPARISON.md** - Before/after benchmark methodology
- **PROFILING_ANALYSIS.md** - Initial profiling analysis
- **CPU_PROFILE_POOLED_MAPS.md** - Map pooling experiments
- **BENCHMARK_SUMMARY.md** - Original benchmark summary

### CPU Profiles
- **cpu_double_lookup_fix.prof** - Latest profile (after double lookup fix)
- **cpu_clear_optimization.prof** - After clear() optimization
- **cpu_array_based.prof** - Pure array baseline (Oct 19)
- **cpu_final.prof** - Legacy final profile
- **cpu_optimized.prof** - Legacy optimized profile

### Profile Analysis Text Files
- **cpu_double_lookup_fix_analysis.txt** - Latest analysis output
- **cpu_clear_optimization_analysis.txt** - Clear optimization analysis
- **cpu_array_based_analysis.txt** - Array baseline analysis
- **cpu_hash_positions_detail.txt** - Detailed hash function profile
- **cpu_get_bit_detail.txt** - Detailed bit check profile

### Benchmark Results
- **benchmark_clear_optimization_full.txt** - Comprehensive benchmarks (latest)
- **benchmark_vs_willf.txt** - Competitive benchmarking results
- **benchmark_array_based.txt** - Array-only benchmark
- **benchmark_pooled_maps.txt** - Map pooling experiments

---

## 🔬 How to Use These Results

### View Interactive Profile
```bash
cd results/
go tool pprof -http=:8080 cpu_double_lookup_fix.prof
```
Then open http://localhost:8080 and select "Flame Graph" from the VIEW menu.

### Compare Latest Optimizations
```bash
cd results/
# Compare before and after double lookup fix
go tool pprof -base=cpu_clear_optimization.prof cpu_double_lookup_fix.prof
```

### View Text Analysis
```bash
cd results/
# Top CPU consumers
go tool pprof -top cpu_double_lookup_fix.prof

# Detailed function analysis
go tool pprof -list=getHashPositionsOptimized cpu_double_lookup_fix.prof
```

### Run Current Benchmarks
```bash
# Hybrid mode benchmarks
go test -bench=BenchmarkHybridModes -benchmem -run=^$

# SIMD comparison
go test -bench=BenchmarkSIMDvsScalar -benchmem -run=^$

# Full suite with profiling
go test -bench=BenchmarkBloomFilterWithSIMD -cpuprofile=results/cpu_latest.prof -benchtime=2s -run=^$
```

---

## 📈 Optimization Timeline

### 2025 Optimization History

| Date | Optimization | Key Impact | Document |
|------|-------------|------------|----------|
| **Oct 26** | Remove double map lookups | 16.7% map overhead reduction | DOUBLE_LOOKUP_FIX_ANALYSIS.md |
| **Oct 26** | Remove competitive benchmarks | Zero dependencies | - |
| **Oct 25** | clear() built-in | 41% CPU improvement | CLEAR_OPTIMIZATION_PROFILING.md |
| **Oct 25** | Competitive analysis | Market positioning | COMPETITIVE_ANALYSIS.md |
| **Oct 19** | Hybrid array/map | 95% memory reduction | HYBRID_FINAL_COMPARISON.md |
| Oct 18 | Slice pre-allocation | 32% speedup | OPTIMIZATION_RESULTS.md |
| Oct 18 | fmt.Sprintf removal | 12.8x (benchmark fix) | PROFILING_COMPARISON.md |

### Cumulative Improvements

**From Pure Array (Oct 19) to Hybrid + Optimizations (Oct 26):**

Small Filters (10K-100K):
- Speed: **1.5x faster** than alternatives
- Memory: **95% reduction** (14.4 MB → 720 KB)
- Allocations: **Zero per-operation**

Large Filters (1M+):
- Scalability: **Unlimited** (was capped at 12.8 MB)
- Performance: **Competitive** with simple implementations
- Memory: **Dynamic** allocation

Code Quality:
- **Zero external dependencies**
- **No legacy code**
- **Comprehensive documentation**
- **Production-ready**

---

## 🎯 Architecture Evolution

### Phase 1: Pure Array (Legacy)
```
✗ Fixed 200K cache line limit (12.8 MB max filter)
✗ 14.4 MB overhead per instance
✓ Fast for small filters
✗ Not scalable
```

### Phase 2: Hybrid Array/Map (Current)
```
✓ Array mode: ≤10K cache lines (small filters)
  - Zero allocations
  - 720 KB overhead
  - 1.5x faster than alternatives

✓ Map mode: >10K cache lines (large filters)
  - Unlimited scalability
  - Dynamic memory
  - Competitive performance

✓ Automatic mode selection
✓ No configuration needed
```

### Phase 3: Future (Planned)
```
○ Paged array mode (documented in FUTURE_PAGED_ARRAY_OPTIMIZATION.md)
○ 100K-10M element sweet spot
○ 2-3x speedup for large filters
○ Bridge gap between array and map
```

---

## 🔬 Key Insights from Profiling

### Hot Paths (from latest profile)
1. **getHashPositionsOptimized** - 38.10% CPU (was 59.81%, improved 41%)
2. **getBitCacheOptimized** - 28.02% CPU (was 30.01%, improved 18%)
3. **mapassign_fast64** - 3.99% CPU (map writes)
4. **prefetchCacheLines** - 3.99% CPU (cache optimization)

### Optimization Impact
- **clear() vs manual deletion**: 10-50x faster map clearing
- **Single vs double lookup**: 16.7% reduction in map access overhead
- **Array vs map (small filters)**: 8-10x faster, zero allocations
- **SIMD vs scalar**: 2-4x speedup on bulk operations

### Memory Characteristics
- **Array mode**: Zero per-op allocations, 720 KB fixed overhead
- **Map mode**: 144 bytes per operation, dynamic growth
- **SIMD alignment**: 64-byte cache line alignment maintained
- **Total allocations**: 832 MB for benchmark suite

---

## 🔄 Running New Benchmarks

### Standard Benchmark Suite
```bash
# Full benchmark suite with memory stats
go test -bench=. -benchmem -run=^$ | tee results/benchmark_$(date +%Y%m%d).txt

# CPU profiling
go test -bench=BenchmarkBloomFilterWithSIMD \
  -cpuprofile=results/cpu_$(date +%Y%m%d).prof \
  -benchtime=2s -run=^$

# Memory profiling
go test -bench=BenchmarkBloomFilterWithSIMD \
  -memprofile=results/mem_$(date +%Y%m%d).prof \
  -benchtime=2s -run=^$
```

### Hybrid-Specific Benchmarks
```bash
# Array mode performance
go test -bench=BenchmarkHybridModes/.*Array -benchmem -run=^$

# Map mode performance
go test -bench=BenchmarkHybridModes/.*Map -benchmem -run=^$

# Mode crossover analysis
go test -bench=BenchmarkHybridCrossoverPoint -benchmem -run=^$
```

### Analysis Commands
```bash
# Top functions by CPU time
go tool pprof -top -nodecount=30 results/cpu_latest.prof

# Detailed function profile
go tool pprof -list=getHashPositionsOptimized results/cpu_latest.prof

# Memory allocations
go tool pprof -top results/mem_latest.prof

# Interactive web UI
go tool pprof -http=:8080 results/cpu_latest.prof
```

---

## 📚 Document Reading Guide

### For Understanding Current Implementation
1. **HYBRID_FINAL_COMPARISON.md** - Why hybrid architecture?
2. **CLEAR_OPTIMIZATION_PROFILING.md** - Latest performance improvements
3. **DOUBLE_LOOKUP_FIX_ANALYSIS.md** - Code quality optimizations
4. **COMPETITIVE_ANALYSIS.md** - Market positioning

### For Historical Context
1. **ARRAY_OPTIMIZATION_RESULTS.md** - Legacy pure array approach
2. **OPTIMIZATION_RESULTS.md** - Early optimizations
3. **FLAMEGRAPH_ANALYSIS.md** - Initial profiling insights

### For Future Planning
1. **COMPETITIVE_ANALYSIS.md** - Optimization roadmap
2. **FUTURE_PAGED_ARRAY_OPTIMIZATION.md** (in parent dir) - Next steps

---

## 🧹 Maintenance Notes

### Current State (Oct 26, 2025)
- Latest profile: `cpu_double_lookup_fix.prof`
- Latest benchmarks: `benchmark_clear_optimization_full.txt`
- Active architecture: Hybrid array/map
- Dependencies: Zero external
- Production status: Ready

### Cleanup Completed
- ✅ Removed competitive benchmark code (moved to separate project)
- ✅ Removed unused MaxCacheLines constant
- ✅ Removed external dependencies (willf/bloom)
- ✅ Cleaned up legacy code comments

### Next Maintenance
- Archive profiles older than 6 months
- Update this README with new optimizations
- Add paged array results when implemented

---

## 📊 Quick Reference

**Best Use Cases:**
- **Array Mode (10K-100K)**: Microservices, session filters, rate limiting
- **Map Mode (1M+)**: Large-scale deduplication, distributed systems

**Performance Expectations:**
- Small filters: 55-65 ns/op, zero allocations
- Large filters: 450-520 ns/op, 144 B/op
- SIMD operations: 2-4x faster than scalar

**Documentation:**
- Implementation details: See parent README.md
- API reference: See parent README.md
- Optimization history: This document
