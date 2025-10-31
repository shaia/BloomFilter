#!/bin/bash

# Comprehensive Profiling Script
# Generates CPU, memory, goroutine, block, and mutex profiles for the Bloom filter

set -e

# Configuration
RESULTS_BASE="results"
DATE=$(date +%Y%m%d_%H%M%S)
PROFILE_DIR="${RESULTS_BASE}/profile_${DATE}_analysis"

# Benchmark to profile (can be customized)
BENCHMARK_PATTERN="${1:-BenchmarkBloomFilterWithSIMD}"
BENCHMARK_TIME="${2:-5s}"

# Create results directory
mkdir -p "$PROFILE_DIR"

echo "================================================================================"
echo "COMPREHENSIVE PROFILING SCRIPT"
echo "================================================================================"
echo ""
echo "Profile Directory: $PROFILE_DIR"
echo "Benchmark Pattern: $BENCHMARK_PATTERN"
echo "Benchmark Time: $BENCHMARK_TIME"
echo ""
echo "Generating profiles..."
echo ""

# =============================================================================
# CPU PROFILE
# =============================================================================
echo "[1/5] Generating CPU profile..."
go test -bench="$BENCHMARK_PATTERN" \
    -cpuprofile="$PROFILE_DIR/cpu.prof" \
    -run=^$ \
    -benchtime="$BENCHMARK_TIME" \
    > "$PROFILE_DIR/benchmark_output.txt"
echo "✓ CPU profile saved: $PROFILE_DIR/cpu.prof"

# Generate CPU profile analysis
echo "      Analyzing CPU profile..."
go tool pprof -text -nodecount=50 "$PROFILE_DIR/cpu.prof" > "$PROFILE_DIR/cpu_analysis.txt"
go tool pprof -tree -nodecount=30 "$PROFILE_DIR/cpu.prof" > "$PROFILE_DIR/cpu_tree.txt"
go tool pprof -top -nodecount=20 "$PROFILE_DIR/cpu.prof" > "$PROFILE_DIR/cpu_top.txt"
echo "      ✓ CPU analysis files generated"
echo ""

# =============================================================================
# MEMORY PROFILE (Heap Allocations)
# =============================================================================
echo "[2/5] Generating memory (heap) profile..."
go test -bench="$BENCHMARK_PATTERN" \
    -memprofile="$PROFILE_DIR/mem.prof" \
    -run=^$ \
    -benchtime="$BENCHMARK_TIME" \
    > /dev/null 2>&1
echo "✓ Memory profile saved: $PROFILE_DIR/mem.prof"

# Generate memory profile analysis
echo "      Analyzing memory profile..."
go tool pprof -text -nodecount=50 "$PROFILE_DIR/mem.prof" > "$PROFILE_DIR/mem_analysis.txt"
go tool pprof -tree -nodecount=30 "$PROFILE_DIR/mem.prof" > "$PROFILE_DIR/mem_tree.txt"
go tool pprof -top -nodecount=20 "$PROFILE_DIR/mem.prof" > "$PROFILE_DIR/mem_top.txt"
# Allocation analysis
go tool pprof -alloc_space -text -nodecount=30 "$PROFILE_DIR/mem.prof" > "$PROFILE_DIR/mem_alloc_space.txt"
go tool pprof -alloc_objects -text -nodecount=30 "$PROFILE_DIR/mem.prof" > "$PROFILE_DIR/mem_alloc_objects.txt"
echo "      ✓ Memory analysis files generated"
echo ""

# =============================================================================
# BLOCK PROFILE (Contention on blocking operations)
# =============================================================================
echo "[3/5] Generating block profile..."
go test -bench="$BENCHMARK_PATTERN" \
    -blockprofile="$PROFILE_DIR/block.prof" \
    -run=^$ \
    -benchtime="$BENCHMARK_TIME" \
    > /dev/null 2>&1
echo "✓ Block profile saved: $PROFILE_DIR/block.prof"

# Generate block profile analysis if file has content
if [ -s "$PROFILE_DIR/block.prof" ]; then
    echo "      Analyzing block profile..."
    go tool pprof -text -nodecount=50 "$PROFILE_DIR/block.prof" > "$PROFILE_DIR/block_analysis.txt" 2>/dev/null || echo "      (No blocking detected)"
else
    echo "      (No blocking detected - profile empty)"
fi
echo ""

# =============================================================================
# MUTEX PROFILE (Lock contention)
# =============================================================================
echo "[4/5] Generating mutex profile..."
go test -bench="$BENCHMARK_PATTERN" \
    -mutexprofile="$PROFILE_DIR/mutex.prof" \
    -run=^$ \
    -benchtime="$BENCHMARK_TIME" \
    > /dev/null 2>&1
echo "✓ Mutex profile saved: $PROFILE_DIR/mutex.prof"

# Generate mutex profile analysis if file has content
if [ -s "$PROFILE_DIR/mutex.prof" ]; then
    echo "      Analyzing mutex profile..."
    go tool pprof -text -nodecount=50 "$PROFILE_DIR/mutex.prof" > "$PROFILE_DIR/mutex_analysis.txt" 2>/dev/null || echo "      (No mutex contention detected)"
else
    echo "      (No mutex contention detected - profile empty)"
fi
echo ""

# =============================================================================
# GOROUTINE PROFILE
# =============================================================================
echo "[5/5] Generating goroutine profile (from trace)..."
# Note: Goroutine profile requires a running program or execution trace
# For benchmarks, we can generate an execution trace
go test -bench="$BENCHMARK_PATTERN" \
    -trace="$PROFILE_DIR/trace.out" \
    -run=^$ \
    -benchtime="$BENCHMARK_TIME" \
    > /dev/null 2>&1
echo "✓ Execution trace saved: $PROFILE_DIR/trace.out"
echo "      (View with: go tool trace $PROFILE_DIR/trace.out)"
echo ""

# =============================================================================
# SUMMARY AND ANALYSIS
# =============================================================================
echo "================================================================================"
echo "PROFILE SUMMARY"
echo "================================================================================"
echo ""

# Display file sizes
echo "Generated files:"
ls -lh "$PROFILE_DIR/" | awk 'NR>1 {printf "  %-40s %8s\n", $9, $5}'
echo ""

# Quick CPU analysis summary
echo "--------------------------------------------------------------------------------"
echo "TOP 10 CPU CONSUMERS:"
echo "--------------------------------------------------------------------------------"
head -n 15 "$PROFILE_DIR/cpu_top.txt" | tail -n 11
echo ""

# Quick memory analysis summary
echo "--------------------------------------------------------------------------------"
echo "TOP 10 MEMORY ALLOCATORS:"
echo "--------------------------------------------------------------------------------"
head -n 15 "$PROFILE_DIR/mem_top.txt" | tail -n 11
echo ""

# Benchmark results summary
echo "--------------------------------------------------------------------------------"
echo "BENCHMARK RESULTS:"
echo "--------------------------------------------------------------------------------"
grep "^Benchmark" "$PROFILE_DIR/benchmark_output.txt" | head -10
echo ""

echo "================================================================================"
echo "PROFILE COMPLETE"
echo "================================================================================"
echo ""
echo "All profiles saved to: $PROFILE_DIR/"
echo ""
echo "View profiles interactively:"
echo "  CPU:    go tool pprof -http=:8080 $PROFILE_DIR/cpu.prof"
echo "  Memory: go tool pprof -http=:8081 $PROFILE_DIR/mem.prof"
echo "  Trace:  go tool trace $PROFILE_DIR/trace.out"
echo ""
echo "Generate flamegraphs (if you have go-torch installed):"
echo "  go-torch --file=$PROFILE_DIR/cpu_flamegraph.svg $PROFILE_DIR/cpu.prof"
echo "  go-torch --alloc_objects --file=$PROFILE_DIR/mem_flamegraph.svg $PROFILE_DIR/mem.prof"
echo ""
echo "Compare profiles (if you have a baseline):"
echo "  go tool pprof -base=baseline.prof -http=:8080 $PROFILE_DIR/cpu.prof"
echo ""
