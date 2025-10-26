# Publishing Guide for go-simd-bloomfilter

## Current Status

- **Repository**: https://github.com/shaia/BloomFilter.git
- **Module Path**: `github.com/shaia/BloomFilter`
- **Current Branch**: `feature/hybrid-map-array-optimization`
- **Main Branch**: `main`

## Pre-Publishing Checklist

### 1. Fix Failing Tests

**Current Issue**: Flaky SIMD performance test
```
--- FAIL: TestSIMDPerformanceImprovement/Size_4096/VectorOr
```

**Action Required**:
```bash
# Option 1: Make performance test more lenient
# Edit simd_comparison_test.go to allow more variance

# Option 2: Skip flaky performance tests in CI
# Add build tag: // +build performance

# Option 3: Increase benchmark time for more stable results
```

### 2. Ensure All Tests Pass

```bash
# Run full test suite
go test ./... -v

# Run with race detector
go test ./... -race

# Run benchmarks to verify
go test -bench=. -run=^$
```

### 3. Verify Code Quality

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Check for issues
go mod tidy
go mod verify
```

### 4. Update Documentation

- [x] README.md - Updated with hybrid architecture
- [x] CHANGELOG.md - Has [Unreleased] section ready
- [x] Documentation reflects current state
- [ ] Update CHANGELOG version from [Unreleased] to [0.2.0]

## Publishing Steps

### Step 1: Merge Feature Branch to Main

```bash
# Ensure you're on the feature branch
git checkout feature/hybrid-map-array-optimization

# Make sure all changes are committed
git status

# Push latest changes to remote
git push origin feature/hybrid-map-array-optimization

# Switch to main
git checkout main

# Pull latest main
git pull origin main

# Merge feature branch (no fast-forward to preserve history)
git merge --no-ff feature/hybrid-map-array-optimization

# Resolve any conflicts if they occur
```

### Step 2: Update CHANGELOG for Release

```bash
# Edit CHANGELOG.md
# Change:
## [Unreleased]

# To:
## [0.2.0] - 2025-10-26

# Add at bottom:
[0.2.0]: https://github.com/shaia/BloomFilter/releases/tag/v0.2.0

# Commit the change
git add CHANGELOG.md
git commit -m "Prepare v0.2.0 release"
```

### Step 3: Create and Push Git Tag

```bash
# Create annotated tag
git tag -a v0.2.0 -m "Release v0.2.0: Hybrid array/map optimization

Major Features:
- Hybrid array/map architecture with automatic mode selection
- Zero-allocation array mode for small filters (10K-100K elements)
- Unlimited scalability with map mode for large filters
- 95% memory reduction for small filters
- 41% CPU improvement with clear() optimization
- 16.7% map overhead reduction

Performance:
- Small filters: 1.5x faster than alternatives, 0 B/op
- Large filters: Competitive with unlimited scalability
- Zero external dependencies

See CHANGELOG.md for complete details."

# Verify tag was created
git tag -l -n9 v0.2.0

# Push main branch
git push origin main

# Push tag
git push origin v0.2.0
```

### Step 4: Create GitHub Release

**Option A: Using GitHub Web Interface**
1. Go to https://github.com/shaia/BloomFilter/releases
2. Click "Create a new release"
3. Select tag: `v0.2.0`
4. Title: `v0.2.0 - Hybrid Array/Map Optimization`
5. Description: Copy from CHANGELOG.md [0.2.0] section
6. Click "Publish release"

**Option B: Using GitHub CLI**
```bash
# Install gh if needed
# https://cli.github.com/

# Create release from tag
gh release create v0.2.0 \
  --title "v0.2.0 - Hybrid Array/Map Optimization" \
  --notes-file <(sed -n '/## \[0.2.0\]/,/## \[0.1.0\]/p' CHANGELOG.md | head -n -1)
```

### Step 5: Verify Package is Published

```bash
# Wait a few minutes for Go proxy to index

# Test package can be fetched
GOPROXY=proxy.golang.org go list -m github.com/shaia/BloomFilter@v0.2.0

# Test installation in a temp directory
cd /tmp
mkdir test-install && cd test-install
go mod init test
go get github.com/shaia/BloomFilter@v0.2.0

# Check pkg.go.dev
# Visit: https://pkg.go.dev/github.com/shaia/BloomFilter
# (May take 10-15 minutes to appear)
```

### Step 6: Update Benchmark Project

```bash
# Update bloomfilter-benchmarks to use published version
cd ../bloomfilter-benchmarks

# Edit go.mod - remove replace directive
# Change to use published version
go mod edit -require=github.com/shaia/BloomFilter@v0.2.0

# Remove replace line manually or:
# (just delete the replace line from go.mod)

# Update dependencies
go mod tidy

# Test benchmarks still work
go test -bench=BenchmarkComparisonAdd/Size_10K -run=^$ -benchtime=500ms

# Commit the change
git add go.mod go.sum
git commit -m "Use published version v0.2.0 of go-simd-bloomfilter"
git push
```

## Post-Publishing Tasks

### Update Repository Settings

1. **Add topics** on GitHub:
   - `bloom-filter`
   - `simd`
   - `go`
   - `golang`
   - `performance`
   - `cache-optimization`
   - `avx2`
   - `neon`

2. **Add description**:
   "High-performance SIMD-optimized Bloom filter for Go with hybrid architecture. Zero-allocation array mode for small filters, unlimited scalability for large filters."

3. **Update README badges** (optional):
   ```markdown
   ![Go Version](https://img.shields.io/github/go-mod/go-version/shaia/BloomFilter)
   ![Release](https://img.shields.io/github/v/release/shaia/BloomFilter)
   ![License](https://img.shields.io/github/license/shaia/BloomFilter)
   ```

### Announce Release

Consider announcing on:
- Reddit: r/golang
- Twitter/X
- Hacker News
- Your blog/website

Sample announcement:
```
go-simd-bloomfilter v0.2.0 released!

Major update with hybrid architecture:
- Zero allocations for small filters (10K-100K elements)
- 1.5x faster than popular alternatives
- 95% memory reduction
- Unlimited scalability with map mode

Perfect for microservices, rate limiting, and session management.

go get github.com/shaia/BloomFilter@v0.2.0
```

## Troubleshooting

### Tag Already Exists

```bash
# Delete local tag
git tag -d v0.2.0

# Delete remote tag
git push origin :refs/tags/v0.2.0

# Recreate tag
git tag -a v0.2.0 -m "..."
git push origin v0.2.0
```

### Package Not Appearing on pkg.go.dev

```bash
# Request indexing (after 10 minutes)
curl https://proxy.golang.org/github.com/shaia/BloomFilter/@v/v0.2.0.info

# Or visit directly and it will trigger indexing:
# https://pkg.go.dev/github.com/shaia/BloomFilter@v0.2.0
```

### Tests Fail After Merge

```bash
# Identify which tests fail
go test ./... -v

# Fix issues
# Commit fixes to main
git add .
git commit -m "Fix post-merge test failures"
git push origin main

# May need to delete and recreate tag if critical
```

## Version Numbering Guide

Follow Semantic Versioning (SemVer):

- **v0.2.0** (Current) - Major feature: Hybrid architecture
- **v0.2.1** - Bug fixes, documentation updates
- **v0.3.0** - Next major feature (e.g., paged array mode)
- **v1.0.0** - Stable API, production-ready, no breaking changes expected

## Next Release Checklist

For future releases:

1. Create feature branch
2. Make changes
3. Update CHANGELOG.md [Unreleased] section
4. Merge to main
5. Update CHANGELOG version
6. Create tag
7. Push tag
8. Create GitHub release
9. Verify on pkg.go.dev
10. Update benchmarks project

## Notes

- Go modules are automatically published when you push a tag
- No registration or manual submission required
- pkg.go.dev indexes automatically (may take 10-15 min)
- Users can install immediately after tag push:
  ```bash
  go get github.com/shaia/BloomFilter@v0.2.0
  ```

## Important: Fix Before Publishing

**MUST FIX**: Flaky SIMD performance test
- Location: `simd_comparison_test.go`
- Issue: TestSIMDPerformanceImprovement sometimes fails under load
- Options:
  1. Increase tolerance for speedup assertion
  2. Increase benchmark time for more stable results
  3. Mark as flaky with build tag
  4. Skip in CI, run manually

Do NOT publish with failing tests!
