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

func (f *FFArm64) generateMul() {
	f.Comment("mul(res, x, y *Element)")
	f.Comment("Algorithm 2 of Faster Montgomery Multiplication and Multi-Scalar-Multiplication for SNARKS")
	f.Comment("by Y. El Housni and G. Botrel https://doi.org/10.46586/tches.v2023.i3.504-521")
	registers := f.FnHeader("mul", 0, 24)
	defer f.AssertCleanStack(0, 0)

	fatModulus := f.NbWords > 6

	xPtr := registers.Pop()
	// yPtr := registers.Pop()
	bi := registers.Pop()
	a := registers.PopN(f.NbWords)
	t := registers.PopN(f.NbWords + 1)

	var qInv0, m, yPtr arm64.Register
	var q []arm64.Register
	ax := xPtr
	if fatModulus {
		qInv0 = ax
		m = bi
		yPtr = bi
	} else {
		qInv0 = registers.Pop()
		m = registers.Pop()
		q = registers.PopN(f.NbWords)
		yPtr = registers.Pop()
	}

	qAt := func(i int) arm64.Register {
		if !fatModulus {
			return q[i]
		}
		f.MOVD(f.qAt(i), ax)
		return ax
	}

	divShift := f.Define("divShift", 0, func(args ...arm64.Register) {
		if fatModulus {
			f.MOVD(f.qInv0(), qInv0)
			f.MUL(t[0], qInv0, m)
		}
		// for j=0 to N-1
		//	(C,t[j-1]) := t[j] + m*q[j] + C

		for j := 0; j < f.NbWords; j++ {
			f.MUL(qAt(j), m, ax)
			f.add0m(j)(ax, t[j], t[j])
		}
		f.add0m(f.NbWords)(t[f.NbWords], "ZR", t[f.NbWords])

		// propagate high bits
		f.UMULH(qAt(0), m, ax)
		for j := 1; j <= f.NbWords; j++ {
			f.add1m(j, true)(ax, t[j], t[j-1])
			if j != f.NbWords {
				f.UMULH(qAt(j), m, ax)
			}
		}
	})

	mulWordN := f.Define("MUL_WORD_N", 0, func(args ...arm64.Register) {
		// for j=0 to N-1
		//    (C,t[j])  := t[j] + a[j]*b[i] + C

		// lo bits
		for j := 0; j < f.NbWords; j++ {
			f.MUL(a[j], bi, ax)
			f.add0m(j)(ax, t[j], t[j])

			if j == 0 && !fatModulus {
				f.MUL(t[0], qInv0, m)
			}
		}
		f.add0m(f.NbWords)("ZR", "ZR", t[f.NbWords])

		// propagate high bits
		f.UMULH(a[0], bi, ax)
		for j := 1; j <= f.NbWords; j++ {
			f.add1m(j)(ax, t[j], t[j])
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
		for j := 1; j < f.NbWords; j++ {
			f.add1m(j)(ax, t[j], t[j])
			f.UMULH(a[j], bi, ax)
		}
		f.add1m(f.NbWords)(ax, "ZR", t[f.NbWords])
		if !fatModulus {
			f.MUL(t[0], qInv0, m)
		}
		divShift()
	})

	f.MOVD("y+16(FP)", yPtr)
	f.MOVD("x+8(FP)", xPtr)
	f.load(xPtr, a)

	for i := 0; i < f.NbWords; i++ {
		if fatModulus {
			f.MOVD("y+16(FP)", yPtr)
		}
		f.MOVD(yPtr.At(i), bi)

		if i == 0 {
			// load qInv0 and q at first iteration.
			if !fatModulus {
				f.MOVD(f.qInv0(), qInv0)
				for i := 0; i < f.NbWords; i += 2 {
					f.LDP(f.qAt(i), q[i], q[i+1])
				}
			}
			mulWord0()
		} else {
			mulWordN()
		}
	}

	if fatModulus {
		q = a
		for i := 0; i < f.NbWords; i += 2 {
			f.LDP(f.qAt(i), q[i], q[i+1])
		}
	}

	f.Comment("reduce if necessary")
	f.SUBS(q[0], t[0], q[0])
	for i := 1; i < f.NbWords; i++ {
		f.SBCS(q[i], t[i], q[i])
	}

	f.MOVD("res+0(FP)", ax)
	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", q[i], t[i], t[i])
		if i%2 == 1 {
			f.STP(t[i-1], t[i], ax.At(i-1))
		}
	}

	f.RET()
}

func (f *FFArm64) generateReduce() {
	f.Comment("reduce(res *Element)")
	registers := f.FnHeader("reduce", 0, 8)
	defer f.AssertCleanStack(0, 0)

	// registers
	t := registers.PopN(f.NbWords)
	q := registers.PopN(f.NbWords)
	rPtr := registers.Pop()

	for i := 0; i < f.NbWords; i += 2 {
		f.LDP(f.qAt(i), q[i], q[i+1])
	}

	f.MOVD("res+0(FP)", rPtr)
	f.load(rPtr, t)
	f.reduceAndStore(t, q, rPtr)

	f.RET()
}

func (f *FFArm64) load(zPtr arm64.Register, z []arm64.Register) {
	for i := 0; i < f.NbWords; i += 2 {
		f.LDP(zPtr.At(i), z[i], z[i+1])
	}
}

// q must contain the modulus
// q is modified
// t = t mod q (t must be less than 2q)
// t is stored in zPtr
func (f *FFArm64) reduceAndStore(t, q []arm64.Register, zPtr arm64.Register) {
	f.Comment("q = t - q")
	f.SUBS(q[0], t[0], q[0])
	for i := 1; i < f.NbWords; i++ {
		f.SBCS(q[i], t[i], q[i])
	}

	f.Comment("if no borrow, return q, else return t")
	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", q[i], t[i], t[i])
		if i%2 == 1 {
			f.STP(t[i-1], t[i], zPtr.At(i-1))
		}
	}
}

func (f *FFArm64) add0n(i int) func(op1, op2, dst interface{}, comment ...string) {
	switch {
	case i == 0:
		return f.ADDS
	case i == f.NbWordsLastIndex:
		return f.ADC
	default:
		return f.ADCS
	}
}

func (f *FFArm64) add0m(i int) func(op1, op2, dst interface{}, comment ...string) {
	switch {
	case i == 0:
		return f.ADDS
	case i == f.NbWordsLastIndex+1:
		return f.ADC
	default:
		return f.ADCS
	}
}

func (f *FFArm64) add1m(i int, dumb ...bool) func(op1, op2, dst interface{}, comment ...string) {
	switch {
	case i == 1:
		return f.ADDS
	case i == f.NbWordsLastIndex+1:
		if len(dumb) == 1 && dumb[0] {
			// odd, but it performs better on c8g instances.
			return f.ADCS
		}
		return f.ADC
	default:
		return f.ADCS
	}
}
