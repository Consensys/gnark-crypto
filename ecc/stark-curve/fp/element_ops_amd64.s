// +build !purego

#define q0 $0x0000000000000001
#define q1 $0x0000000000000000
#define q2 $0x0000000000000000
#define q3 $0x0800000000000011

#include "../../../field/asm/element_4w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $1
DATA q<>+8(SB)/8, $0
DATA q<>+16(SB)/8, $0
DATA q<>+24(SB)/8, $0x0800000000000011
GLOBL q<>(SB), (RODATA+NOPTR), $32

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0xffffffffffffffff
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// Mu
DATA mu<>(SB)/8, $0x0000001fffffffff
GLOBL mu<>(SB), (RODATA+NOPTR), $8

