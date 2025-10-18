//go:build arm64 && !purego

#include "textflag.h"

// neonPopCount performs SIMD population count using ARM NEON
// func neonPopCount(data unsafe.Pointer, length int) int
TEXT ·neonPopCount(SB), NOSPLIT, $0-24
    MOVD data+0(FP), R0      // Load data pointer
    MOVD length+8(FP), R1    // Load length in bytes
    MOVD $0, R2              // Initialize count accumulator
    MOVD $0, R3              // Initialize loop counter

    // Check if length is less than 8 bytes 
    CMP $8, R1
    BLT scalar_loop

uint64_loop:
    CMP R3, R1
    BEQ done
    
    // Check if we have at least 8 bytes remaining
    SUB R3, R1, R4
    CMP $8, R4
    BLT scalar_loop
    
    // Load 8 bytes (uint64) and count bits
    MOVD (R0), R4
    
    // Efficient popcount for uint64 using bit manipulation
    // x = x - ((x >> 1) & 0x5555555555555555)
    MOVD $0x5555555555555555, R5
    LSR $1, R4, R6
    AND R5, R6
    SUB R6, R4
    
    // x = (x & 0x3333333333333333) + ((x >> 2) & 0x3333333333333333)
    MOVD $0x3333333333333333, R5
    LSR $2, R4, R6
    AND R5, R6
    AND R5, R4
    ADD R6, R4
    
    // x = (x + (x >> 4)) & 0x0f0f0f0f0f0f0f0f
    LSR $4, R4, R6
    ADD R6, R4
    MOVD $0x0f0f0f0f0f0f0f0f, R5
    AND R5, R4
    
    // return (x * 0x0101010101010101) >> 56
    MOVD $0x0101010101010101, R5
    MUL R5, R4
    LSR $56, R4
    
    ADD R4, R2               // Add to accumulator
    ADD $8, R0               // Advance pointer by 8 bytes
    ADD $8, R3               // Advance counter
    B uint64_loop

scalar_loop:
    CMP R3, R1
    BEQ done
    
    MOVBU (R0), R4           // Load one byte
    
    // Count bits in byte using lookup table approach
    AND $0x0F, R4, R5        // Lower nibble
    LSR $4, R4, R6           // Upper nibble to R6
    AND $0x0F, R6            // Mask upper nibble
    
    // Lookup table for 4-bit popcount (0-4)
    MOVD $0x4332322132212110, R7  // Packed lookup table
    
    // Lookup lower nibble
    LSL $2, R5, R8           // R8 = R5 << 2
    LSR R8, R7, R9           // R9 = R7 >> R8
    AND $0xF, R9             // Mask result
    
    // Lookup upper nibble  
    LSL $2, R6, R8           // R8 = R6 << 2
    LSR R8, R7, R10          // R10 = R7 >> R8
    AND $0xF, R10            // Mask result
    
    ADD R10, R9              // Add upper and lower nibble counts
    
    ADD R9, R2               // Add to accumulator
    ADD $1, R0               // Advance pointer
    ADD $1, R3               // Advance counter
    B scalar_loop

done:
    MOVD R2, ret+16(FP)      // Store result
    RET

// neonVectorOr performs SIMD OR operation using ARM NEON  
// func neonVectorOr(dst, src unsafe.Pointer, length int)
TEXT ·neonVectorOr(SB), NOSPLIT, $0-24
    MOVD dst+0(FP), R0       // Load dst pointer
    MOVD src+8(FP), R1       // Load src pointer  
    MOVD length+16(FP), R2   // Load length in bytes
    MOVD $0, R3              // Initialize loop counter

uint64_or_loop:
    CMP R3, R2
    BEQ or_done
    
    SUB R3, R2, R4           // Calculate remaining bytes
    CMP $8, R4               // Check if we have at least 8 bytes
    BLT or_scalar
    
    // Load 8 bytes from both src and dst
    MOVD (R0), R5            // Load dst
    MOVD (R1), R6            // Load src
    
    // Perform OR operation
    ORR R6, R5, R5           // dst = dst | src
    
    // Store result back to dst
    MOVD R5, (R0)
    
    ADD $8, R0               // Advance dst pointer
    ADD $8, R1               // Advance src pointer
    ADD $8, R3               // Advance counter
    B uint64_or_loop

or_scalar:
    CMP R3, R2
    BEQ or_done
    
    MOVBU (R0), R4           // Load dst byte
    MOVBU (R1), R5           // Load src byte
    ORR R5, R4, R4           // dst = dst | src
    MOVB R4, (R0)            // Store result
    
    ADD $1, R0               // Advance dst pointer
    ADD $1, R1               // Advance src pointer  
    ADD $1, R3               // Advance counter
    B or_scalar

or_done:
    RET

// neonVectorAnd performs SIMD AND operation using ARM NEON
// func neonVectorAnd(dst, src unsafe.Pointer, length int)  
TEXT ·neonVectorAnd(SB), NOSPLIT, $0-24
    MOVD dst+0(FP), R0       // Load dst pointer
    MOVD src+8(FP), R1       // Load src pointer
    MOVD length+16(FP), R2   // Load length in bytes
    MOVD $0, R3              // Initialize loop counter

uint64_and_loop:
    CMP R3, R2
    BEQ and_done
    
    SUB R3, R2, R4           // Calculate remaining bytes
    CMP $8, R4               // Check if we have at least 8 bytes
    BLT and_scalar
    
    // Load 8 bytes from both src and dst
    MOVD (R0), R5            // Load dst
    MOVD (R1), R6            // Load src
    
    // Perform AND operation
    AND R6, R5, R5           // dst = dst & src
    
    // Store result back to dst  
    MOVD R5, (R0)
    
    ADD $8, R0               // Advance dst pointer
    ADD $8, R1               // Advance src pointer
    ADD $8, R3               // Advance counter
    B uint64_and_loop

and_scalar:
    CMP R3, R2
    BEQ and_done
    
    MOVBU (R0), R4           // Load dst byte
    MOVBU (R1), R5           // Load src byte
    AND R5, R4, R4           // dst = dst & src
    MOVB R4, (R0)            // Store result
    
    ADD $1, R0               // Advance dst pointer
    ADD $1, R1               // Advance src pointer
    ADD $1, R3               // Advance counter
    B and_scalar

and_done:
    RET

// neonVectorClear performs SIMD clear operation using ARM NEON
// func neonVectorClear(data unsafe.Pointer, length int)
TEXT ·neonVectorClear(SB), NOSPLIT, $0-16
    MOVD data+0(FP), R0      // Load data pointer
    MOVD length+8(FP), R1    // Load length in bytes
    MOVD $0, R2              // Initialize loop counter

uint64_clear_loop:
    CMP R2, R1
    BEQ clear_done
    
    SUB R2, R1, R3           // Calculate remaining bytes
    CMP $8, R3               // Check if we have at least 8 bytes
    BLT clear_scalar
    
    // Store 8 zeros
    MOVD $0, (R0)
    
    ADD $8, R0               // Advance pointer
    ADD $8, R2               // Advance counter
    B uint64_clear_loop

clear_scalar:
    CMP R2, R1
    BEQ clear_done
    
    MOVB $0, (R0)            // Store zero byte
    ADD $1, R0               // Advance pointer
    ADD $1, R2               // Advance counter
    B clear_scalar

clear_done:
    RET
