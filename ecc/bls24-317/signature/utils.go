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

package signature

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
)

// HashToInt converts a hash value to an integer. Per FIPS 186-4, Section 6.4,
// we use the left-most bits of the hash to match the bit-length of the order of
// the curve. This also performs Step 5 of SEC 1, Version 2.0, Section 4.1.3.
func HashToInt(hash []byte) *big.Int {
	if len(hash) > fr.Bytes {
		hash = hash[:fr.Bytes]
	}
	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - fr.Bytes
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}
