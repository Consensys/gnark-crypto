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

func (f *FFArm64) generateMul() {
	f.Comment("mul(res, x, y *Element)")
	registers := f.FnHeader("mul", 0, 24)
	defer f.AssertCleanStack(0, 0)

	xPtr := registers.Pop()
	yPtr := registers.Pop()
	bi := registers.Pop()
	a := registers.PopN(f.NbWords)
	t := registers.PopN(f.NbWords + 1)
	q := registers.PopN(f.NbWords)

	f.LDP("x+8(FP)", xPtr, yPtr)

	f.load(xPtr, a)
	ax := xPtr
	// f.load(yPtr, y)

	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.qAt(i), q[i], q[i+1])
	}

	divShift := f.Define("divShift", 0, func(args ...arm64.Register) {
		m := bi
		f.MOVD(f.qInv0(), m)
		f.MUL(t[0], m, m)

		// for j=0 to N-1
		//	(C,t[j-1]) := t[j] + m*q[j] + C

		for j := 0; j < f.NbWords; j++ {
			f.MUL(q[j], m, ax)
			if j == 0 {
				f.ADDS(ax, t[j], t[j])
			} else {
				f.ADCS(ax, t[j], t[j])
			}
		}
		f.ADCS("ZR", t[f.NbWords], t[f.NbWords])

		// propagate high bits
		f.UMULH(q[0], m, ax)
		for j := 1; j <= f.NbWords; j++ {
			if j == 1 {
				f.ADDS(ax, t[j], t[j-1])
			} else {
				f.ADCS(ax, t[j], t[j-1])
			}
			if j != f.NbWords {
				f.UMULH(q[j], m, ax)
			}
		}
	})

	mulWordN := f.Define("MUL_WORD_N", 0, func(args ...arm64.Register) {
		// for j=0 to N-1
		//    (C,t[j])  := t[j] + a[j]*b[i] + C

		// lo bits
		for j := 0; j < f.NbWords; j++ {
			f.MUL(a[j], bi, ax)

			if j == 0 {
				f.ADDS(ax, t[j], t[j])
			} else {
				f.ADCS(ax, t[j], t[j])
			}
		}

		f.ADCS("ZR", "ZR", t[f.NbWords])

		// propagate high bits
		f.UMULH(a[0], bi, ax)
		for j := 1; j <= f.NbWords; j++ {
			if j == 1 {
				f.ADDS(ax, t[j], t[j])
			} else {
				f.ADCS(ax, t[j], t[j])
			}
			if j != f.NbWords {
				f.UMULH(a[j], bi, ax)
			}
		}
		divShift()
	})

	mulWord0 := f.Define("MUL_WORD_0", 0, func(args ...arm64.Register) {
		// for j=0 to N-1
		//    (C,t[j])  := t[j] + a[j]*b[i] + C

		// lo bits
		for j := 0; j < f.NbWords; j++ {
			f.MUL(a[j], bi, t[j])
		}

		// propagate high bits
		f.UMULH(a[0], bi, ax)
		for j := 1; j <= f.NbWords; j++ {
			if j == 1 {
				f.ADDS(ax, t[j], t[j])
			} else {
				if j == f.NbWords {
					f.ADCS("ZR", ax, t[j])
				} else {
					f.ADCS(ax, t[j], t[j])
				}
			}
			if j != f.NbWords {
				f.UMULH(a[j], bi, ax)
			}
		}
		divShift()
	})

	f.Comment("mul body")

	for i := 0; i < f.NbWords; i++ {
		f.MOVD(yPtr.At(i), bi)

		if i == 0 {
			mulWord0()
		} else {
			mulWordN()
		}
	}
	f.Comment("reduce if necessary")
	f.SUBS(q[0], t[0], q[0])
	for i := 1; i < f.NbWords; i++ {
		f.SBCS(q[i], t[i], q[i])
	}
	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", q[i], t[i], t[i])
	}

	f.MOVD("res+0(FP)", xPtr)
	f.store(t[:f.NbWords], xPtr)

	f.RET()
}

func (f *FFArm64) reduce(t, q []arm64.Register) {

	if len(t) != f.NbWords || len(q) != f.NbWords {
		panic("need 2*nbWords registers")
	}

	f.Comment("load modulus and subtract")

	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.qAt(i), q[i], q[i+1])
	}
	f.SUBS(q[0], t[0], q[0])
	for i := 1; i < f.NbWords; i++ {
		f.SBCS(q[i], t[i], q[i])
	}

	f.Comment("reduce if necessary")
	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", q[i], t[i], t[i])
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
