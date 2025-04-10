#include "textflag.h"

// func LowerBound(slice []int64, value int64) int64
TEXT ·LowerBound(SB), NOSPLIT, $0
    LDP slice+0(FP), (R0, R1)
    MOVD value+24(FP), R2

    MOVD $0, R3
    MOVD R1, R4
    SUB $1, R4
    MOVD $-1, R5

loop:
    CMP     R4, R3
    BGT     end_loop

    MOVD    R3, R6
    ADD     R4, R6
    LSR     $1, R6

    MOVD $8, R7
    MUL     R6, R7
    ADD     R0, R7
    MOVD    (R7), R7

    CMP     R2, R7
    BLE     search_left

    CMP     R2, R7
    BGT     search_right

search_left:
    MOVD R6, R5
    ADD $1, R6
    MOVD R6, R3
    B loop

search_right:
    SUB $1, R6, R4
    B loop

end_loop:
    MOVD    R5, ret+32(FP)
    RET
