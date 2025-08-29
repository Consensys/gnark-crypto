// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"

	"github.com/consensys/bavard/amd64"
)

// Butterfly sets
// a = a + b
// b = a - b
//
//	func Butterfly(a, b *{{.ElementName}}) {
//		t := *a
//		a.Add(a, b)
//		b.Sub(&t, b)
//	}
func (f *FFAmd64) generateButterfly() {
	f.Comment("Butterfly(a, b *Element) sets a = a + b; b = a - b")

	nbRegisters := f.NbWords*3 + 1
	if f.NbWords > 6 {
		nbRegisters = 2*f.NbWords + 1
	}
	stackSize := f.StackSize(nbRegisters, 0, 0)
	registers := f.FnHeader("Butterfly", stackSize, 16)
	defer f.AssertCleanStack(stackSize, 0)

	if f.NbWords <= 6 {
		// registers.UnsafePush(amd64.R15)
		// registers
		a := amd64.R15 //f.Pop(&registers)
		b := f.Pop(&registers)
		t0 := f.PopN(&registers)
		t1 := f.PopN(&registers)
		q := f.PopN(&registers)

		// t = a
		f.MOVQ("a+0(FP)", a)
		f.Mov(a, t0)
		f.Mov(t0, t1)
		f.XORQ(a, a) // set a to zero for later reduction

		f.MOVQ("b+8(FP)", b)
		f.Add(b, t0) // t0 = a + b
		f.Sub(b, t1) // t1 = a - b

		// reduce t1
		if f.NbWords >= 5 {
			// q is on the stack, can't use for CMOVQCC
			f.Mov(t0, q) // save t0
			for i := 0; i < f.NbWords; i++ {
				f.MOVQ(fmt.Sprintf("$const_q%d", i), t0[i])
			}
			for i := 0; i < f.NbWords; i++ {
				f.CMOVQCC(a, t0[i])
			}
			// add registers (q or 0) to t, and set to result
			f.Add(t0, t1)
			f.Mov(q, t0) // restore t0
		} else {
			for i := 0; i < f.NbWords; i++ {
				f.MOVQ(fmt.Sprintf("$const_q%d", i), q[i])
			}
			for i := 0; i < f.NbWords; i++ {
				f.CMOVQCC(a, q[i])
			}
			// add registers (q or 0) to t, and set to result
			f.Add(q, t1)
		}

		f.UnsafePush(&registers, q...)

		// save t1
		f.Mov(t1, b)

		// reduce t0
		f.ReduceElement(t0, t1, false)

		// save t0
		f.MOVQ("a+0(FP)", a)
		f.Mov(t0, a)

		f.RET()

		f.UnsafePush(&registers, t0...)
		f.UnsafePush(&registers, t1...)
		f.UnsafePush(&registers, a, b)
	} else {
		// registers
		r := f.Pop(&registers)
		t0 := f.PopN(&registers)
		t1 := f.PopN(&registers)

		// t = a
		f.MOVQ("b+8(FP)", r)
		f.Mov(r, t0)

		f.MOVQ("a+0(FP)", r)
		f.Add(r, t0)  // t0 = a + b
		f.Mov(t0, t1) // save t1 = t0
		f.Mov(r, t0)  // t0 = a
		f.MOVQ("b+8(FP)", r)
		f.Sub(r, t0) // t0 = a - b

		// reduce t0
		noReduce := f.NewLabel("noReduce")
		f.JCC(noReduce)
		q := r
		f.MOVQ("$const_q0", q)

		f.ADDQ(q, t0[0])
		for i := 1; i < f.NbWords; i++ {
			f.MOVQ(fmt.Sprintf("$const_q%d", i), q)
			f.ADCQ(q, t0[i])
		}
		f.LABEL(noReduce)

		// save t1
		f.MOVQ("b+8(FP)", r)
		f.Mov(t0, r)

		// reduce t0
		f.Mov(t1, t0)
		f.ReduceElement(t0, t1, false)

		// save t0
		f.MOVQ("a+0(FP)", r)
		f.Mov(t0, r)

		f.RET()

		f.Push(&registers, t0...)
		f.Push(&registers, t1...)
		f.Push(&registers, r)
	}

}
