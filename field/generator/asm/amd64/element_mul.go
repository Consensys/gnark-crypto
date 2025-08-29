// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"fmt"

	"github.com/consensys/bavard/amd64"
)

// MulADX uses AX, DX and BP
// sets x * y into t, without modular reduction
// x() will have more accesses than y()
// (caller should store x in registers, if possible)
// if no (tmp) register is available, this uses one PUSH/POP on the stack in the hot loop.
func (_f *FFAmd64) MulADX(registers *amd64.Registers, x, y func(int) string, t []amd64.Register) []amd64.Register {
	// registers
	var tr amd64.Register // temporary register
	A := amd64.BP

	hasFreeRegister := registers.Available() > 0
	if hasFreeRegister {
		tr = registers.Pop()
	} else {
		tr = A
	}

	_f.LabelRegisters("A", A)
	_f.LabelRegisters("t", t...)

	mac := _f.Define("MACC", 3, func(args ...any) {
		in0 := args[0]
		in1 := args[1]
		in2 := args[2]
		_f.ADCXQ(in0, in1)
		_f.MULXQ(in2, amd64.AX, in0)
		_f.ADOXQ(amd64.AX, in1)
	}, true)

	divShift := _f.Define("DIV_SHIFT", 0, func(_ ...any) {
		if !hasFreeRegister {
			_f.PUSHQ(A)
		}
		// m := t[0]*q'[0] mod W
		m := amd64.DX
		_f.MOVQ(_f.qInv0(), m)
		_f.IMULQ(t[0], m)

		// clear the carry flags
		_f.XORQ(amd64.AX, amd64.AX)

		// C,_ := t[0] + m*q[0]
		_f.MULXQ(_f.qAt(0), amd64.AX, tr)
		_f.ADCXQ(t[0], amd64.AX)
		_f.MOVQ(tr, t[0])

		if !hasFreeRegister {
			_f.POPQ(A)
		}

		// for j=1 to N-1
		//
		//	(C,t[j-1]) := t[j] + m*q[j] + C
		for j := 1; j < _f.NbWords; j++ {
			mac(t[j], t[j-1], amd64.Register(_f.qAt(j)))
		}

		_f.MOVQ(0, amd64.AX)
		_f.ADCXQ(amd64.AX, t[_f.NbWordsLastIndex])
		_f.ADOXQ(A, t[_f.NbWordsLastIndex])

	}, true)

	mulWord0 := _f.Define("MUL_WORD_0", 0, func(_ ...any) {
		_f.XORQ(amd64.AX, amd64.AX)
		// for j=0 to N-1
		//    (A,t[j])  := t[j] + x[j]*y[i] + A
		for j := 0; j < _f.NbWords; j++ {
			if j == 0 && _f.NbWords == 1 {
				_f.MULXQ(x(j), t[j], A)
			} else if j == 0 {
				_f.MULXQ(x(j), t[j], t[j+1])
			} else {
				highBits := A
				if j != _f.NbWordsLastIndex {
					highBits = t[j+1]
				}
				_f.MULXQ(x(j), amd64.AX, highBits)
				_f.ADOXQ(amd64.AX, t[j])
			}
		}
		_f.MOVQ(0, amd64.AX)
		_f.ADOXQ(amd64.AX, A)
		divShift()
	}, true)

	mulWordN := _f.Define("MUL_WORD_N", 0, func(args ...any) {
		_f.XORQ(amd64.AX, amd64.AX)
		// for j=0 to N-1
		//    (A,t[j])  := t[j] + x[j]*y[i] + A
		_f.MULXQ(x(0), amd64.AX, A)
		_f.ADOXQ(amd64.AX, t[0])
		for j := 1; j < _f.NbWords; j++ {
			mac(A, t[j], amd64.Register(x(j)))
		}
		_f.MOVQ(0, amd64.AX)
		_f.ADCXQ(amd64.AX, A)
		_f.ADOXQ(amd64.AX, A)
		divShift()
	}, true)

	_f.Comment("mul body")

	for i := 0; i < _f.NbWords; i++ {
		_f.MOVQ(y(i), amd64.DX)

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

func (_f *FFAmd64) generateMul(_ bool) {
	_f.Comment("mul(res, x, y *Element)")

	const argSize = 3 * 8
	const minStackSize = argSize

	nbRegistersNeeded := (_f.NbWords * 2) - 2

	// we need to use R15 register, and to avoid issue with dynamic linking
	// see https://github.com/Consensys/gnark-crypto/issues/707
	// we avoid using global variables in this particular instance.
	needR15 := _f.NbWords >= 12
	if needR15 {
		nbRegistersNeeded += _f.NbWords // we need to store Q on the stack.
		nbRegistersNeeded++             // account for R15, and the path of available registers == 0 below
	}

	stackSize := _f.StackSize(nbRegistersNeeded, 2, minStackSize)
	registers := _f.FnHeader("mul", stackSize, argSize, amd64.AX, amd64.DX)
	defer _f.AssertCleanStack(stackSize, minStackSize)

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

	_f.WriteLn(`
	// Algorithm 2 of "Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS" 
	// by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521
	// See github.com/Consensys/gnark-crypto/field/generator for more comments.
	`)
	if stackSize > 0 {
		_f.WriteLn("NO_LOCAL_POINTERS")
	}

	noAdx := _f.NewLabel("noAdx")

	{
		// check ADX instruction support
		_f.CMPB("·supportAdx(SB)", 1)
		_f.JNE(noAdx)
	}

	{
		// we need to access x and y words, per index
		var xat, yat func(int) string
		var gc func()

		// we need NbWords registers for t, plus optionally one for tmp register in mulADX if we want to avoid PUSH/POP
		nbRegisters := registers.Available()
		if nbRegisters < _f.NbWords {
			panic(fmt.Sprintf("not enough registers, not supported: %d < %d", nbRegisters, _f.NbWords))
		}
		t := registers.PopN(_f.NbWords)

		nbRegisters = registers.Available()
		switch nbRegisters {
		case 0:

			// y is access through use of AX/DX
			yat = func(i int) string {
				y := amd64.AX
				_f.MOVQ("y+16(FP)", y)
				return y.At(i)
			}

			// we move x on the stack.
			_f.MOVQ("x+8(FP)", amd64.AX)
			_x := _f.PopN(&registers, true)
			_f.LabelRegisters("x", _x...)
			_f.Mov(amd64.AX, t)
			_f.Mov(t, _x)
			xat = func(i int) string {
				return string(_x[i])
			}

			gc = func() {
				_f.Push(&registers, _x...)
			}
		case 1:
			// y is access through use of AX/DX
			yat = func(i int) string {
				y := amd64.AX
				_f.MOVQ("y+16(FP)", y)
				return y.At(i)
			}
			// x uses the register
			x := registers.Pop()
			_f.MOVQ("x+8(FP)", x)
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

			_f.MOVQ("x+8(FP)", x)
			_f.MOVQ("y+16(FP)", y)

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

			if nbRegisters >= _f.NbWords {
				// we store x fully in registers
				x := registers.Pop()
				_f.MOVQ("x+8(FP)", x)
				_x := registers.PopN(_f.NbWords)
				_f.LabelRegisters("x", _x...)
				_f.Mov(x, _x)

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
				_f.MOVQ("x+8(FP)", x)

				// and use the rest for x0...xn
				_x := registers.PopN(nbRegisters)
				_f.LabelRegisters("x", _x...)
				for i := 0; i < len(_x); i++ {
					_f.MOVQ(x.At(i), _x[i])
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

			_f.MOVQ("y+16(FP)", y)
			yat = func(i int) string {
				return y.At(i)
			}

		}

		_f.MulADX(&registers, xat, yat, t)
		gc()
		_f.Push(&registers, amd64.AX, amd64.DX)

		// ---------------------------------------------------------------------------------------------
		// reduce
		_f.Reduce(&registers, t, needR15)

		_f.MOVQ("res+0(FP)", amd64.AX)
		_f.Mov(t, amd64.AX)
		_f.RET()
	}

	// ---------------------------------------------------------------------------------------------
	// no MULX, ADX instructions
	{
		_f.LABEL(noAdx)

		_f.MOVQ("res+0(FP)", amd64.AX)
		_f.MOVQ(amd64.AX, "(SP)")
		_f.MOVQ("x+8(FP)", amd64.AX)
		_f.MOVQ(amd64.AX, "8(SP)")
		_f.MOVQ("y+16(FP)", amd64.AX)
		_f.MOVQ(amd64.AX, "16(SP)")
		_f.WriteLn("CALL ·_mulGeneric(SB)")
		_f.RET()

	}
}
