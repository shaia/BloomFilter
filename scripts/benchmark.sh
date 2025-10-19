#!/bin/bash

# Benchmark Runner Script
# Automatically saves all results to the results/ folder

set -e

RESULTS_DIR="results"
DATE=$(date +%Y%m%d_%H%M%S)

# Create results directory if it doesn't exist
mkdir -p "$RESULTS_DIR"

echo "Running benchmarks and profiling..."
echo "Results will be saved to: $RESULTS_DIR/"
echo ""

# Run full benchmark suite
echo "Running full benchmark suite..."
go test -bench=. -benchmem -run=^$ -benchtime=1s > "$RESULTS_DIR/benchmark_results_${DATE}.txt"
echo "Saved to: $RESULTS_DIR/benchmark_results_${DATE}.txt"

# Run SIMD comparison benchmarks
echo "Running SIMD comparison benchmarks..."
go test -bench=BenchmarkSIMDvsScalar -benchmem -run=^$ -benchtime=1s > "$RESULTS_DIR/simd_comparison_${DATE}.txt"
echo "Saved to: $RESULTS_DIR/simd_comparison_${DATE}.txt"

# Run bloom filter benchmarks with CPU profiling
echo "Running bloom filter benchmarks with CPU profiling..."
go test -bench=BenchmarkBloomFilterWithSIMD -cpuprofile="$RESULTS_DIR/cpu_${DATE}.prof" -run=^$ -benchtime=2s > "$RESULTS_DIR/bloom_bench_${DATE}.txt"
echo "Saved CPU profile to: $RESULTS_DIR/cpu_${DATE}.prof"
echo "Saved benchmark to: $RESULTS_DIR/bloom_bench_${DATE}.txt"

# Generate profile analysis
echo "Generating profile analysis..."
go tool pprof -text -nodecount=30 "$RESULTS_DIR/cpu_${DATE}.prof" > "$RESULTS_DIR/profile_${DATE}.txt"
echo "Saved to: $RESULTS_DIR/profile_${DATE}.txt"

# Generate call tree
echo "Generating call tree..."
go tool pprof -tree -nodecount=30 "$RESULTS_DIR/cpu_${DATE}.prof" > "$RESULTS_DIR/profile_tree_${DATE}.txt"
echo "Saved to: $RESULTS_DIR/profile_tree_${DATE}.txt"

echo ""
echo "Benchmark complete! Results saved to $RESULTS_DIR/"
echo ""
echo "Quick Summary:"
grep "BenchmarkBloomFilterWithSIMD" "$RESULTS_DIR/bloom_bench_${DATE}.txt" | tail -3
echo ""
echo "To view interactive flamegraph:"
echo "   go tool pprof -http=:8080 $RESULTS_DIR/cpu_${DATE}.prof"
echo ""
echo "All results in: $RESULTS_DIR/"
ls -lh "$RESULTS_DIR"/*_${DATE}.* 2>/dev/null || true
