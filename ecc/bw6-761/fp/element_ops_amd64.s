// +build !purego

#define q0 $0xf49d00000000008b
#define q1 $0xe6913e6870000082
#define q2 $0x160cf8aeeaf0a437
#define q3 $0x98a116c25667a8f8
#define q4 $0x71dcd3dc73ebff2e
#define q5 $0x8689c8ed12f9fd90
#define q6 $0x03cebaff25b42304
#define q7 $0x707ba638e584e919
#define q8 $0x528275ef8087be41
#define q9 $0xb926186a81d14688
#define q10 $0xd187c94004faff3e
#define q11 $0x0122e824fb83ce0a
#include "../../../field/asm/element_12w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0xf49d00000000008b
DATA q<>+8(SB)/8, $0xe6913e6870000082
DATA q<>+16(SB)/8, $0x160cf8aeeaf0a437
DATA q<>+24(SB)/8, $0x98a116c25667a8f8
DATA q<>+32(SB)/8, $0x71dcd3dc73ebff2e
DATA q<>+40(SB)/8, $0x8689c8ed12f9fd90
DATA q<>+48(SB)/8, $0x03cebaff25b42304
DATA q<>+56(SB)/8, $0x707ba638e584e919
DATA q<>+64(SB)/8, $0x528275ef8087be41
DATA q<>+72(SB)/8, $0xb926186a81d14688
DATA q<>+80(SB)/8, $0xd187c94004faff3e
DATA q<>+88(SB)/8, $0x0122e824fb83ce0a
GLOBL q<>(SB), (RODATA+NOPTR), $96

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x0a5593568fa798dd
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

