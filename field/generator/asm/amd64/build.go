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

// Package amd64 contains syntactic sugar to generate amd64 assembly code
package amd64

import (
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/bavard/amd64"
	"github.com/consensys/gnark-crypto/field/generator/config"
)

const SmallModulus = 6
const (
	ElementASMFileName = "element_%dw_amd64.s"
)

func NewFFAmd64(w io.Writer, nbWords int) *FFAmd64 {
	F := &FFAmd64{
		amd64.NewAmd64(w),
		0,
		0,
		nbWords,
		nbWords - 1,
		make([]int, nbWords),
		make([]int, nbWords-1),
		make(map[string]defineFn),
	}

	// indexes (template helpers)
	for i := 0; i < F.NbWords; i++ {
		F.NbWordsIndexesFull[i] = i
		if i > 0 {
			F.NbWordsIndexesNoZero[i-1] = i
		}
	}

	return F
}

type FFAmd64 struct {
	// *config.FieldConfig
	*amd64.Amd64
	nbElementsOnStack    int
	maxOnStack           int
	NbWords              int
	NbWordsLastIndex     int
	NbWordsIndexesFull   []int
	NbWordsIndexesNoZero []int
	mDefines             map[string]defineFn
}

type defineFn func(args ...amd64.Register)

func (f *FFAmd64) StackSize(maxNbRegistersNeeded, nbRegistersReserved, minStackSize int) int {
	got := amd64.NbRegisters - nbRegistersReserved
	r := got - maxNbRegistersNeeded
	if r >= 0 {
		return minStackSize
	}
	r *= -8
	return max(r, minStackSize)
}

func (f *FFAmd64) DefineFn(name string) (fn defineFn, err error) {
	fn, ok := f.mDefines[name]
	if !ok {
		return nil, fmt.Errorf("function %s not defined", name)
	}
	return fn, nil
}

func (f *FFAmd64) Define(name string, nbInputs int, fn defineFn) defineFn {

	inputs := make([]string, nbInputs)
	for i := 0; i < nbInputs; i++ {
		inputs[i] = fmt.Sprintf("in%d", i)
	}
	name = strings.ToUpper(name)

	for _, ok := f.mDefines[name]; ok; {
		// name already exist, for code generation purpose we add a suffix
		// should happen only with e2 deprecated functions
		fmt.Println("WARNING: function name already defined, adding suffix")
		i := 0
		for {
			newName := fmt.Sprintf("%s_%d", name, i)
			if _, ok := f.mDefines[newName]; !ok {
				name = newName
				goto startDefine
			}
			i++
		}
	}
startDefine:

	f.StartDefine()
	f.WriteLn("#define " + name + "(" + strings.Join(inputs, ", ") + ")")
	inputsRegisters := make([]amd64.Register, nbInputs)
	for i := 0; i < nbInputs; i++ {
		inputsRegisters[i] = amd64.Register(inputs[i])
	}
	fn(inputsRegisters...)
	f.EndDefine()
	f.WriteLn("")

	toReturn := func(args ...amd64.Register) {
		if len(args) != nbInputs {
			panic("invalid number of arguments")
		}
		inputsStr := make([]string, len(args))
		for i := 0; i < len(args); i++ {
			inputsStr[i] = string(args[i])
		}
		f.WriteLn(name + "(" + strings.Join(inputsStr, ", ") + ")")
	}

	f.mDefines[name] = toReturn

	return toReturn
}

func (f *FFAmd64) AssertCleanStack(reservedStackSize, minStackSize int) {
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

func (f *FFAmd64) Push(registers *amd64.Registers, rIn ...amd64.Register) {
	for _, r := range rIn {
		if strings.HasPrefix(string(r), "s") {
			// it's on the stack, decrease the offset
			f.nbElementsOnStack--
			continue
		}
		registers.Push(r)
	}
}

func (f *FFAmd64) Pop(registers *amd64.Registers, forceStack ...bool) amd64.Register {
	if registers.Available() >= 1 && !(len(forceStack) > 0 && forceStack[0]) {
		return registers.Pop()
	}
	r := amd64.Register(fmt.Sprintf("s%d-%d(SP)", f.nbElementsOnStack, 8+f.nbElementsOnStack*8))
	f.nbElementsOnStack++
	if f.nbElementsOnStack > f.maxOnStack {
		f.maxOnStack = f.nbElementsOnStack
	}
	return r
}

func (f *FFAmd64) PopN(registers *amd64.Registers, forceStack ...bool) []amd64.Register {
	if len(forceStack) > 0 && forceStack[0] {
		nbStack := f.NbWords
		var u []amd64.Register

		for i := f.nbElementsOnStack; i < nbStack+f.nbElementsOnStack; i++ {
			u = append(u, amd64.Register(fmt.Sprintf("s%d-%d(SP)", i, 8+i*8)))
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
		u = append(u, amd64.Register(fmt.Sprintf("s%d-%d(SP)", i, 8+i*8)))
	}
	f.nbElementsOnStack += nbStack
	if f.nbElementsOnStack > f.maxOnStack {
		f.maxOnStack = f.nbElementsOnStack
	}
	return u
}

func (f *FFAmd64) qAt(index int) string {
	return fmt.Sprintf("·qElement+%d(SB)", index*8)
}

func (f *FFAmd64) qAt_bcst(index int) string {
	return fmt.Sprintf("·qElement+%d(SB)", index*4)
}

func (f *FFAmd64) qInv0() string {
	return "$const_qInvNeg"
}

func (f *FFAmd64) mu() string {
	return "$const_mu"
}

func GenerateFieldWrapper(w io.Writer, F *config.FieldConfig, asmDirBuildPath, asmDirIncludePath string) error {
	// for each field we generate the defines for the modulus and the montgomery constant
	f := NewFFAmd64(w, F.NbWords)

	// we add the defines first, then the common asm, then the global variable section
	// to enable correct compilations with #include in order.
	f.WriteLn("")

	hashAndInclude := func(fileName string) error {
		// we hash the file content and include the hash in comment of the generated file
		// to force the Go compiler to recompile the file if the content has changed
		fData, err := os.ReadFile(filepath.Join(asmDirBuildPath, fileName))
		if err != nil {
			return err
		}
		// hash the file using FNV
		hasher := fnv.New64()
		hasher.Write(fData)
		hash := hasher.Sum64()

		f.WriteLn("// Code generated by gnark-crypto/generator. DO NOT EDIT.")
		f.WriteLn(fmt.Sprintf("// We include the hash to force the Go compiler to recompile: %d", hash))
		includePath := filepath.Join(asmDirIncludePath, fileName)
		// on windows, we replace the "\" by "/"
		if filepath.Separator == '\\' {
			includePath = strings.ReplaceAll(includePath, "\\", "/")
		}
		f.WriteLn(fmt.Sprintf("#include \"%s\"\n", includePath))

		return nil
	}

	toInclude := fmt.Sprintf(ElementASMFileName, F.NbWords)
	if err := hashAndInclude(toInclude); err != nil {
		return err
	}

	return nil
}

// GenerateCommonASM generates assembly code for the base field provided to goff
// see internal/templates/ops*
func GenerateCommonASM(w io.Writer, nbWords int, hasVector bool) error {
	f := NewFFAmd64(w, nbWords)
	f.Comment("Code generated by gnark-crypto/generator. DO NOT EDIT.")

	f.WriteLn("#include \"textflag.h\"")
	f.WriteLn("#include \"funcdata.h\"")
	f.WriteLn("#include \"go_asm.h\"")
	f.WriteLn("")

	if nbWords == 1 {
		return GenerateF31ASM(f, hasVector)
	}

	f.GenerateReduceDefine()

	// reduce
	f.generateReduce()

	// mul by constants
	f.generateMulBy3()
	f.generateMulBy5()
	f.generateMulBy13()

	// fft butterflies
	f.generateButterfly()

	// mul
	f.generateMul(false)

	// from mont
	f.generateFromMont(false)

	if hasVector {
		f.WriteLn("")
		f.Comment("Vector operations are partially derived from Dag Arne Osvik's work in github.com/a16z/vectorized-fields")
		f.WriteLn("")

		f.generateAddVec()
		f.generateSubVec()
		f.generateSumVec()
		f.generateInnerProduct()
		f.generateMulVec("scalarMulVec")
		f.generateMulVec("mulVec")
	}

	return nil
}

func GenerateF31ASM(f *FFAmd64, hasVector bool) error {
	f.Comment("TODO: implement F31 assembly code")
	return nil
}
