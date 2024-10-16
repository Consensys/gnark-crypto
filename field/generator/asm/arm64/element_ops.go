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

import (
	"github.com/consensys/bavard/arm64"
)

func (f *FFArm64) generateAdd() {
	f.Comment("add(res, x, y *Element)")
	registers := f.FnHeader("add", 0, 24)
	defer f.AssertCleanStack(0, 0)

	// registers
	t := registers.PopN(f.NbWords)
	z := registers.PopN(f.NbWords)
	x := registers.PopN(f.NbWords)
	xPtr := registers.Pop()
	yPtr := registers.Pop()
	zPtr := registers.Pop()

	f.LDP("x+8(FP)", xPtr, yPtr)

	f.load(xPtr, x)
	f.load(yPtr, z)

	f.ADDS(x[0], z[0], z[0])
	for i := 1; i < f.NbWords; i++ {
		f.ADCS(x[i], z[i], z[i])
	}

	f.reduce(z, t)

	f.Comment("store")

	f.MOVD("res+0(FP)", zPtr)
	f.store(z, zPtr)

	f.RET()

}

func (f *FFArm64) generateDouble() {
	f.Comment("double(res, x *Element)")
	registers := f.FnHeader("double", 0, 16)
	defer f.AssertCleanStack(0, 0)

	// registers
	xPtr := registers.Pop()
	zPtr := registers.Pop()
	z := registers.PopN(f.NbWords)
	t := registers.PopN(f.NbWords)

	f.LDP("res+0(FP)", zPtr, xPtr)

	f.load(xPtr, z)

	f.ADDS(z[0], z[0], z[0])
	for i := 1; i < f.NbWords; i++ {
		f.ADCS(z[i], z[i], z[i])
	}

	f.reduce(z, t)

	f.store(z, zPtr)

	f.RET()

}

// generateSub NO LONGER uses one more register than generateAdd, but that's okay since we have 29 registers available.
func (f *FFArm64) generateSub() {
	f.Comment("sub(res, x, y *Element)")

	registers := f.FnHeader("sub", 0, 24)
	defer f.AssertCleanStack(0, 0)

	// registers
	z := registers.PopN(f.NbWords)
	x := registers.PopN(f.NbWords)
	t := registers.PopN(f.NbWords)
	xPtr := registers.Pop()
	yPtr := registers.Pop()
	zPtr := registers.Pop()

	f.LDP("x+8(FP)", xPtr, yPtr)

	f.load(xPtr, x)
	f.load(yPtr, z)

	f.SUBS(z[0], x[0], z[0])
	for i := 1; i < f.NbWords; i++ {
		f.SBCS(z[i], x[i], z[i])
	}

	f.Comment("load modulus and select")

	zero := arm64.Register("ZR")

	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.qAt(i), t[i], t[i+1])
	}
	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", zero, t[i], t[i])
	}
	f.Comment("add q if underflow, 0 if not")
	f.ADDS(z[0], t[0], z[0])
	for i := 1; i < f.NbWords; i++ {
		f.ADCS(z[i], t[i], z[i])
	}

	f.MOVD("res+0(FP)", zPtr)
	f.store(z, zPtr)

	f.RET()

}

func (f *FFArm64) generateButterfly() {
	f.Comment("butterfly(x, y *Element)")
	registers := f.FnHeader("Butterfly", 0, 16)
	defer f.AssertCleanStack(0, 0)
	// Butterfly sets
	//  a = a + b (mod q)
	//  b = a - b (mod q)
	// registers
	a := registers.PopN(f.NbWords)
	b := registers.PopN(f.NbWords)
	aRes := registers.PopN(f.NbWords)
	t := registers.PopN(f.NbWords)
	aPtr := registers.Pop()
	bPtr := registers.Pop()

	f.LDP("x+0(FP)", aPtr, bPtr)
	f.load(aPtr, a)
	f.load(bPtr, b)

	f.ADDS(a[0], b[0], aRes[0])
	for i := 1; i < f.NbWords; i++ {
		f.ADCS(a[i], b[i], aRes[i])
	}

	f.reduce(aRes, t)

	f.Comment("store")

	f.store(aRes, aPtr)

	bRes := b

	f.SUBS(b[0], a[0], bRes[0])
	for i := 1; i < f.NbWords; i++ {
		f.SBCS(b[i], a[i], bRes[i])
	}

	f.Comment("load modulus and select")

	zero := arm64.Register("ZR")

	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.qAt(i), t[i], t[i+1])
	}
	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", zero, t[i], t[i])
	}
	f.Comment("add q if underflow, 0 if not")
	f.ADDS(bRes[0], t[0], bRes[0])
	for i := 1; i < f.NbWords; i++ {
		f.ADCS(bRes[i], t[i], bRes[i])
	}

	f.Comment("store")

	f.store(bRes, bPtr)

	f.RET()
}

func (f *FFArm64) reduce(z, t []arm64.Register) {

	if len(z) != f.NbWords || len(t) != f.NbWords {
		panic("need 2*nbWords registers")
	}

	f.Comment("load modulus and subtract")

	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.qAt(i), t[i], t[i+1])
	}
	f.SUBS(t[0], z[0], t[0])
	for i := 1; i < f.NbWords; i++ {
		f.SBCS(t[i], z[i], t[i])
	}

	f.Comment("reduce if necessary")
	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", t[i], z[i], z[i])
	}
}

func (f *FFArm64) load(zPtr arm64.Register, z []arm64.Register) {
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(zPtr.At(i), z[i], z[i+1])
	}
}

func (f *FFArm64) store(z []arm64.Register, zPtr arm64.Register) {
	for i := 0; i < f.NbWords-1; i += 2 {
		f.STP(z[i], z[i+1], zPtr.At(i))
	}
}
