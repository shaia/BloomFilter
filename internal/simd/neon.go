package simd

import (
	"unsafe"

	"github.com/shaia/BloomFilter/internal/simd/arm64"
)

// NEONOperations implements SIMD operations using ARM NEON
type NEONOperations struct{}

func (n *NEONOperations) PopCount(data unsafe.Pointer, length int) int {
	return arm64.PopCount(data, length)
}

func (n *NEONOperations) VectorOr(dst, src unsafe.Pointer, length int) {
	arm64.VectorOr(dst, src, length)
}

func (n *NEONOperations) VectorAnd(dst, src unsafe.Pointer, length int) {
	arm64.VectorAnd(dst, src, length)
}

func (n *NEONOperations) VectorClear(data unsafe.Pointer, length int) {
	arm64.VectorClear(data, length)
}
