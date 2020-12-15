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

import "github.com/consensys/bavard/amd64"

func (fq2 *Fq2Amd64) generateMulByNonResidueE2BN256() {
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
	fq2.Reduce(&registers, a, a)

	fq2.Add(a, a)
	fq2.Reduce(&registers, a, a)

	fq2.Add(a, a)
	fq2.Reduce(&registers, a, a)

	fq2.Add(x, a)
	fq2.Reduce(&registers, a, a)

	fq2.Mov(x, b, fq2.NbWords) // b = a1
	fq2.Sub(b, a)
	fq2.ReduceAfterSub(&registers, a, true)

	fq2.Add(b, b)
	fq2.Reduce(&registers, b, b)

	fq2.Add(b, b)
	fq2.Reduce(&registers, b, b)

	fq2.Add(b, b)
	fq2.Reduce(&registers, b, b)

	fq2.Add(x, b, fq2.NbWords)
	fq2.Reduce(&registers, b, b)
	fq2.Add(x, b)
	fq2.Reduce(&registers, b, b)

	fq2.MOVQ("res+0(FP)", x)
	fq2.Mov(a, x)
	fq2.Mov(b, x, 0, fq2.NbWords)

	fq2.RET()
}

func (fq2 *Fq2Amd64) generateSquareE2BN256() {
	// var a, b fp.Element
	// a.fq2.Add(&x.A0, &x.A1)
	// b.fq2.Sub(&x.A0, &x.A1)
	// a.Mul(&a, &b)
	// b.Mul(&x.A0, &x.A1).Double(&b)
	// z.A0.Set(&a)
	// z.A1.Set(&b)
	registers := fq2.FnHeader("squareAdxE2", 16, 16, amd64.DX, amd64.AX)
	fq2.WriteLn("NO_LOCAL_POINTERS")
	noAdx := fq2.NewLabel()
	// check ADX instruction support
	fq2.CMPB("·supportAdx(SB)", 1)
	fq2.JNE(noAdx)

	a := registers.PopN(fq2.NbWords)
	b := registers.PopN(fq2.NbWords)
	{
		x := registers.Pop()

		fq2.MOVQ("x+8(FP)", x)
		fq2.Mov(x, a, fq2.NbWords) // a = a1
		fq2.Mov(x, b)              // b = a0

		// a = a0 + a1
		fq2.Add(b, a)
		fq2.Reduce(&registers, a, a)

		// b = a0 - a1
		fq2.Sub(x, b, fq2.NbWords)
		registers.Push(x)
		fq2.ReduceAfterSub(&registers, b, true)
	}

	// a = a * b
	{
		yat := func(i int) string {
			return string(b[i])
		}
		xat := func(i int) string {
			return string(a[i])
		}
		uglyHook := func(i int) {
			registers.Push(b[i])
		}
		t := fq2.MulADX(&registers, yat, xat, uglyHook)
		fq2.Reduce(&registers, t, a)

		registers.Push(t...)
	}

	// b = a0 * a1 * 2
	{
		r := registers.Pop()
		fq2.MOVQ("x+8(FP)", r)
		yat := func(i int) string {
			return r.At(i + fq2.NbWords)
		}
		xat := func(i int) string {
			return r.At(i)
		}
		b = fq2.MulADX(&registers, yat, xat, nil)
		registers.Push(r)

		// reduce b
		fq2.Reduce(&registers, b, b)

		// double b (no reduction)
		fq2.Add(b, b)
	}

	// result.a1 = b
	r := registers.Pop()
	fq2.MOVQ("res+0(FP)", r)
	fq2.Reduce(&registers, b, r, fq2.NbWords)

	// result.a0 = a
	fq2.Mov(a, r)

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

func (fq2 *Fq2Amd64) generateMulE2BN256() {
	// var a, b, c fp.Element
	// a.fq2.Add(&x.A0, &x.A1)
	// b.fq2.Add(&y.A0, &y.A1)
	// a.Mul(&a, &b)
	// b.Mul(&x.A0, &y.A0)
	// c.Mul(&x.A1, &y.A1)
	// z.A1.fq2.Sub(&a, &b).fq2.Sub(&z.A1, &c)
	// z.A0.fq2.Sub(&b, &c)
	registers := fq2.FnHeader("mulAdxE2", 24, 24, amd64.DX, amd64.AX)
	fq2.WriteLn("NO_LOCAL_POINTERS")
	noAdx := fq2.NewLabel()
	// check ADX instruction support
	fq2.CMPB("·supportAdx(SB)", 1)
	fq2.JNE(noAdx)

	a := registers.PopN(fq2.NbWords)
	b := registers.PopN(fq2.NbWords)
	{
		x := registers.Pop()

		fq2.MOVQ("x+8(FP)", x)

		fq2.Mov(x, a, fq2.NbWords) // a = x.a1
		fq2.Add(x, a)              // a = x.a0 + x.a1
		fq2.Reduce(&registers, a, a)

		fq2.MOVQ("y+16(FP)", x)
		fq2.Mov(x, b, fq2.NbWords) // b = y.a1
		fq2.Add(x, b)              // b = y.a0 + y.a1
		fq2.Reduce(&registers, b, b)

		registers.Push(x)
	}

	// a = a * b
	{
		yat := func(i int) string {
			return string(b[i])
		}
		xat := func(i int) string {
			return string(a[i])
		}
		uglyHook := func(i int) {
			registers.Push(b[i])
		}
		t := fq2.MulADX(&registers, yat, xat, uglyHook)
		fq2.Reduce(&registers, t, a)

		registers.Push(t...)
	}

	// b = x.A0 * y.AO
	{
		r := registers.Pop()
		yat := func(i int) string {
			fq2.MOVQ("y+16(FP)", r)
			return r.At(i)
		}
		xat := func(i int) string {
			fq2.MOVQ("x+8(FP)", r)
			return r.At(i)
		}
		b = fq2.MulADX(&registers, yat, xat, nil)
		registers.Push(r)
		fq2.Reduce(&registers, b, b)
	}
	// a - = b
	fq2.Sub(b, a)
	fq2.ReduceAfterSub(&registers, a, true)

	// push a to the stack for later use
	for i := 0; i < fq2.NbWords; i++ {
		fq2.PUSHQ(a[i])
	}
	registers.Push(a...)

	var c []amd64.Register
	// c = x.A1 * y.A1
	{
		r := registers.Pop()
		yat := func(i int) string {
			fq2.MOVQ("y+16(FP)", r)
			return r.At(i + fq2.NbWords)
		}
		xat := func(i int) string {
			fq2.MOVQ("x+8(FP)", r)
			return r.At(i + fq2.NbWords)
		}
		c = fq2.MulADX(&registers, yat, xat, nil)
		registers.Push(r)
		fq2.Reduce(&registers, c, c)
	}

	// b = b - c
	fq2.Sub(c, b)
	fq2.ReduceAfterSub(&registers, b, true)

	// dereference result
	r := registers.Pop()
	fq2.MOVQ("res+0(FP)", r)

	// z.A0 = b
	fq2.Mov(b, r)

	// restore a
	a = b
	for i := fq2.NbWords - 1; i >= 0; i-- {
		fq2.POPQ(a[i])
	}

	// a = a - c
	fq2.Sub(c, a)
	registers.Push(c...)

	// reduce a
	fq2.ReduceAfterSub(&registers, a, true)

	// z.A1 = a
	fq2.Mov(a, r, 0, fq2.NbWords)

	fq2.RET()

	// No adx
	fq2.LABEL(noAdx)
	fq2.MOVQ("res+0(FP)", amd64.AX)
	fq2.MOVQ(amd64.AX, "(SP)")
	fq2.MOVQ("x+8(FP)", amd64.AX)
	fq2.MOVQ(amd64.AX, "8(SP)")
	fq2.MOVQ("y+16(FP)", amd64.AX)
	fq2.MOVQ(amd64.AX, "16(SP)")
	fq2.WriteLn("CALL ·mulGenericE2(SB)")
	fq2.RET()
}
