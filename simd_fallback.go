//go:build purego || (!amd64 && !arm64)

package bloomfilter

// Generic fallback implementations for architectures without specific SIMD support

func (bf *CacheOptimizedBloomFilter) popCountAVX2() uint64 {
	return bf.PopCountOptimized()
}

func (bf *CacheOptimizedBloomFilter) popCountNEON() uint64 {
	return bf.PopCountOptimized()
}

func (bf *CacheOptimizedBloomFilter) unionAVX2(other *CacheOptimizedBloomFilter) error {
	return bf.UnionOptimized(other)
}

func (bf *CacheOptimizedBloomFilter) unionNEON(other *CacheOptimizedBloomFilter) error {
	return bf.UnionOptimized(other)
}

func (bf *CacheOptimizedBloomFilter) intersectionAVX2(other *CacheOptimizedBloomFilter) error {
	return bf.IntersectionOptimized(other)
}

func (bf *CacheOptimizedBloomFilter) intersectionNEON(other *CacheOptimizedBloomFilter) error {
	return bf.IntersectionOptimized(other)
}

func (bf *CacheOptimizedBloomFilter) clearAVX2() {
	bf.ClearOptimized()
}

func (bf *CacheOptimizedBloomFilter) clearNEON() {
	bf.ClearOptimized()
}
