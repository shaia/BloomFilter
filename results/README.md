# Benchmark Results

This directory contains benchmark results and performance analysis data for the BloomFilter project.

> **Note:** Benchmark result files and folders are NOT tracked in git. Only this README.md and analysis tools are version controlled.

## Latest Benchmark Run

**Last Updated:** 2025-10-31 16:45:39

**Results Folder:** run_20251031_163225_analysis

**Status:** Benchmark completed successfully

---

## Performance Metrics

### Bloom Filter Operations

*Metrics will be automatically updated after each benchmark run*

**Small Filters (Array Mode - < 10K cache lines):**
- Add Operation: TBD
- Contains Operation: TBD
- Allocations: TBD
- Memory Overhead: TBD

**Large Filters (Map Mode - >= 10K cache lines):**
- Add Operation: TBD
- Contains Operation: TBD
- Allocations: TBD
- Memory Overhead: TBD

### SIMD Performance

*SIMD speedup measurements will be updated here*

- PopCount: TBD
- VectorOr: TBD
- VectorAnd: TBD
- VectorClear: TBD

---

## Running Benchmarks

### Automated Benchmark Suite

```bash
# Run full benchmark suite with profiling
bash scripts/benchmark.sh
```

This creates a timestamped folder: `results/run_YYYYMMDD_HHMMSS_analysis/` containing:
- Full benchmark suite results
- SIMD comparison benchmarks
- CPU profile data (separate profiles for main and SIMD benchmarks)
- Profile analysis (text and tree format for each profile)

### Comprehensive Profiling

```bash
# Run all profiling types (CPU, memory, goroutine, block, mutex)
bash scripts/profile.sh

# Profile specific benchmark
bash scripts/profile.sh BenchmarkInsertion 10s

# Profile SIMD comparison
bash scripts/profile.sh BenchmarkSIMDvsScalar 5s
```

### Viewing Results

**Latest benchmark folder:**
```bash
cd results/run_*/
ls -la
```

**Interactive profile viewer:**
```bash
# Main benchmarks CPU profile in browser
go tool pprof -http=:8080 results/run_*/cpu_profile_main.prof

# SIMD benchmarks CPU profile in browser
go tool pprof -http=:8081 results/run_*/cpu_profile_simd.prof

# Memory profile
go tool pprof -http=:8082 results/profile_*/mem.prof

# Execution trace viewer
go tool trace results/profile_*/trace.out
```

**Text analysis:**
```bash
# View main benchmarks profile analysis
cat results/run_*/profile_main_text.txt

# View main benchmarks call tree
cat results/run_*/profile_main_tree.txt

# View SIMD benchmarks profile analysis
cat results/run_*/profile_simd_text.txt

# View SIMD benchmarks call tree
cat results/run_*/profile_simd_tree.txt
```

---

## Analysis Tools

### Jupyter Notebook
[benchmark_analyzer.ipynb](benchmark_analyzer.ipynb) - Interactive Python-based analysis tool for benchmark results

### Profile Comparison
```bash
# Compare two profile runs
go tool pprof -base=results/baseline_profile/cpu.prof \
    -http=:8080 results/profile_*/cpu.prof
```

---

## Documentation

Detailed documentation is available in the following locations:

- **[scripts/BENCHMARK_WORKFLOW.md](../scripts/BENCHMARK_WORKFLOW.md)** - Comprehensive workflow guide
- **[scripts/README.md](../scripts/README.md)** - Scripts documentation with examples
- **[TESTING.md](../TESTING.md)** - Complete testing guide
- **[tests/README.md](../tests/README.md)** - Integration tests documentation

---

## Directory Structure

```
results/
├── README.md                           # This file (tracked in git)
├── benchmark_analyzer.ipynb            # Analysis tool (tracked in git)
├── run_YYYYMMDD_HHMMSS_analysis/      # Benchmark runs (not tracked)
│   ├── benchmark_full_suite.txt       # All benchmarks results
│   ├── simd_comparison.txt            # SIMD vs scalar comparison
│   ├── profiled_benchmarks.txt        # Profiled benchmark output
│   ├── cpu_profile_main.prof          # Main benchmarks CPU profile
│   ├── cpu_profile_simd.prof          # SIMD benchmarks CPU profile
│   ├── profile_main_text.txt          # Main profile analysis (text)
│   ├── profile_main_tree.txt          # Main profile call tree
│   ├── profile_simd_text.txt          # SIMD profile analysis (text)
│   └── profile_simd_tree.txt          # SIMD profile call tree
└── profile_YYYYMMDD_HHMMSS_analysis/  # Profile runs (not tracked)
    ├── cpu.prof, mem.prof, etc.
    └── various analysis files
```

---

*This README is automatically updated by [scripts/benchmark.sh](../scripts/benchmark.sh). Last manual edit: 2025-10-31*
