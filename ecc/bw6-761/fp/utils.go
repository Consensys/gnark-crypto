// Copyright 2020 ConsenSys AG
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

import "math/bits"

// Halve divides a fp.Element by 2
func (z *Element) Halve(x *Element) *Element {

	v := *x

	var carry uint64
	if v[0]&1 == 0 {
		// v = v >> 1
		v[0] = v[0]>>1 | v[1]<<63
		v[1] = v[1]>>1 | v[2]<<63
		v[2] = v[2]>>1 | v[3]<<63
		v[3] = v[3]>>1 | v[4]<<63
		v[4] = v[4]>>1 | v[5]<<63
		v[5] = v[5]>>1 | v[6]<<63
		v[6] = v[6]>>1 | v[7]<<63
		v[7] = v[7]>>1 | v[8]<<63
		v[8] = v[8]>>1 | v[9]<<63
		v[9] = v[9]>>1 | v[10]<<63
		v[10] = v[10]>>1 | v[11]<<63
		v[11] >>= 1
	} else {
		// v = v + q
		v[0], carry = bits.Add64(v[0], 17626244516597989515, 0)
		v[1], carry = bits.Add64(v[1], 16614129118623039618, carry)
		v[2], carry = bits.Add64(v[2], 1588918198704579639, carry)
		v[3], carry = bits.Add64(v[3], 10998096788944562424, carry)
		v[4], carry = bits.Add64(v[4], 8204665564953313070, carry)
		v[5], carry = bits.Add64(v[5], 9694500593442880912, carry)
		v[6], carry = bits.Add64(v[6], 274362232328168196, carry)
		v[7], carry = bits.Add64(v[7], 8105254717682411801, carry)
		v[8], carry = bits.Add64(v[8], 5945444129596489281, carry)
		v[9], carry = bits.Add64(v[9], 13341377791855249032, carry)
		v[10], carry = bits.Add64(v[10], 15098257552581525310, carry)
		v[11], _ = bits.Add64(v[11], 81882988782276106, carry)
		// v = v >> 1
		v[0] = v[0]>>1 | v[1]<<63
		v[1] = v[1]>>1 | v[2]<<63
		v[2] = v[2]>>1 | v[3]<<63
		v[3] = v[3]>>1 | v[4]<<63
		v[4] = v[4]>>1 | v[5]<<63
		v[5] = v[5]>>1 | v[6]<<63
		v[6] = v[6]>>1 | v[7]<<63
		v[7] = v[7]>>1 | v[8]<<63
		v[8] = v[8]>>1 | v[9]<<63
		v[9] = v[9]>>1 | v[10]<<63
		v[10] = v[10]>>1 | v[11]<<63
		v[11] >>= 1
	}

	z.Set(&v)

	return z
}
