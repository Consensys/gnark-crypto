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
		v[4] >>= 1
	} else {
		// v = v + q
		v[0], carry = bits.Add64(v[0], 8063698428123676673, 0)
		v[1], carry = bits.Add64(v[1], 4764498181658371330, carry)
		v[2], carry = bits.Add64(v[2], 16051339359738796768, carry)
		v[3], carry = bits.Add64(v[3], 15273757526516850351, carry)
		v[4], _ = bits.Add64(v[4], 342900304943437392, carry)
		// v = v >> 1
		v[0] = v[0]>>1 | v[1]<<63
		v[1] = v[1]>>1 | v[2]<<63
		v[2] = v[2]>>1 | v[3]<<63
		v[3] = v[3]>>1 | v[4]<<63
		v[4] >>= 1
	}

	z.Set(&v)

	return z
}
