#!/bin/bash

# Benchmark Runner Script
# Automatically saves all results to timestamped folders under results/

set -e

RESULTS_BASE="results"
DATE=$(date +%Y%m%d_%H%M%S)
RESULTS_DIR="${RESULTS_BASE}/run_${DATE}_analysis"

# Create results directory for this run
mkdir -p "$RESULTS_DIR"

echo "Running benchmarks and profiling..."
echo "Results will be saved to: $RESULTS_DIR/"
echo ""

# Run full benchmark suite
echo "Running full benchmark suite..."
go test -bench=. -benchmem -run=^$ -benchtime=1s > "$RESULTS_DIR/benchmark_full_suite.txt"
echo "Saved to: $RESULTS_DIR/benchmark_full_suite.txt"

# Run SIMD comparison benchmarks (integration tests)
echo "Running SIMD comparison benchmarks..."
go test -tags=simd_comparison ./tests/integration -bench=BenchmarkSIMDvsScalar -benchmem -run=^$ -benchtime=1s > "$RESULTS_DIR/simd_comparison.txt"
echo "Saved to: $RESULTS_DIR/simd_comparison.txt"

# Run bloom filter benchmarks with CPU profiling
echo "Running bloom filter benchmarks with CPU profiling..."
go test -bench=BenchmarkBloomFilterWithSIMD -cpuprofile="$RESULTS_DIR/cpu_profile.prof" -run=^$ -benchtime=2s > "$RESULTS_DIR/bloom_benchmark.txt"
echo "Saved CPU profile to: $RESULTS_DIR/cpu_profile.prof"
echo "Saved benchmark to: $RESULTS_DIR/bloom_benchmark.txt"

# Generate profile analysis
echo "Generating profile analysis..."
go tool pprof -text -nodecount=30 "$RESULTS_DIR/cpu_profile.prof" > "$RESULTS_DIR/profile_text.txt"
echo "Saved to: $RESULTS_DIR/profile_text.txt"

# Generate call tree
echo "Generating call tree..."
go tool pprof -tree -nodecount=30 "$RESULTS_DIR/cpu_profile.prof" > "$RESULTS_DIR/profile_tree.txt"
echo "Saved to: $RESULTS_DIR/profile_tree.txt"

echo ""
echo "Benchmark complete! Results saved to $RESULTS_DIR/"
echo ""
echo "Quick Summary:"
grep "BenchmarkBloomFilterWithSIMD" "$RESULTS_DIR/bloom_benchmark.txt" | tail -3
echo ""

# Update results/README.md with latest results
echo "Updating results/README.md..."
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
FOLDER_NAME=$(basename "$RESULTS_DIR")

# Extract key metrics from benchmark results
BLOOM_METRICS=$(grep "BenchmarkBloomFilterWithSIMD" "$RESULTS_DIR/bloom_benchmark.txt" | head -3)
SIMD_METRICS=$(grep "BenchmarkSIMDvsScalar" "$RESULTS_DIR/simd_comparison.txt" 2>/dev/null | head -4)

# Update README with sed (create backup, then replace)
README_PATH="${RESULTS_BASE}/README.md"

# Update timestamp
sed -i "s/\*\*Last Updated:\*\*.*/\*\*Last Updated:\*\* $TIMESTAMP/" "$README_PATH"

# Update results folder
sed -i "s/\*\*Results Folder:\*\*.*/\*\*Results Folder:\*\* $FOLDER_NAME/" "$README_PATH"

# Update status
sed -i "s/\*\*Status:\*\*.*/\*\*Status:\*\* Benchmark completed successfully/" "$README_PATH"

echo "✓ README.md updated with latest results"
echo ""

echo "To view interactive flamegraph:"
echo "   go tool pprof -http=:8080 $RESULTS_DIR/cpu_profile.prof"
echo ""
echo "All results in: $RESULTS_DIR/"
ls -lh "$RESULTS_DIR/"
