# Future Optimization: Paged/Sparse Array Implementation

## Concept

A **paged array** (also called sparse array) could provide a middle ground between the current hybrid approach, offering:
- O(1) array access speed (like pure arrays)
- Lazy allocation (like maps)
- No hash overhead (unlike maps)
- Scalable to millions of cache lines

## Design

### Data Structure

```go
const PageSize = 1024  // 1024 entries per page

type (
    // Sub-page: 1024 cache line indices
    opDetailSubPage [PageSize][]opDetail

    // Root page: 1024 pointers to sub-pages
    // Supports: 1024 × 1024 = 1M cache lines = ~536 MB bloom filters
    PagedOpArray [PageSize]*opDetailSubPage
)

type CacheOptimizedBloomFilter struct {
    // Sparse paged arrays (lazy allocated)
    pagedOps    PagedOpArray
    pagedOpsSet PagedOpSetArray
    pagedMap    PagedHashMap

    // Track used pages for O(used) clearing
    usedPages [][2]uint64  // [rootIdx, subIdx] pairs
}
```

### Access Pattern

```go
func (bf *CacheOptimizedBloomFilter) getBitCacheOptimized(positions []uint64) bool {
    // 1. Clear only used slices - O(used)
    for _, idxs := range bf.usedPagesGet {
        bf.pagedOps[idxs[0]][idxs[1]] = bf.pagedOps[idxs[0]][idxs[1]][:0]
    }
    bf.usedPagesGet = bf.usedPagesGet[:0]

    // 2. Access with O(1) array indexing
    for _, bitPos := range positions {
        cacheLineIdx := bitPos / BitsPerCacheLine

        // Calculate page indices
        rootIdx := cacheLineIdx / PageSize  // e.g., 5000 / 1024 = 4
        subIdx  := cacheLineIdx % PageSize  // e.g., 5000 % 1024 = 904

        // Lazy allocate sub-page if needed (1 alloc per 1024 cache lines)
        if bf.pagedOps[rootIdx] == nil {
            bf.pagedOps[rootIdx] = new([PageSize][]opDetail)
        }

        // Track for clearing
        if len(bf.pagedOps[rootIdx][subIdx]) == 0 {
            bf.usedPagesGet = append(bf.usedPagesGet, [2]uint64{rootIdx, subIdx})
        }

        // Direct array access - O(1)!
        bf.pagedOps[rootIdx][subIdx] = append(
            bf.pagedOps[rootIdx][subIdx],
            opDetail{wordIdx, bitOffset},
        )
    }

    // 3. Process operations
    for _, idxs := range bf.usedPagesGet {
        ops := bf.pagedOps[idxs[0]][idxs[1]]
        // ... check bits
    }
}
```

## Performance Analysis

### Memory Footprint

```
Empty filter:
  Root arrays: 3 × 1024 × 8 bytes = 24 KB (just pointers)

Small filter (100 cache lines):
  Root arrays: 24 KB
  Sub-page: 1 × (1024 × 24 bytes) = 24 KB
  Total: ~48 KB (vs 720 KB array mode, vs ~5 KB map mode)

Medium filter (10K cache lines):
  Root arrays: 24 KB
  Sub-pages: ~10 × 24 KB = 240 KB
  Total: ~264 KB (vs 720 KB array mode, vs ~50 KB map mode)

Large filter (1M cache lines):
  Root arrays: 24 KB
  Sub-pages: 1024 × 24 KB = 24 MB
  Total: ~24 MB (vs IMPOSSIBLE array mode, vs ~500 KB map mode)
```

### Access Performance

```
Array mode:     1 array access         (fastest)
Paged array:    2 array accesses       (very fast)
Map mode:       1 hash + 1 map access  (slower)

Estimated timing:
  Array:  1 ns
  Paged:  2 ns  (2× slower than array, but still O(1))
  Map:    15-20 ns (hash + map lookup)
```

### Allocation Pattern

```
Array mode:
  - 1 allocation at init (720 KB)
  - 0 allocations in hot path

Paged mode:
  - ~24 KB at init (root arrays)
  - 1 allocation per 1024 cache lines (lazy)
  - 0 allocations in hot path after warmup

Map mode:
  - Minimal at init
  - Many allocations in hot path (144 B/op)
```

## Comparison Matrix

| Metric | Pure Array | Paged Array | Map Mode |
|--------|-----------|-------------|----------|
| **Init Memory** | 720 KB | 24 KB | ~1 KB |
| **10K elements** | 720 KB | ~264 KB | ~50 KB |
| **1M elements** | IMPOSSIBLE | ~24 MB | ~500 KB |
| **Access Speed** | **1×** | 2× | 15-20× |
| **Hot Path Allocs** | **0** | **0** (after warmup) | 144 B/op |
| **Max Size** | 5 MB | **536 MB** | Unlimited |
| **Complexity** | Simple | Medium | Simple |

## When to Use Each Mode

### Paged Array Sweet Spot

Best for:
- **Medium to large filters** (10K - 1M cache lines / 5-500 MB)
- **Performance-critical applications** needing O(1) access
- **Moderate memory constraints** (can't waste 720 KB)
- **Predictable access patterns** (repeated operations warm up pages)

Not ideal for:
- Very small filters (<1K cache lines) - pure array is fine
- Huge filters (>1M cache lines) - maps scale better
- Extremely memory-constrained - maps use less memory

### Recommendation for Future

Could implement a **three-mode system**:

```go
if cacheLineCount <= 10000 {
    // Use pure arrays (720 KB overhead, fastest)
    useArrayMode = true
} else if cacheLineCount <= 1000000 {
    // Use paged arrays (dynamic, O(1) access, good memory)
    usePagedMode = true
} else {
    // Use maps (unlimited scaling, slower but works)
    useMapMode = true
}
```

## Implementation Considerations

### Pros
✅ O(1) access (2 array lookups vs 1 map hash)
✅ Lazy allocation (only allocates what's needed)
✅ Zero hot-path allocations after warmup
✅ Scales to 1M cache lines (536 MB filters)
✅ Better memory than pure arrays for medium filters
✅ Much faster than maps

### Cons
⚠️ More complex than pure array/map approaches
⚠️ Slightly slower than pure arrays (2× vs 1×, but still very fast)
⚠️ More memory than maps for medium filters
⚠️ Fixed upper limit (1M cache lines)
⚠️ Requires tuning page size for optimal performance

### Potential Issues

1. **Cache Misses**: Two array accesses = two potential cache misses
   - Mitigation: Root array fits in L1 cache (24 KB)

2. **Page Size Tuning**: Wrong page size hurts performance
   - Too small: More root pages, more indirection
   - Too large: Wasted allocations
   - Sweet spot: 1024 seems optimal

3. **Complexity**: More code to maintain
   - Three modes vs two modes
   - More test cases needed

## Benchmark Estimates

```go
// Small filter (10K elements)
Array:  60 ns/op,  0 B/op, 0 allocs/op  ← Current best
Paged:  120 ns/op, 0 B/op, 0 allocs/op  ← 2× slower
Map:    450 ns/op, 144 B/op, 11 allocs/op

// Large filter (1M elements)
Array:  IMPOSSIBLE
Paged:  120 ns/op, 0 B/op, 0 allocs/op  ← Much faster than map!
Map:    450 ns/op, 144 B/op, 11 allocs/op  ← Current best (only option)
```

## Implementation Priority

**Priority: Medium-Low**

### Why not now?
1. Hybrid implementation is already excellent
2. Covers most use cases well
3. Additional complexity not justified yet
4. Need real-world performance data first

### When to implement?
- If profiling shows map overhead is bottleneck
- If users request filters in the 10K-1M cache line range
- If memory usage of pure arrays is problematic
- After v0.1.0 is stable and we have usage patterns

## Testing Strategy

If implemented, test:

1. **Correctness**: All sizes (1, 1023, 1024, 1025, 10K, 100K, 1M cache lines)
2. **Memory**: Verify lazy allocation works
3. **Performance**: Benchmark vs array and map modes
4. **Stress**: Large scale with random access patterns
5. **Edge cases**: Exactly at page boundaries

## Code Skeleton

```go
// bloomfilter.go

type CacheOptimizedBloomFilter struct {
    // Mode selection (one of three)
    useArrayMode bool
    usePagedMode bool
    useMapMode   bool

    // Array mode (small filters)
    arrayOps *[10000][]opDetail

    // Paged mode (medium filters)
    pagedOps PagedOpArray

    // Map mode (large filters)
    mapOps map[uint64][]opDetail

    // Tracking
    usedIndices  []uint64    // For array mode
    usedPages    [][2]uint64 // For paged mode
}

func NewCacheOptimizedBloomFilter(...) *CacheOptimizedBloomFilter {
    if cacheLineCount <= 10000 {
        // Initialize array mode
    } else if cacheLineCount <= MaxPagedCacheLines {
        // Initialize paged mode
    } else {
        // Initialize map mode
    }
}
```

## Conclusion

The paged array approach is a **promising optimization** that could provide:
- Near-array performance (2× vs 15× slower than array)
- Better memory than pure arrays (264 KB vs 720 KB for 10K)
- Scalability beyond current array limits (1M vs 10K cache lines)

However, it adds complexity and is **not necessary for v0.1.0**. The current hybrid approach already provides excellent performance and unlimited scalability.

**Recommendation**: Document for future consideration, implement only if profiling shows it's needed.

## References

- Sparse Array: https://en.wikipedia.org/wiki/Sparse_array
- Two-level paging: Similar to CPU virtual memory page tables
- Trade-off: 1 extra indirection for lazy allocation benefits
