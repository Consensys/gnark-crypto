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

func (f *FFArm64) generateButterfly() {
	const argSize = 2 * 8
	f.Comment("Butterfly(a, b *Element) sets a = a + b; b = a - b")

	stackSize := f.StackSize(f.NbWords*2, 0, 0)
	registers := f.FnHeader("Butterfly", stackSize, argSize)
	defer f.AssertCleanStack(stackSize, 0)

	if stackSize > 0 {
		f.WriteLn("NO_LOCAL_POINTERS")
	}

	a := registers.PopN(f.NbWords)
	b := registers.PopN(f.NbWords)
	q := registers.PopN(f.NbWords)
	t := registers.PopN(f.NbWords)

	f.LabelRegisters("a", a...)
	f.LabelRegisters("b", a...)
	f.LabelRegisters("t", t...)

	// we store x fully in registers
	ra := registers.Pop()
	rb := registers.Pop()
	f.LDP("a+0(FP)", ra, rb)
	f.Mov(ra, a)
	f.Mov(a, t)
	f.Mov(rb, b)

	// registers
	f.LabelRegisters("q", q...)
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.qAt(i), q[i], q[i+1])
	}
	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(f.qAt(i), q[i])
	}

	// a = a + b
	f.ADDS(a[0], b[0], a[0])
	for i := 1; i < f.NbWords; i++ {
		f.ADCS(a[i], b[i], a[i])
	}

	// a = a mod q
	// reduce
	f.SUBS(q[0], a[0], q[0])
	for i := 1; i < f.NbWords; i++ {
		f.SBCS(q[i], a[i], q[i])
	}

	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", q[i], a[i], a[i])
	}

	// restore q
	for i := 0; i < f.NbWords-1; i += 2 {
		f.LDP(f.qAt(i), q[i], q[i+1])
	}
	if f.NbWords%2 == 1 {
		i := f.NbWords - 1
		f.MOVD(f.qAt(i), q[i])
	}

	// save a
	f.Mov(a, ra)

	// b = t - b
	f.SUBS(b[0], t[0], b[0])
	for i := 1; i < f.NbWords; i++ {
		f.SBCS(b[i], t[i], b[i])
	}

	for i := 0; i < f.NbWords; i++ {
		f.CSEL("CS", "ZR", q[i], a[i])
	}

	// b = b mod q
	// reduce
	f.ADDS(a[0], b[0], b[0])
	for i := 1; i < f.NbWords; i++ {
		f.ADCS(a[i], b[i], b[i])
	}

	f.Mov(b, rb)

	// for i=0 to N-1
	// 		for j=0 to N-1
	// 		    (A,t[j])  := t[j] + x[j]*y[i%2] + A
	// 		m := t[0]*q'[0] mod W
	// 		C,_ := t[0] + m*q[0]
	// 		for j=1 to N-1
	// 		    (C,t[j-1]) := t[j] + m*q[j] + C
	// 		t[N-1] = C + A

	f.RET()

}
