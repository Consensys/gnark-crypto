// Copyright 2020-2025 Consensys Software Inc.
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
// If useQGlobal is true, the global qElement variable is used for reduction.
// If false, the reduction is done without using the global variable (using defines / IMMs)
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
	f.Push(registers, scratch...)
}
