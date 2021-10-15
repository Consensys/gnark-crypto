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
		v[5] >>= 1
	} else {
		// v = v + q
		v[0], carry = bits.Add64(v[0], 13402431016077863595, 0)
		v[1], carry = bits.Add64(v[1], 2210141511517208575, carry)
		v[2], carry = bits.Add64(v[2], 7435674573564081700, carry)
		v[3], carry = bits.Add64(v[3], 7239337960414712511, carry)
		v[4], carry = bits.Add64(v[4], 5412103778470702295, carry)
		v[5], _ = bits.Add64(v[5], 1873798617647539866, carry)
		// v = v >> 1
		v[0] = v[0]>>1 | v[1]<<63
		v[1] = v[1]>>1 | v[2]<<63
		v[2] = v[2]>>1 | v[3]<<63
		v[3] = v[3]>>1 | v[4]<<63
		v[4] = v[4]>>1 | v[5]<<63
		v[5] >>= 1
	}

	z.Set(&v)

	return z
}
