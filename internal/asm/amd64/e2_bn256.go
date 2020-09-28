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
	"strings"

	. "github.com/consensys/bavard/amd64"
)

func (fq2 *Fq2Amd64) generateMulByNonResidueE2BN256() {
	// 	var a, b fp.Element
	// 	a.Double(&x.A0).Double(&a).Double(&a).fq2.F.Add(&a, &x.A0).fq2.F.Sub(&a, &x.A1)
	// 	b.Double(&x.A1).Double(&b).Double(&b).fq2.F.Add(&b, &x.A1).fq2.F.Add(&b, &x.A0)
	// 	z.A0.Set(&a)
	// 	z.A1.Set(&b)
	registers := FnHeader("mulNonRes"+strings.ToUpper(fq2.F.ElementName), 0, 16)

	a := registers.PopN(fq2.F.NbWords)
	b := registers.PopN(fq2.F.NbWords)
	x := registers.Pop()

	MOVQ("x+8(FP)", x)
	fq2.F.Mov(x, a) // a = a0

	fq2.F.Add(a, a)
	fq2.F.Reduce(&registers, a, a)

	fq2.F.Add(a, a)
	fq2.F.Reduce(&registers, a, a)

	fq2.F.Add(a, a)
	fq2.F.Reduce(&registers, a, a)

	fq2.F.Add(x, a)
	fq2.F.Reduce(&registers, a, a)

	fq2.F.Mov(x, b, fq2.F.NbWords) // b = a1
	fq2.F.Sub(b, a)
	fq2.F.ReduceAfterSub(&registers, a, true)

	fq2.F.Add(b, b)
	fq2.F.Reduce(&registers, b, b)

	fq2.F.Add(b, b)
	fq2.F.Reduce(&registers, b, b)

	fq2.F.Add(b, b)
	fq2.F.Reduce(&registers, b, b)

	fq2.F.Add(x, b, fq2.F.NbWords)
	fq2.F.Reduce(&registers, b, b)
	fq2.F.Add(x, b)
	fq2.F.Reduce(&registers, b, b)

	MOVQ("res+0(FP)", x)
	fq2.F.Mov(a, x)
	fq2.F.Mov(b, x, 0, fq2.F.NbWords)

	RET()
}

func (fq2 *Fq2Amd64) generateSquareE2BN256() {
	// var a, b fp.Element
	// a.fq2.F.Add(&x.A0, &x.A1)
	// b.fq2.F.Sub(&x.A0, &x.A1)
	// a.Mul(&a, &b)
	// b.Mul(&x.A0, &x.A1).Double(&b)
	// z.A0.Set(&a)
	// z.A1.Set(&b)
	registers := FnHeader("squareAdx"+strings.ToUpper(fq2.F.ElementName), 16, 16, DX, AX)

	noAdx := NewLabel()
	// check ADX instruction support
	CMPB("·supportAdx(SB)", 1)
	JNE(noAdx)

	a := registers.PopN(fq2.F.NbWords)
	b := registers.PopN(fq2.F.NbWords)
	{
		x := registers.Pop()

		MOVQ("x+8(FP)", x)
		fq2.F.Mov(x, a, fq2.F.NbWords) // a = a1
		fq2.F.Mov(x, b)                // b = a0

		// a = a0 + a1
		fq2.F.Add(b, a)
		fq2.F.Reduce(&registers, a, a)

		// b = a0 - a1
		fq2.F.Sub(x, b, fq2.F.NbWords)
		registers.Push(x)
		fq2.F.ReduceAfterSub(&registers, b, true)
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
		t := fq2.F.MulADX(&registers, yat, xat, uglyHook)
		fq2.F.Reduce(&registers, t, a)

		registers.Push(t...)
	}

	// b = a0 * a1 * 2
	{
		r := registers.Pop()
		MOVQ("x+8(FP)", r)
		yat := func(i int) string {
			return r.At(i + fq2.F.NbWords)
		}
		xat := func(i int) string {
			return r.At(i)
		}
		b = fq2.F.MulADX(&registers, yat, xat, nil)
		registers.Push(r)

		// reduce b
		fq2.F.Reduce(&registers, b, b)

		// double b (no reduction)
		fq2.F.Add(b, b)
	}

	// result.a1 = b
	r := registers.Pop()
	MOVQ("res+0(FP)", r)
	fq2.F.Reduce(&registers, b, r, fq2.F.NbWords)

	// result.a0 = a
	fq2.F.Mov(a, r)

	RET()

	// No adx
	LABEL(noAdx)
	MOVQ("res+0(FP)", AX)
	MOVQ(AX, "(SP)")
	MOVQ("x+8(FP)", AX)
	MOVQ(AX, "8(SP)")
	WriteLn("CALL ·squareGenericE2(SB)")
	RET()
}

func (fq2 *Fq2Amd64) generateMulE2BN256() {
	// var a, b, c fp.Element
	// a.fq2.F.Add(&x.A0, &x.A1)
	// b.fq2.F.Add(&y.A0, &y.A1)
	// a.Mul(&a, &b)
	// b.Mul(&x.A0, &y.A0)
	// c.Mul(&x.A1, &y.A1)
	// z.A1.fq2.F.Sub(&a, &b).fq2.F.Sub(&z.A1, &c)
	// z.A0.fq2.F.Sub(&b, &c)
	registers := FnHeader("mulAdx"+strings.ToUpper(fq2.F.ElementName), 24, 24, DX, AX)

	noAdx := NewLabel()
	// check ADX instruction support
	CMPB("·supportAdx(SB)", 1)
	JNE(noAdx)

	a := registers.PopN(fq2.F.NbWords)
	b := registers.PopN(fq2.F.NbWords)
	{
		x := registers.Pop()

		MOVQ("x+8(FP)", x)

		fq2.F.Mov(x, a, fq2.F.NbWords) // a = x.a1
		fq2.F.Add(x, a)                // a = x.a0 + x.a1
		fq2.F.Reduce(&registers, a, a)

		MOVQ("y+16(FP)", x)
		fq2.F.Mov(x, b, fq2.F.NbWords) // b = y.a1
		fq2.F.Add(x, b)                // b = y.a0 + y.a1
		fq2.F.Reduce(&registers, b, b)

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
		t := fq2.F.MulADX(&registers, yat, xat, uglyHook)
		fq2.F.Reduce(&registers, t, a)

		registers.Push(t...)
	}

	// b = x.A0 * y.AO
	{
		r := registers.Pop()
		yat := func(i int) string {
			MOVQ("y+16(FP)", r)
			return r.At(i)
		}
		xat := func(i int) string {
			MOVQ("x+8(FP)", r)
			return r.At(i)
		}
		b = fq2.F.MulADX(&registers, yat, xat, nil)
		registers.Push(r)
		fq2.F.Reduce(&registers, b, b)
	}
	// a - = b
	fq2.F.Sub(b, a)
	fq2.F.ReduceAfterSub(&registers, a, true)

	// push a to the stack for later use
	for i := 0; i < fq2.F.NbWords; i++ {
		PUSHQ(a[i])
	}
	registers.Push(a...)

	var c []Register
	// c = x.A1 * y.A1
	{
		r := registers.Pop()
		yat := func(i int) string {
			MOVQ("y+16(FP)", r)
			return r.At(i + fq2.F.NbWords)
		}
		xat := func(i int) string {
			MOVQ("x+8(FP)", r)
			return r.At(i + fq2.F.NbWords)
		}
		c = fq2.F.MulADX(&registers, yat, xat, nil)
		registers.Push(r)
		fq2.F.Reduce(&registers, c, c)
	}

	// b = b - c
	fq2.F.Sub(c, b)
	fq2.F.ReduceAfterSub(&registers, b, true)

	// dereference result
	r := registers.Pop()
	MOVQ("res+0(FP)", r)

	// z.A0 = b
	fq2.F.Mov(b, r)

	// restore a
	a = b
	for i := fq2.F.NbWords - 1; i >= 0; i-- {
		POPQ(a[i])
	}

	// a = a - c
	fq2.F.Sub(c, a)
	registers.Push(c...)

	// reduce a
	fq2.F.ReduceAfterSub(&registers, a, true)

	// z.A1 = a
	fq2.F.Mov(a, r, 0, fq2.F.NbWords)

	RET()

	// No adx
	LABEL(noAdx)
	MOVQ("res+0(FP)", AX)
	MOVQ(AX, "(SP)")
	MOVQ("x+8(FP)", AX)
	MOVQ(AX, "8(SP)")
	MOVQ("y+16(FP)", AX)
	MOVQ(AX, "16(SP)")
	WriteLn("CALL ·mulGenericE2(SB)")
	RET()
}
