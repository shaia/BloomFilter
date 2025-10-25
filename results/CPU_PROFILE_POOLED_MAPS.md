# CPU Profile Analysis - After Map Pooling Optimization

## Summary

After implementing persistent map pooling, the CPU profile reveals that **map operations still dominate** the runtime, consuming ~80% of CPU time. However, allocations have been reduced by 99.9%, eliminating GC overhead.

---

## CPU Profile Breakdown

### Top Hotspots (12.63s total)

| Function | Flat Time | Flat % | Cumulative Time | Cum % | Analysis |
|----------|-----------|--------|-----------------|-------|----------|
| `runtime.mapiternext` | 4.98s | **39.43%** | 5.77s | 45.68% | **Map iteration overhead** - clearing maps via `for k := range` |
| `runtime.mapassign_fast64` | 2.16s | **17.10%** | 2.83s | 22.41% | **Map assignment** - appending to map values |
| `runtime.mapaccess1_fast64` | 1.51s | **11.96%** | 2.37s | 18.76% | **Map access** - reading from maps |
| `runtime.add` | 0.84s | 6.65% | 0.84s | 6.65% | Pointer arithmetic |
| `getHashPositionsOptimized` | 0.76s | 6.02% | 6.66s | **52.73%** | Our code - hash position generation |
| `runtime.memhash64` | 0.67s | 5.30% | 0.67s | 5.30% | Map key hashing |
| `getBitCacheOptimized` | 0.42s | 3.33% | 4.56s | **36.10%** | Our code - bit checking |
| `setBitCacheOptimized` | 0.13s | 1.03% | 1.13s | 8.95% | Our code - bit setting |

---

## Key Findings

### 1. Map Operations Consume 80% of CPU Time

**Map runtime overhead:**
- `mapiternext` (39.43%) - Iterating to clear maps
- `mapassign_fast64` (17.10%) - Assigning values to map entries
- `mapaccess1_fast64` (11.96%) - Reading from maps
- `memhash64` (5.30%) - Hashing map keys
- **Total: ~74% of CPU time**

### 2. Map Clearing is the New Bottleneck (39.43%)

**Current implementation:**
```go
// Clear persistent map (reset slices to length 0, keep capacity)
for k := range bf.cacheLineOps {
    bf.cacheLineOps[k] = bf.cacheLineOps[k][:0]  // 39.43% CPU time here!
}
```

**Why it's slow:**
- `runtime.mapiternext` must walk the entire hash table
- For each bucket, check if occupied
- Return key-value pairs one by one
- This is O(capacity) not O(entries)

### 3. Our Functions Are Fast (Total: 7.41% flat time)

| Function | Flat Time | Actual Work |
|----------|-----------|-------------|
| `getHashPositionsOptimized` | 0.76s (6.02%) | Hash calculation, position generation |
| `getBitCacheOptimized` | 0.42s (3.33%) | Bit operations, cache line access |
| `setBitCacheOptimized` | 0.13s (1.03%) | Bit setting operations |

**Total actual work: 7.41% of CPU time**
**Map overhead: 74% of CPU time**

**Conclusion: Only 7.41% of CPU time is actual bloom filter work!**

---

## Comparison: Before vs After Map Pooling

### Memory & Allocations (Massive Improvement ✅)

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| **Allocations (10K)** | 12,000 | **1** | **99.99% reduction** |
| **Memory (10K)** | 145 KB | **103 B** | **99.93% reduction** |
| **Allocations (100K)** | 11,996 | **61** | **99.5% reduction** |
| **Memory (100K)** | 144 KB | **11.8 KB** | **91.8% reduction** |

### CPU Time Distribution (Map Operations Still Dominant ⚠️)

**Before map pooling:**
- Slice growing: 34.5%
- Map operations: 32.8%
- Allocations: 23.5%

**After map pooling:**
- **Map clearing (mapiternext): 39.43%** ← New bottleneck
- Map assignment: 17.10%
- Map access: 11.96%
- Map hashing: 5.30%
- **Total map overhead: ~74%**

---

## Why Map Pooling Succeeded (Allocations) But Not CPU Time

### What We Fixed ✅
- **Eliminated 99.9% of allocations** - maps created once in constructor
- **Eliminated GC pressure** - no per-query allocations
- **Reduced memory by 99.93%** - reusing map capacity

### What We Didn't Fix ⚠️
- **Map clearing overhead** - must iterate entire hash table
- **Map hash operations** - still need to hash keys for access/assignment
- **Map indirection** - still pointer chasing through buckets

---

## Next Optimization: Replace Maps with Arrays

### Current Problem

Maps have inherent overhead:
```go
cacheLineOps map[uint64][]opDetail  // 74% CPU overhead
```

**Why maps are slow:**
1. **Hash computation**: Hash each uint64 key (5.30% CPU)
2. **Bucket search**: Find bucket, check for collisions (11.96% CPU)
3. **Assignment overhead**: Allocate map entries (17.10% CPU)
4. **Iteration overhead**: Walk entire hash table to clear (39.43% CPU)

### Proposed Solution: Direct Array Indexing

Replace maps with fixed-size arrays using cache line index directly:

```go
type CacheOptimizedBloomFilter struct {
    // Direct array indexing - O(1) access, no hashing
    cacheLineOps    [MAX_CACHE_LINES][]opDetail  // Instead of map[uint64][]opDetail
    cacheLineOpsSet [MAX_CACHE_LINES][]struct{...}
    cacheLineMap    [MAX_CACHE_LINES][]uint64

    // Track which indices are in use for fast clearing
    usedIndices []uint64
}
```

**Benefits:**
1. **No hashing** - cache line index is already a uint64, use it directly
2. **O(1) access** - `array[cacheLineIdx]` instead of `map[cacheLineIdx]`
3. **Fast clearing** - only clear used indices, not entire map
4. **Better cache locality** - contiguous memory instead of scattered buckets

### Implementation Pattern

**Before (map-based):**
```go
func (bf *Filter) getBitCacheOptimized(positions []uint64) bool {
    // Clear map - O(capacity) iteration
    for k := range bf.cacheLineOps {  // 39.43% CPU!
        bf.cacheLineOps[k] = bf.cacheLineOps[k][:0]
    }

    // Use map
    for _, bitPos := range positions {
        cacheLineIdx := bitPos / BitsPerCacheLine
        bf.cacheLineOps[cacheLineIdx] = append(...)  // Hash + bucket search
    }
}
```

**After (array-based):**
```go
func (bf *Filter) getBitCacheOptimized(positions []uint64) bool {
    // Clear only used indices - O(used entries) not O(capacity)
    for _, idx := range bf.usedIndices {
        bf.cacheLineOps[idx] = bf.cacheLineOps[idx][:0]
    }
    bf.usedIndices = bf.usedIndices[:0]

    // Use array - direct indexing, no hashing
    for _, bitPos := range positions {
        cacheLineIdx := bitPos / BitsPerCacheLine
        if len(bf.cacheLineOps[cacheLineIdx]) == 0 {
            bf.usedIndices = append(bf.usedIndices, cacheLineIdx)
        }
        bf.cacheLineOps[cacheLineIdx] = append(...)  // Direct array access
    }
}
```

### Expected Improvement

**Eliminate map overhead (74% of CPU):**
- Remove `mapiternext` (39.43%) → Clear only used indices
- Remove `mapassign_fast64` (17.10%) → Direct array assignment
- Remove `mapaccess1_fast64` (11.96%) → Direct array access
- Remove `memhash64` (5.30%) → No key hashing needed

**Potential speedup: 2-4x** (reducing 74% overhead to ~10% array overhead)

### Array Size Calculation

```go
// Maximum cache lines for typical bloom filter sizes:
// - 1M bits = 1,000,000 / 512 = 1,953 cache lines
// - 10M bits = 10,000,000 / 512 = 19,531 cache lines
// - 100M bits = 100,000,000 / 512 = 195,313 cache lines

const MAX_CACHE_LINES = 200_000  // Support up to ~100M bit bloom filters

type CacheOptimizedBloomFilter struct {
    cacheLineOps [MAX_CACHE_LINES][]opDetail
    usedIndices  []uint64  // Typically 8-32 entries
}
```

**Memory cost:**
- Array of slices: 200K * 24 bytes (slice header) = 4.8 MB per bloom filter
- Worth it for 2-4x speedup by eliminating all map overhead

---

## Optimization Priority Roadmap

| Priority | Optimization | Expected Gain | Status |
|----------|--------------|---------------|--------|
| 1 | Slice pre-allocation | 30% faster | ✅ **Done** |
| 2 | Persistent map pooling | 99.9% fewer allocations | ✅ **Done** |
| 3 | **Replace maps with arrays** | **2-4x faster** | ⚠️ **Recommended Next** |
| 4 | SIMD bit operations | 10-15% faster | 💡 Future |

---

## Conclusion

The persistent map pooling optimization was **highly successful for memory/allocations** but revealed that **map operations are the true CPU bottleneck** (74% of runtime).

### What We Achieved ✅
- 99.99% allocation reduction (12,000 → 1 allocation)
- 99.93% memory reduction (145KB → 103 bytes)
- Near-zero GC pressure on hot path

### What We Discovered 🔍
- Map clearing via iteration: 39.43% CPU time
- Map operations total: 74% CPU time
- Actual bloom filter work: Only 7.41% CPU time

### Recommendation 🚀
**Replace maps with direct array indexing** to eliminate the 74% map overhead and achieve 2-4x additional speedup. This would make the bloom filter spend most of its time on actual work instead of map operations.
