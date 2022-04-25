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

package fptower

import "github.com/consensys/gnark-crypto/ecc/bls12-39/fp"

var _p = fp.Modulus()

// TODO: optimize

// Frobenius set z to Frobenius(x), return z
func (z *E12) Frobenius(x *E12) *E12 {
	z.Exp(x, *_p)
	return z
}

// FrobeniusSquare set z to Frobenius^2(x), and return z
func (z *E12) FrobeniusSquare(x *E12) *E12 {
	z.Exp(x, *_p).Exp(z, *_p)
	return z
}

// FrobeniusCube set z to Frobenius^3(x), return z
func (z *E12) FrobeniusCube(x *E12) *E12 {
	z.Exp(x, *_p).Exp(z, *_p).Exp(z, *_p)
	return z
}
