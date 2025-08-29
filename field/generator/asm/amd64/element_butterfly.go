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
func (_f *FFAmd64) generateButterfly() {
	_f.Comment("Butterfly(a, b *Element) sets a = a + b; b = a - b")

	nbRegisters := _f.NbWords*3 + 1
	if _f.NbWords > 6 {
		nbRegisters = 2*_f.NbWords + 1
	}
	stackSize := _f.StackSize(nbRegisters, 0, 0)
	registers := _f.FnHeader("Butterfly", stackSize, 16)
	defer _f.AssertCleanStack(stackSize, 0)

	if _f.NbWords <= 6 {
		// registers.UnsafePush(amd64.R15)
		// registers
		a := amd64.R15 //f.Pop(&registers)
		b := _f.Pop(&registers)
		t0 := _f.PopN(&registers)
		t1 := _f.PopN(&registers)
		q := _f.PopN(&registers)

		// t = a
		_f.MOVQ("a+0(FP)", a)
		_f.Mov(a, t0)
		_f.Mov(t0, t1)
		_f.XORQ(a, a) // set a to zero for later reduction

		_f.MOVQ("b+8(FP)", b)
		_f.Add(b, t0) // t0 = a + b
		_f.Sub(b, t1) // t1 = a - b

		// reduce t1
		if _f.NbWords >= 5 {
			// q is on the stack, can't use for CMOVQCC
			_f.Mov(t0, q) // save t0
			for i := 0; i < _f.NbWords; i++ {
				_f.MOVQ(fmt.Sprintf("$const_q%d", i), t0[i])
			}
			for i := 0; i < _f.NbWords; i++ {
				_f.CMOVQCC(a, t0[i])
			}
			// add registers (q or 0) to t, and set to result
			_f.Add(t0, t1)
			_f.Mov(q, t0) // restore t0
		} else {
			for i := 0; i < _f.NbWords; i++ {
				_f.MOVQ(fmt.Sprintf("$const_q%d", i), q[i])
			}
			for i := 0; i < _f.NbWords; i++ {
				_f.CMOVQCC(a, q[i])
			}
			// add registers (q or 0) to t, and set to result
			_f.Add(q, t1)
		}

		_f.UnsafePush(&registers, q...)

		// save t1
		_f.Mov(t1, b)

		// reduce t0
		_f.ReduceElement(t0, t1, false)

		// save t0
		_f.MOVQ("a+0(FP)", a)
		_f.Mov(t0, a)

		_f.RET()

		_f.UnsafePush(&registers, t0...)
		_f.UnsafePush(&registers, t1...)
		_f.UnsafePush(&registers, a, b)
	} else {
		// registers
		r := _f.Pop(&registers)
		t0 := _f.PopN(&registers)
		t1 := _f.PopN(&registers)

		// t = a
		_f.MOVQ("b+8(FP)", r)
		_f.Mov(r, t0)

		_f.MOVQ("a+0(FP)", r)
		_f.Add(r, t0)  // t0 = a + b
		_f.Mov(t0, t1) // save t1 = t0
		_f.Mov(r, t0)  // t0 = a
		_f.MOVQ("b+8(FP)", r)
		_f.Sub(r, t0) // t0 = a - b

		// reduce t0
		noReduce := _f.NewLabel("noReduce")
		_f.JCC(noReduce)
		q := r
		_f.MOVQ("$const_q0", q)

		_f.ADDQ(q, t0[0])
		for i := 1; i < _f.NbWords; i++ {
			_f.MOVQ(fmt.Sprintf("$const_q%d", i), q)
			_f.ADCQ(q, t0[i])
		}
		_f.LABEL(noReduce)

		// save t1
		_f.MOVQ("b+8(FP)", r)
		_f.Mov(t0, r)

		// reduce t0
		_f.Mov(t1, t0)
		_f.ReduceElement(t0, t1, false)

		// save t0
		_f.MOVQ("a+0(FP)", r)
		_f.Mov(t0, r)

		_f.RET()

		_f.Push(&registers, t0...)
		_f.Push(&registers, t1...)
		_f.Push(&registers, r)
	}

}
