# Benchmark Workflow Guide

This guide explains how to run benchmarks, organize results, and analyze them using the Jupyter notebook.

## Quick Start

### 1. Run Benchmarks

Use the benchmark script to run a complete benchmark suite:

```bash
bash scripts/benchmark.sh
```

This will:
- Create a new timestamped folder: `results/run_YYYYMMDD_HHMMSS_analysis/`
- Run all benchmarks and save results
- Generate CPU profiles
- Create profile analysis files
- All related files are kept together in one folder

**Output structure:**
```
results/run_20251030_143022_analysis/
├── benchmark_full_suite.txt      # Full benchmark results
├── simd_comparison.txt           # SIMD vs scalar comparison
├── bloom_benchmark.txt           # Bloom filter specific benchmarks
├── cpu_profile.prof              # CPU profile data
├── profile_text.txt              # Text profile analysis
└── profile_tree.txt              # Call tree analysis
```

### 2. Analyze Results

Open the Jupyter notebook:

```bash
jupyter notebook benchmark_analyzer.ipynb
```

Update the configuration in cell 3:

```python
# Option 1: Analyze the latest benchmark run
RUN_FOLDER = r"results\run_20251030_143022_analysis"  # Update timestamp
BENCHMARK_FILE = os.path.join(RUN_FOLDER, "benchmark_full_suite.txt")
PROFILE_FILE = os.path.join(RUN_FOLDER, "cpu_profile.prof")
PROFILE_TREE_FILE = os.path.join(RUN_FOLDER, "profile_tree.txt")
```

Run all cells to generate:
- Performance visualizations
- SIMD speedup analysis
- Memory allocation analysis
- CPU profile charts
- Interactive flamegraph
- HTML report with all results

**Analysis outputs (saved in the run folder):**
```
results/run_20251030_143022_analysis/
├── graphs/
│   ├── simd_speedup_analysis.png
│   ├── performance_scaling.png
│   ├── memory_allocation_analysis.png
│   ├── hybrid_mode_comparison.png
│   └── cpu_profile_analysis.png
├── data/
│   ├── all_benchmarks.csv
│   ├── simd_comparison.csv
│   ├── cpu_profile.csv
│   └── metadata.csv
├── flamegraph.html
└── benchmark_analysis_report.html
```

## Organizing Old Results

If you have existing unorganized benchmark files in the `results/` folder, use the organizer script:

```bash
# Preview what will be organized (dry run)
python organize_results.py --dry-run

# Actually organize the files
python organize_results.py
```

This will:
- Group related files by their analysis name
- Create organized folders like `array_based_20251026_analysis/`
- Move related files together
- Keep `.md` documentation files at the root level

## Manual Benchmark Runs

For custom benchmark runs with specific options:

```bash
# Create a timestamped folder
DATE=$(date +%Y%m%d_%H%M%S)
RESULTS_DIR="results/custom_${DATE}_analysis"
mkdir -p "$RESULTS_DIR"

# Run specific benchmarks
go test -bench=BenchmarkSIMDvsScalar \
  -benchmem -run=^$ -benchtime=1s \
  > "$RESULTS_DIR/simd_comparison.txt"

# Run with CPU profiling
go test -bench=BenchmarkBloomFilterWithSIMD \
  -cpuprofile="$RESULTS_DIR/cpu_profile.prof" \
  -run=^$ -benchtime=2s \
  > "$RESULTS_DIR/bloom_benchmark.txt"

# Generate profile analysis
go tool pprof -text -nodecount=30 \
  "$RESULTS_DIR/cpu_profile.prof" \
  > "$RESULTS_DIR/profile_text.txt"

go tool pprof -tree -nodecount=30 \
  "$RESULTS_DIR/cpu_profile.prof" \
  > "$RESULTS_DIR/profile_tree.txt"
```

Then analyze using the notebook by pointing to your custom folder.

## SIMD Comparison Tests

To run comprehensive SIMD comparison tests:

```bash
# Run SIMD comparison benchmarks
go test -tags=simd_comparison \
  -bench=BenchmarkSIMDvsScalar \
  -benchmem -run=^$ -benchtime=2s \
  > "results/run_$(date +%Y%m%d_%H%M%S)_analysis/simd_comparison.txt"
```

## Folder Naming Convention

The system uses these folder naming patterns:

- **Benchmark runs:** `run_YYYYMMDD_HHMMSS_analysis/`
  - Created by `scripts/benchmark.sh`
  - Contains complete benchmark suite for one run

- **Organized files:** `{analysis_name}_YYYYMMDD_analysis/`
  - Created by `organize_results.py`
  - Groups related files by common prefix (e.g., `array_based`, `clear_optimization`)

- **Manual analysis:** `{custom_name}_analysis/`
  - Created by Jupyter notebook when analyzing individual files
  - User-defined name

## Best Practices

1. **Use the benchmark script** for complete analysis runs
   ```bash
   bash scripts/benchmark.sh
   ```

2. **Keep each run's files together** - don't split related files across folders

3. **Use descriptive folder names** for manual runs
   ```bash
   RESULTS_DIR="results/fix_memory_leak_20251030_analysis"
   ```

4. **Archive old results** to keep the results folder manageable
   ```bash
   mkdir -p results/archive/2024-10
   mv results/run_202410*_analysis results/archive/2024-10/
   ```

5. **Document significant runs** - add a README.md in important analysis folders
   ```bash
   echo "# Array-based optimization results" > results/array_based_20251026_analysis/README.md
   ```

## Troubleshooting

### "Benchmark file not found" error
- Check that the path in the notebook configuration is correct
- Verify the file exists: `ls results/run_*_analysis/`
- Use tab completion or copy the full path

### No SIMD comparison data
- Ensure the benchmark file contains SIMD tests
- Run with the full suite: `bash scripts/benchmark.sh`
- Check the file contains lines with "SIMDvsScalar"

### CPU profile analysis fails
- Verify Go is installed and in PATH: `go version`
- Check the .prof file exists and is not empty
- Try generating manually: `go tool pprof -text cpu_profile.prof`

### Flamegraph not generating
- Ensure `profile_tree_parser.py` is in the same directory as the notebook
- Check that `profile_tree.txt` contains valid pprof tree output
- Verify the file is not empty: `wc -l results/*/profile_tree.txt`

## Additional Resources

- [benchmark_analyzer.ipynb](benchmark_analyzer.ipynb) - Main analysis notebook
- [organize_results.py](organize_results.py) - Results organization script
- [profile_tree_parser.py](profile_tree_parser.py) - Profile parsing module
- [FLAMEGRAPH_USAGE.md](FLAMEGRAPH_USAGE.md) - Flamegraph guide
- [scripts/benchmark.sh](scripts/benchmark.sh) - Automated benchmark runner
