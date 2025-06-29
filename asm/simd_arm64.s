//go:build arm64 && !purego

#include "textflag.h"

// popCountNEON performs vectorized population count using ARM NEON
// func (bf *CacheOptimizedBloomFilter) popCountNEON() uint64
TEXT ·(*CacheOptimizedBloomFilter).popCountNEON(SB), NOSPLIT, $0-16
    MOVD bf+0(FP), R0
    MOVD 8(R0), R1          // bf.cacheLines slice data
    MOVD 32(R0), R2         // bf.cacheLineCount
    
    MOVD $0, R3             // total count accumulator
    MOVD $0, R4             // loop counter
    
    // Check if we have any cache lines to process
    CMP $0, R2
    BEQ done
    
loop:
    // Load 16 bytes at a time using NEON
    VLD1 (R1), [V0.2D]      // Load first 16 bytes
    VLD1 16(R1), [V1.2D]    // Load second 16 bytes
    VLD1 32(R1), [V2.2D]    // Load third 16 bytes
    VLD1 48(R1), [V3.2D]    // Load fourth 16 bytes
    
    // Count bits in each register using CNT instruction
    VCNT V0.16B, V4.16B
    VCNT V1.16B, V5.16B  
    VCNT V2.16B, V6.16B
    VCNT V3.16B, V7.16B
    
    // Sum up the counts
    VADDV V4.16B, S8
    VADDV V5.16B, S9
    VADDV V6.16B, S10
    VADDV V7.16B, S11
    
    FMOVS S8, R5
    ADD R5, R3, R3
    FMOVS S9, R5
    ADD R5, R3, R3
    FMOVS S10, R5
    ADD R5, R3, R3
    FMOVS S11, R5
    ADD R5, R3, R3
    
    ADD $64, R1             // Move to next cache line
    ADD $1, R4
    CMP R2, R4
    BLT loop
    
done:
    MOVD R3, ret+8(FP)
    RET

// unionNEON performs vectorized union using ARM NEON
// func (bf *CacheOptimizedBloomFilter) unionNEON(other *CacheOptimizedBloomFilter) error
TEXT ·(*CacheOptimizedBloomFilter).unionNEON(SB), NOSPLIT, $0-24
    MOVD bf+0(FP), R0
    MOVD other+8(FP), R1
    
    MOVD 8(R0), R2          // bf.cacheLines data
    MOVD 8(R1), R3          // other.cacheLines data
    MOVD 32(R0), R4         // bf.cacheLineCount
    
    MOVD $0, R5             // loop counter
    
    CMP $0, R4
    BEQ union_done
    
union_loop:
    // Load 16 bytes from both bloom filters
    VLD1 (R2), [V0.2D]      // Load from bf
    VLD1 (R3), [V1.2D]      // Load from other
    VORR V0.16B, V1.16B, V0.16B // Bitwise OR
    VST1 [V0.2D], (R2)      // Store back to bf
    
    VLD1 16(R2), [V0.2D]    // Load next 16 bytes from bf
    VLD1 16(R3), [V1.2D]    // Load next 16 bytes from other
    VORR V0.16B, V1.16B, V0.16B // Bitwise OR
    VST1 [V0.2D], 16(R2)    // Store back to bf
    
    VLD1 32(R2), [V0.2D]    // Load next 16 bytes from bf
    VLD1 32(R3), [V1.2D]    // Load next 16 bytes from other
    VORR V0.16B, V1.16B, V0.16B // Bitwise OR
    VST1 [V0.2D], 32(R2)    // Store back to bf
    
    VLD1 48(R2), [V0.2D]    // Load next 16 bytes from bf
    VLD1 48(R3), [V1.2D]    // Load next 16 bytes from other
    VORR V0.16B, V1.16B, V0.16B // Bitwise OR
    VST1 [V0.2D], 48(R2)    // Store back to bf
    
    ADD $64, R2             // Move to next cache line
    ADD $64, R3
    ADD $1, R5
    CMP R4, R5
    BLT union_loop
    
union_done:
    MOVD $0, ret+16(FP)     // Return nil error
    RET

// intersectionNEON performs vectorized intersection using ARM NEON
// func (bf *CacheOptimizedBloomFilter) intersectionNEON(other *CacheOptimizedBloomFilter) error
TEXT ·(*CacheOptimizedBloomFilter).intersectionNEON(SB), NOSPLIT, $0-24
    MOVD bf+0(FP), R0
    MOVD other+8(FP), R1
    
    MOVD 8(R0), R2          // bf.cacheLines data
    MOVD 8(R1), R3          // other.cacheLines data
    MOVD 32(R0), R4         // bf.cacheLineCount
    
    MOVD $0, R5             // loop counter
    
    CMP $0, R4
    BEQ intersect_done
    
intersect_loop:
    // Load 16 bytes from both bloom filters
    VLD1 (R2), [V0.2D]      // Load from bf
    VLD1 (R3), [V1.2D]      // Load from other
    VAND V0.16B, V1.16B, V0.16B // Bitwise AND
    VST1 [V0.2D], (R2)      // Store back to bf
    
    VLD1 16(R2), [V0.2D]    // Load next 16 bytes from bf
    VLD1 16(R3), [V1.2D]    // Load next 16 bytes from other
    VAND V0.16B, V1.16B, V0.16B // Bitwise AND
    VST1 [V0.2D], 16(R2)    // Store back to bf
    
    VLD1 32(R2), [V0.2D]    // Load next 16 bytes from bf
    VLD1 32(R3), [V1.2D]    // Load next 16 bytes from other
    VAND V0.16B, V1.16B, V0.16B // Bitwise AND
    VST1 [V0.2D], 32(R2)    // Store back to bf
    
    VLD1 48(R2), [V0.2D]    // Load next 16 bytes from bf
    VLD1 48(R3), [V1.2D]    // Load next 16 bytes from other
    VAND V0.16B, V1.16B, V0.16B // Bitwise AND
    VST1 [V0.2D], 48(R2)    // Store back to bf
    
    ADD $64, R2             // Move to next cache line
    ADD $64, R3
    ADD $1, R5
    CMP R4, R5
    BLT intersect_loop
    
intersect_done:
    MOVD $0, ret+16(FP)     // Return nil error
    RET

// clearNEON performs vectorized clear using ARM NEON
// func (bf *CacheOptimizedBloomFilter) clearNEON()
TEXT ·(*CacheOptimizedBloomFilter).clearNEON(SB), NOSPLIT, $0-8
    MOVD bf+0(FP), R0
    MOVD 8(R0), R1          // bf.cacheLines data
    MOVD 32(R0), R2         // bf.cacheLineCount
    
    VEOR V0.16B, V0.16B, V0.16B // Zero vector
    MOVD $0, R3             // loop counter
    
    CMP $0, R2
    BEQ clear_done
    
clear_loop:
    VST1 [V0.2D], (R1)      // Clear first 16 bytes
    VST1 [V0.2D], 16(R1)    // Clear second 16 bytes
    VST1 [V0.2D], 32(R1)    // Clear third 16 bytes
    VST1 [V0.2D], 48(R1)    // Clear fourth 16 bytes
    
    ADD $64, R1             // Move to next cache line
    ADD $1, R3
    CMP R2, R3
    BLT clear_loop
    
clear_done:
    RET
