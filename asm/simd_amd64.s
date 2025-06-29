//go:build amd64 && !purego

#include "textflag.h"

// popCountAVX2 performs vectorized population count using AVX2
// func (bf *CacheOptimizedBloomFilter) popCountAVX2() uint64
TEXT ·(*CacheOptimizedBloomFilter).popCountAVX2(SB), NOSPLIT, $0-16
    MOVQ bf+0(FP), AX
    MOVQ 8(AX), BX          // bf.cacheLines slice data
    MOVQ 32(AX), CX         // bf.cacheLineCount
    
    XORQ R8, R8             // total count accumulator
    XORQ R9, R9             // loop counter
    
    // Check if we have any cache lines to process
    TESTQ CX, CX
    JZ done
    
loop:
    // Load 4 uint64 values (32 bytes) using AVX2
    VMOVDQU (BX), Y0        // Load first 32 bytes
    VMOVDQU 32(BX), Y1      // Load second 32 bytes
    
    // Use POPCNT instruction on each 64-bit part
    MOVQ X0, R10
    POPCNTQ R10, R10
    ADDQ R10, R8
    
    VPEXTRQ $1, X0, R10
    POPCNTQ R10, R10
    ADDQ R10, R8
    
    VEXTRACTF128 $1, Y0, X2
    MOVQ X2, R10
    POPCNTQ R10, R10
    ADDQ R10, R8
    
    VPEXTRQ $1, X2, R10
    POPCNTQ R10, R10
    ADDQ R10, R8
    
    // Process second half
    MOVQ X1, R10
    POPCNTQ R10, R10
    ADDQ R10, R8
    
    VPEXTRQ $1, X1, R10
    POPCNTQ R10, R10
    ADDQ R10, R8
    
    VEXTRACTF128 $1, Y1, X2
    MOVQ X2, R10
    POPCNTQ R10, R10
    ADDQ R10, R8
    
    VPEXTRQ $1, X2, R10
    POPCNTQ R10, R10
    ADDQ R10, R8
    
    ADDQ $64, BX            // Move to next cache line
    INCQ R9
    CMPQ R9, CX
    JL loop
    
done:
    VZEROUPPER              // Clean up AVX state
    MOVQ R8, ret+8(FP)
    RET

// unionAVX2 performs vectorized union using AVX2
// func (bf *CacheOptimizedBloomFilter) unionAVX2(other *CacheOptimizedBloomFilter) error
TEXT ·(*CacheOptimizedBloomFilter).unionAVX2(SB), NOSPLIT, $0-24
    MOVQ bf+0(FP), AX
    MOVQ other+8(FP), BX
    
    MOVQ 8(AX), R8          // bf.cacheLines data
    MOVQ 8(BX), R9          // other.cacheLines data
    MOVQ 32(AX), CX         // bf.cacheLineCount
    
    XORQ R10, R10           // loop counter
    
    TESTQ CX, CX
    JZ union_done
    
union_loop:
    // Load 32 bytes from both bloom filters
    VMOVDQU (R8), Y0        // Load from bf
    VMOVDQU (R9), Y1        // Load from other
    VPORQ Y0, Y1, Y0        // Bitwise OR
    VMOVDQU Y0, (R8)        // Store back to bf
    
    VMOVDQU 32(R8), Y0      // Load next 32 bytes from bf
    VMOVDQU 32(R9), Y1      // Load next 32 bytes from other
    VPORQ Y0, Y1, Y0        // Bitwise OR
    VMOVDQU Y0, 32(R8)      // Store back to bf
    
    ADDQ $64, R8            // Move to next cache line
    ADDQ $64, R9
    INCQ R10
    CMPQ R10, CX
    JL union_loop
    
union_done:
    VZEROUPPER
    MOVQ $0, ret+16(FP)     // Return nil error
    RET

// intersectionAVX2 performs vectorized intersection using AVX2
// func (bf *CacheOptimizedBloomFilter) intersectionAVX2(other *CacheOptimizedBloomFilter) error
TEXT ·(*CacheOptimizedBloomFilter).intersectionAVX2(SB), NOSPLIT, $0-24
    MOVQ bf+0(FP), AX
    MOVQ other+8(FP), BX
    
    MOVQ 8(AX), R8          // bf.cacheLines data
    MOVQ 8(BX), R9          // other.cacheLines data
    MOVQ 32(AX), CX         // bf.cacheLineCount
    
    XORQ R10, R10           // loop counter
    
    TESTQ CX, CX
    JZ intersect_done
    
intersect_loop:
    // Load 32 bytes from both bloom filters
    VMOVDQU (R8), Y0        // Load from bf
    VMOVDQU (R9), Y1        // Load from other
    VPANDQ Y0, Y1, Y0       // Bitwise AND
    VMOVDQU Y0, (R8)        // Store back to bf
    
    VMOVDQU 32(R8), Y0      // Load next 32 bytes from bf
    VMOVDQU 32(R9), Y1      // Load next 32 bytes from other
    VPANDQ Y0, Y1, Y0       // Bitwise AND
    VMOVDQU Y0, 32(R8)      // Store back to bf
    
    ADDQ $64, R8            // Move to next cache line
    ADDQ $64, R9
    INCQ R10
    CMPQ R10, CX
    JL intersect_loop
    
intersect_done:
    VZEROUPPER
    MOVQ $0, ret+16(FP)     // Return nil error
    RET

// clearAVX2 performs vectorized clear using AVX2
// func (bf *CacheOptimizedBloomFilter) clearAVX2()
TEXT ·(*CacheOptimizedBloomFilter).clearAVX2(SB), NOSPLIT, $0-8
    MOVQ bf+0(FP), AX
    MOVQ 8(AX), BX          // bf.cacheLines data
    MOVQ 32(AX), CX         // bf.cacheLineCount
    
    VPXORQ Y0, Y0, Y0       // Zero vector
    XORQ R8, R8             // loop counter
    
    TESTQ CX, CX
    JZ clear_done
    
clear_loop:
    VMOVDQU Y0, (BX)        // Clear first 32 bytes
    VMOVDQU Y0, 32(BX)      // Clear second 32 bytes
    
    ADDQ $64, BX            // Move to next cache line
    INCQ R8
    CMPQ R8, CX
    JL clear_loop
    
clear_done:
    VZEROUPPER
    RET
