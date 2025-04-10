#include "textflag.h"

// func SumSlice(s []int32) int64
TEXT ·SumSlice(SB), NOSPLIT, $0
    LDP slice_base+0(FP), (R0, R1)
    MOVD $0, R2

loop:
    CBZ R1, end_loop
    MOVW.P 4(R0), R3
    ADD R3, R2
    SUB $1, R1
    B loop

end_loop:
    MOVD R2, ret+24(FP)
    RET
