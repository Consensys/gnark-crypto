// Copyright 2020 Consensys Software Inc.
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

//cf https://eprint.iacr.org/2020/081.pdf

package shplonk

import (
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/kzg"
)

// OpeningProof KZG proof for opening (fᵢ)_{i} at a different points (xᵢ)_{i}.
//
// implements io.ReaderFrom and io.WriterTo
type OpeningProof struct {

	// W = ∑ᵢ γⁱZ_{T\xᵢ}(f_i(X)-f(x_i))
	W bn254.G1Affine

	// (fᵢ(xᵢ))_{i}
	ClaimedValues []fr.Element
}

func BatchOpen(p [][]fr.Element, points []fr.Element, pk kzg.ProvingKey) {}

// ------------------------------
// utils

func eval(f []fr.Element, x fr.Element) fr.Element {
	var y fr.Element
	for i := len(f) - 1; i >= 0; i-- {
		y.Mul(&y, &x).Add(&y, &f[i])
	}
	return y
}

// computes f <- (x-a)*f (in place if the capacity of f is correctly set)
func multiplyLinearFactor(f []fr.Element, a fr.Element) []fr.Element {
	s := len(f)
	var tmp fr.Element
	f = append(f, fr.NewElement(0))
	f[s] = f[s-1]
	for i := s - 1; i >= 1; i-- {
		tmp.Mul(&f[i], &a)
		f[i].Sub(&f[i-1], &tmp)
	}
	f[0].Mul(&f[0], &a).Neg(&f[0])
	return f
}

// returns πᵢ(X-xᵢ)
func buildVanishingPoly(x []fr.Element) []fr.Element {
	res := make([]fr.Element, 1, len(x)+1)
	for i := 0; i < len(x); i++ {
		res = multiplyLinearFactor(res, x[i])
	}
	return res
}

// returns f/g (assuming g divides f)
// OK to not use fft if deg(g) is small
// g's leading coefficient is assumed to be 1
// f memory is re-used for the result
func div(f, g []fr.Element) []fr.Element {
	sizef := len(f)
	sizeg := len(g)
	stop := sizeg - +1
	var t fr.Element
	for i := sizef - 2; i >= stop; i-- {
		for j := 0; j < sizeg-1; j++ {
			t.Mul(&f[i+1], &g[sizeg-2-j])
			f[i-j].Sub(&f[i-j], &t)
		}
	}
	return f[sizeg-1:]
}
