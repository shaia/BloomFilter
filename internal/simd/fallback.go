package simd

import "unsafe"

// FallbackOperations implements SIMD operations using optimized scalar code
type FallbackOperations struct{}

func (f *FallbackOperations) PopCount(data unsafe.Pointer, length int) int {
	// Use optimized scalar popcount
	ptr := (*[1 << 30]uint64)(data)[:length/8]
	count := 0
	for i := 0; i < len(ptr); i++ {
		count += popcount64(ptr[i])
	}

	// Handle remaining bytes
	remaining := length % 8
	if remaining > 0 {
		lastBytes := (*[8]byte)(unsafe.Pointer(uintptr(data) + uintptr(length-remaining)))
		var lastWord uint64
		for i := 0; i < remaining; i++ {
			lastWord |= uint64(lastBytes[i]) << (i * 8)
		}
		count += popcount64(lastWord)
	}

	return count
}

func (f *FallbackOperations) VectorOr(dst, src unsafe.Pointer, length int) {
	// Process 8 bytes at a time
	dstPtr := (*[1 << 30]uint64)(dst)[:length/8]
	srcPtr := (*[1 << 30]uint64)(src)[:length/8]

	for i := 0; i < len(dstPtr); i++ {
		dstPtr[i] |= srcPtr[i]
	}

	// Handle remaining bytes
	remaining := length % 8
	if remaining > 0 {
		dstBytes := (*[8]byte)(unsafe.Pointer(uintptr(dst) + uintptr(length-remaining)))
		srcBytes := (*[8]byte)(unsafe.Pointer(uintptr(src) + uintptr(length-remaining)))
		for i := 0; i < remaining; i++ {
			dstBytes[i] |= srcBytes[i]
		}
	}
}

func (f *FallbackOperations) VectorAnd(dst, src unsafe.Pointer, length int) {
	// Process 8 bytes at a time
	dstPtr := (*[1 << 30]uint64)(dst)[:length/8]
	srcPtr := (*[1 << 30]uint64)(src)[:length/8]

	for i := 0; i < len(dstPtr); i++ {
		dstPtr[i] &= srcPtr[i]
	}

	// Handle remaining bytes
	remaining := length % 8
	if remaining > 0 {
		dstBytes := (*[8]byte)(unsafe.Pointer(uintptr(dst) + uintptr(length-remaining)))
		srcBytes := (*[8]byte)(unsafe.Pointer(uintptr(src) + uintptr(length-remaining)))
		for i := 0; i < remaining; i++ {
			dstBytes[i] &= srcBytes[i]
		}
	}
}

func (f *FallbackOperations) VectorClear(data unsafe.Pointer, length int) {
	// Process 8 bytes at a time
	ptr := (*[1 << 30]uint64)(data)[:length/8]

	for i := 0; i < len(ptr); i++ {
		ptr[i] = 0
	}

	// Handle remaining bytes
	remaining := length % 8
	if remaining > 0 {
		bytes := (*[8]byte)(unsafe.Pointer(uintptr(data) + uintptr(length-remaining)))
		for i := 0; i < remaining; i++ {
			bytes[i] = 0
		}
	}
}

// popcount64 implements efficient popcount for uint64
func popcount64(x uint64) int {
	// Use the same algorithm as bits.OnesCount64 but inline for performance
	x = x - ((x >> 1) & 0x5555555555555555)
	x = (x & 0x3333333333333333) + ((x >> 2) & 0x3333333333333333)
	x = (x + (x >> 4)) & 0x0f0f0f0f0f0f0f0f
	return int((x * 0x0101010101010101) >> 56)
}
