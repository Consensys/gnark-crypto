// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"github.com/consensys/bavard/amd64"
)

func (f *FFAmd64) generateReduce() {
	stackSize := f.StackSize(1+f.NbWords*2, 0, 0)
	registers := f.FnHeader("reduce", stackSize, 8)
	defer f.AssertCleanStack(stackSize, 0)

	// registers
	r := registers.Pop()
	t := registers.PopN(f.NbWords)

	f.MOVQ("res+0(FP)", r)
	f.Mov(r, t)
	f.Reduce(&registers, t, false)
	f.Mov(t, r)
	f.RET()

	f.UnsafePush(&registers, r)
	f.UnsafePush(&registers, t...) // ensure the stack is clean
}

// Reduce scratch can be on the stack or a set of registers.
// If avoidGlobal is true:
// 1. f.qStack is set, in which case f.qAt should return the address of q on the stack
// 2. f.qStack is not set, in which case we use an extra register to move immediates values from
// constants.
// If avoidGlobal is false: we use f.qAt (which by default fetch q from global memory)
func (f *FFAmd64) Reduce(registers *amd64.Registers, t []amd64.Register, avoidGlobal bool) {
	var spare amd64.Register
	if avoidGlobal && f.qStack == nil {
		spare = f.Pop(registers)
	}
	scratch := f.PopN(registers)
	if avoidGlobal && f.qStack == nil {
		scratch = append(scratch, spare)
	}
	f.ReduceElement(t, scratch, avoidGlobal)
	f.UnsafePush(registers, scratch...)
}
