#include "textflag.h"

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24   //what is nosplit? What are the resigter numbers?

    MOVD.P -8(x), R8
    //MOV x+8(FP), X0
    //LDR x0, [x]
    //ldr x1, [x+8]


// addLessGeneric(res, x, y *Element)
TEXT ·addLessGeneric(SB), NOSPLIT, $0-24   //No variables, 3 arguments

    LDP x+8(FP), (R0, R1)  // R0 = x, R1 = y

    LDP 0(R0), (R2, R3) // R2 = x[0], R3 = x[1]
    LDP 0(R1), (R4, R5) // R4 = y[0], R5 = y[1]
    ADDS R2, R4, R2 // R2 = z[0]
    ADCS R3, R5, R3 // R3 = z[1]

    LDP 16(R0), (R4, R5)   //R4 = x[2], R5 = x[3]
    LDP 16(R1), (R6, R7)   //R6 = y[2], R7 = y[3]
    ADCS R4, R6, R4 // R4 = z[2]
    ADCS R5, R7, R5 // R5 = z[3]

    // Now load q
    LDP q<>(SB), (R0, R1)  //R0 = q[0], R1 = q[1]
    // TODO: Can this be done in pairs?
    //LDP q<>+16(SB) (R6, R7)    //R6 = q[2], R7 = q[3]
    MOVD q<>+16(SB), R6
    MOVD q<>+24(SB), R7

    // Subtract q
    SUBS R0, R2, R0 // R0 = z[0] - q[0]  TODO: Make sure arguments are lined up correctly
    SBCS R1, R3, R1 // R1 = z[1] - q[1]
    SBCS R6, R4, R6 // R6 = z[2] - q[2]
    SBCS R7, R5, R7 // R7 = z[3] - q[3]

    // If borrow not needed, select subtraction result
    CSEL CS, R0, R2, R0  //TODO: Wrong way around?
    CSEL CS, R1, R3, R1
    CSEL CS, R6, R4, R6
    CSEL CS, R7, R5, R7

    //Store
    MOVD res+0(FP), R2 // R2 = z
    STP (R0, R1), 0(R2)
    STP (R6, R7), 16(R2)

    RET
