// +build !purego

#define q0 $0xffffffff00000001
#define q1 $0x53bda402fffe5bfe
#define q2 $0x3339d80809a1d805
#define q3 $0x73eda753299d7d48
#include "../../../field/asm/element_4w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0xffffffff00000001
DATA q<>+8(SB)/8, $0x53bda402fffe5bfe
DATA q<>+16(SB)/8, $0x3339d80809a1d805
DATA q<>+24(SB)/8, $0x73eda753299d7d48
GLOBL q<>(SB), (RODATA+NOPTR), $32

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0xfffffffeffffffff
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// Mu
DATA mu<>(SB)/8, $0x00000002355094ed
GLOBL mu<>(SB), (RODATA+NOPTR), $8

