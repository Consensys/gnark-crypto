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

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fptower

func (z *E12) Select(cond int, caseZ *E12, caseNz *E12) *E12 {
	//Might be able to save a nanosecond or two by an aggregate implementation

	z.C0.Select(cond, &caseZ.C0, &caseNz.C0)
	z.C1.Select(cond, &caseZ.C1, &caseNz.C1)

	return z
}

func (z *E12) Div(x *E12, y *E12) *E12 {
	var r E12
	r.Inverse(y).Mul(x, y)
	return z.Set(&r)
}