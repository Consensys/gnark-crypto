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

// addVec res = a + b
// func addVec(res, a, b *{{.ElementName}}, n uint64)
func (f *FFAmd64) generateAddVec() {
	f.Comment("addVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] + b[0...n]")

	nbRegisters := f.NbWords*2 + 4
	stackSize := f.StackSize(nbRegisters, 0, 0)
	registers := f.FnHeader("addVec", stackSize, 36)
	defer f.AssertCleanStack(stackSize, 0)

	// registers
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)

	a := f.PopN(&registers)
	t := f.PopN(&registers)

	loop := f.NewLabel()
	done := f.NewLabel()

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done)

	// a = a + b
	f.Mov(addrA, a)
	f.Add(addrB, a)

	// reduce a
	f.ReduceElement(a, t)

	// save a into res
	f.Mov(a, addrRes)

	f.ADDQ("$32", addrA)
	f.ADDQ("$32", addrB)
	f.ADDQ("$32", addrRes)
	f.DECQ(len)
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, a...)
	f.Push(&registers, t...)
	f.Push(&registers, addrA, addrB, addrRes, len)

}
