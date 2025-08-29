// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"

	"github.com/consensys/bavard/amd64"
)

func (_f *FFAmd64) generateFromMont(_ bool) {
	const argSize = 8
	const minStackSize = argSize
	nbRegistersNeeded := (_f.NbWords * 2) - 2

	// we need to use R15 register, and to avoid issue with dynamic linking
	// see https://github.com/Consensys/gnark-crypto/issues/707
	// we avoid using global variables in this particular instance.
	needR15 := _f.NbWords >= 12
	if needR15 {
		nbRegistersNeeded += _f.NbWords // we need to store Q on the stack.
		nbRegistersNeeded--             // account for R15
	}

	stackSize := _f.StackSize(nbRegistersNeeded, 2, minStackSize)
	registers := _f.FnHeader("fromMont", stackSize, argSize, amd64.DX, amd64.AX)
	defer _f.AssertCleanStack(stackSize, minStackSize)

	if stackSize > 0 {
		_f.WriteLn("NO_LOCAL_POINTERS")
	}

	_f.WriteLn(`
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

	if needR15 {
		registers.UnsafePush(amd64.R15)
		_q := _f.PopN(&registers, true)
		for i := 0; i < _f.NbWords; i++ {
			_f.MOVQ(fmt.Sprintf("$const_q%d", i), amd64.AX)
			_f.MOVQ(amd64.AX, _q[i])
		}
		_f.SetQStack(_q)
		defer func() {
			_f.Push(&registers, _q...)
			_f.UnsetQStack()
		}()
	}

	noAdx := _f.NewLabel("noAdx")
	{
		// check ADX instruction support
		_f.CMPB("·supportAdx(SB)", 1)
		_f.JNE(noAdx)
	}

	// registers
	t := registers.PopN(_f.NbWords)

	_f.MOVQ("res+0(FP)", amd64.DX)

	// 	for i=0 to N-1
	//     t[i] = a[i]
	_f.Mov(amd64.DX, t)

	for i := 0; i < _f.NbWords; i++ {

		_f.XORQ(amd64.DX, amd64.DX)

		// m := t[0]*q'[0] mod W
		_f.Comment("m := t[0]*q'[0] mod W")
		m := amd64.DX
		_f.MOVQ(_f.qInv0(), m)
		_f.IMULQ(t[0], m)

		// clear the carry flags
		_f.XORQ(amd64.AX, amd64.AX)

		// C,_ := t[0] + m*q[0]
		_f.Comment("C,_ := t[0] + m*q[0]")

		_f.MULXQ(_f.qAt(0), amd64.AX, amd64.BP)
		_f.ADCXQ(t[0], amd64.AX)
		_f.MOVQ(amd64.BP, t[0])

		// for j=1 to N-1
		//    (C,t[j-1]) := t[j] + m*q[j] + C
		for j := 1; j < _f.NbWords; j++ {
			_f.Comment(fmt.Sprintf("(C,t[%[1]d]) := t[%[2]d] + m*q[%[2]d] + C", j-1, j))
			_f.ADCXQ(t[j], t[j-1])
			_f.MULXQ(_f.qAt(j), amd64.AX, t[j])
			_f.ADOXQ(amd64.AX, t[j-1])
		}
		_f.MOVQ(0, amd64.AX)
		_f.ADCXQ(amd64.AX, t[_f.NbWordsLastIndex])
		_f.ADOXQ(amd64.AX, t[_f.NbWordsLastIndex])

	}

	// ---------------------------------------------------------------------------------------------
	// reduce
	_f.Push(&registers, amd64.DX, amd64.AX)
	_f.Reduce(&registers, t, needR15)
	_f.MOVQ("res+0(FP)", amd64.AX)
	_f.Mov(t, amd64.AX)
	_f.RET()

	// No adx
	{
		_f.LABEL(noAdx)
		_f.MOVQ("res+0(FP)", amd64.AX)
		_f.MOVQ(amd64.AX, "(SP)")
		_f.WriteLn("CALL ·_fromMontGeneric(SB)")
		_f.RET()
	}

}
