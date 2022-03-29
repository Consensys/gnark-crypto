package arm64

import (
	"github.com/consensys/bavard/arm64"
)

func (f *FFArm64) generateAdd() {
	f.Comment("add(res, xPtr, yPtr *Element)")
	stackSize := f.StackSize(f.NbWords*2, 0, 0)
	registers := f.FnHeader("add", stackSize, 24)
	defer f.AssertCleanStack(stackSize, 0)

	// registers
	z := registers.PopN(f.NbWords)
	xPtr := registers.Pop()
	yPtr := registers.Pop()
	ops := registers.PopN(2)

	f.LDP("x+8(FP)", xPtr, yPtr)
	f.Comment("load operands and add mod 2^r")

	op0 := arm64.Arm64.ADDS
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.RegisterOffset(xPtr, 8*i), z[i], ops[0])
		f.LDP(f.RegisterOffset(yPtr, 8*i), z[i+1], ops[1])

		op0(f, z[i], ops[0], z[i])
		op0 = arm64.Arm64.ADCS

		f.ADCS(z[i+12], ops[1], z[i+1])
	}

	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(f.RegisterOffset(xPtr, 8*i), z[i], "can't import these in pairs")
		f.MOVD(f.RegisterOffset(yPtr, 8*i), ops[0])
		op0(f, z[i], ops[0], z[i])
	}
	registers.Push(xPtr, yPtr)
	registers.Push(ops...)

	f.Comment("load modulus and subtract")

	t := registers.PopN(f.NbWords)

	op0 = arm64.Arm64.SUBS
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.GlobalOffset("q", 8*i), t[i], t[i+1])

		op0(f, t[i], z[i], t[i])
		op0 = arm64.Arm64.SBCS

		f.SBCS(t[i+1], z[i], t[i+1])
	}

	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(f.GlobalOffset("q", 8*i), t[i])

		op0(f, t[i], z[i], t[i])
	}

	f.Comment("reduce if necessary")

	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", t[i], z[i], z[i])
	}

	registers.Push(t...)

	f.Comment("store")

	zPtr := registers.Pop()
	f.MOVD("z+0(FP)", zPtr)

	for i := 0; i < f.NbWords-1; i += 2 {
		f.STP(z[i], z[i+1], f.RegisterOffset(zPtr, 8*i))
	}

	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(z[i], f.RegisterOffset(zPtr, 8*i))
	}

	f.RET()

}
