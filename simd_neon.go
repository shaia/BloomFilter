package bloomfilter

import "unsafe"

// NEONOperations implements SIMD operations using ARM NEON
type NEONOperations struct{}

func (n *NEONOperations) PopCount(data unsafe.Pointer, length int) int {
	return neonPopCount(data, length)
}

func (n *NEONOperations) VectorOr(dst, src unsafe.Pointer, length int) {
	neonVectorOr(dst, src, length)
}

func (n *NEONOperations) VectorAnd(dst, src unsafe.Pointer, length int) {
	neonVectorAnd(dst, src, length)
}

func (n *NEONOperations) VectorClear(data unsafe.Pointer, length int) {
	neonVectorClear(data, length)
}
