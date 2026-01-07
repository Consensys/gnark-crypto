// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

func (f *FFAmd64) generateMulBy3() {
	f.Comment("MulBy3(x *Element)")
	stackSize := f.StackSize(1+f.NbWords*2, 0, 0)
	registers := f.FnHeader("MulBy3", stackSize, 8)
	defer f.AssertCleanStack(stackSize, 0)
	// registers
	x := registers.Pop()
	t := registers.PopN(f.NbWords)

	f.MOVQ("x+0(FP)", x)

	f.Mov(x, t)
	f.Add(t, t)

	f.Reduce(&registers, t, false)
	f.Add(x, t)
	f.Reduce(&registers, t, false)
	f.Mov(t, x)

	f.RET()
}

func (f *FFAmd64) generateMulBy5() {
	f.Comment("MulBy5(x *Element)")
	stackSize := f.StackSize(1+f.NbWords*2, 0, 0)
	registers := f.FnHeader("MulBy5", stackSize, 8)
	defer f.AssertCleanStack(stackSize, 0)

	// registers
	x := registers.Pop()
	t := registers.PopN(f.NbWords)

	f.MOVQ("x+0(FP)", x)

	f.Mov(x, t)
	f.Add(t, t)
	f.Reduce(&registers, t, false)
	f.Add(t, t)
	f.Reduce(&registers, t, false)
	f.Add(x, t)
	f.Reduce(&registers, t, false)

	f.Mov(t, x)
	f.RET()
}

func (f *FFAmd64) generateMulBy13() {
	f.Comment("MulBy13(x *Element)")
	stackSize := f.StackSize(1+f.NbWords*3, 0, 0)
	registers := f.FnHeader("MulBy13", stackSize, 8)
	defer f.AssertCleanStack(stackSize, 0)

	// registers
	x := f.Pop(&registers)
	t := f.PopN(&registers)
	s := f.PopN(&registers)
	u := f.PopN(&registers)

	f.MOVQ("x+0(FP)", x)

	f.Mov(x, t)

	f.Add(t, t)
	f.ReduceElement(t, s, false)
	f.Add(t, t)
	f.ReduceElement(t, u, false)

	f.Mov(t, u) // u == 4

	f.Add(t, t) // t == 8
	f.ReduceElement(t, s, false)

	f.Add(u, t) // t == 12
	f.ReduceElement(t, s, false)

	f.Add(x, t) // t == 13
	f.ReduceElement(t, s, false)

	f.Mov(t, x)
	f.RET()

	f.Push(&registers, x)
	f.Push(&registers, t...)
	f.Push(&registers, u...)
	f.Push(&registers, s...)
}
