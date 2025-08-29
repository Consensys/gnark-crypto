// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"github.com/consensys/bavard/amd64"
)

func (_f *FFAmd64) generateReduce() {
	stackSize := _f.StackSize(1+_f.NbWords*2, 0, 0)
	registers := _f.FnHeader("reduce", stackSize, 8)
	defer _f.AssertCleanStack(stackSize, 0)

	// registers
	r := registers.Pop()
	t := registers.PopN(_f.NbWords)

	_f.MOVQ("res+0(FP)", r)
	_f.Mov(r, t)
	_f.Reduce(&registers, t, false)
	_f.Mov(t, r)
	_f.RET()

	_f.UnsafePush(&registers, r)
	_f.UnsafePush(&registers, t...) // ensure the stack is clean
}

// Reduce scratch can be on the stack or a set of registers.
// If avoidGlobal is true:
// 1. f.qStack is set, in which case f.qAt should return the address of q on the stack
// 2. f.qStack is not set, in which case we use an extra register to move immediates values from
// constants.
// If avoidGlobal is false: we use f.qAt (which by default fetch q from global memory)
func (_f *FFAmd64) Reduce(registers *amd64.Registers, t []amd64.Register, avoidGlobal bool) {
	var spare amd64.Register
	if avoidGlobal && _f.qStack == nil {
		spare = _f.Pop(registers)
	}
	scratch := _f.PopN(registers)
	if avoidGlobal && _f.qStack == nil {
		scratch = append(scratch, spare)
	}
	_f.ReduceElement(t, scratch, avoidGlobal)
	_f.UnsafePush(registers, scratch...)
}
