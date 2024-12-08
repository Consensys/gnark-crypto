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
	"github.com/consensys/bavard/amd64"
)

// addVec res = a + b
// func addVec(res, a, b *{{.ElementName}}, n uint64)
func (f *FFAmd64) generateAddVecF31() {
	f.Comment("addVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] + b[0...n]")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("addVec", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)

	// AVX512 registers
	a := amd64.Register("Z0")
	b := amd64.Register("Z1")
	t := amd64.Register("Z2")
	q := amd64.Register("Z3")

	// load q in Z3
	f.WriteLn("MOVD $const_q, AX")
	f.VPBROADCASTD("AX", q)

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// a = a + b
	f.VMOVDQU32(addrA.At(0), a)
	f.VMOVDQU32(addrB.At(0), b)
	f.VPADDD(a, b, a)
	// t = a - q
	f.VPSUBD(q, a, t)
	// b = min(t, a)
	f.VPMINUD(a, t, b)

	// move b to res
	f.VMOVDQU32(b, addrRes.At(0))

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$64", addrA)
	f.ADDQ("$64", addrB)
	f.ADDQ("$64", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)

}

// subVec res = a - b
// func subVec(res, a, b *{{.ElementName}}, n uint64)
func (f *FFAmd64) generateSubVecF31() {
	f.Comment("subVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] - b[0...n]")

	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 0, 0)
	registers := f.FnHeader("subVec", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	// registers & labels we need
	addrA := f.Pop(&registers)
	addrB := f.Pop(&registers)
	addrRes := f.Pop(&registers)
	len := f.Pop(&registers)

	// AVX512 registers
	a := amd64.Register("Z0")
	b := amd64.Register("Z1")
	t := amd64.Register("Z2")
	q := amd64.Register("Z3")

	// load q in Z3
	f.WriteLn("MOVD $const_q, AX")
	f.VPBROADCASTD("AX", q)

	loop := f.NewLabel("loop")
	done := f.NewLabel("done")

	// load arguments
	f.MOVQ("res+0(FP)", addrRes)
	f.MOVQ("a+8(FP)", addrA)
	f.MOVQ("b+16(FP)", addrB)
	f.MOVQ("n+24(FP)", len)

	f.LABEL(loop)

	f.TESTQ(len, len)
	f.JEQ(done, "n == 0, we are done")

	// a = a - b
	f.VMOVDQU32(addrA.At(0), a)
	f.VMOVDQU32(addrB.At(0), b)

	f.VPSUBD(b, a, a)

	// t = a + q
	f.VPADDD(q, a, t)

	// b = min(t, a)
	f.VPMINUD(a, t, b)

	// move b to res
	f.VMOVDQU32(b, addrRes.At(0))

	f.Comment("increment pointers to visit next element")
	f.ADDQ("$64", addrA)
	f.ADDQ("$64", addrB)
	f.ADDQ("$64", addrRes)
	f.DECQ(len, "decrement n")
	f.JMP(loop)

	f.LABEL(done)

	f.RET()

	f.Push(&registers, addrA, addrB, addrRes, len)

}

// // subVec res = a - b
// // func subVec(res, a, b *{{.ElementName}}, n uint64)
// func (f *FFAmd64) generateSubVecW4() {
// 	f.Comment("subVec(res, a, b *Element, n uint64) res[0...n] = a[0...n] - b[0...n]")

// 	const argSize = 4 * 8
// 	stackSize := f.StackSize(f.NbWords*2+5, 0, 0)
// 	registers := f.FnHeader("subVec", stackSize, argSize)
// 	defer f.AssertCleanStack(stackSize, 0)

// 	// registers
// 	addrA := f.Pop(&registers)
// 	addrB := f.Pop(&registers)
// 	addrRes := f.Pop(&registers)
// 	len := f.Pop(&registers)
// 	zero := f.Pop(&registers)

// 	a := f.PopN(&registers)
// 	q := f.PopN(&registers)

// 	loop := f.NewLabel("loop")
// 	done := f.NewLabel("done")

// 	// load arguments
// 	f.MOVQ("res+0(FP)", addrRes)
// 	f.MOVQ("a+8(FP)", addrA)
// 	f.MOVQ("b+16(FP)", addrB)
// 	f.MOVQ("n+24(FP)", len)

// 	f.XORQ(zero, zero)

// 	f.LABEL(loop)

// 	f.TESTQ(len, len)
// 	f.JEQ(done, "n == 0, we are done")

// 	// a = a - b
// 	f.LabelRegisters("a", a...)
// 	f.Mov(addrA, a)
// 	f.Sub(addrB, a)
// 	f.WriteLn(fmt.Sprintf("PREFETCHT0 2048(%[1]s)", addrA))
// 	f.WriteLn(fmt.Sprintf("PREFETCHT0 2048(%[1]s)", addrB))

// 	// reduce a
// 	f.Comment("reduce (a-b) mod q")
// 	f.LabelRegisters("q", q...)
// 	for i := 0; i < f.NbWords; i++ {
// 		f.MOVQ(fmt.Sprintf("$const_q%d", i), q[i])
// 	}
// 	for i := 0; i < f.NbWords; i++ {
// 		f.CMOVQCC(zero, q[i])
// 	}
// 	// add registers (q or 0) to a, and set to result
// 	f.Comment("add registers (q or 0) to a, and set to result")
// 	f.Add(q, a)

// 	// save a into res
// 	f.Mov(a, addrRes)

// 	f.Comment("increment pointers to visit next element")
// 	f.ADDQ("$32", addrA)
// 	f.ADDQ("$32", addrB)
// 	f.ADDQ("$32", addrRes)
// 	f.DECQ(len, "decrement n")
// 	f.JMP(loop)

// 	f.LABEL(done)

// 	f.RET()

// 	f.Push(&registers, a...)
// 	f.Push(&registers, q...)
// 	f.Push(&registers, addrA, addrB, addrRes, len, zero)

// }
