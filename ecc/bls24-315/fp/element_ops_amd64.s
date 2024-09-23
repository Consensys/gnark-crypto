// +build !purego

#define q0 $0x6fe802ff40300001
#define q1 $0x421ee5da52bde502
#define q2 $0xdec1d01aa27a1ae0
#define q3 $0xd3f7498be97c5eaf
#define q4 $0x04c23a02b586d650
#include "../../../field/asm/element_5w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0x6fe802ff40300001
DATA q<>+8(SB)/8, $0x421ee5da52bde502
DATA q<>+16(SB)/8, $0xdec1d01aa27a1ae0
DATA q<>+24(SB)/8, $0xd3f7498be97c5eaf
DATA q<>+32(SB)/8, $0x04c23a02b586d650
GLOBL q<>(SB), (RODATA+NOPTR), $40

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x702ff9ff402fffff
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

