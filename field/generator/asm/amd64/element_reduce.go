// Copyright 2020 Consensys Software Inc.
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
	f.Reduce(&registers, t)
	f.Mov(t, r)
	f.RET()
}

// Reduce scratch can be on the stack or a set of registers.
func (f *FFAmd64) Reduce(registers *amd64.Registers, t []amd64.Register) {
	scratch := f.PopN(registers)
	f.ReduceElement(t, scratch)
	f.Push(registers, scratch...)
}
