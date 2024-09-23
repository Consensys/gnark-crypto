// +build !purego

#define q0 $0x0a11800000000001
#define q1 $0x59aa76fed0000001
#define q2 $0x60b44d1e5c37b001
#define q3 $0x12ab655e9a2ca556

#include "../../../field/asm/element_4w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0x0a11800000000001
DATA q<>+8(SB)/8, $0x59aa76fed0000001
DATA q<>+16(SB)/8, $0x60b44d1e5c37b001
DATA q<>+24(SB)/8, $0x12ab655e9a2ca556
GLOBL q<>(SB), (RODATA+NOPTR), $32

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x0a117fffffffffff
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// Mu
DATA mu<>(SB)/8, $0x0000000db65247b1
GLOBL mu<>(SB), (RODATA+NOPTR), $8

