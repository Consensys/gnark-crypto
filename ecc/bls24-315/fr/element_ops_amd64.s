// +build !purego

#define q0 $0x19d0c5fd00c00001
#define q1 $0xc8c480ece644e364
#define q2 $0x25fc7ec9cf927a98
#define q3 $0x196deac24a9da12b

#include "../../../field/asm/element_4w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0x19d0c5fd00c00001
DATA q<>+8(SB)/8, $0xc8c480ece644e364
DATA q<>+16(SB)/8, $0x25fc7ec9cf927a98
DATA q<>+24(SB)/8, $0x196deac24a9da12b
GLOBL q<>(SB), (RODATA+NOPTR), $32

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0x1e5035fd00bfffff
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
// Mu
DATA mu<>(SB)/8, $0x0000000a112d9c09
GLOBL mu<>(SB), (RODATA+NOPTR), $8

