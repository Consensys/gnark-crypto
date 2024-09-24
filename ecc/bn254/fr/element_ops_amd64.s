// +build !purego

#define q0 $0x43e1f593f0000001
#define q1 $0x2833e84879b97091
#define q2 $0xb85045b68181585d
#define q3 $0x30644e72e131a029

#include "../../../field/asm/element_4w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0x43e1f593f0000001
DATA q<>+8(SB)/8, $0x2833e84879b97091
DATA q<>+16(SB)/8, $0xb85045b68181585d
DATA q<>+24(SB)/8, $0x30644e72e131a029
GLOBL q<>(SB), (RODATA+NOPTR), $32

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0xc2e1f593efffffff
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// Mu
DATA mu<>(SB)/8, $0x000000054a474626
GLOBL mu<>(SB), (RODATA+NOPTR), $8

