// +build !purego

#define q0 $0x8508c00000000001
#define q1 $0x170b5d4430000000
#define q2 $0x1ef3622fba094800
#define q3 $0x1a22d9f300f5138f
#define q4 $0xc63b05c06ca1493b
#define q5 $0x01ae3a4617c510ea

#include "../../../field/asm/element_6w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0x8508c00000000001
DATA q<>+8(SB)/8, $0x170b5d4430000000
DATA q<>+16(SB)/8, $0x1ef3622fba094800
DATA q<>+24(SB)/8, $0x1a22d9f300f5138f
DATA q<>+32(SB)/8, $0xc63b05c06ca1493b
DATA q<>+40(SB)/8, $0x01ae3a4617c510ea
GLOBL q<>(SB), (RODATA+NOPTR), $48

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x8508bfffffffffff
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

