# Benchmark and Profiling Results

This folder contains all benchmark results, CPU profiles, and performance analysis documents.

## 📊 Latest Results Summary

**Last Updated:** October 18, 2025

### Current Performance (After Slice Pre-allocation Optimization)

| Metric | Value |
|--------|-------|
| Query Time (10K) | 441µs |
| Query Time (100K) | 445µs |
| Query Time (1M) | 474µs |
| Throughput | 2.27M queries/sec |
| Memory per Query | 144KB |
| Allocations | 12,000/query |

### SIMD Speedups

| Operation | SIMD Time | Scalar Time | Speedup |
|-----------|-----------|-------------|---------|
| PopCount (65KB) | 2.64µs | 7.85µs | 2.97x |
| VectorOr (65KB) | 1.15µs | 2.53µs | 2.19x |
| VectorAnd (65KB) | 1.12µs | 2.54µs | 2.28x |
| VectorClear (65KB) | 862ns | 1.21µs | 1.40x |

---

## 📁 File Organization

### Benchmark Results
- **benchmark_results_final.txt** - Latest benchmark run (after optimizations)
- **benchmark_results_optimized.txt** - After fmt.Sprintf removal
- **benchmark_results.txt** - Original baseline

### CPU Profiles
- **cpu_final.prof** - Latest CPU profile (pprof binary format)
- **cpu_optimized.prof** - After fmt.Sprintf removal
- **cpu.prof** - Original profile with fmt.Sprintf overhead

### Profile Analysis
- **profile_final.txt** - Text summary of final profile
- **profile_final_tree.txt** - Call tree of final profile
- **profile_optimized.txt** - Text summary after fmt.Sprintf removal
- **profile_optimized_tree.txt** - Call tree after fmt.Sprintf removal
- **profile_text.txt** - Original profile text
- **profile_tree.txt** - Original call tree
- **profile_raw.txt** - Raw profile data
- **simd_profile.txt** - SIMD-specific analysis

### Analysis Documents
- **OPTIMIZATION_RESULTS.md** - Slice pre-allocation optimization results
- **FLAMEGRAPH_ANALYSIS.md** - CPU profiling and flamegraph analysis
- **PROFILING_COMPARISON.md** - Before/after benchmark methodology
- **PROFILING_ANALYSIS.md** - Initial profiling analysis

---

## 🔬 How to Use These Results

### View Interactive Flamegraph
```bash
cd results/
go tool pprof -http=:8080 cpu_final.prof
```
Then open http://localhost:8080 and select "Flame Graph" from the VIEW menu.

### Generate SVG Flamegraph (requires Graphviz)
```bash
cd results/
go tool pprof -svg cpu_final.prof > flamegraph.svg
```

### View Text Profile
```bash
cd results/
go tool pprof -text cpu_final.prof
```

### View Call Tree
```bash
cd results/
go tool pprof -tree cpu_final.prof
```

### Compare Profiles
```bash
cd results/
# Compare before and after optimization
go tool pprof -base=cpu_optimized.prof cpu_final.prof
```

---

## 📈 Performance History

### Optimization Timeline

| Date | Optimization | Query Time | Speedup | Document |
|------|-------------|------------|---------|----------|
| Oct 18, 2025 | Slice Pre-allocation | 441µs | 1.48x | OPTIMIZATION_RESULTS.md |
| Oct 18, 2025 | Benchmark Fix (fmt.Sprintf removal) | 650µs | 1.0x | PROFILING_COMPARISON.md |
| Oct 18, 2025 | Baseline (with fmt overhead) | 6,064µs | - | PROFILING_ANALYSIS.md |

### Improvement Summary

**Total improvement from baseline to current:**
- Query time: 6,064µs → 441µs (**13.7x faster**)
  - 12.8x from removing fmt.Sprintf overhead (benchmark fix)
  - 1.48x from slice pre-allocation (actual optimization)

**Actual code optimization impact:**
- Query time: 651µs → 441µs (**1.48x faster, 32% improvement**)
- Memory: 337KB → 144KB (**57% reduction**)
- Allocations: 18K → 12K (**33% reduction**)

---

## 🎯 Next Optimization Targets

Based on flamegraph analysis, remaining opportunities:

1. **Replace maps with arrays** - Potential 30% speedup
2. **Implement sync.Pool** - Potential 10-15% speedup
3. **Custom memory allocator** - Potential 5-10% speedup

**Projected final performance:** ~270µs per query (2.4x faster than current)

---

## 🔄 Running New Benchmarks

All new benchmark and profiling results should be saved to this folder:

```bash
# Run benchmarks and save to results/
go test -bench=. -benchmem -run=^$ > results/benchmark_results_$(date +%Y%m%d).txt

# Generate CPU profile
go test -bench=BenchmarkBloomFilterWithSIMD -cpuprofile=results/cpu_$(date +%Y%m%d).prof -run=^$ -benchtime=2s

# Generate profile analysis
go tool pprof -text results/cpu_$(date +%Y%m%d).prof > results/profile_$(date +%Y%m%d).txt
```

---

## 📚 Document Index

### Primary Documents (Read in Order)
1. **PROFILING_ANALYSIS.md** - Initial profiling showing fmt.Sprintf overhead
2. **PROFILING_COMPARISON.md** - Benchmark methodology improvements
3. **FLAMEGRAPH_ANALYSIS.md** - Detailed CPU profiling analysis
4. **OPTIMIZATION_RESULTS.md** - Slice pre-allocation optimization results

### Reference Documents
- **benchmark_results_*.txt** - Raw benchmark data
- **profile_*.txt** - Raw profile analysis
- **cpu_*.prof** - Binary profiles for pprof

---

## 🧹 Maintenance

This folder contains historical results for reference. Periodically:
- Archive old profiles (>1 month)
- Keep latest 3 optimization cycles
- Update this README with new optimization results
