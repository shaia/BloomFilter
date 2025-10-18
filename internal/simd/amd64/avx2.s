//go:build amd64 && !purego

#include "textflag.h"

// hasAVX2Support checks if AVX2 is supported
// func hasAVX2Support() bool
TEXT ·hasAVX2Support(SB), NOSPLIT, $0-1
    MOVQ $7, AX
    MOVQ $0, CX
    CPUID
    SHRQ $5, BX
    ANDQ $1, BX
    MOVB BX, ret+0(FP)
    RET

// avx2PopCount performs SIMD population count using AVX2
// func avx2PopCount(data unsafe.Pointer, length int) int
TEXT ·avx2PopCount(SB), NOSPLIT, $0-24
    MOVQ data+0(FP), SI      // Load data pointer
    MOVQ length+8(FP), CX    // Load length in bytes
    XORQ AX, AX              // Initialize count accumulator
    XORQ DX, DX              // Initialize loop counter

    // Check if length is less than 32 bytes (256 bits)
    CMPQ CX, $32
    JL scalar_loop

    // Prepare for AVX2 processing - align to 32-byte chunks
    MOVQ CX, R8
    SUBQ DX, R8              // Remaining bytes
    SHRQ $5, R8              // Number of 32-byte chunks
    SHLQ $5, R8              // Aligned length for AVX2

avx2_loop:
    CMPQ DX, R8
    JGE scalar_loop

    // Load 32 bytes (256 bits) using AVX2
    VMOVDQU (SI)(DX*1), Y0

    // Count bits using a simpler method: process as uint64s using POPCNT
    // Extract and process each 64-bit chunk
    VMOVQ X0, R9
    POPCNTQ R9, R9
    ADDQ R9, AX

    VPEXTRQ $1, X0, R9
    POPCNTQ R9, R9
    ADDQ R9, AX

    VEXTRACTI128 $1, Y0, X1
    VMOVQ X1, R9
    POPCNTQ R9, R9
    ADDQ R9, AX

    VPEXTRQ $1, X1, R9
    POPCNTQ R9, R9
    ADDQ R9, AX

    ADDQ $32, DX
    JMP avx2_loop

scalar_loop:
    CMPQ DX, CX
    JGE done

    MOVBQZX (SI)(DX*1), R9   // Load one byte

    // Use POPCNT if available (part of SSE4.2)
    POPCNTQ R9, R9
    ADDQ R9, AX

    INCQ DX
    JMP scalar_loop

done:
    VZEROUPPER                // Clear upper AVX state
    MOVQ AX, ret+16(FP)       // Store result
    RET

// avx2VectorOr performs SIMD OR operation using AVX2
// func avx2VectorOr(dst, src unsafe.Pointer, length int)
TEXT ·avx2VectorOr(SB), NOSPLIT, $0-24
    MOVQ dst+0(FP), DI       // Load dst pointer
    MOVQ src+8(FP), SI       // Load src pointer
    MOVQ length+16(FP), CX   // Load length in bytes
    XORQ DX, DX              // Initialize loop counter

    // Check if we have at least 32 bytes
    CMPQ CX, $32
    JL scalar_or_loop

    // Calculate number of 32-byte chunks
    MOVQ CX, R8
    SHRQ $5, R8
    SHLQ $5, R8              // Aligned length

avx2_or_loop:
    CMPQ DX, R8
    JGE scalar_or_loop

    // Load 32 bytes from src and dst
    VMOVDQU (SI)(DX*1), Y0   // Load src
    VMOVDQU (DI)(DX*1), Y1   // Load dst

    // Perform OR operation
    VPOR Y0, Y1, Y1          // dst = dst | src

    // Store result back to dst
    VMOVDQU Y1, (DI)(DX*1)

    ADDQ $32, DX
    JMP avx2_or_loop

scalar_or_loop:
    CMPQ DX, CX
    JGE or_done

    MOVBQZX (DI)(DX*1), AX   // Load dst byte
    MOVBQZX (SI)(DX*1), R9   // Load src byte
    ORQ R9, AX               // dst = dst | src
    MOVB AX, (DI)(DX*1)      // Store result

    INCQ DX
    JMP scalar_or_loop

or_done:
    VZEROUPPER
    RET

// avx2VectorAnd performs SIMD AND operation using AVX2
// func avx2VectorAnd(dst, src unsafe.Pointer, length int)
TEXT ·avx2VectorAnd(SB), NOSPLIT, $0-24
    MOVQ dst+0(FP), DI       // Load dst pointer
    MOVQ src+8(FP), SI       // Load src pointer
    MOVQ length+16(FP), CX   // Load length in bytes
    XORQ DX, DX              // Initialize loop counter

    // Check if we have at least 32 bytes
    CMPQ CX, $32
    JL scalar_and_loop

    // Calculate number of 32-byte chunks
    MOVQ CX, R8
    SHRQ $5, R8
    SHLQ $5, R8              // Aligned length

avx2_and_loop:
    CMPQ DX, R8
    JGE scalar_and_loop

    // Load 32 bytes from src and dst
    VMOVDQU (SI)(DX*1), Y0   // Load src
    VMOVDQU (DI)(DX*1), Y1   // Load dst

    // Perform AND operation
    VPAND Y0, Y1, Y1         // dst = dst & src

    // Store result back to dst
    VMOVDQU Y1, (DI)(DX*1)

    ADDQ $32, DX
    JMP avx2_and_loop

scalar_and_loop:
    CMPQ DX, CX
    JGE and_done

    MOVBQZX (DI)(DX*1), AX   // Load dst byte
    MOVBQZX (SI)(DX*1), R9   // Load src byte
    ANDQ R9, AX              // dst = dst & src
    MOVB AX, (DI)(DX*1)      // Store result

    INCQ DX
    JMP scalar_and_loop

and_done:
    VZEROUPPER
    RET

// avx2VectorClear performs SIMD clear operation using AVX2
// func avx2VectorClear(data unsafe.Pointer, length int)
TEXT ·avx2VectorClear(SB), NOSPLIT, $0-16
    MOVQ data+0(FP), DI      // Load data pointer
    MOVQ length+8(FP), CX    // Load length in bytes
    XORQ DX, DX              // Initialize loop counter

    // Zero out YMM register for clearing
    VPXOR Y0, Y0, Y0

    // Check if we have at least 32 bytes
    CMPQ CX, $32
    JL scalar_clear_loop

    // Calculate number of 32-byte chunks
    MOVQ CX, R8
    SHRQ $5, R8
    SHLQ $5, R8              // Aligned length

avx2_clear_loop:
    CMPQ DX, R8
    JGE scalar_clear_loop

    // Store 32 zeros
    VMOVDQU Y0, (DI)(DX*1)

    ADDQ $32, DX
    JMP avx2_clear_loop

scalar_clear_loop:
    CMPQ DX, CX
    JGE clear_done

    MOVB $0, (DI)(DX*1)      // Store zero byte
    INCQ DX
    JMP scalar_clear_loop

clear_done:
    VZEROUPPER
    RET
