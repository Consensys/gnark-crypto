// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

func (_f *FFAmd64) generateMulBy3() {
	_f.Comment("MulBy3(x *Element)")
	stackSize := _f.StackSize(1+_f.NbWords*2, 0, 0)
	registers := _f.FnHeader("MulBy3", stackSize, 8)
	defer _f.AssertCleanStack(stackSize, 0)
	// registers
	x := registers.Pop()
	t := registers.PopN(_f.NbWords)

	_f.MOVQ("x+0(FP)", x)

	_f.Mov(x, t)
	_f.Add(t, t)

	_f.Reduce(&registers, t, false)
	_f.Add(x, t)
	_f.Reduce(&registers, t, false)
	_f.Mov(t, x)

	_f.RET()
}

func (_f *FFAmd64) generateMulBy5() {
	_f.Comment("MulBy5(x *Element)")
	stackSize := _f.StackSize(1+_f.NbWords*2, 0, 0)
	registers := _f.FnHeader("MulBy5", stackSize, 8)
	defer _f.AssertCleanStack(stackSize, 0)

	// registers
	x := registers.Pop()
	t := registers.PopN(_f.NbWords)

	_f.MOVQ("x+0(FP)", x)

	_f.Mov(x, t)
	_f.Add(t, t)
	_f.Reduce(&registers, t, false)
	_f.Add(t, t)
	_f.Reduce(&registers, t, false)
	_f.Add(x, t)
	_f.Reduce(&registers, t, false)

	_f.Mov(t, x)
	_f.RET()
}

func (_f *FFAmd64) generateMulBy13() {
	_f.Comment("MulBy13(x *Element)")
	stackSize := _f.StackSize(1+_f.NbWords*3, 0, 0)
	registers := _f.FnHeader("MulBy13", stackSize, 8)
	defer _f.AssertCleanStack(stackSize, 0)

	// registers
	x := _f.Pop(&registers)
	t := _f.PopN(&registers)
	s := _f.PopN(&registers)
	u := _f.PopN(&registers)

	_f.MOVQ("x+0(FP)", x)

	_f.Mov(x, t)

	_f.Add(t, t)
	_f.ReduceElement(t, s, false)
	_f.Add(t, t)
	_f.ReduceElement(t, u, false)

	_f.Mov(t, u) // u == 4

	_f.Add(t, t) // t == 8
	_f.ReduceElement(t, s, false)

	_f.Add(u, t) // t == 12
	_f.ReduceElement(t, s, false)

	_f.Add(x, t) // t == 13
	_f.ReduceElement(t, s, false)

	_f.Mov(t, x)
	_f.RET()

	_f.Push(&registers, x)
	_f.Push(&registers, t...)
	_f.Push(&registers, u...)
	_f.Push(&registers, s...)
}
