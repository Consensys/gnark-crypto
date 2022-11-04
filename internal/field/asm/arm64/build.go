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

package arm64

import (
	"fmt"
	"io"
	"strings"

	"github.com/consensys/bavard"

	"github.com/consensys/bavard/arm64"
	"github.com/consensys/gnark-crypto/internal/field"
)

const SmallModulus = 6

func NewFFArm64(w io.Writer, F *field.FieldConfig) *FFArm64 {
	return &FFArm64{F, arm64.NewArm64(w), 0, 0}
}

type FFArm64 struct {
	*field.FieldConfig
	*arm64.Arm64
	nbElementsOnStack int
	maxOnStack        int
}

func (f *FFArm64) StackSize(maxNbRegistersNeeded, nbRegistersReserved, minStackSize int) int {
	got := arm64.NbRegisters - nbRegistersReserved
	r := got - maxNbRegistersNeeded
	if r >= 0 {
		return minStackSize
	}
	r *= -8
	return max(r, minStackSize)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (f *FFArm64) AssertCleanStack(reservedStackSize, minStackSize int) {
	if f.nbElementsOnStack != 0 {
		panic("missing f.Push stack elements")
	}
	if reservedStackSize < minStackSize {
		panic("invalid minStackSize or reservedStackSize")
	}
	usedStackSize := f.maxOnStack * 8
	if usedStackSize > reservedStackSize {
		panic("using more stack size than reserved")
	} else if max(usedStackSize, minStackSize) < reservedStackSize {
		// this panic is for dev purposes as this may be by design for aligment
		panic("reserved more stack size than needed")
	}

	f.maxOnStack = 0
}

func (f *FFArm64) Push(registers *arm64.Registers, rIn ...arm64.Register) {
	for _, r := range rIn {
		if strings.HasPrefix(string(r), "s") {
			// it's on the stack, decrease the offset
			f.nbElementsOnStack--
			continue
		}
		registers.Push(r)
	}
}

func (f *FFArm64) Pop(registers *arm64.Registers, forceStack ...bool) arm64.Register {
	if registers.Available() >= 1 && !(len(forceStack) > 0 && forceStack[0]) {
		return registers.Pop()
	}
	r := arm64.Register(fmt.Sprintf("s%d-%d(SP)", f.nbElementsOnStack, 8+f.nbElementsOnStack*8))
	f.nbElementsOnStack++
	if f.nbElementsOnStack > f.maxOnStack {
		f.maxOnStack = f.nbElementsOnStack
	}
	return r
}

func (f *FFArm64) PopN(registers *arm64.Registers, forceStack ...bool) []arm64.Register {
	if len(forceStack) > 0 && forceStack[0] {
		nbStack := f.NbWords
		var u []arm64.Register

		for i := f.nbElementsOnStack; i < nbStack+f.nbElementsOnStack; i++ {
			u = append(u, arm64.Register(fmt.Sprintf("s%d-%d(SP)", i, 8+i*8)))
		}
		f.nbElementsOnStack += nbStack
		if f.nbElementsOnStack > f.maxOnStack {
			f.maxOnStack = f.nbElementsOnStack
		}
		return u
	}
	if registers.Available() >= f.NbWords {
		return registers.PopN(f.NbWords)
	}
	nbStack := f.NbWords - registers.Available()
	u := registers.PopN(registers.Available())

	for i := f.nbElementsOnStack; i < nbStack+f.nbElementsOnStack; i++ {
		u = append(u, arm64.Register(fmt.Sprintf("s%d-%d(SP)", i, 8+i*8)))
	}
	f.nbElementsOnStack += nbStack
	if f.nbElementsOnStack > f.maxOnStack {
		f.maxOnStack = f.nbElementsOnStack
	}
	return u
}

func (f *FFArm64) qAt(index int) string {
	return fmt.Sprintf("q<>+%d(SB)", index*8)
}

// Generate generates assembly code for the base field provided to goff
// see internal/templates/ops*
func Generate(w io.Writer, F *field.FieldConfig) error {
	f := NewFFArm64(w, F)
	f.WriteLn(bavard.Apache2Header("ConsenSys Software Inc.", 2020))

	f.WriteLn("#include \"textflag.h\"")
	f.WriteLn("#include \"funcdata.h\"")
	f.WriteLn("")

	f.GenerateDefines()

	// butterfly
	f.generateButterfly()

	return nil
}
