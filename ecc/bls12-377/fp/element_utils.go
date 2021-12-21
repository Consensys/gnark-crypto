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

package fp

// MulByNonResidueInv ...
func (z *Element) MulByNonResidueInv(x *Element) *Element {
	qnrInv := Element{
		330620507644336508,
		9878087358076053079,
		11461392860540703536,
		6973035786057818995,
		8846909097162646007,
		104838758629667239,
	}
	z.Mul(x, &qnrInv).
		Neg(z)
	return z
}
