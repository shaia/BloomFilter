package simd

import (
	"unsafe"

	"github.com/shaia/go-simd-bloomfilter/internal/simd/amd64"
)

// AVX2Operations implements SIMD operations using Intel/AMD AVX2
type AVX2Operations struct{}

func (a *AVX2Operations) PopCount(data unsafe.Pointer, length int) int {
	return amd64.PopCount(data, length)
}

func (a *AVX2Operations) VectorOr(dst, src unsafe.Pointer, length int) {
	amd64.VectorOr(dst, src, length)
}

func (a *AVX2Operations) VectorAnd(dst, src unsafe.Pointer, length int) {
	amd64.VectorAnd(dst, src, length)
}

func (a *AVX2Operations) VectorClear(data unsafe.Pointer, length int) {
	amd64.VectorClear(data, length)
}
