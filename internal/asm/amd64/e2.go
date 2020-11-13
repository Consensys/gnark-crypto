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
	. "github.com/consensys/bavard/amd64"
	"github.com/consensys/goff/asm/amd64"
	"github.com/consensys/goff/field"
)

type Fq2Amd64 struct {
	F         *amd64.FFAmd64
	curveName string
}

func NewFq2Amd64(w io.Writer, F *field.Field, curveName string) *Fq2Amd64 {
	ffamd64 := amd64.NewFFAmd64(w, F)
	return &Fq2Amd64{ffamd64, curveName}
}

func (fq2 *Fq2Amd64) Generate() error {
	WriteLn(bavard.Apache2Header("ConsenSys Software Inc.", 2020))

	WriteLn("#include \"textflag.h\"")
	WriteLn("#include \"funcdata.h\"")

	fq2.generateAddE2()
	fq2.generateDoubleE2()
	fq2.generateSubE2()
	fq2.generateNegE2()

	if strings.ToLower(fq2.curveName) == "bn256" {
		fq2.generateMulE2BN256()
		fq2.generateSquareE2BN256()
		fq2.generateMulByNonResidueE2BN256()
	}

	return nil
}

func (fq2 *Fq2Amd64) generateAddE2() {
	stackSize := 0
	if fq2.F.NbWords > amd64.SmallModulus {
		stackSize = fq2.F.NbWords * 8
	}
	registers := FnHeader("addE2", stackSize, 24)

	// registers
	x := registers.Pop()
	y := registers.Pop()
	r := registers.Pop()
	t := registers.PopN(fq2.F.NbWords)

	MOVQ("x+8(FP)", x)

	// move t = x
	fq2.F.Mov(x, t)

	MOVQ("y+16(FP)", y)

	// t = t + y = x + y
	fq2.F.Add(y, t)

	// reduce
	MOVQ("res+0(FP)", r)
	fq2.F.Reduce(&registers, t, r)

	// move x+offset(fq2.F.NbWords) into t
	fq2.F.Mov(x, t, fq2.F.NbWords)

	// add y+offset(fq2.F.NbWords) into t
	fq2.F.Add(y, t, fq2.F.NbWords)

	// reduce t into r with offset fq2.F.NbWords
	fq2.F.Reduce(&registers, t, r, fq2.F.NbWords)

	RET()

}

func (fq2 *Fq2Amd64) generateDoubleE2() {
	// func header
	stackSize := 0
	if fq2.F.NbWords > amd64.SmallModulus {
		stackSize = fq2.F.NbWords * 8
	}
	registers := FnHeader("doubleE2", stackSize, 16)

	// registers
	x := registers.Pop()
	r := registers.Pop()
	t := registers.PopN(fq2.F.NbWords)

	MOVQ("res+0(FP)", r)
	MOVQ("x+8(FP)", x)

	fq2.F.Mov(x, t)
	fq2.F.Add(t, t)
	fq2.F.Reduce(&registers, t, r)
	fq2.F.Mov(x, t, fq2.F.NbWords)
	fq2.F.Add(t, t)
	fq2.F.Reduce(&registers, t, r, fq2.F.NbWords)

	RET()
}

func (fq2 *Fq2Amd64) generateNegE2() {
	registers := FnHeader("negE2", 0, 16)

	nonZeroA := NewLabel()
	nonZeroB := NewLabel()
	B := NewLabel()

	// registers
	x := registers.Pop()
	r := registers.Pop()
	q := registers.Pop()
	t := registers.PopN(fq2.F.NbWords)

	MOVQ("res+0(FP)", r)
	MOVQ("x+8(FP)", x)

	// t = x
	fq2.F.Mov(x, t)

	// x = t[0] | ... | t[n]
	MOVQ(t[0], x)
	for i := 1; i < fq2.F.NbWords; i++ {
		ORQ(t[i], x)
	}

	TESTQ(x, x)

	// if x != 0, we jump to nonzero label
	JNE(nonZeroA)

	// if x == 0, we set the result to zero and continue
	for i := 0; i < fq2.F.NbWords; i++ {
		MOVQ(x, r.At(i+fq2.F.NbWords))
	}
	JMP(B)

	LABEL(nonZeroA)

	// z = x - q
	for i := 0; i < fq2.F.NbWords; i++ {
		MOVQ(fq2.F.Q[i], q)
		if i == 0 {
			SUBQ(t[i], q)
		} else {
			SBBQ(t[i], q)
		}
		MOVQ(q, r.At(i))
	}

	LABEL(B)
	MOVQ("x+8(FP)", x)
	fq2.F.Mov(x, t, fq2.F.NbWords)

	// x = t[0] | ... | t[n]
	MOVQ(t[0], x)
	for i := 1; i < fq2.F.NbWords; i++ {
		ORQ(t[i], x)
	}

	TESTQ(x, x)

	// if x != 0, we jump to nonzero label
	JNE(nonZeroB)

	// if x == 0, we set the result to zero and return
	for i := 0; i < fq2.F.NbWords; i++ {
		MOVQ(x, r.At(i+fq2.F.NbWords))
	}
	RET()

	LABEL(nonZeroB)

	// z = x - q
	for i := 0; i < fq2.F.NbWords; i++ {
		MOVQ(fq2.F.Q[i], q)
		if i == 0 {
			SUBQ(t[i], q)
		} else {
			SBBQ(t[i], q)
		}
		MOVQ(q, r.At(i+fq2.F.NbWords))
	}

	RET()

}

func (fq2 *Fq2Amd64) generateSubE2() {
	registers := FnHeader("subE2", 0, 24)

	// registers
	t := registers.PopN(fq2.F.NbWords)
	x := registers.Pop()
	y := registers.Pop()

	MOVQ("x+8(FP)", x)
	MOVQ("y+16(FP)", y)

	fq2.F.Mov(x, t)

	// z = x - y mod q
	// move t = x
	fq2.F.Sub(y, t)

	if fq2.F.NbWords > 6 {
		fq2.F.ReduceAfterSub(&registers, t, false)
	} else {
		fq2.F.ReduceAfterSub(&registers, t, true)
	}

	r := registers.Pop()
	MOVQ("res+0(FP)", r)
	fq2.F.Mov(t, r)
	registers.Push(r)

	fq2.F.Mov(x, t, fq2.F.NbWords)

	// z = x - y mod q
	// move t = x
	fq2.F.Sub(y, t, fq2.F.NbWords)

	if fq2.F.NbWords > 6 {
		fq2.F.ReduceAfterSub(&registers, t, false)
	} else {
		fq2.F.ReduceAfterSub(&registers, t, true)
	}

	r = x
	MOVQ("res+0(FP)", r)

	fq2.F.Mov(t, r, 0, fq2.F.NbWords)

	RET()

}
