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
	x := registers.Pop()

	fq2.MOVQ("x+8(FP)", x)
	fq2.Mov(x, a) // a = a0

	// a = x.A0 - x.A1
	fq2.Sub(x, a, fq2.NbWords)
	fq2.reduceAfterSubWithJump(&registers, a)

	// b = x.A0 + x.A1
	b := registers.PopN(fq2.NbWords)
	fq2.Mov(x, b, fq2.NbWords) // b = a1
	fq2.Add(x, b)

	fq2.MOVQ("res+0(FP)", x)
	fq2.Mov(a, x)
	registers.Push(a...)
	fq2.Reduce(&registers, b)
	fq2.Mov(b, x, 0, fq2.NbWords)

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
	noAdx := fq2.NewLabel()
	// check ADX instruction support
	fq2.CMPB("·supportAdx(SB)", 1)
	fq2.JNE(noAdx)

	a0SP := fq2.PopN(&registers, true)
	// a0SP := make([]string, fq2.NbWords)
	// offsetSP := 16 // we use one push/pop in the mul, need some extra space.
	// for i := 0; i < fq2.NbWords; i++ {
	// 	a0SP[i] = fmt.Sprintf("-%d(SP)", offsetSP)
	// 	offsetSP += 8
	// }

	var a []amd64.Register
	{
		x := amd64.DX
		fq2.MOVQ("x+8(FP)", x)

		a0 := registers.PopN(fq2.NbWords)

		fq2.Mov(x, a0) // a0

		zero := amd64.BP
		fq2.XORQ(zero, zero)
		// a0 = a0 - a1
		fq2.Sub(x, a0, fq2.NbWords)
		fq2.reduceAfterSubNoJump(&registers, zero, a0)
		for i := 0; i < len(a0); i++ {
			// fq2.PUSHQ(a0[i])
			fq2.MOVQ(a0[i], a0SP[i])
		}

		a1 := registers.PopN(fq2.NbWords)
		a = a0
		fq2.Mov(x, a)               // a = a0
		fq2.Mov(x, a1, fq2.NbWords) // a = a0

		// a = a0 + a1
		fq2.Add(a1, a)
		registers.Push(a1...)

		fq2.Reduce(&registers, a)
	}

	// a = a * b
	{
		a10 := amd64.BP
		fq2.MOVQ(a0SP[0], a10)
		fq2.MOVQ(a[0], a0SP[0])
		registers.Push(a[0])
		yat := func(i int) string {
			if i == 0 {
				return string(a10)
			}
			return string(a0SP[i])
		}
		xat := func(i int) string {
			if i == 0 {
				return string(a0SP[0])
			}
			return string(a[i])
		}
		t := registers.PopN(fq2.NbWords)
		fq2.MulADX(&registers, xat, yat, t)
		registers.Push(a[1:]...)
		fq2.Push(&registers, a0SP...)
		a = t
		fq2.Reduce(&registers, a)
	}

	// // result.a1 = b
	r := amd64.DX
	fq2.MOVQ("res+0(FP)", r)

	// we need to save x.A0 in case z == x
	{
		x := amd64.BP
		fq2.MOVQ("x+8(FP)", amd64.BP)
		b := registers.PopN(fq2.NbWords)
		fq2.Mov(x, b) // b = a0
		// registers.Push(x)
		// result.a0 = a
		fq2.Mov(a, r)
		registers.Push(a...)

		// b = a0 * a1 * 2
		yat := func(i int) string {
			ry := amd64.DX
			fq2.MOVQ("x+8(FP)", ry)
			return ry.At(i + fq2.NbWords)
		}

		b00 := fq2.Pop(&registers, true)
		fq2.MOVQ(b[0], b00)
		registers.Push(b[0])
		xat := func(i int) string {
			if i == 0 {
				return string(b00)
			}
			return string(b[i])
		}
		t := registers.PopN(fq2.NbWords)
		fq2.MulADX(&registers, xat, yat, t)
		fq2.Push(&registers, b00)
		registers.Push(b[1:]...)
		// reduce b
		fq2.Reduce(&registers, t)

		// double b (no reduction)
		fq2.Add(t, t)

		// result.a1 = b
		r = amd64.DX

		fq2.Reduce(&registers, t)
		fq2.MOVQ("res+0(FP)", r)
		fq2.Mov(t, r, 0, fq2.NbWords)
	}

	fq2.RET()

	// No adx
	fq2.LABEL(noAdx)
	fq2.MOVQ("res+0(FP)", amd64.AX)
	fq2.MOVQ(amd64.AX, "(SP)")
	fq2.MOVQ("x+8(FP)", amd64.AX)
	fq2.MOVQ(amd64.AX, "8(SP)")
	fq2.WriteLn("CALL ·squareGenericE2(SB)")
	fq2.RET()
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
