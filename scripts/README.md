# Scripts Directory

This directory contains automation scripts for benchmarking and profiling the BloomFilter project.

## Available Scripts

### 1. benchmark.sh - Comprehensive Benchmark Suite

Runs a complete benchmark suite with CPU profiling and analysis.

**Usage:**
```bash
bash scripts/benchmark.sh
```

**What it does:**
1. Runs full benchmark suite (all benchmarks)
2. Runs SIMD comparison benchmarks (integration tests)
3. Runs bloom filter benchmarks with CPU profiling
4. Generates CPU profile analysis (text and tree format)
5. Saves all results to timestamped folder

**Output structure:**
```
results/run_YYYYMMDD_HHMMSS_analysis/
├── benchmark_full_suite.txt    # All benchmark results
├── simd_comparison.txt          # SIMD vs fallback comparison
├── bloom_benchmark.txt          # Bloom filter specific benchmarks
├── cpu_profile.prof             # CPU profile data
├── profile_text.txt             # CPU profile analysis (text)
└── profile_tree.txt             # CPU profile call tree
```

**Interactive viewing:**
```bash
# View CPU profile in browser
go tool pprof -http=:8080 results/run_*/cpu_profile.prof

# View specific analysis
cat results/run_*/profile_text.txt
```

**See:** [BENCHMARK_WORKFLOW.md](BENCHMARK_WORKFLOW.md) for detailed workflow guide.

---

### 2. profile.sh - Comprehensive Profiling Script

Generates detailed profiles for CPU, memory, goroutines, blocking, and mutex contention.

**Usage:**
```bash
# Profile default benchmark (BenchmarkBloomFilterWithSIMD) for 5 seconds
bash scripts/profile.sh

# Profile specific benchmark
bash scripts/profile.sh BenchmarkInsertion

# Profile with custom duration
bash scripts/profile.sh BenchmarkInsertion 10s

# Profile SIMD comparison
bash scripts/profile.sh BenchmarkSIMDvsScalar 3s
```

**What it does:**
1. **CPU Profile** - Where time is spent
2. **Memory Profile** - Heap allocations and memory usage
3. **Block Profile** - Contention on blocking operations
4. **Mutex Profile** - Lock contention
5. **Execution Trace** - Detailed runtime events

**Output structure:**
```
results/profile_YYYYMMDD_HHMMSS_analysis/
├── benchmark_output.txt        # Benchmark results
├── cpu.prof                    # CPU profile (binary)
├── cpu_analysis.txt            # CPU analysis (text)
├── cpu_tree.txt                # CPU call tree
├── cpu_top.txt                 # Top CPU consumers
├── mem.prof                    # Memory profile (binary)
├── mem_analysis.txt            # Memory analysis
├── mem_tree.txt                # Memory call tree
├── mem_top.txt                 # Top memory allocators
├── mem_alloc_space.txt         # Total allocation space
├── mem_alloc_objects.txt       # Allocation count
├── block.prof                  # Block profile
├── block_analysis.txt          # Block analysis (if applicable)
├── mutex.prof                  # Mutex profile
├── mutex_analysis.txt          # Mutex analysis (if applicable)
└── trace.out                   # Execution trace
```

**Interactive viewing:**
```bash
# CPU profile in browser
go tool pprof -http=:8080 results/profile_*/cpu.prof

# Memory profile in browser
go tool pprof -http=:8081 results/profile_*/mem.prof

# Execution trace viewer
go tool trace results/profile_*/trace.out

# Compare with baseline
go tool pprof -base=baseline_cpu.prof -http=:8080 results/profile_*/cpu.prof
```

**Analysis commands:**
```bash
# Top CPU consumers
go tool pprof -top results/profile_*/cpu.prof

# Top memory allocators
go tool pprof -top results/profile_*/mem.prof

# Allocation sites
go tool pprof -alloc_space -top results/profile_*/mem.prof

# Generate flamegraph (requires graphviz)
go tool pprof -svg results/profile_*/cpu.prof > cpu_flamegraph.svg
```

---

## Profile Types Explained

### CPU Profile (`cpu.prof`)
- **What:** Where your program spends CPU time
- **Use for:** Finding performance bottlenecks, slow functions
- **Look for:** Hot paths, unexpected function calls, inefficient algorithms

### Memory Profile (`mem.prof`)
- **What:** Heap allocations and memory usage
- **Use for:** Finding memory leaks, excessive allocations
- **Look for:** High allocation counts, large memory consumers
- **Modes:**
  - `alloc_space` - Total bytes allocated
  - `alloc_objects` - Total number of allocations
  - `inuse_space` - Currently used memory
  - `inuse_objects` - Current object count

### Block Profile (`block.prof`)
- **What:** Time spent blocked on synchronization primitives
- **Use for:** Finding contention on channels, mutex locks
- **Look for:** High blocking time, channel bottlenecks

### Mutex Profile (`mutex.prof`)
- **What:** Contention on mutexes
- **Use for:** Finding lock contention issues
- **Look for:** Contested locks, long mutex hold times

### Execution Trace (`trace.out`)
- **What:** Detailed timeline of program execution
- **Use for:** Visualizing goroutine execution, scheduler behavior
- **Look for:** Goroutine blocking, scheduler issues, GC pauses

---

## Common Workflows

### Workflow 1: Initial Performance Analysis
```bash
# 1. Run benchmarks to see overall performance
bash scripts/benchmark.sh

# 2. Check results
cat results/run_*/benchmark_full_suite.txt

# 3. If performance issues, generate profiles
bash scripts/profile.sh
```

### Workflow 2: Memory Optimization
```bash
# 1. Generate memory profile
bash scripts/profile.sh BenchmarkInsertion 10s

# 2. View memory allocations
go tool pprof -alloc_space -top results/profile_*/mem.prof

# 3. Find allocation sites
go tool pprof -alloc_objects -list=FunctionName results/profile_*/mem.prof

# 4. Interactive exploration
go tool pprof -http=:8080 results/profile_*/mem.prof
```

### Workflow 3: CPU Optimization
```bash
# 1. Generate CPU profile
bash scripts/profile.sh BenchmarkLookup 10s

# 2. Find hot functions
go tool pprof -top results/profile_*/cpu.prof

# 3. Examine specific function
go tool pprof -list=FunctionName results/profile_*/cpu.prof

# 4. View flamegraph
go tool pprof -http=:8080 results/profile_*/cpu.prof
```

### Workflow 4: Comparative Analysis
```bash
# 1. Create baseline profile
bash scripts/profile.sh BenchmarkInsertion 10s
mv results/profile_* results/baseline_profile

# 2. Make optimizations to code

# 3. Generate new profile
bash scripts/profile.sh BenchmarkInsertion 10s

# 4. Compare
go tool pprof -base=results/baseline_profile/cpu.prof \
    -http=:8080 results/profile_*/cpu.prof
```

---

## Tips and Best Practices

### For Accurate Profiling:

1. **Run for sufficient time** - At least 5-10 seconds for meaningful data
   ```bash
   bash scripts/profile.sh BenchmarkName 10s
   ```

2. **Close other applications** - Reduce system noise

3. **Run multiple times** - Check for consistency
   ```bash
   for i in {1..3}; do bash scripts/profile.sh BenchmarkName 10s; done
   ```

4. **Profile in production-like conditions** - Realistic data sizes

5. **Use `-benchtime`** to control duration:
   - `1s` - Quick check
   - `5s` - Standard profiling
   - `10s` - Detailed analysis

### Interpreting Results:

- **Flat %** - Time spent in this function alone
- **Cum %** - Time spent in this function + its callees
- **High flat %** - Direct optimization opportunity
- **High cum %** - Look at callees for optimization

### Flamegraph Interpretation:

- **Width** - Proportion of CPU time
- **Height** - Call stack depth
- **Color** - Different functions/packages
- **Wide bars** - Hot paths (optimize these!)

---

## Troubleshooting

### "no profile data" or empty profiles
- Benchmark may be too short
- Try longer `-benchtime`: `bash scripts/profile.sh BenchmarkName 10s`

### Block/Mutex profiles empty
- Good news! No contention detected
- These profiles only show data when there's actual blocking

### "undefined: BenchmarkName"
- Check benchmark name exists: `go test -bench=. -run=^$ -list=.`
- Use correct pattern: `bash scripts/profile.sh BenchmarkInsertion`

### pprof command not found
- Ensure Go is installed and in PATH
- Try: `go version`

---

## Best Practices

### Test Binary Management

**Our scripts automatically handle output directories:**
- `benchmark.sh` outputs to `results/run_YYYYMMDD_HHMMSS_analysis/`
- `profile.sh` outputs to `results/profile_YYYYMMDD_HHMMSS_analysis/`
- All profiles and results are automatically organized

**If manually building test binaries** with `go test -c`, use designated directories:

```bash
# Bad: Creates test binary in current directory
go test -c

# Good: Output to designated directory
mkdir -p testbin
go test -c -o testbin/bloomfilter.test

# Or use build directory
mkdir -p build
go test -c -o build/bloomfilter.test
```

**Note:** The `.gitignore` is configured to:
- Ignore all `*.exe` and `*.test` files anywhere in the project
- Ignore entire `/build/`, `/output/`, and `/testbin/` directories
- Ignore `results/*` except README.md and analysis tools

This ensures test binaries and temporary files won't be accidentally committed.

---

## Additional Resources

- [Go Profiling Documentation](https://go.dev/blog/pprof)
- [Go Execution Tracer](https://go.dev/blog/execution-tracer)
- [Profiling Go Programs](https://go.dev/blog/profiling-go-programs)
- [BENCHMARK_WORKFLOW.md](BENCHMARK_WORKFLOW.md) - Benchmark automation guide
- [../TESTING.md](../TESTING.md) - Complete testing guide
