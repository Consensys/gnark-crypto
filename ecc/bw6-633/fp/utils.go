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
		v[9] >>= 1
	} else {
		// v = v + q
		v[0], carry = bits.Add64(v[0], 15512955586897510413, 0)
		v[1], carry = bits.Add64(v[1], 4410884215886313276, carry)
		v[2], carry = bits.Add64(v[2], 15543556715411259941, carry)
		v[3], carry = bits.Add64(v[3], 9083347379620258823, carry)
		v[4], carry = bits.Add64(v[4], 13320134076191308873, carry)
		v[5], carry = bits.Add64(v[5], 9318693926755804304, carry)
		v[6], carry = bits.Add64(v[6], 5645674015335635503, carry)
		v[7], carry = bits.Add64(v[7], 12176845843281334983, carry)
		v[8], carry = bits.Add64(v[8], 18165857675053050549, carry)
		v[9], _ = bits.Add64(v[9], 82862755739295587, carry)
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
		v[9] >>= 1
	}

	z.Set(&v)

	return z
}
