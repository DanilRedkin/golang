#include "textflag.h"

// func Fibonacci(n uint64) uint64
TEXT ·Fibonacci(SB), NOSPLIT, $0
    MOVD    number+0(FP), R0

    MOVD    $0, R1
    MOVD    $1, R2

    CMP $0, R0
    BEQ return_zero

    MOVD    $1, R3

loop:
    CMP R0, R3
    BGE return_fib

    MOVD R2, R4
    ADD R1, R2
    MOVD R4, R1
    ADD $1, R3

    B loop

return_zero:
    MOVD    R1, ret+8(FP)
    RET

return_fib:
    MOVD    R2, ret+8(FP)
    RET
