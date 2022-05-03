// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://wwwApache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fptower

import (
	"github.com/consensys/gnark-crypto/ecc/bls24-317/fp"
)

// TODO: optimize Frobenius maps with stored coefficients

var _p = fp.Modulus()

// Frobenius sets z in E4 to x^q, returns z
func (z *E4) Frobenius(x *E4) *E4 {
	z.Exp(x, *_p)
	return z
}

// Frobenius set z to Frobenius(x), return z
func (z *E24) Frobenius(x *E24) *E24 {
	z.Exp(x, *_p)
	return z
}

// FrobeniusSquare set z to Frobenius^2(x), return z
func (z *E24) FrobeniusSquare(x *E24) *E24 {
	z.Exp(x, *_p).Exp(z, *_p)
	return z
}

// FrobeniusQuad set z to Frobenius^4(x), return z
func (z *E24) FrobeniusQuad(x *E24) *E24 {
	z.Exp(x, *_p).Exp(z, *_p).Exp(z, *_p).Exp(z, *_p)
	return z
}
