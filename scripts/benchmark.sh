#!/bin/bash

# Benchmark Runner Script
# Automatically saves all results to timestamped folders under results/

# Note: Not using set -e to allow script to continue if SIMD comparison fails

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
if go test -tags=simd_comparison ./tests/integration -bench=BenchmarkSIMDvsScalar -benchmem -run=^$ -benchtime=1s > "$RESULTS_DIR/simd_comparison.txt" 2>&1; then
    echo "Saved to: $RESULTS_DIR/simd_comparison.txt"
else
    echo "WARNING: SIMD comparison benchmarks failed, continuing with other benchmarks..."
    echo "SIMD comparison benchmarks failed" > "$RESULTS_DIR/simd_comparison.txt"
fi

# Run selected benchmarks with CPU profiling (excluding very slow ones)
echo "Running benchmarks with CPU profiling..."
# Exclude BenchmarkHybridMemoryAllocation and BenchmarkHybridThroughput as they take too long
# Run main package benchmarks first, then integration tests with separate profiles
go test -bench='Benchmark(Cache|Insertion|Lookup|FalsePositives|Comprehensive|HybridModes|HybridCrossover)' -cpuprofile="$RESULTS_DIR/cpu_profile_main.prof" -run=^$ -benchtime=1s > "$RESULTS_DIR/profiled_benchmarks.txt" 2>&1
MAIN_EXIT_CODE=$?

# Run SIMD benchmarks with separate profile
go test -tags=simd_comparison ./tests/integration -bench=BenchmarkBloomFilterWithSIMD -cpuprofile="$RESULTS_DIR/cpu_profile_simd.prof" -run=^$ -benchtime=1s >> "$RESULTS_DIR/profiled_benchmarks.txt" 2>&1
SIMD_EXIT_CODE=$?

# Use main profile for overall analysis (it typically has more comprehensive data)
PROFILE_EXIT_CODE=$MAIN_EXIT_CODE
echo "Saved benchmark to: $RESULTS_DIR/profiled_benchmarks.txt"

# Check both profile files
MAIN_PROFILE_OK=0
SIMD_PROFILE_OK=0

if [ -s "$RESULTS_DIR/cpu_profile_main.prof" ]; then
    PROFILE_SIZE=$(stat -f%z "$RESULTS_DIR/cpu_profile_main.prof" 2>/dev/null || stat -c%s "$RESULTS_DIR/cpu_profile_main.prof" 2>/dev/null || echo "0")
    echo "Saved main CPU profile to: $RESULTS_DIR/cpu_profile_main.prof ($PROFILE_SIZE bytes)"
    MAIN_PROFILE_OK=1
else
    echo "WARNING: Main CPU profile is empty or not generated"
    if [ $MAIN_EXIT_CODE -ne 0 ]; then
        echo "Main benchmark command exited with code: $MAIN_EXIT_CODE"
    fi
fi

if [ -s "$RESULTS_DIR/cpu_profile_simd.prof" ]; then
    PROFILE_SIZE=$(stat -f%z "$RESULTS_DIR/cpu_profile_simd.prof" 2>/dev/null || stat -c%s "$RESULTS_DIR/cpu_profile_simd.prof" 2>/dev/null || echo "0")
    echo "Saved SIMD CPU profile to: $RESULTS_DIR/cpu_profile_simd.prof ($PROFILE_SIZE bytes)"
    SIMD_PROFILE_OK=1
else
    echo "WARNING: SIMD CPU profile is empty or not generated"
    if [ $SIMD_EXIT_CODE -ne 0 ]; then
        echo "SIMD benchmark command exited with code: $SIMD_EXIT_CODE"
    fi
fi

# Generate profile analysis for main benchmarks if available
if [ $MAIN_PROFILE_OK -eq 1 ]; then
    echo "Generating main profile analysis..."
    go tool pprof -text -nodecount=30 "$RESULTS_DIR/cpu_profile_main.prof" > "$RESULTS_DIR/profile_main_text.txt"
    echo "Saved to: $RESULTS_DIR/profile_main_text.txt"

    echo "Generating main call tree..."
    go tool pprof -tree -nodecount=30 "$RESULTS_DIR/cpu_profile_main.prof" > "$RESULTS_DIR/profile_main_tree.txt"
    echo "Saved to: $RESULTS_DIR/profile_main_tree.txt"
fi

# Generate profile analysis for SIMD benchmarks if available
if [ $SIMD_PROFILE_OK -eq 1 ]; then
    echo "Generating SIMD profile analysis..."
    go tool pprof -text -nodecount=30 "$RESULTS_DIR/cpu_profile_simd.prof" > "$RESULTS_DIR/profile_simd_text.txt"
    echo "Saved to: $RESULTS_DIR/profile_simd_text.txt"

    echo "Generating SIMD call tree..."
    go tool pprof -tree -nodecount=30 "$RESULTS_DIR/cpu_profile_simd.prof" > "$RESULTS_DIR/profile_simd_tree.txt"
    echo "Saved to: $RESULTS_DIR/profile_simd_tree.txt"
fi

echo ""
echo "Benchmark complete! Results saved to $RESULTS_DIR/"
echo ""
echo "Quick Summary:"
echo "Total benchmarks run:"
grep "^Benchmark" "$RESULTS_DIR/benchmark_full_suite.txt" | wc -l
echo ""
echo "Sample results:"
grep "^Benchmark" "$RESULTS_DIR/benchmark_full_suite.txt" | head -5
echo ""

# Update results/README.md with latest results
echo "Updating results/README.md..."
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
FOLDER_NAME=$(basename "$RESULTS_DIR")
README_PATH="${RESULTS_BASE}/README.md"

# Update timestamp
sed -i "s/\*\*Last Updated:\*\*.*/\*\*Last Updated:\*\* $TIMESTAMP/" "$README_PATH"

# Update results folder
sed -i "s/\*\*Results Folder:\*\*.*/\*\*Results Folder:\*\* $FOLDER_NAME/" "$README_PATH"

# Update status
sed -i "s/\*\*Status:\*\*.*/\*\*Status:\*\* Benchmark completed successfully/" "$README_PATH"

echo "✓ README.md updated with latest results"
echo ""

echo "To view interactive flamegraphs:"
if [ $MAIN_PROFILE_OK -eq 1 ]; then
    echo "   Main benchmarks: go tool pprof -http=:8080 $RESULTS_DIR/cpu_profile_main.prof"
fi
if [ $SIMD_PROFILE_OK -eq 1 ]; then
    echo "   SIMD benchmarks: go tool pprof -http=:8081 $RESULTS_DIR/cpu_profile_simd.prof"
fi
echo ""
echo "All results in: $RESULTS_DIR/"
ls -lh "$RESULTS_DIR/"
