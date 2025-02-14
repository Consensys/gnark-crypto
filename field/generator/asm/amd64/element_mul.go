// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"

	"github.com/consensys/bavard/amd64"
)

// Registers used when f.NbWords == 4
// for the multiplication.
// They are re-referenced in defines in the vectorized operations.
var mul4Registers = []amd64.Register{
	// t
	amd64.R14, amd64.R13, amd64.CX, amd64.BX,
	// x
	amd64.DI, amd64.R8, amd64.R9, amd64.R10,
	// tr
	amd64.R12,
}

// MulADX uses AX, DX and BP
// sets x * y into t, without modular reduction
// x() will have more accesses than y()
// (caller should store x in registers, if possible)
// if no (tmp) register is available, this uses one PUSH/POP on the stack in the hot loop.
func (f *FFAmd64) MulADX(registers *amd64.Registers, x, y func(int) string, t []amd64.Register) []amd64.Register {
	// registers
	var tr amd64.Register // temporary register
	A := amd64.BP

	hasFreeRegister := registers.Available() > 0
	if hasFreeRegister {
		tr = registers.Pop()
	} else {
		tr = A
	}

	f.LabelRegisters("A", A)
	f.LabelRegisters("t", t...)

	if f.NbWords == 4 && hasFreeRegister {
		// ensure the registers match the "hardcoded ones" in mul4Registers for the vecops
		match := true
		for i := 0; i < 4; i++ {
			if mul4Registers[i] != t[i] {
				match = false
				fmt.Printf("expected %s, got t[%d] %s\n", mul4Registers[i], i, t[i])
			}
			if mul4Registers[i+4] != amd64.Register(x(i)) {
				match = false
				fmt.Printf("expected %s, got x[%d] %s\n", mul4Registers[i+4], i, x(i))
			}
		}
		if tr != mul4Registers[8] {
			match = false
			fmt.Printf("expected %s, got tr %s\n", mul4Registers[8], tr)
		}
		if !match {
			panic("registers do not match hardcoded ones")
		}
	}

	mac := f.Define("MACC", 3, func(args ...any) {
		in0 := args[0]
		in1 := args[1]
		in2 := args[2]
		f.ADCXQ(in0, in1)
		f.MULXQ(in2, amd64.AX, in0)
		f.ADOXQ(amd64.AX, in1)
	})

	divShift := f.Define("DIV_SHIFT", 0, func(_ ...any) {
		if !hasFreeRegister {
			f.PUSHQ(A)
		}
		// m := t[0]*q'[0] mod W
		m := amd64.DX
		f.MOVQ(f.qInv0(), m)
		f.IMULQ(t[0], m)

		// clear the carry flags
		f.XORQ(amd64.AX, amd64.AX)

		// C,_ := t[0] + m*q[0]
		f.MULXQ(f.qAt(0), amd64.AX, tr)
		f.ADCXQ(t[0], amd64.AX)
		f.MOVQ(tr, t[0])

		if !hasFreeRegister {
			f.POPQ(A)
		}

		// for j=1 to N-1
		//
		//	(C,t[j-1]) := t[j] + m*q[j] + C
		for j := 1; j < f.NbWords; j++ {
			mac(t[j], t[j-1], amd64.Register(f.qAt(j)))
		}

		f.MOVQ(0, amd64.AX)
		f.ADCXQ(amd64.AX, t[f.NbWordsLastIndex])
		f.ADOXQ(A, t[f.NbWordsLastIndex])

	})

	mulWord0 := f.Define("MUL_WORD_0", 0, func(_ ...any) {
		f.XORQ(amd64.AX, amd64.AX)
		// for j=0 to N-1
		//    (A,t[j])  := t[j] + x[j]*y[i] + A
		for j := 0; j < f.NbWords; j++ {
			if j == 0 && f.NbWords == 1 {
				f.MULXQ(x(j), t[j], A)
			} else if j == 0 {
				f.MULXQ(x(j), t[j], t[j+1])
			} else {
				highBits := A
				if j != f.NbWordsLastIndex {
					highBits = t[j+1]
				}
				f.MULXQ(x(j), amd64.AX, highBits)
				f.ADOXQ(amd64.AX, t[j])
			}
		}
		f.MOVQ(0, amd64.AX)
		f.ADOXQ(amd64.AX, A)
		divShift()
	})

	mulWordN := f.Define("MUL_WORD_N", 0, func(args ...any) {
		f.XORQ(amd64.AX, amd64.AX)
		// for j=0 to N-1
		//    (A,t[j])  := t[j] + x[j]*y[i] + A
		f.MULXQ(x(0), amd64.AX, A)
		f.ADOXQ(amd64.AX, t[0])
		for j := 1; j < f.NbWords; j++ {
			mac(A, t[j], amd64.Register(x(j)))
		}
		f.MOVQ(0, amd64.AX)
		f.ADCXQ(amd64.AX, A)
		f.ADOXQ(amd64.AX, A)
		divShift()
	})

	f.Comment("mul body")

	for i := 0; i < f.NbWords; i++ {
		f.MOVQ(y(i), amd64.DX)

		if i == 0 {
			mulWord0()
		} else {
			mulWordN()
		}
	}

	if hasFreeRegister {
		registers.Push(tr)
	}

	return t
}

func (f *FFAmd64) generateMul(forceADX bool) {
	f.Comment("mul(res, x, y *Element)")

	const argSize = 3 * 8
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
	registers := f.FnHeader("mul", stackSize, argSize, reserved...)
	defer f.AssertCleanStack(stackSize, minStackSize)

	f.WriteLn(`
	// Algorithm 2 of "Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS" 
	// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521
	// See github.com/Consensys/gnark-crypto/field/generator for more comments.
	`)
	if stackSize > 0 {
		f.WriteLn("NO_LOCAL_POINTERS")
	}

	noAdx := f.NewLabel("noAdx")

	if !forceADX {
		// check ADX instruction support
		f.CMPB("·supportAdx(SB)", 1)
		f.JNE(noAdx)
	}

	{
		// we need to access x and y words, per index
		var xat, yat func(int) string
		var gc func()

		// we need NbWords registers for t, plus optionally one for tmp register in mulADX if we want to avoid PUSH/POP
		nbRegisters := registers.Available()
		if nbRegisters < f.NbWords {
			panic("not enough registers, not supported.")
		}

		t := registers.PopN(f.NbWords)
		nbRegisters = registers.Available()
		switch nbRegisters {
		case 0:
			// y is access through use of AX/DX
			yat = func(i int) string {
				y := amd64.AX
				f.MOVQ("y+16(FP)", y)
				return y.At(i)
			}

			// we move x on the stack.
			f.MOVQ("x+8(FP)", amd64.AX)
			_x := f.PopN(&registers, true)
			f.LabelRegisters("x", _x...)
			f.Mov(amd64.AX, t)
			f.Mov(t, _x)
			xat = func(i int) string {
				return string(_x[i])
			}
			gc = func() {
				f.Push(&registers, _x...)
			}
		case 1:
			// y is access through use of AX/DX
			yat = func(i int) string {
				y := amd64.AX
				f.MOVQ("y+16(FP)", y)
				return y.At(i)
			}
			// x uses the register
			x := registers.Pop()
			f.MOVQ("x+8(FP)", x)
			xat = func(i int) string {
				return x.At(i)
			}
			gc = func() {
				registers.Push(x)
			}
		case 2, 3:
			// x, y uses registers
			x := registers.Pop()
			y := registers.Pop()

			f.MOVQ("x+8(FP)", x)
			f.MOVQ("y+16(FP)", y)

			xat = func(i int) string {
				return x.At(i)
			}

			yat = func(i int) string {
				return y.At(i)
			}
			gc = func() {
				registers.Push(x, y)
			}
		default:
			// we have a least 4 registers.
			// 1 for tmp.
			nbRegisters--
			// 1 for y
			nbRegisters--
			var y amd64.Register

			if nbRegisters >= f.NbWords {
				// we store x fully in registers
				x := registers.Pop()
				f.MOVQ("x+8(FP)", x)
				_x := registers.PopN(f.NbWords)
				f.LabelRegisters("x", _x...)
				f.Mov(x, _x)

				xat = func(i int) string {
					return string(_x[i])
				}
				registers.Push(x)
				gc = func() {
					registers.Push(y)
					registers.Push(_x...)
				}
			} else {
				// we take at least 1 register for x addr
				nbRegisters--
				x := registers.Pop()
				y = registers.Pop() // temporary lock 1 for y
				f.MOVQ("x+8(FP)", x)

				// and use the rest for x0...xn
				_x := registers.PopN(nbRegisters)
				f.LabelRegisters("x", _x...)
				for i := 0; i < len(_x); i++ {
					f.MOVQ(x.At(i), _x[i])
				}
				xat = func(i int) string {
					if i < len(_x) {
						return string(_x[i])
					}
					return x.At(i)
				}
				registers.Push(y)

				gc = func() {
					registers.Push(x, y)
					registers.Push(_x...)
				}

			}
			y = registers.Pop()

			f.MOVQ("y+16(FP)", y)
			yat = func(i int) string {
				return y.At(i)
			}

		}

		f.MulADX(&registers, xat, yat, t)
		gc()

		// ---------------------------------------------------------------------------------------------
		// reduce
		f.Reduce(&registers, t)

		f.MOVQ("res+0(FP)", amd64.AX)
		f.Mov(t, amd64.AX)
		f.RET()
	}

	// ---------------------------------------------------------------------------------------------
	// no MULX, ADX instructions
	if !forceADX {
		f.LABEL(noAdx)

		f.MOVQ("res+0(FP)", amd64.AX)
		f.MOVQ(amd64.AX, "(SP)")
		f.MOVQ("x+8(FP)", amd64.AX)
		f.MOVQ(amd64.AX, "8(SP)")
		f.MOVQ("y+16(FP)", amd64.AX)
		f.MOVQ(amd64.AX, "16(SP)")
		f.WriteLn("CALL ·_mulGeneric(SB)")
		f.RET()

	}
}
