// +build !purego

#define q0 $0xf000000000000001
#define q1 $0x1cd1e79196bf0e7a
#define q2 $0xd0b097f28d83cd49
#define q3 $0x443f917ea68dafc2

#include "../../../field/asm/element_4w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0xf000000000000001
DATA q<>+8(SB)/8, $0x1cd1e79196bf0e7a
DATA q<>+16(SB)/8, $0xd0b097f28d83cd49
DATA q<>+24(SB)/8, $0x443f917ea68dafc2
GLOBL q<>(SB), (RODATA+NOPTR), $32

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0xefffffffffffffff
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// Mu
DATA mu<>(SB)/8, $0x00000003c0421687
GLOBL mu<>(SB), (RODATA+NOPTR), $8

