// +build !purego

#define q0 $0x1e66a241adc64d2f
#define q1 $0xb781126dcae7b232
#define q2 $0xffffffffffffffff
#define q3 $0x0800000000000010
#include "../../../field/asm/element_4w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0x1e66a241adc64d2f
DATA q<>+8(SB)/8, $0xb781126dcae7b232
DATA q<>+16(SB)/8, $0xffffffffffffffff
DATA q<>+24(SB)/8, $0x0800000000000010
GLOBL q<>(SB), (RODATA+NOPTR), $32

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0xbb6b3c4ce8bde631
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// Mu
DATA mu<>(SB)/8, $0x0000001fffffffff
GLOBL mu<>(SB), (RODATA+NOPTR), $8

