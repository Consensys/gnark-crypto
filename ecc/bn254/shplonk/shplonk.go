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
