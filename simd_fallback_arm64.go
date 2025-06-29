//go:build (arm64 && purego) || !arm64

package bloomfilter

// Fallback implementations when NEON is not available

func (bf *CacheOptimizedBloomFilter) popCountNEON() uint64 {
	return bf.PopCountOptimized()
}

func (bf *CacheOptimizedBloomFilter) unionNEON(other *CacheOptimizedBloomFilter) error {
	return bf.UnionOptimized(other)
}

func (bf *CacheOptimizedBloomFilter) intersectionNEON(other *CacheOptimizedBloomFilter) error {
	return bf.IntersectionOptimized(other)
}

func (bf *CacheOptimizedBloomFilter) clearNEON() {
	bf.ClearOptimized()
}
