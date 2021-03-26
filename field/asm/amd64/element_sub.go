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

import "github.com/consensys/bavard/amd64"

func (f *FFAmd64) generateSub() {
	f.Comment("sub(res, x, y *Element)")
	registers := f.FnHeader("sub", 0, 24)
	defer f.AssertCleanStack(0, 0)

	// registers
	var zero amd64.Register
	t := registers.PopN(f.NbWords)
	xy := registers.Pop()

	// set a register to zero if needed
	if f.NbWords <= SmallModulus {
		zero = registers.Pop()
		f.XORQ(zero, zero)
	}

	f.MOVQ("x+8(FP)", xy)
	f.Mov(xy, t)

	// z = x - y mod q
	f.MOVQ("y+16(FP)", xy)
	f.Sub(xy, t)
	registers.Push(xy)

	if f.NbWords > SmallModulus {
		noReduce := f.NewLabel()
		f.JCC(noReduce)
		q := registers.Pop()
		f.MOVQ(f.Q[0], q)
		f.ADDQ(q, t[0])
		for i := 1; i < f.NbWords; i++ {
			f.MOVQ(f.Q[i], q)
			f.ADCQ(q, t[i])
		}
		f.LABEL(noReduce)
		registers.Push(q)
	} else {
		q := registers.PopN(f.NbWords)
		f.Mov(f.Q, q)
		for i := 0; i < f.NbWords; i++ {
			f.CMOVQCC(zero, q[i])
		}
		// add registers (q or 0) to t, and set to result
		f.Add(q, t)
		registers.Push(q...)
		registers.Push(zero)
	}

	r := registers.Pop()
	f.MOVQ("res+0(FP)", r)
	f.Mov(t, r)

	f.RET()

}
