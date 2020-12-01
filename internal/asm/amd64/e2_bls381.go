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

func (fq2 *Fq2Amd64) generateMulByNonResidueE2BLS381() {
	// // MulByNonResidue multiplies a E2 by (1,1)
	// func (z *E2) MulByNonResidue(x *E2) *E2 {
	// 	var a fp.Element
	// 	a.Sub(&x.A0, &x.A1)
	// 	z.A1.Add(&x.A0, &x.A1)
	// 	z.A0.Set(&a)
	// 	return z
	// }
	registers := fq2.FnHeader("mulNonResE2", 0, 16)

	a := registers.PopN(fq2.NbWords)
	x := registers.Pop()

	fq2.MOVQ("x+8(FP)", x)
	fq2.Mov(x, a) // a = a0

	// a = x.A0 - x.A1
	fq2.Sub(x, a, fq2.NbWords)
	fq2.ReduceAfterSub(&registers, a, true)

	// b = x.A0 + x.A1
	b := registers.PopN(fq2.NbWords)
	fq2.Mov(x, b, fq2.NbWords) // b = a1
	fq2.Add(x, b)

	fq2.MOVQ("res+0(FP)", x)
	fq2.Mov(a, x)
	registers.Push(a...)
	fq2.Reduce(&registers, b, b)

	fq2.Mov(b, x, 0, fq2.NbWords)

	fq2.RET()
}
