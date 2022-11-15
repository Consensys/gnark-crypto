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
// limitations under the License.s

package utils

import "math/big"

// Decompose interpret rawBytes as a bigInt x in big endian,
// and returns the digits of x (from LSB to MSB) when x is written
// in basis modulo.
func Decompose(rawBytes []byte, modulo *big.Int) (decomposed []byte) {
	raw := big.NewInt(0).SetBytes(rawBytes)

	var chunk [32]byte
	decomposed = make([]byte, 0, len(rawBytes))
	for raw.Cmp(modulo) >= 0 {
		mod := big.NewInt(0).Mod(raw, modulo)
		mod.FillBytes(chunk[:])
		decomposed = append(decomposed, chunk[:]...)

		raw.Div(raw, modulo)
	}

	raw.FillBytes(chunk[:])
	decomposed = append(decomposed, chunk[:]...)
	return decomposed
}
