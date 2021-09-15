package amd64

import "github.com/consensys/bavard/amd64"

func (f *FFAmd64) generateInverse() {
	f.Comment("inverse(res, x *Element)")

	// we need r, s, u, v registers, + one set for subs or reductions
	stackSize := f.StackSize(f.NbWords*5+1, 0, 0)
	registers := f.FnHeader("inverse", stackSize, 16)
	defer f.AssertCleanStack(stackSize, 0)

	t := f.PopN(&registers)
	zero := f.Pop(&registers)
	loopCounter := amd64.BP

	// order is important here; for NbWords <= 6, u is going to fit into registers.
	u := f.PopN(&registers)
	v := f.PopN(&registers)
	r := f.PopN(&registers)
	s := f.PopN(&registers)

	uOnStack := f.NbWords > 6

	// labels
	startLoop := f.NewLabel()
	vBigger := f.NewLabel()
	endLoop := f.NewLabel()
	returnR := f.NewLabel()
	returnS := f.NewLabel()
	returnZero := f.NewLabel()

	// u = q
	f.Comment("u = q")
	f.LabelRegisters("u", u...)
	for i := 0; i < f.NbWords; i++ {
		if !uOnStack {
			// u is on registers
			f.MOVQ(f.qAt(i), u[i])
		} else {
			f.MOVQ(f.qAt(i), zero)
			f.MOVQ(zero, u[i])
		}
	}

	// s = r^2
	f.Comment("s = r^2")
	f.LabelRegisters("s", s...)
	for i := 0; i < f.NbWords; i++ {
		f.MOVQ(f.RSquare[i], zero)
		f.MOVQ(zero, s[i])
	}

	// v = x
	f.Comment("v = x")
	f.LabelRegisters("v", v...)
	f.MOVQ("x+8(FP)", zero)
	f.Mov(zero, t)
	f.Mov(t, v)

	f.Comment("if x is 0, returns 0")
	f.MOVQ(t[0], zero)
	for i := 1; i < len(t); i++ {
		f.ORQ(t[i], zero)
	}
	f.JEQ(returnZero)

	//r = 0
	f.Comment("r = 0")
	f.LabelRegisters("r", r...)
	for i := 0; i < len(r); i++ {
		f.MOVQ(0, r[i])
	}

	// rshOne set a and b such that
	// for a[0]&1 == 0 {
	// 		a <<= 1
	// 		if b[0] & 1 == 1 {
	//			b += q
	// 		}
	//		b <<= 1
	// }
	// t must be a set of registers.
	rshOne := func(a, b, t []amd64.Register) {
		end := f.NewLabel()
		firstLoop := f.NewLabel()
		secondLoop := f.NewLabel()

		// this is done before by the caller
		// f.Mov(a, t)
		f.BTQ(0, t[0])
		f.JCS(end)

		f.MOVQ(0, loopCounter)
		f.XORQ(zero, zero)
		f.LABEL(firstLoop)
		f.INCQ(loopCounter)

		f.SHRQw(1, t[0], zero)
		for i := 1; i < len(t); i++ {
			f.SHRQw(1, t[i], t[i-1])
		}
		f.SHRQ(1, t[len(t)-1])

		f.BTQ(0, t[0])
		f.JCC(firstLoop)

		// we need to save the result of the first loop
		f.Mov(t, a)
		f.Mov(b, t)
		// we need to shift r (t) loopCOunter times
		f.LABEL(secondLoop)

		f.BTQ(0, t[0]) // if r[0] is odd, we add modulus
		f.reduceIfBorrow(t)
		f.SHRQw(1, t[0], zero)
		for i := 1; i < len(t); i++ {
			f.SHRQw(1, t[i], t[i-1])
		}
		f.SHRQ(1, t[len(t)-1])

		f.DECQ(loopCounter)
		f.JNE(secondLoop)

		// save result of second loop
		f.Mov(t, b)

		f.LABEL(end)
	}

	f.LABEL(startLoop)

	// note: t always contains v here
	rshOne(v, s, t)

	f.Mov(u, t)
	rshOne(u, r, t)

	// f.Push(&registers, loopCounter)

	// v = v - u
	f.Comment("v = v - u")
	f.Mov(v, t)

	f.Sub(u, t)
	f.JCC(vBigger)

	// here v is smaller
	// u = u - v
	if !uOnStack {
		f.Sub(v, u)
	} else {
		f.Mov(u, t)
		f.Sub(v, t)
		f.Mov(t, u)
	}

	// r = r - s
	f.Mov(r, t)
	f.Sub(s, t)
	f.reduceIfBorrow(t)
	f.Mov(t, r)
	f.JMP(endLoop)

	// here v is bigger
	f.LABEL(vBigger)
	// v = v - u
	f.Mov(t, v)
	// s = s - r
	f.Mov(s, t)
	f.Sub(r, t)
	f.reduceIfBorrow(t)
	f.Mov(t, s)
	f.LABEL(endLoop)

	// if (u[0] == 1) && (u[5]|u[4]|u[3]|u[2]|u[1]) == 0 {
	// 		return z.Set(&r)
	// 	}
	// 	if (v[0] == 1) && (v[5]|v[4]|v[3]|v[2]|v[1]) == 0 {
	// 		return z.Set(&s)
	// 	}
	if !uOnStack {
		f.MOVQ(u[0], zero)
		f.SUBQ(1, zero)
		for i := 1; i < f.NbWords; i++ {
			f.ORQ(u[i], zero)
		}
	} else {
		f.Mov(u, t)
		f.SUBQ(1, t[0])
		last := len(t) - 1
		for i := 0; i < f.NbWords-1; i++ {
			f.ORQ(t[i], t[last])
		}
	}

	f.JEQ(returnR)

	f.Mov(v, t)
	f.MOVQ(t[0], zero)
	f.SUBQ(1, zero)
	f.JNE(startLoop)
	for i := 1; i < f.NbWords; i++ {
		f.ORQ(t[i], zero)
	}
	f.JEQ(returnS)

	f.JMP(startLoop)

	f.LABEL(returnR)
	f.MOVQ("res+0(FP)", zero)
	f.Mov(r, t)
	f.Mov(t, zero)
	f.RET()

	f.LABEL(returnS)
	f.MOVQ("res+0(FP)", zero)
	f.Mov(s, t)
	f.Mov(t, zero)
	f.RET()

	f.LABEL(returnZero)
	f.MOVQ("res+0(FP)", zero)
	for i := 0; i < len(t); i++ {
		f.MOVQ(0, zero.At(i))
	}
	f.RET()

	// f.Push(&registers, flagBorrow)
	f.Push(&registers, u...)
	f.Push(&registers, r...)
	f.Push(&registers, v...)
	f.Push(&registers, s...)
	f.Push(&registers, t...)
	f.Push(&registers, zero)
}

func (f *FFAmd64) reduceIfBorrow(t []amd64.Register) {
	noReduce := f.NewLabel()
	f.JCC(noReduce)
	f.ADDQ(f.qAt(0), t[0])
	for i := 1; i < f.NbWords; i++ {
		f.ADCQ(f.qAt(i), t[i])
	}
	f.LABEL(noReduce)
}
