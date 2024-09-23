// +build !purego

#define q0 $0xd74916ea4570000d
#define q1 $0x3d369bd31147f73c
#define q2 $0xd7b5ce7ab839c225
#define q3 $0x7e0e8850edbda407
#define q4 $0xb8da9f5e83f57c49
#define q5 $0x8152a6c0fadea490
#define q6 $0x4e59769ad9bbda2f
#define q7 $0xa8fcd8c75d79d2c7
#define q8 $0xfc1a174f01d72ab5
#define q9 $0x0126633cc0f35f63
#include "../../../field/asm/element_10w_amd64.h"

// modulus q
DATA q<>+0(SB)/8, $0xd74916ea4570000d
DATA q<>+8(SB)/8, $0x3d369bd31147f73c
DATA q<>+16(SB)/8, $0xd7b5ce7ab839c225
DATA q<>+24(SB)/8, $0x7e0e8850edbda407
DATA q<>+32(SB)/8, $0xb8da9f5e83f57c49
DATA q<>+40(SB)/8, $0x8152a6c0fadea490
DATA q<>+48(SB)/8, $0x4e59769ad9bbda2f
DATA q<>+56(SB)/8, $0xa8fcd8c75d79d2c7
DATA q<>+64(SB)/8, $0xfc1a174f01d72ab5
DATA q<>+72(SB)/8, $0x0126633cc0f35f63
GLOBL q<>(SB), (RODATA+NOPTR), $80

// qInv0 q'[0]
DATA qInv0<>(SB)/8, $0xb50f29ab0b03b13b
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8

