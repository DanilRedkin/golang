#include "textflag.h"

// func WordCount(data []rune) int32
TEXT ·WordCount(SB), NOSPLIT, $0
    LDP slice+0(FP), (R0, R1)
    MOVW $0, R2 // R2 := counter
    MOVW $0, R3 // R3 := flag

loop:
    SUBS $1, R1
    BLT end

    MOVW R1, R4
    MOVW $4, R5
    MUL R5, R4
    ADD R0, R4
    MOVWU (R4), R6 //R6 := single char from string

single_checks:
    MOVW $0x20, R7
    CMP R6, R7
    BEQ space_encountered

    MOVW $0x85, R7
    CMP R6, R7
    BEQ space_encountered

    MOVW $0xA0, R7
    CMP R6, R7
    BEQ space_encountered

    MOVW $0x202F, R7
    CMP R6, R7
    BEQ space_encountered

    MOVW $0x205F, R7
    CMP R6, R7
    BEQ space_encountered

    MOVW $0x3000, R7
    CMP R6, R7
    BEQ space_encountered

    MOVW $0x2028, R7
    CMP R6, R7
    BEQ space_encountered

    MOVW $0x2029, R7
    CMP R6, R7
    BEQ space_encountered

    MOVW $0x1680, R7
    CMP R6, R7
    BEQ space_encountered

range_check_1:
    MOVW $0x0009, R4
    MOVW $0x000D, R5
    CMP R4, R6
    BGE range_check_1_next

range_check_2:
    MOVW $0x2000, R4
    MOVW $0x200A, R5
    CMP R4, R6
    BGE range_check_2_next

not_space_encountered:
    MOVW $0, R7
    CMP R3, R7
    BNE loop

    MOVW $1, R3
    ADD $1, R2
    B loop

space_encountered:
    MOVW $0, R7
    CMP R3, R7
    BEQ loop

    MOVW $0, R3
    B loop

range_check_1_next:
    CMP R5, R6
    BLE space_encountered
    B range_check_2

range_check_2_next:
    CMP R5, R6
    BLE space_encountered
    B not_space_encountered

end:
    MOVW R2, ret+24(FP)
    RET
