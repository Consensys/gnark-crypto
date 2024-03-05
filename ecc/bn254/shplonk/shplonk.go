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
	"errors"
	"hash"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
)

var (
	ErrInvalidNumberOfPoints = errors.New("number of digests should be equal to the number of points")
)

// OpeningProof KZG proof for opening (fᵢ)_{i} at a different points (xᵢ)_{i}.
//
// implements io.ReaderFrom and io.WriterTo
type OpeningProof struct {

	// W = ∑ᵢ γⁱZ_{T\xᵢ}(f_i(X)-f(x_i)) where Z_{T} is the vanishing polynomial on the (xᵢ)_{i}
	W bn254.G1Affine

	// L(X)/(X-z) where L(X)=∑ᵢγⁱZ_{T\xᵢ}(f_i(X)-rᵢ) - Z_{T}W(X)
	WPrime bn254.G1Affine

	// (fᵢ(xᵢ))_{i}
	ClaimedValues []fr.Element
}

// func BatchOpen(polynomials [][]fr.Element, digests []kzg.Digest, points []fr.Element, hf hash.Hash, pk kzg.ProvingKey, dataTranscript ...[]byte) (OpeningProof, error) {

// 	var res OpeningProof

// 	if len(polynomials) != len(points) {
// 		return res, ErrInvalidNumberOfPoints
// 	}

// 	// derive γ
// 	gamma, err := deriveGamma(points, digests, hf, dataTranscript...)
// 	if err != nil {
// 		return res, err
// 	}

// 	// compute the claimed evaluations
// 	maxSize := len(polynomials[0])
// 	for i := 1; i < len(polynomials); i++ {
// 		if maxSize < len(polynomials[i]) {
// 			maxSize = len(polynomials[i])
// 		}
// 	}

// 	totalSize := maxSize + len(points) // maxSize+len(points)-1 is the max degree among the polynomials Z_{T\xᵢ}fᵢ
// 	buf := make([]fr.Element, totalSize)
// 	f := make([]fr.Element, totalSize)
// 	copy(buf, polynomials[0])
// 	v := buildVanishingPoly(points[1:])

// 	for i := 1; i<len(polynomials); i++ {

// 	}

// 	// derive z

// 	return res, nil
// }

// BatchVerify uses proof to check that the commitments correctly open to proof.ClaimedValues
// at points. The order mattes: the proof validates that the i-th commitment is correctly opened
// at the i-th point
func BatchVerify(proof OpeningProof, commitments []kzg.Digest, points []fr.Element) error {

	// compute γ

	return nil
}

// deriveGamma derives a challenge using Fiat Shamir to polynomials.
func deriveGamma(points []fr.Element, digests []kzg.Digest, hf hash.Hash, dataTranscript ...[]byte) (fr.Element, error) {

	// derive the challenge gamma, binded to the point and the commitments
	fs := fiatshamir.NewTranscript(hf, "gamma")
	for i := range points {
		if err := fs.Bind("gamma", points[i].Marshal()); err != nil {
			return fr.Element{}, err
		}
	}
	for i := range digests {
		if err := fs.Bind("gamma", digests[i].Marshal()); err != nil {
			return fr.Element{}, err
		}
	}

	for i := 0; i < len(dataTranscript); i++ {
		if err := fs.Bind("gamma", dataTranscript[i]); err != nil {
			return fr.Element{}, err
		}
	}

	gammaByte, err := fs.ComputeChallenge("gamma")
	if err != nil {
		return fr.Element{}, err
	}
	var gamma fr.Element
	gamma.SetBytes(gammaByte)

	return gamma, nil
}

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
	res[0].SetOne()
	for i := 0; i < len(x); i++ {
		res = multiplyLinearFactor(res, x[i])
	}
	return res
}

// returns f*g using naive multiplication
// deg(big)>>deg(small), deg(small) =~ 10 max
// buf is used as a buffer
func mul(big, small []fr.Element, buf []fr.Element) []fr.Element {

	sizeRes := len(big) + len(small) - 1
	if len(buf) < sizeRes {
		s := make([]fr.Element, sizeRes-len(buf))
		buf = append(buf, s...)
	}
	for i := 0; i < len(buf); i++ {
		buf[i].SetZero()
	}

	var tmp fr.Element
	for i := 0; i < len(small); i++ {
		for j := 0; j < len(big); j++ {
			tmp.Mul(&big[j], &small[i])
			buf[j+i].Add(&buf[j+i], &tmp)
		}
	}
	return buf
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
