// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"

	"github.com/consensys/bavard/amd64"
)

func (f *FFAmd64) generateFromMont(forceADX bool) {
	const argSize = 8
	minStackSize := argSize
	if forceADX {
		minStackSize = 0
	}
	stackSize := f.StackSize(f.NbWords*2, 2, minStackSize)

	reserved := []amd64.Register{amd64.DX, amd64.AX}
	if f.NbWords <= 5 {
		// when dynamic linking, R15 is clobbered by a global variable access
		// this is a temporary workaround --> don't use R15 when we can avoid it.
		// see https://github.com/Consensys/gnark-crypto/issues/113
		reserved = append(reserved, amd64.R15)
	}
	registers := f.FnHeader("fromMont", stackSize, argSize, reserved...)
	defer f.AssertCleanStack(stackSize, minStackSize)

	if stackSize > 0 {
		f.WriteLn("NO_LOCAL_POINTERS")
	}
	f.WriteLn(`
	// Algorithm 2 of "Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS" 
	// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521
	// when y = 1 we have: 
	// for i=0 to N-1
	// 		t[i] = x[i]
	// for i=0 to N-1
	// 		m := t[0]*q'[0] mod W
	// 		C,_ := t[0] + m*q[0]
	// 		for j=1 to N-1
	// 		    (C,t[j-1]) := t[j] + m*q[j] + C
	// 		t[N-1] = C`)

	noAdx := f.NewLabel("noAdx")
	if !forceADX {
		// check ADX instruction support
		f.CMPB("·supportAdx(SB)", 1)
		f.JNE(noAdx)
	}

	// registers
	t := registers.PopN(f.NbWords)

	f.MOVQ("res+0(FP)", amd64.DX)

	// 	for i=0 to N-1
	//     t[i] = a[i]
	f.Mov(amd64.DX, t)

	for i := 0; i < f.NbWords; i++ {

		f.XORQ(amd64.DX, amd64.DX)

		// m := t[0]*q'[0] mod W
		f.Comment("m := t[0]*q'[0] mod W")
		m := amd64.DX
		f.MOVQ(f.qInv0(), m)
		f.IMULQ(t[0], m)

		// clear the carry flags
		f.XORQ(amd64.AX, amd64.AX)

		// C,_ := t[0] + m*q[0]
		f.Comment("C,_ := t[0] + m*q[0]")

		f.MULXQ(f.qAt(0), amd64.AX, amd64.BP)
		f.ADCXQ(t[0], amd64.AX)
		f.MOVQ(amd64.BP, t[0])

		// for j=1 to N-1
		//    (C,t[j-1]) := t[j] + m*q[j] + C
		for j := 1; j < f.NbWords; j++ {
			f.Comment(fmt.Sprintf("(C,t[%[1]d]) := t[%[2]d] + m*q[%[2]d] + C", j-1, j))
			f.ADCXQ(t[j], t[j-1])
			f.MULXQ(f.qAt(j), amd64.AX, t[j])
			f.ADOXQ(amd64.AX, t[j-1])
		}
		f.MOVQ(0, amd64.AX)
		f.ADCXQ(amd64.AX, t[f.NbWordsLastIndex])
		f.ADOXQ(amd64.AX, t[f.NbWordsLastIndex])

	}

	// ---------------------------------------------------------------------------------------------
	// reduce
	f.Reduce(&registers, t)
	f.MOVQ("res+0(FP)", amd64.AX)
	f.Mov(t, amd64.AX)
	f.RET()

	// No adx
	if !forceADX {
		f.LABEL(noAdx)
		f.MOVQ("res+0(FP)", amd64.AX)
		f.MOVQ(amd64.AX, "(SP)")
		f.WriteLn("CALL ·_fromMontGeneric(SB)")
		f.RET()
	}

}
