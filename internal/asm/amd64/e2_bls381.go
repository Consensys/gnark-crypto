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
	fq2.ReduceAfterSub(&registers, a, true)

	// b = x.A0 + x.A1
	b := registers.PopN(fq2.NbWords)
	fq2.Mov(x, b, fq2.NbWords) // b = a1
	fq2.Add(x, b)

	fq2.MOVQ("res+0(FP)", x)
	fq2.Mov(a, x)
	registers.Push(a...)
	fq2.Reduce(&registers, b, b)

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
	registers := fq2.FnHeader("squareAdxE2", 16, 16, amd64.DX, amd64.AX)

	noAdx := fq2.NewLabel()
	// check ADX instruction support
	fq2.CMPB("·supportAdx(SB)", 1)
	fq2.JNE(noAdx)

	var a []amd64.Register
	{
		x := amd64.DX
		fq2.MOVQ("x+8(FP)", x)

		a0 := registers.PopN(fq2.NbWords)

		fq2.Mov(x, a0) // a0

		// a0 = a0 - a1
		fq2.Sub(x, a0, fq2.NbWords)
		fq2.ReduceAfterSub(&registers, a0, true)
		for i := len(a0) - 1; i >= 0; i-- {
			fq2.PUSHQ(a0[i])
		}

		a1 := registers.PopN(fq2.NbWords)
		a = a0
		fq2.Mov(x, a)               // a = a0
		fq2.Mov(x, a1, fq2.NbWords) // a = a0

		// a = a0 + a1
		fq2.Add(a1, a)
		registers.Push(a1...)

		fq2.Reduce(&registers, a, a)
	}

	// a = a * b
	{
		xat := func(i int) string {
			return string(a[i])
		}
		// dirty: yat = nil --> will POPQ() from stack the values for y[i]
		t := fq2.MulADX(&registers, nil, xat, nil)
		registers.Push(a...)
		a = t
		fq2.Reduce(&registers, a, a)
	}

	// // result.a1 = b
	r := amd64.DX
	fq2.MOVQ("res+0(FP)", r)

	// we need to save x.A0 in case z == x
	{
		x := registers.Pop()
		fq2.MOVQ("x+8(FP)", x)
		b := registers.PopN(fq2.NbWords)
		fq2.Mov(x, b) // b = a0
		registers.Push(x)
		// result.a0 = a
		fq2.Mov(a, r)
		registers.Push(a...)

		// b = a0 * a1 * 2
		yat := func(i int) string {
			ry := amd64.DX
			fq2.MOVQ("x+8(FP)", ry)
			return ry.At(i + fq2.NbWords)
		}
		xat := func(i int) string {
			return string(b[i])
		}
		t := fq2.MulADX(&registers, yat, xat, nil)

		registers.Push(b...)
		// reduce b
		fq2.Reduce(&registers, t, t)

		// double b (no reduction)
		fq2.Add(t, t)

		// result.a1 = b
		r = amd64.DX
		fq2.MOVQ("res+0(FP)", r)
		fq2.Reduce(&registers, t, r, fq2.NbWords)
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
