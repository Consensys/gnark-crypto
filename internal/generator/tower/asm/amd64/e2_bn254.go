// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"github.com/consensys/bavard/amd64"
)

func (fq2 *Fq2Amd64) generateMulByNonResidueE2BN254() {
	// 	var a, b fp.Element
	// 	a.Double(&x.A0).Double(&a).Double(&a).fq2.Add(&a, &x.A0).fq2.Sub(&a, &x.A1)
	// 	b.Double(&x.A1).Double(&b).Double(&b).fq2.Add(&b, &x.A1).fq2.Add(&b, &x.A0)
	// 	z.A0.Set(&a)
	// 	z.A1.Set(&b)
	registers := fq2.FnHeader("mulNonResE2", 0, 16)

	a := registers.PopN(fq2.NbWords)
	b := registers.PopN(fq2.NbWords)
	x := registers.Pop()

	fq2.MOVQ("x+8(FP)", x)
	fq2.Mov(x, a) // a = a0

	fq2.Add(a, a)
	fq2.Reduce(&registers, a, false)

	fq2.Add(a, a)
	fq2.Reduce(&registers, a, false)

	fq2.Add(a, a)
	fq2.Reduce(&registers, a, false)

	fq2.Add(x, a)
	fq2.Reduce(&registers, a, false)

	fq2.Mov(x, b, fq2.NbWords) // b = a1
	zero := amd64.R15
	fq2.XORQ(zero, zero)
	fq2.Sub(b, a)
	fq2.modReduceAfterSub(&registers, zero, a)

	fq2.Add(b, b)
	fq2.Reduce(&registers, b, false)

	fq2.Add(b, b)
	fq2.Reduce(&registers, b, false)

	fq2.Add(b, b)
	fq2.Reduce(&registers, b, false)

	fq2.Add(x, b, fq2.NbWords)
	fq2.Reduce(&registers, b, false)
	fq2.Add(x, b)
	fq2.Reduce(&registers, b, false)

	fq2.MOVQ("res+0(FP)", x)
	fq2.Mov(a, x)
	fq2.Mov(b, x, 0, fq2.NbWords)

	fq2.RET()
}

func (fq2 *Fq2Amd64) generateMulE2BN254(forceCheck bool) {
	// var a, b, c fp.Element
	// a.Add(&x.A0, &x.A1)
	// b.Add(&y.A0, &y.A1)
	// a.Mul(&a, &b)
	// b.Mul(&x.A0, &y.A0)
	// c.Mul(&x.A1, &y.A1)
	// z.A1.Sub(&a, &b).Sub(&z.A1, &c)
	// z.A0.Sub(&b, &c)

	// we need a bit of stack space to store the results of the xA0yA0 and xA1yA1 multiplications
	const argSize = 24
	minStackSize := (4 + 2) * 8 // aStack and cStack are (partially) on the stack
	stackSize := fq2.StackSize(fq2.NbWords*4, 2, minStackSize)
	registers := fq2.FnHeader("mulAdxE2", stackSize, argSize, amd64.DX, amd64.AX)
	defer fq2.AssertCleanStack(stackSize, minStackSize)

	fq2.WriteLn("NO_LOCAL_POINTERS")

	// check ADX instruction support
	lblNoAdx := fq2.NewLabel()
	if forceCheck {
		fq2.CMPB("·supportAdx(SB)", 1)
		fq2.JNE(lblNoAdx)
	}

	fq2.WriteLn(`
	// var a, b, c fp.Element
	// a.Add(&x.A0, &x.A1)
	// b.Add(&y.A0, &y.A1)
	// a.Mul(&a, &b)
	// b.Mul(&x.A0, &y.A0)
	// c.Mul(&x.A1, &y.A1)
	// z.A1.Sub(&a, &b).Sub(&z.A1, &c)
	// z.A0.Sub(&b, &c)
	`)

	// used in the mul operation
	op1 := registers.PopN(fq2.NbWords)
	res := registers.PopN(fq2.NbWords)

	xat := func(i int) string {
		return string(op1[i])
	}

	ax := amd64.AX
	dx := amd64.DX

	_r0 := fq2.Pop(&registers) // reserve this one for the MulADX
	aStack := fq2.PopN(&registers)
	cStack := fq2.PopN(&registers)
	fq2.Push(&registers, _r0)

	fq2.MOVQ("x+8(FP)", ax)

	// c = x.A1 * y.A1
	fq2.Mov(ax, op1, fq2.NbWords)
	fq2.MulADX(&registers, xat, func(i int) string {
		fq2.MOVQ("y+16(FP)", dx)
		return dx.At(i + fq2.NbWords)
	}, res)
	fq2.ReduceElement(res, op1, false)
	// res = x.A1 * y.A1
	// pushing on stack for later use.
	fq2.Mov(res, cStack)

	fq2.MOVQ("x+8(FP)", ax)
	fq2.MOVQ("y+16(FP)", dx)

	// a = x.a0 + x.a1
	fq2.Mov(ax, op1, fq2.NbWords)
	fq2.Add(ax, op1)
	fq2.Mov(op1, aStack)

	// b = y.a0 + y.a1
	fq2.Mov(dx, op1)
	fq2.Add(dx, op1, fq2.NbWords)
	// --> note, we don't reduce, as this is used as input to the mul which accept input of size D-1/2 -1

	// a = 	a * b = (x.a0 + x.a1) *  (y.a0 + y.a1)
	fq2.MulADX(&registers, xat, func(i int) string {
		return string(aStack[i])
	}, res)
	fq2.ReduceElement(res, op1, false)

	// moving result to the stack.
	fq2.Mov(res, aStack)

	// b = x.A0 * y.AO
	fq2.MOVQ("x+8(FP)", ax)

	fq2.Mov(ax, op1)
	fq2.MulADX(&registers, xat, func(i int) string {
		fq2.MOVQ("y+16(FP)", dx)
		return dx.At(i)
	}, res)
	fq2.ReduceElement(res, op1, false)

	zero := dx
	fq2.XORQ(zero, zero)

	// a = a - b -c
	fq2.Mov(aStack, op1)
	fq2.Sub(res, op1) // a -= b
	fq2.Mov(res, aStack)
	fq2.modReduceAfterSubScratch(zero, op1, res)

	fq2.Sub(cStack, op1) // a -= c
	fq2.modReduceAfterSubScratch(zero, op1, res)

	fq2.MOVQ("res+0(FP)", ax)
	fq2.Mov(op1, ax, 0, fq2.NbWords)

	// b = b - c
	fq2.Mov(aStack, res)
	fq2.Sub(cStack, res) // b -= c
	fq2.modReduceAfterSubScratch(zero, res, op1)

	fq2.Mov(res, ax)

	fq2.RET()

	// No adx
	if forceCheck {
		fq2.LABEL(lblNoAdx)
		fq2.MOVQ("res+0(FP)", amd64.AX)
		fq2.MOVQ(amd64.AX, "(SP)")
		fq2.MOVQ("x+8(FP)", amd64.AX)
		fq2.MOVQ(amd64.AX, "8(SP)")
		fq2.MOVQ("y+16(FP)", amd64.AX)
		fq2.MOVQ(amd64.AX, "16(SP)")
		fq2.WriteLn("CALL ·mulGenericE2(SB)")
		fq2.RET()
	}

	fq2.Push(&registers, aStack...)
	fq2.Push(&registers, cStack...)
}
