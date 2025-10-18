package bloomfilter

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/shaia/go-simd-bloomfilter/internal/simd"
)

// BenchmarkSIMDvsScalar compares SIMD implementations against scalar fallback
func BenchmarkSIMDvsScalar(b *testing.B) {
	sizes := []int{64, 256, 1024, 4096, 16384, 65536}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("PopCount_Size_%d", size), func(b *testing.B) {
			data := make([]byte, size)
			// Fill with some pattern
			for i := range data {
				data[i] = byte(i % 256)
			}
			ptr := unsafe.Pointer(&data[0])

			b.Run("SIMD", func(b *testing.B) {
				simdOps := simd.Get()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = simdOps.PopCount(ptr, size)
				}
			})

			b.Run("Fallback", func(b *testing.B) {
				fallbackOps := &simd.FallbackOperations{}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = fallbackOps.PopCount(ptr, size)
				}
			})
		})

		b.Run(fmt.Sprintf("VectorOr_Size_%d", size), func(b *testing.B) {
			dst := make([]byte, size)
			src := make([]byte, size)
			for i := range src {
				src[i] = byte(i % 256)
			}
			dstPtr := unsafe.Pointer(&dst[0])
			srcPtr := unsafe.Pointer(&src[0])

			b.Run("SIMD", func(b *testing.B) {
				simdOps := simd.Get()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					simdOps.VectorOr(dstPtr, srcPtr, size)
				}
			})

			b.Run("Fallback", func(b *testing.B) {
				fallbackOps := &simd.FallbackOperations{}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					fallbackOps.VectorOr(dstPtr, srcPtr, size)
				}
			})
		})

		b.Run(fmt.Sprintf("VectorAnd_Size_%d", size), func(b *testing.B) {
			dst := make([]byte, size)
			src := make([]byte, size)
			for i := range src {
				src[i] = byte(i % 256)
			}
			dstPtr := unsafe.Pointer(&dst[0])
			srcPtr := unsafe.Pointer(&src[0])

			b.Run("SIMD", func(b *testing.B) {
				simdOps := simd.Get()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					simdOps.VectorAnd(dstPtr, srcPtr, size)
				}
			})

			b.Run("Fallback", func(b *testing.B) {
				fallbackOps := &simd.FallbackOperations{}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					fallbackOps.VectorAnd(dstPtr, srcPtr, size)
				}
			})
		})

		b.Run(fmt.Sprintf("VectorClear_Size_%d", size), func(b *testing.B) {
			data := make([]byte, size)
			ptr := unsafe.Pointer(&data[0])

			b.Run("SIMD", func(b *testing.B) {
				simdOps := simd.Get()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					simdOps.VectorClear(ptr, size)
				}
			})

			b.Run("Fallback", func(b *testing.B) {
				fallbackOps := &simd.FallbackOperations{}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					fallbackOps.VectorClear(ptr, size)
				}
			})
		})
	}
}

// TestSIMDPerformanceImprovement validates that SIMD is faster than fallback
func TestSIMDPerformanceImprovement(t *testing.T) {
	// Skip if no SIMD available
	if !HasSIMD() {
		t.Skip("No SIMD support available on this platform")
	}

	sizes := []int{1024, 4096, 16384, 65536}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}
			ptr := unsafe.Pointer(&data[0])

			// Test PopCount performance
			t.Run("PopCount", func(t *testing.T) {
				simdOps := simd.Get()
				fallbackOps := &simd.FallbackOperations{}

				// Warmup
				for i := 0; i < 100; i++ {
					_ = simdOps.PopCount(ptr, size)
					_ = fallbackOps.PopCount(ptr, size)
				}

				// Measure SIMD
				simdStart := testing.Benchmark(func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_ = simdOps.PopCount(ptr, size)
					}
				})

				// Measure Fallback
				fallbackStart := testing.Benchmark(func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_ = fallbackOps.PopCount(ptr, size)
					}
				})

				simdNsPerOp := simdStart.NsPerOp()
				fallbackNsPerOp := fallbackStart.NsPerOp()
				speedup := float64(fallbackNsPerOp) / float64(simdNsPerOp)

				t.Logf("SIMD: %d ns/op", simdNsPerOp)
				t.Logf("Fallback: %d ns/op", fallbackNsPerOp)
				t.Logf("Speedup: %.2fx", speedup)

				if speedup < 1.0 {
					t.Errorf("SIMD should be faster than fallback, got speedup of %.2fx", speedup)
				}

				// For larger sizes, we expect significant speedup
				if size >= 4096 && speedup < 1.2 {
					t.Logf("Warning: Expected at least 1.2x speedup for size %d, got %.2fx", size, speedup)
				}
			})

			// Test VectorOr performance
			t.Run("VectorOr", func(t *testing.T) {
				dst := make([]byte, size)
				src := make([]byte, size)
				for i := range src {
					src[i] = byte(i % 256)
				}
				dstPtr := unsafe.Pointer(&dst[0])
				srcPtr := unsafe.Pointer(&src[0])

				simdOps := simd.Get()
				fallbackOps := &simd.FallbackOperations{}

				// Warmup
				for i := 0; i < 100; i++ {
					simdOps.VectorOr(dstPtr, srcPtr, size)
					fallbackOps.VectorOr(dstPtr, srcPtr, size)
				}

				// Measure SIMD
				simdStart := testing.Benchmark(func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						simdOps.VectorOr(dstPtr, srcPtr, size)
					}
				})

				// Measure Fallback
				fallbackStart := testing.Benchmark(func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						fallbackOps.VectorOr(dstPtr, srcPtr, size)
					}
				})

				simdNsPerOp := simdStart.NsPerOp()
				fallbackNsPerOp := fallbackStart.NsPerOp()
				speedup := float64(fallbackNsPerOp) / float64(simdNsPerOp)

				t.Logf("SIMD: %d ns/op", simdNsPerOp)
				t.Logf("Fallback: %d ns/op", fallbackNsPerOp)
				t.Logf("Speedup: %.2fx", speedup)

				if speedup < 1.0 {
					t.Errorf("SIMD should be faster than fallback, got speedup of %.2fx", speedup)
				}
			})

			// Test VectorAnd performance
			t.Run("VectorAnd", func(t *testing.T) {
				dst := make([]byte, size)
				src := make([]byte, size)
				for i := range src {
					src[i] = byte(i % 256)
				}
				dstPtr := unsafe.Pointer(&dst[0])
				srcPtr := unsafe.Pointer(&src[0])

				simdOps := simd.Get()
				fallbackOps := &simd.FallbackOperations{}

				// Warmup
				for i := 0; i < 100; i++ {
					simdOps.VectorAnd(dstPtr, srcPtr, size)
					fallbackOps.VectorAnd(dstPtr, srcPtr, size)
				}

				// Measure SIMD
				simdStart := testing.Benchmark(func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						simdOps.VectorAnd(dstPtr, srcPtr, size)
					}
				})

				// Measure Fallback
				fallbackStart := testing.Benchmark(func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						fallbackOps.VectorAnd(dstPtr, srcPtr, size)
					}
				})

				simdNsPerOp := simdStart.NsPerOp()
				fallbackNsPerOp := fallbackStart.NsPerOp()
				speedup := float64(fallbackNsPerOp) / float64(simdNsPerOp)

				t.Logf("SIMD: %d ns/op", simdNsPerOp)
				t.Logf("Fallback: %d ns/op", fallbackNsPerOp)
				t.Logf("Speedup: %.2fx", speedup)

				if speedup < 1.0 {
					t.Errorf("SIMD should be faster than fallback, got speedup of %.2fx", speedup)
				}
			})
		})
	}
}

// TestSIMDCorrectness validates that SIMD produces same results as fallback
func TestSIMDCorrectness(t *testing.T) {
	sizes := []int{1, 7, 8, 15, 16, 31, 32, 63, 64, 127, 128, 255, 256, 1023, 1024, 4096}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			// Test PopCount
			t.Run("PopCount", func(t *testing.T) {
				data := make([]byte, size)
				for i := range data {
					data[i] = byte((i * 17) % 256) // Some pattern
				}
				ptr := unsafe.Pointer(&data[0])

				simdOps := simd.Get()
				fallbackOps := &simd.FallbackOperations{}

				simdResult := simdOps.PopCount(ptr, size)
				fallbackResult := fallbackOps.PopCount(ptr, size)

				if simdResult != fallbackResult {
					t.Errorf("PopCount mismatch: SIMD=%d, Fallback=%d", simdResult, fallbackResult)
				}
			})

			// Test VectorOr
			t.Run("VectorOr", func(t *testing.T) {
				dst1 := make([]byte, size)
				dst2 := make([]byte, size)
				src := make([]byte, size)

				for i := range src {
					src[i] = byte((i * 13) % 256)
					dst1[i] = byte((i * 7) % 256)
					dst2[i] = dst1[i]
				}

				simdOps := simd.Get()
				fallbackOps := &simd.FallbackOperations{}

				simdOps.VectorOr(unsafe.Pointer(&dst1[0]), unsafe.Pointer(&src[0]), size)
				fallbackOps.VectorOr(unsafe.Pointer(&dst2[0]), unsafe.Pointer(&src[0]), size)

				for i := 0; i < size; i++ {
					if dst1[i] != dst2[i] {
						t.Errorf("VectorOr mismatch at index %d: SIMD=%d, Fallback=%d", i, dst1[i], dst2[i])
						break
					}
				}
			})

			// Test VectorAnd
			t.Run("VectorAnd", func(t *testing.T) {
				dst1 := make([]byte, size)
				dst2 := make([]byte, size)
				src := make([]byte, size)

				for i := range src {
					src[i] = byte((i * 13) % 256)
					dst1[i] = byte((i * 7) % 256)
					dst2[i] = dst1[i]
				}

				simdOps := simd.Get()
				fallbackOps := &simd.FallbackOperations{}

				simdOps.VectorAnd(unsafe.Pointer(&dst1[0]), unsafe.Pointer(&src[0]), size)
				fallbackOps.VectorAnd(unsafe.Pointer(&dst2[0]), unsafe.Pointer(&src[0]), size)

				for i := 0; i < size; i++ {
					if dst1[i] != dst2[i] {
						t.Errorf("VectorAnd mismatch at index %d: SIMD=%d, Fallback=%d", i, dst1[i], dst2[i])
						break
					}
				}
			})

			// Test VectorClear
			t.Run("VectorClear", func(t *testing.T) {
				data1 := make([]byte, size)
				data2 := make([]byte, size)

				for i := range data1 {
					data1[i] = byte((i * 19) % 256)
					data2[i] = data1[i]
				}

				simdOps := simd.Get()
				fallbackOps := &simd.FallbackOperations{}

				simdOps.VectorClear(unsafe.Pointer(&data1[0]), size)
				fallbackOps.VectorClear(unsafe.Pointer(&data2[0]), size)

				for i := 0; i < size; i++ {
					if data1[i] != data2[i] || data1[i] != 0 {
						t.Errorf("VectorClear mismatch at index %d: SIMD=%d, Fallback=%d", i, data1[i], data2[i])
						break
					}
				}
			})
		})
	}
}

// BenchmarkBloomFilterWithSIMD benchmarks the full bloom filter with SIMD vs without
func BenchmarkBloomFilterWithSIMD(b *testing.B) {
	sizes := []int{10000, 100000, 1000000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			// Test with current SIMD settings
			b.Run("WithSIMD", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					bf := NewCacheOptimizedBloomFilter(uint64(size), 0.01)
					// Insert some elements
					for j := 0; j < 1000; j++ {
						bf.AddString(fmt.Sprintf("test-%d", j))
					}
					// Query some elements
					for j := 0; j < 1000; j++ {
						_ = bf.ContainsString(fmt.Sprintf("test-%d", j))
					}
				}
			})
		})
	}
}
