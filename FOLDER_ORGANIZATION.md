# Folder Organization

This document describes the organization of the BloomFilter project.

## 📁 Project Structure

```
BloomFilter/
├── bloomfilter.go              # Main bloom filter implementation
├── bloomfilter_test.go         # Unit tests
├── benchmark_test.go           # Benchmark suite
├── simd_comparison_test.go     # SIMD vs Scalar comparison tests
├── simd_test.go               # SIMD-specific tests
├── go.mod                     # Go module definition
├── Makefile                   # Build automation
├── README.md                  # Main documentation
├── .gitignore                 # Git ignore rules
│
├── internal/                  # Internal packages (not importable)
│   └── simd/                 # SIMD implementations
│       ├── simd.go           # SIMD interface & detection
│       ├── fallback.go       # Scalar fallback implementation
│       ├── amd64/            # x86-64 specific code
│       │   ├── avx2.go       # AVX2 declarations
│       │   ├── avx2.s        # AVX2 assembly
│       │   └── stub.go       # Stubs for non-amd64
│       └── arm64/            # ARM64 specific code
│           ├── neon_asm.go   # NEON declarations
│           ├── neon.s        # NEON assembly
│           └── stub.go       # Stubs for non-arm64
│
├── docs/                      # Documentation
│   └── examples/             # Example code
│       └── basic/            # Basic usage examples
│
├── results/                   # ⭐ All benchmark & profiling results
│   ├── README.md             # Results documentation
│   ├── benchmark_results_*.txt    # Benchmark outputs
│   ├── cpu_*.prof            # CPU profiles (pprof format)
│   ├── profile_*.txt         # Profile analysis
│   ├── FLAMEGRAPH_ANALYSIS.md     # Performance analysis
│   ├── OPTIMIZATION_RESULTS.md    # Optimization tracking
│   └── PROFILING_*.md        # Profiling documentation
│
├── scripts/                   # Automation scripts
│   └── benchmark.sh          # Automated benchmark runner
│
├── bin/                       # Compiled binaries (gitignored)
└── debug/                     # Debug outputs (gitignored)
```

## 📊 Results Folder

All benchmark results, CPU profiles, and performance analysis documents are stored in `results/`.

### Why a Dedicated Results Folder?

1. **Organization**: Keeps benchmark data separate from source code
2. **History**: Easy to track performance over time
3. **Sharing**: Simple to share/archive performance data
4. **Git**: Can be selectively committed or ignored
5. **Automation**: Scripts know where to save outputs

### What Goes in results/?

✅ **Include:**
- Benchmark results (`benchmark_results_*.txt`)
- CPU profiles (`cpu_*.prof`)
- Profile analysis (`profile_*.txt`, `profile_tree_*.txt`)
- Flamegraphs (`flamegraph.svg`)
- Analysis documents (`FLAMEGRAPH_ANALYSIS.md`, etc.)

❌ **Exclude:**
- Source code
- Test files
- Build artifacts
- Temporary files

## 🔧 Using the Results Folder

### Running Benchmarks

Use the automated script:
```bash
./scripts/benchmark.sh
```

This automatically:
- Runs all benchmarks
- Generates CPU profiles
- Creates analysis reports
- Saves everything to `results/` with timestamps

### Manual Benchmarking

Save to results folder:
```bash
# Benchmark with profiling
go test -bench=BenchmarkBloomFilterWithSIMD \
  -cpuprofile=results/cpu_$(date +%Y%m%d).prof \
  -run=^$ -benchtime=2s > results/benchmark_$(date +%Y%m%d).txt

# Generate analysis
go tool pprof -text results/cpu_$(date +%Y%m%d).prof > results/profile_$(date +%Y%m%d).txt
```

### Viewing Results

```bash
# View latest benchmark
cat results/benchmark_results_*.txt | tail -20

# View interactive flamegraph
go tool pprof -http=:8080 results/cpu_final.prof

# View profile summary
cat results/profile_final.txt
```

## 📋 File Naming Conventions

### Timestamp Format
Use `YYYYMMDD_HHMMSS` for uniqueness:
```
benchmark_results_20251018_192000.txt
cpu_20251018_192000.prof
profile_20251018_192000.txt
```

### Named Versions
For milestone results, use descriptive names:
```
benchmark_results_final.txt       # Latest stable
benchmark_results_optimized.txt   # After optimization
cpu_baseline.prof                 # Baseline profile
cpu_final.prof                    # Latest profile
```

### Analysis Documents
Use descriptive names:
```
FLAMEGRAPH_ANALYSIS.md
OPTIMIZATION_RESULTS.md
PROFILING_COMPARISON.md
```

## 🧹 Cleanup Guidelines

### Keep
- Latest 3 optimization cycles
- Milestone results (baseline, final, major optimizations)
- All analysis documents

### Archive (after 1 month)
- Intermediate benchmark runs
- Experimental profiles
- Debug outputs

### Delete
- Duplicate results
- Failed runs
- Temporary files

## 📝 .gitignore Configuration

The `.gitignore` is configured to:

✅ **Ignore in root:**
```gitignore
/*.prof              # Don't commit profiles to root
/profile_*.txt       # Don't commit analysis to root
/benchmark_results*.txt  # Don't commit benchmarks to root
```

✅ **Keep in results/:**
```
results/             # Track results folder
results/*.md         # Track analysis documents
results/*.prof       # Can optionally track profiles
```

This keeps the root clean while allowing selective tracking of important results.

## 🚀 Quick Reference

### Run All Benchmarks
```bash
./scripts/benchmark.sh
```

### View Latest Results
```bash
ls -lt results/ | head -10
```

### Interactive Flamegraph
```bash
go tool pprof -http=:8080 results/cpu_final.prof
```

### Compare Profiles
```bash
go tool pprof -base=results/cpu_before.prof results/cpu_after.prof
```

### Clean Old Results
```bash
# Keep only last 30 days
find results/ -name "*.prof" -mtime +30 -delete
find results/ -name "benchmark_*" -mtime +30 -delete
```

## 📚 Documentation

- **Main README**: Project overview and usage
- **results/README.md**: Benchmark results and analysis index
- **FLAMEGRAPH_ANALYSIS.md**: CPU profiling analysis
- **OPTIMIZATION_RESULTS.md**: Optimization tracking
- **FOLDER_ORGANIZATION.md**: This document

## ✨ Benefits

1. **Clean Root**: Source files are easy to find
2. **Organized Results**: All performance data in one place
3. **Easy Sharing**: `tar czf results.tar.gz results/`
4. **Git Friendly**: Selective tracking of important results
5. **Automated**: Scripts handle file placement
6. **Searchable**: Easy to find historical data

---

Last Updated: October 18, 2025
