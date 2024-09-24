// +build !purego

#define q0 $0xb9feffffffffaaab
#define q1 $0x1eabfffeb153ffff
#define q2 $0x6730d2a0f6b0f624
#define q3 $0x64774b84f38512bf
#define q4 $0x4b1ba7b6434bacd7
#define q5 $0x1a0111ea397fe69a

#include "../../../field/asm/element_6w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0xb9feffffffffaaab
DATA q<>+8(SB)/8, $0x1eabfffeb153ffff
DATA q<>+16(SB)/8, $0x6730d2a0f6b0f624
DATA q<>+24(SB)/8, $0x64774b84f38512bf
DATA q<>+32(SB)/8, $0x4b1ba7b6434bacd7
DATA q<>+40(SB)/8, $0x1a0111ea397fe69a
GLOBL q<>(SB), (RODATA+NOPTR), $48

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x89f3fffcfffcfffd
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

