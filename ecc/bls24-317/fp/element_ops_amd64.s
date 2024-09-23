// +build !purego

#define q0 $0x8d512e565dab2aab
#define q1 $0xd6f339e43424bf7e
#define q2 $0x169a61e684c73446
#define q3 $0xf28fc5a0b7f9d039
#define q4 $0x1058ca226f60892c
#include "../../../field/asm/element_5w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0x8d512e565dab2aab
DATA q<>+8(SB)/8, $0xd6f339e43424bf7e
DATA q<>+16(SB)/8, $0x169a61e684c73446
DATA q<>+24(SB)/8, $0xf28fc5a0b7f9d039
DATA q<>+32(SB)/8, $0x1058ca226f60892c
GLOBL q<>(SB), (RODATA+NOPTR), $40

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x55b5e0028b047ffd
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

