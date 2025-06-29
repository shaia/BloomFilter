//go:build (amd64 && purego) || !amd64

package bloomfilter

// Fallback implementations when AVX2 is not available

func (bf *CacheOptimizedBloomFilter) popCountAVX2() uint64 {
	return bf.PopCountOptimized()
}

func (bf *CacheOptimizedBloomFilter) unionAVX2(other *CacheOptimizedBloomFilter) error {
	return bf.UnionOptimized(other)
}

func (bf *CacheOptimizedBloomFilter) intersectionAVX2(other *CacheOptimizedBloomFilter) error {
	return bf.IntersectionOptimized(other)
}

func (bf *CacheOptimizedBloomFilter) clearAVX2() {
	bf.ClearOptimized()
}
