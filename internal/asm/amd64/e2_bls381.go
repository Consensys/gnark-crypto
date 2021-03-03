// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package amd64

import (
	"github.com/consensys/bavard/amd64"
)

func (fq2 *Fq2Amd64) generateMulByNonResidueE2BLS381() {
	// // MulByNonResidue multiplies a E2 by (1,1)
	// func (z *E2) MulByNonResidue(x *E2) *E2 {
	// 	var a fp.Element
	// 	a.Sub(&x.A0, &x.A1)
	// 	z.A1.Add(&x.A0, &x.A1)
	// 	z.A0.Set(&a)
	// 	return z
	// }
	registers := fq2.FnHeader("mulNonResE2", 0, 16)

	a := registers.PopN(fq2.NbWords)
	b := registers.PopN(fq2.NbWords)
	x := registers.Pop()
	tr := registers.Pop() // zero or r
	fq2.XORQ(tr, tr)      // set to zero

	fq2.MOVQ("x+8(FP)", x)
	fq2.Mov(x, a) // a = a0

	// a = x.A0 - x.A1
	fq2.Sub(x, a, fq2.NbWords)
	fq2.reduceAfterSubNoJumpScratch(tr, a, b)
	// b = x.A0 + x.A1
	fq2.Mov(x, b, fq2.NbWords) // b = a1
	fq2.Add(x, b)

	fq2.MOVQ("res+0(FP)", tr)
	fq2.Mov(a, tr)
	fq2.ReduceElement(b, a)
	fq2.Mov(b, tr, 0, fq2.NbWords)

	fq2.RET()
}

func (fq2 *Fq2Amd64) generateSquareE2BLS381() {
	// // Square sets z to the E2-product of x,x returns z
	// func (z *E2) Square(x *E2) *E2 {
	// 	// algo 22 https://eprint.iacr.org/2010/354.pdf
	// 	var a, b fp.Element
	// 	a.Add(&x.A0, &x.A1)
	// 	b.Sub(&x.A0, &x.A1)
	// 	a.Mul(&a, &b)
	// 	b.Mul(&x.A0, &x.A1).Double(&b)
	// 	z.A0.Set(&a)
	// 	z.A1.Set(&b)
	// 	return z
	// }
	const minStackSize = 16
	stackSize := fq2.StackSize(fq2.NbWords*3, 2, minStackSize)
	registers := fq2.FnHeader("squareAdxE2", stackSize, 16, amd64.DX, amd64.AX)
	defer fq2.AssertCleanStack(stackSize, minStackSize)
	fq2.WriteLn("NO_LOCAL_POINTERS")

	fq2.WriteLn(`
	// z.A0 = (x.A0 + x.A1) * (x.A0 - x.A1)
	// z.A1 = 2 * x.A0 * x.A1
	`)

	noAdx := fq2.NewLabel()
	// check ADX instruction support
	fq2.CMPB("·supportAdx(SB)", 1)
	fq2.JNE(noAdx)

	// used in the mul operation
	op1 := registers.PopN(fq2.NbWords)
	res := registers.PopN(fq2.NbWords)

	xat := func(i int) string {
		return string(op1[i])
	}

	ax := amd64.AX
	dx := amd64.DX

	// b = a0 * a1 * 2

	fq2.Comment("2 * x.A0 * x.A1")
	fq2.MOVQ("x+8(FP)", ax)

	fq2.LabelRegisters("2 * x.A1", op1...)
	fq2.Mov(ax, op1, fq2.NbWords)
	fq2.Add(op1, op1) // op1, no reduce

	fq2.MulADX(&registers, xat, func(i int) string {
		fq2.MOVQ("x+8(FP)", dx)
		return dx.At(i)
	}, res)
	fq2.ReduceElement(res, op1)

	fq2.MOVQ("x+8(FP)", ax)

	fq2.LabelRegisters("x.A1", op1...)
	fq2.Mov(ax, op1, fq2.NbWords)

	fq2.MOVQ("res+0(FP)", dx)
	fq2.Mov(res, dx, 0, fq2.NbWords)
	fq2.Mov(op1, res)

	// op1 and res both contains x.A1 at this point
	// res+0(FP) (z.A1) must not be referenced.

	// a = a0 + a1
	fq2.Comment("Add(&x.A0, &x.A1)")
	fq2.Add(ax, op1)
	//--> must save on stack
	a0a1 := fq2.PopN(&registers, true)
	fq2.Mov(op1, a0a1)

	zero := amd64.BP
	fq2.XORQ(zero, zero)

	// b = a0 - a1
	fq2.Comment("Sub(&x.A0, &x.A1)")
	fq2.Mov(ax, op1)
	fq2.Sub(res, op1)
	fq2.reduceAfterSubNoJumpScratch(zero, op1, res) // using res as scratch registers

	// a = a * b
	fq2.MulADX(&registers, xat, func(i int) string { return string(a0a1[i]) }, res)
	fq2.ReduceElement(res, op1)

	fq2.MOVQ("res+0(FP)", ax)
	fq2.Mov(res, ax)

	// result.a0 = a
	fq2.RET()

	// No adx
	fq2.LABEL(noAdx)
	fq2.MOVQ("res+0(FP)", amd64.AX)
	fq2.MOVQ(amd64.AX, "(SP)")
	fq2.MOVQ("x+8(FP)", amd64.AX)
	fq2.MOVQ(amd64.AX, "8(SP)")
	fq2.WriteLn("CALL ·squareGenericE2(SB)")
	fq2.RET()

	fq2.Push(&registers, a0a1...)
}

func (fq2 *Fq2Amd64) generateMulE2BLS381() {
	// var a, b, c fp.Element
	// a.Add(&x.A0, &x.A1)
	// b.Add(&y.A0, &y.A1)
	// a.Mul(&a, &b)
	// b.Mul(&x.A0, &y.A0)
	// c.Mul(&x.A1, &y.A1)
	// z.A1.Sub(&a, &b).Sub(&z.A1, &c)
	// z.A0.Sub(&b, &c)

	// we need a bit of stack space to store the results of the xA0yA0 and xA1yA1 multiplications
	_ = fq2.FnHeader("mulAdxE2", 3*8*fq2.NbWords+8, 24, amd64.DX, amd64.AX)
	fq2.WriteLn("NO_LOCAL_POINTERS")
	noAdx := fq2.NewLabel()
	fq2.JMP(noAdx)
	// // check ADX instruction support
	// fq2.CMPB("·supportAdx(SB)", 1)
	// fq2.JNE(noAdx)

	// // first: let's name our local variables on the stack
	// xA0yA0 := make([]string, fq2.NbWords) // Mul(&x.A0, &y.A0)
	// xA1yA1 := make([]string, fq2.NbWords) // Mul(&x.A1, &y.A1)
	// xA0xA1 := make([]string, fq2.NbWords) // Add(&x.A0, &x.A1)

	// // we use one push/pop in the mul, need some extra space, hence +16
	// offsetSP := 16
	// for i := 0; i < fq2.NbWords; i++ {
	// 	xA0yA0[i] = fmt.Sprintf("-%d(SP)", offsetSP)
	// 	offsetSP += 8
	// }
	// for i := 0; i < fq2.NbWords; i++ {
	// 	xA1yA1[i] = fmt.Sprintf("-%d(SP)", offsetSP)
	// 	offsetSP += 8
	// }
	// for i := 0; i < fq2.NbWords; i++ {
	// 	xA0xA1[i] = fmt.Sprintf("-%d(SP)", offsetSP)
	// 	offsetSP += 8
	// }

	// t := registers.PopN(fq2.NbWords)
	// x := amd64.AX
	// y := amd64.DX
	// fq2.MOVQ("x+8(FP)", x)
	// fq2.Mov(x, t)

	// {
	// 	// Mul(&x.A0, &y.A0)
	// 	yat := func(i int) string {
	// 		ry := amd64.DX
	// 		fq2.MOVQ("y+16(FP)", ry)
	// 		return ry.At(i)
	// 	}
	// 	xat := func(i int) string {
	// 		return string(t[i])
	// 	}
	// 	tr := fq2.MulADX(&registers, yat, xat, nil)
	// 	registers.Push(t...)
	// 	fq2.Reduce(&registers, tr)
	// 	t = registers.PopN(fq2.NbWords)
	// 	// save our registers
	// 	for i := 0; i < fq2.NbWords; i++ {
	// 		fq2.MOVQ(tr[i], xA0yA0[i])
	// 	}
	// 	registers.Push(tr...)
	// }

	// fq2.MOVQ("x+8(FP)", x)
	// fq2.Mov(x, t, fq2.NbWords)
	// {
	// 	// Mul(&x.A1, &y.A1)
	// 	yat := func(i int) string {
	// 		ry := amd64.DX
	// 		fq2.MOVQ("y+16(FP)", ry)
	// 		return ry.At(i + fq2.NbWords)
	// 	}
	// 	xat := func(i int) string {
	// 		return string(t[i])
	// 	}
	// 	tr := fq2.MulADX(&registers, yat, xat, nil)
	// 	registers.Push(t...)
	// 	fq2.Reduce(&registers, tr)
	// 	t = registers.PopN(fq2.NbWords)
	// 	// save our registers
	// 	for i := 0; i < fq2.NbWords; i++ {
	// 		fq2.MOVQ(tr[i], xA1yA1[i])
	// 	}
	// 	registers.Push(tr...)
	// }

	// fq2.MOVQ("x+8(FP)", x)
	// fq2.Mov(x, t)
	// fq2.Add(x, t, fq2.NbWords)
	// fq2.Reduce(&registers, t)

	// // save our registers
	// for i := 0; i < len(t); i++ {
	// 	fq2.MOVQ(t[i], xA0xA1[i])
	// }

	// fq2.MOVQ("y+16(FP)", y)
	// fq2.Mov(y, t)
	// fq2.Add(y, t, fq2.NbWords)
	// fq2.Reduce(&registers, t)

	// {
	// 	yat := func(i int) string {
	// 		return string(xA0xA1[i])
	// 	}
	// 	xat := func(i int) string {
	// 		return string(t[i])
	// 	}
	// 	tR := fq2.MulADX(&registers, yat, xat, nil)
	// 	registers.Push(t...)
	// 	t = tR
	// 	fq2.Reduce(&registers, t)
	// }

	// z := amd64.DX
	// fq2.MOVQ("z+0(FP)", z)
	// // z.A1.Sub(&a, &b).Sub(&z.A1, &c)

	// for i := 0; i < fq2.NbWords; i++ {
	// 	if i == 0 {
	// 		fq2.SUBQ(xA0yA0[i], t[i])
	// 	} else {
	// 		fq2.SBBQ(xA0yA0[i], t[i])
	// 	}
	// }
	// fq2.ReduceAfterSub(&registers, t, true)
	// for i := 0; i < fq2.NbWords; i++ {
	// 	if i == 0 {
	// 		fq2.SUBQ(xA1yA1[i], t[i])
	// 	} else {
	// 		fq2.SBBQ(xA1yA1[i], t[i])
	// 	}
	// }
	// fq2.ReduceAfterSub(&registers, t, true)

	// fq2.Mov(t, z, 0, fq2.NbWords)

	// // z.A0.Sub(&b, &c)
	// for i := 0; i < fq2.NbWords; i++ {
	// 	fq2.MOVQ(xA0yA0[i], t[i])
	// }

	// for i := 0; i < fq2.NbWords; i++ {
	// 	if i == 0 {
	// 		fq2.SUBQ(xA1yA1[i], t[i])
	// 	} else {
	// 		fq2.SBBQ(xA1yA1[i], t[i])
	// 	}
	// }
	// fq2.ReduceAfterSub(&registers, t, true)
	// fq2.Mov(t, z)
	// fq2.RET()

	// No adx
	fq2.LABEL(noAdx)
	fq2.MOVQ("z+0(FP)", amd64.AX)
	fq2.MOVQ(amd64.AX, "(SP)")
	fq2.MOVQ("x+8(FP)", amd64.AX)
	fq2.MOVQ(amd64.AX, "8(SP)")
	fq2.MOVQ("y+16(FP)", amd64.AX)
	fq2.MOVQ(amd64.AX, "16(SP)")
	fq2.WriteLn("CALL ·mulGenericE2(SB)")
	fq2.RET()

}
