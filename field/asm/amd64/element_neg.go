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

func (f *FFAmd64) generateNeg() {
	f.Comment("neg(res, x *Element)")
	registers := f.FnHeader("neg", 0, 16)
	defer f.AssertCleanStack(0, 0)

	// labels
	zero := f.NewLabel()

	// registers
	x := registers.Pop()
	t := registers.PopN(f.NbWords)
	r := registers.Pop()

	f.MOVQ("res+0(FP)", r)
	f.MOVQ("x+8(FP)", x)

	// t = x
	f.Mov(x, t)

	// x = t[0] | ... | t[n]
	f.MOVQ(t[0], x)
	for i := 1; i < f.NbWords; i++ {
		f.ORQ(t[i], x)
	}

	f.TESTQ(x, x)

	// if x == 0, we jump to nonzero label
	f.JEQ(zero)
	registers.Push(x)
	q := registers.Pop()
	// z = x - q
	for i := 0; i < f.NbWords; i++ {
		f.MOVQ(f.Q[i], q)
		if i == 0 {
			f.SUBQ(t[i], q)
		} else {
			f.SBBQ(t[i], q)
		}
		f.MOVQ(q, r.At(i))
	}
	registers.Push(q)
	f.RET()

	f.LABEL(zero)
	// if x == 0, we set the result to zero and return
	for i := 0; i < f.NbWords; i++ {
		f.MOVQ(x, r.At(i))
	}
	f.RET()

}
