// Copyright 2022 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package arm64

import "github.com/consensys/bavard/arm64"

func (f *FFArm64) generateAdd() {
	f.Comment("add(res, x, y *Element)")
	//stackSize := f.StackSize(f.NbWords*2, 0, 0)
	registers := f.FnHeader("add", 0, 24)
	defer f.AssertCleanStack(0, 0)

	// registers
	z := registers.PopN(f.NbWords)
	xPtr := registers.Pop()
	yPtr := registers.Pop()
	ops := registers.PopN(2)

	f.LDP("x+8(FP)", xPtr, yPtr)
	f.Comment("load operands and add mod 2^r")

	op0 := f.ADDS
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.RegisterOffset(xPtr, 8*i), z[i], ops[0])
		f.LDP(f.RegisterOffset(yPtr, 8*i), z[i+1], ops[1])

		op0(z[i], z[i+1], z[i])
		op0 = f.ADCS

		f.ADCS(ops[0], ops[1], z[i+1])
	}

	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(f.RegisterOffset(xPtr, 8*i), z[i], "can't import these in pairs")
		f.MOVD(f.RegisterOffset(yPtr, 8*i), ops[0])
		op0(z[i], ops[0], z[i])
	}
	registers.Push(xPtr, yPtr)
	registers.Push(ops...)

	t := registers.PopN(f.NbWords)
	f.reduce(z, t)
	registers.Push(t...)

	f.Comment("store")
	zPtr := registers.Pop()
	f.MOVD("res+0(FP)", zPtr)
	f.storeVector(z, zPtr)

	f.RET()

}

func (f *FFArm64) generateDouble() {
	f.Comment("double(res, x *Element)")
	registers := f.FnHeader("double", 0, 16)
	defer f.AssertCleanStack(0, 0)

	// registers
	z := registers.PopN(f.NbWords)
	xPtr := registers.Pop()
	zPtr := registers.Pop()
	//ops := registers.PopN(2)

	f.LDP("res+0(FP)", zPtr, xPtr)
	f.Comment("load operands and add mod 2^r")

	op0 := f.ADDS
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.RegisterOffset(xPtr, 8*i), z[i], z[i+1])

		op0(z[i], z[i], z[i])
		op0 = f.ADCS

		f.ADCS(z[i+1], z[i+1], z[i+1])
	}

	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(f.RegisterOffset(xPtr, 8*i), z[i])
		op0(z[i], z[i], z[i])
	}
	registers.Push(xPtr)

	t := registers.PopN(f.NbWords)
	f.reduce(z, t)
	registers.Push(t...)

	f.Comment("store")
	f.storeVector(z, zPtr)

	f.RET()

}

// generateSub uses one more register than generateAdd, but that's okay since we have 29 registers available.
func (f *FFArm64) generateSub() {
	f.Comment("sub(res, x, y *Element)")

	registers := f.FnHeader("sub", 0, 24)
	defer f.AssertCleanStack(0, 0)

	// registers
	z := registers.PopN(f.NbWords)
	xPtr := registers.Pop()
	yPtr := registers.Pop()
	ops := registers.PopN(2)

	f.LDP("x+8(FP)", xPtr, yPtr)
	f.Comment("load operands and subtract mod 2^r")

	op0 := f.SUBS
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.RegisterOffset(xPtr, 8*i), z[i], ops[0])
		f.LDP(f.RegisterOffset(yPtr, 8*i), z[i+1], ops[1])

		op0(z[i+1], z[i], z[i])
		op0 = f.SBCS

		f.SBCS(ops[1], ops[0], z[i+1])
	}

	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(f.RegisterOffset(xPtr, 8*i), z[i], "can't import these in pairs")
		f.MOVD(f.RegisterOffset(yPtr, 8*i), ops[0])
		op0(ops[0], z[i], z[i])
	}
	registers.Push(xPtr, yPtr)
	registers.Push(ops...)

	f.Comment("Store borrow TODO: Can it be done with one instruction?")
	borrow := registers.Pop()
	f.MOVD(0, borrow)
	f.ADC(0, borrow, borrow)
	//f.ADC(0, 0, borrow, "store borrow")

	f.Comment("load modulus and add")

	t := registers.PopN(f.NbWords)

	op0 = f.ADDS
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.GlobalOffset("q", 8*i), t[i], t[i+1])

		op0(t[i], z[i], t[i])
		op0 = f.ADCS

		f.ADCS(t[i+1], z[i+1], t[i+1])
	}

	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(f.GlobalOffset("q", 8*i), t[i])

		op0(t[i], z[i], t[i])
	}

	f.Comment("augment if necessary")

	f.CMP(1, borrow, "\"recall\" the borrow")
	registers.Push(borrow)

	for i := 0; i < f.NbWords; i++ {
		f.CSEL("NE", t[i], z[i], z[i])
	}

	registers.Push(t...)

	f.Comment("store")
	zPtr := registers.Pop()
	f.MOVD("res+0(FP)", zPtr)
	f.storeVector(z, zPtr)

	f.RET()

}

func (f *FFArm64) generateNeg() {
	f.Comment("neg(res, x *Element)")
	registers := f.FnHeader("neg", 0, 16)
	defer f.AssertCleanStack(0, 0)

	// registers
	z := registers.PopN(f.NbWords)
	xPtr := registers.Pop()
	zPtr := registers.Pop()
	ops := registers.PopN(2)

	f.LDP("res+0(FP)", zPtr, xPtr)
	f.Comment("load operands and subtract")

	op0 := f.SUBS
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.RegisterOffset(xPtr, 8*i), z[i], z[i+1])
		f.LDP(f.GlobalOffset("q", 8*i), ops[0], ops[1])

		op0(z[i], ops[0], z[i])
		op0 = f.SBCS

		f.SBCS(z[i+1], ops[1], z[i+1])
	}

	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(f.RegisterOffset(xPtr, 8*i), z[i], "can't import these in pairs")
		f.MOVD(f.GlobalOffset("q", 8*i), ops[0])
		op0(z[i], ops[0], z[i])
	}

	registers.Push(xPtr)
	registers.Push(ops...)

	// One might think reduction is not necessary here. One would be mistaken.
	// q - 0 = q but -0 = 0
	// Perhaps we could instead OR all of x together and select the output based on that.
	// Considered sequentially, an OR takes just as long as an ADD
	// But the OR strategy doesn't depend on the subtraction result.
	// Also, OR might consume less power even if taking the same amount of time?
	// Any difference would probably be minute anyway
	t := registers.PopN(f.NbWords)
	f.reduce(z, t)
	registers.Push(t...)

	f.Comment("store")
	f.storeVector(z, zPtr)

	f.RET()

}

// MACROS?
//TODO: Put it in a macro
func (f *FFArm64) reduce(z, t []arm64.Register) {

	if len(z) != f.NbWords || len(t) != f.NbWords {
		panic("need 2*nbWords registers")
	}

	f.Comment("load modulus and subtract")

	op0 := f.SUBS
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.GlobalOffset("q", 8*i), t[i], t[i+1])

		op0(t[i], z[i], t[i])
		op0 = f.SBCS

		f.SBCS(t[i+1], z[i+1], t[i+1])
	}

	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(f.GlobalOffset("q", 8*i), t[i])

		op0(t[i], z[i], t[i])
	}

	f.Comment("reduce if necessary")

	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", t[i], z[i], z[i])
	}
}

func (f *FFArm64) storeVector(v []arm64.Register, baseAddress arm64.Register) {
	for i := 0; i < f.NbWords-1; i += 2 {
		f.STP(v[i], v[i+1], f.RegisterOffset(baseAddress, 8*i))
	}

	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(v[i], f.RegisterOffset(baseAddress, 8*i))
	}
}

// </macros>
