// Copyright 2020 ConsenSys Software Inc.
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

package amd64

import (
	"io"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/goff/asm/amd64"
	"github.com/consensys/goff/field"
)

// Fq2Amd64 ...
type Fq2Amd64 struct {
	*amd64.FFAmd64
	curveName string
}

// NewFq2Amd64 ...
func NewFq2Amd64(w io.Writer, F *field.Field, curveName string) *Fq2Amd64 {
	return &Fq2Amd64{amd64.NewFFAmd64(w, F), curveName}
}

// Generate ...
func (fq2 *Fq2Amd64) Generate() error {
	fq2.WriteLn(bavard.Apache2Header("ConsenSys Software Inc.", 2020))

	fq2.WriteLn("#include \"textflag.h\"")
	fq2.WriteLn("#include \"funcdata.h\"")

	fq2.generateAddE2()
	fq2.generateDoubleE2()
	fq2.generateSubE2()
	fq2.generateNegE2()

	switch strings.ToLower(fq2.curveName) {
	case "bn256":
		fq2.generateMulE2BN256()
		fq2.generateSquareE2BN256()
		fq2.generateMulByNonResidueE2BN256()
	case "bls381":
		fq2.generateMulByNonResidueE2BLS381()
		fq2.generateSquareE2BLS381()
	}

	return nil
}

func (fq2 *Fq2Amd64) generateAddE2() {
	stackSize := 0
	if fq2.NbWords > amd64.SmallModulus {
		stackSize = fq2.NbWords * 8
	}
	registers := fq2.FnHeader("addE2", stackSize, 24)

	// registers
	x := registers.Pop()
	y := registers.Pop()
	r := registers.Pop()
	t := registers.PopN(fq2.NbWords)

	fq2.MOVQ("x+8(FP)", x)

	// move t = x
	fq2.Mov(x, t)

	fq2.MOVQ("y+16(FP)", y)

	// t = t + y = x + y
	fq2.Add(y, t)

	// reduce
	fq2.MOVQ("res+0(FP)", r)
	fq2.Reduce(&registers, t, r)

	// move x+offset(fq2.NbWords) into t
	fq2.Mov(x, t, fq2.NbWords)

	// add y+offset(fq2.NbWords) into t
	fq2.Add(y, t, fq2.NbWords)

	// reduce t into r with offset fq2.NbWords
	fq2.Reduce(&registers, t, r, fq2.NbWords)

	fq2.RET()

}

func (fq2 *Fq2Amd64) generateDoubleE2() {
	// func header
	stackSize := 0
	if fq2.NbWords > amd64.SmallModulus {
		stackSize = fq2.NbWords * 8
	}
	registers := fq2.FnHeader("doubleE2", stackSize, 16)

	// registers
	x := registers.Pop()
	r := registers.Pop()
	t := registers.PopN(fq2.NbWords)

	fq2.MOVQ("res+0(FP)", r)
	fq2.MOVQ("x+8(FP)", x)

	fq2.Mov(x, t)
	fq2.Add(t, t)
	fq2.Reduce(&registers, t, r)
	fq2.Mov(x, t, fq2.NbWords)
	fq2.Add(t, t)
	fq2.Reduce(&registers, t, r, fq2.NbWords)

	fq2.RET()
}

func (fq2 *Fq2Amd64) generateNegE2() {
	registers := fq2.FnHeader("negE2", 0, 16)

	nonZeroA := fq2.NewLabel()
	nonZeroB := fq2.NewLabel()
	B := fq2.NewLabel()

	// registers
	x := registers.Pop()
	r := registers.Pop()
	q := registers.Pop()
	t := registers.PopN(fq2.NbWords)

	fq2.MOVQ("res+0(FP)", r)
	fq2.MOVQ("x+8(FP)", x)

	// t = x
	fq2.Mov(x, t)

	// x = t[0] | ... | t[n]
	fq2.MOVQ(t[0], x)
	for i := 1; i < fq2.NbWords; i++ {
		fq2.ORQ(t[i], x)
	}

	fq2.TESTQ(x, x)

	// if x != 0, we jump to nonzero label
	fq2.JNE(nonZeroA)

	// if x == 0, we set the result to zero and continue
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(x, r.At(i+fq2.NbWords))
	}
	fq2.JMP(B)

	fq2.LABEL(nonZeroA)

	// z = x - q
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(fq2.Q[i], q)
		if i == 0 {
			fq2.SUBQ(t[i], q)
		} else {
			fq2.SBBQ(t[i], q)
		}
		fq2.MOVQ(q, r.At(i))
	}

	fq2.LABEL(B)
	fq2.MOVQ("x+8(FP)", x)
	fq2.Mov(x, t, fq2.NbWords)

	// x = t[0] | ... | t[n]
	fq2.MOVQ(t[0], x)
	for i := 1; i < fq2.NbWords; i++ {
		fq2.ORQ(t[i], x)
	}

	fq2.TESTQ(x, x)

	// if x != 0, we jump to nonzero label
	fq2.JNE(nonZeroB)

	// if x == 0, we set the result to zero and return
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(x, r.At(i+fq2.NbWords))
	}
	fq2.RET()

	fq2.LABEL(nonZeroB)

	// z = x - q
	for i := 0; i < fq2.NbWords; i++ {
		fq2.MOVQ(fq2.Q[i], q)
		if i == 0 {
			fq2.SUBQ(t[i], q)
		} else {
			fq2.SBBQ(t[i], q)
		}
		fq2.MOVQ(q, r.At(i+fq2.NbWords))
	}

	fq2.RET()

}

func (fq2 *Fq2Amd64) generateSubE2() {
	registers := fq2.FnHeader("subE2", 0, 24)

	// registers
	t := registers.PopN(fq2.NbWords)
	x := registers.Pop()
	y := registers.Pop()

	fq2.MOVQ("x+8(FP)", x)
	fq2.MOVQ("y+16(FP)", y)

	fq2.Mov(x, t)

	// z = x - y mod q
	// move t = x
	fq2.Sub(y, t)

	if fq2.NbWords > 6 {
		fq2.ReduceAfterSub(&registers, t, false)
	} else {
		fq2.ReduceAfterSub(&registers, t, true)
	}

	r := registers.Pop()
	fq2.MOVQ("res+0(FP)", r)
	fq2.Mov(t, r)
	registers.Push(r)

	fq2.Mov(x, t, fq2.NbWords)

	// z = x - y mod q
	// move t = x
	fq2.Sub(y, t, fq2.NbWords)

	if fq2.NbWords > 6 {
		fq2.ReduceAfterSub(&registers, t, false)
	} else {
		fq2.ReduceAfterSub(&registers, t, true)
	}

	r = x
	fq2.MOVQ("res+0(FP)", r)

	fq2.Mov(t, r, 0, fq2.NbWords)

	fq2.RET()

}
