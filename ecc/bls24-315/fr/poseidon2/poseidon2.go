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

// Code generated by consensys/gnark-crypto DO NOT EDIT

package poseidon2

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	"golang.org/x/crypto/sha3"
)

// poseidon
// https://github.com/argumentcomputer/neptune/blob/main/spec/poseidon_spec.pdf

// poseidon2 ref implem
// https://github.com/HorizenLabs/poseidon2/blob/main/plain_implementations/src/poseidon2/poseidon2.rs

// M ∈ {80,128,256}, security level in bits

var (
	ErrInvalidSizebuffer = errors.New("the size of the input should match the size of the hash buffer")
)

// parameters describing the poseidon2 implementation
type parameters struct {
	// len(preimage)+len(digest)=len(preimage)+ceil(log(2*<security_level>/r))
	t int

	// sbox degree
	d int

	// number of full rounds (even number)
	rF int

	// number of partial rounds
	rP int

	// diagonal elements of the internal matrices, minus one
	diagInternalMatrices []fr.Element

	// round keys
	roundKeys [][]fr.Element
}

// Hash stores the buffer of the poseidon2 permutation and provides poseidon2 permutation
// methods on the buffer
type Hash struct {

	// parameters describing the
	params parameters

	// buffer
	buffer []fr.Element
}

func NewHash(t, d, rf, rp int, seed string) Hash {
	params := parameters{t: t, d: d, rF: rf, rP: rp}
	res := Hash{params: params}
	initRC(seed, &params)
	return res
}

// GetT returns the size of the state
func (h *Hash) GetT() int {
	return h.params.t
}

// GetRF returns the total number of full rounds (even number)
func (h *Hash) GetRF() int {
	return h.params.rF
}

// GetRP retursn the total number of partial rounds
func (h *Hash) GetRP() int {
	return h.params.rP
}

// GetD returns the degree of the sbox
func (h *Hash) GetD() int {
	return h.params.d
}

// GetInternalMatrix returns the diagonal elements of the internal matrix, minus one
func (h *Hash) GetInternalMatrix() []fr.Element {
	return h.params.diagInternalMatrices
}

// GetRC returns the round constants
func (h *Hash) GetRC() [][]fr.Element {
	return h.params.roundKeys
}

// initRC initiate round keys. Only one entry is non zero for the internal
// rounds, cf https://eprint.iacr.org/2023/323.pdf page 9
func initRC(seed string, params *parameters) {

	bseed := ([]byte)(seed)
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write(bseed)
	rnd := hash.Sum(nil) // pre hash before use
	hash.Reset()
	_, _ = hash.Write(rnd)

	params.roundKeys = make([][]fr.Element, params.rF+params.rP)
	for i := 0; i < params.rF/2; i++ {
		params.roundKeys[i] = make([]fr.Element, params.t)
		for j := 0; j < params.t; j++ {
			rnd = hash.Sum(nil)
			params.roundKeys[i][j].SetBytes(rnd)
			hash.Reset()
			_, _ = hash.Write(rnd)
		}
	}
	for i := 0; i < params.rP; i++ {
		params.roundKeys[i] = make([]fr.Element, 1)
		rnd = hash.Sum(nil)
		params.roundKeys[i][0].SetBytes(rnd)
		hash.Reset()
		_, _ = hash.Write(rnd)
	}
	for i := 0; i < params.rF/2; i++ {
		params.roundKeys[i] = make([]fr.Element, params.t)
		for j := 0; j < params.t; j++ {
			rnd = hash.Sum(nil)
			params.roundKeys[i][j].SetBytes(rnd)
			hash.Reset()
			_, _ = hash.Write(rnd)
		}
	}

}

// Reset empty the state
func (h *Hash) Reset() {
	h.buffer = make([]fr.Element, 0, len(h.buffer))
}

// Write populate the buffer of the hash
func (h *Hash) Write(elmts []fr.Element) {
	h.buffer = append(h.buffer, elmts...)
}

// sBox applies the sBox on buffer[index]
func (h *Hash) sBox(index int, input []fr.Element) {
	var tmp fr.Element
	tmp.Set(&input[index])
	if h.params.d == 3 {
		input[index].Square(&input[index]).
			Mul(&input[index], &tmp)
	} else if h.params.d == 5 {
		input[index].Square(&input[index]).
			Square(&input[index]).
			Mul(&input[index], &tmp)
	} else if h.params.d == 7 {
		input[index].Square(&input[index]).
			Mul(&input[index], &tmp).
			Square(&input[index]).
			Mul(&input[index], &tmp)
	}
}

// matMulM4 computes
// s <- M4*s
// where M4=
// (5 7 1 3)
// (4 6 1 1)
// (1 3 5 7)
// (1 1 4 6)
// on chunks of 4 elemts on each part of the buffer
// see https://eprint.iacr.org/2023/323.pdf appendix B for the addition chain
func (h *Hash) matMulM4InPlace(s []fr.Element) {
	c := len(s) / 4
	for i := 0; i < c; i++ {
		var t0, t1, t2, t3, t4, t5, t6, t7 fr.Element
		t0.Add(&s[4*i], &s[4*i+1])               // s0+s1
		t1.Add(&s[4*i+2], &s[4*i+3])             // s2+s3
		t2.Double(&s[4*i+1]).Add(&t2, &t1)       // 2s1+t1
		t3.Double(&s[4*i+3]).Add(&t3, &t0)       // 2s3+t0
		t4.Double(&t1).Double(&t4).Add(&t4, &t3) // 4t1+t3
		t5.Double(&t0).Double(&t5).Add(&t5, &t2) // 4t0+t2
		t6.Add(&t3, &t5)                         // t3+t4
		t7.Add(&t2, &t4)                         // t2+t4
		s[4*i].Set(&t6)
		s[4*i+1].Set(&t5)
		s[4*i+2].Set(&t7)
		s[4*i+3].Set(&t4)
	}
}

// when t=2,3 the buffer is multiplied by circ(2,1) and circ(2,1,1)
// see https://eprint.iacr.org/2023/323.pdf page 15, case t=2,3
//
// when t=0[4], the buffer is multiplied by circ(2M4,M4,..,M4)
// see https://eprint.iacr.org/2023/323.pdf
func (h *Hash) matMulExternalInPlace(input []fr.Element) {

	if h.params.t == 2 {
		var tmp fr.Element
		tmp.Add(&input[0], &input[1])
		input[0].Add(&tmp, &input[0])
		input[1].Add(&tmp, &input[1])
	} else if h.params.t == 3 {
		var tmp fr.Element
		tmp.Add(&input[0], &input[1]).
			Add(&tmp, &input[2])
		input[0].Add(&tmp, &input[0])
		input[1].Add(&tmp, &input[1])
		input[2].Add(&tmp, &input[2])
	} else if h.params.t == 4 {
		h.matMulM4InPlace(input)
	} else {
		// at this stage t is supposed to be a multiple of 4
		// the MDS matrix is circ(2M4,M4,..,M4)
		h.matMulM4InPlace(input)
		tmp := make([]fr.Element, 4)
		for i := 0; i < h.params.t/4; i++ {
			tmp[0].Add(&tmp[0], &input[4*i])
			tmp[1].Add(&tmp[1], &input[4*i+1])
			tmp[2].Add(&tmp[2], &input[4*i+2])
			tmp[3].Add(&tmp[3], &input[4*i+3])
		}
		for i := 0; i < h.params.t/4; i++ {
			input[4*i].Add(&input[4*i], &tmp[0])
			input[4*i+1].Add(&input[4*i], &tmp[1])
			input[4*i+2].Add(&input[4*i], &tmp[2])
			input[4*i+3].Add(&input[4*i], &tmp[3])
		}
	}
}

// when t=2,3 the matrix are respectibely [[2,1][1,3]] and [[2,1,1][1,2,1][1,1,3]]
// otherwise the matrix is filled with ones except on the diagonal,
func (h *Hash) matMulInternalInPlace(input []fr.Element) {
	if h.params.t == 2 {
		var sum fr.Element
		sum.Add(&input[0], &input[1])
		input[0].Add(&input[0], &sum)
		input[1].Double(&input[1]).Add(&input[1], &sum)
	} else if h.params.t == 3 {
		var sum fr.Element
		sum.Add(&input[0], &input[1]).Add(&sum, &input[2])
		input[0].Add(&input[0], &sum)
		input[1].Add(&input[1], &sum)
		input[2].Double(&input[2]).Add(&input[2], &sum)
	} else {
		var sum fr.Element
		sum.Set(&input[0])
		for i := 1; i < h.params.t; i++ {
			sum.Add(&sum, &input[i])
		}
		for i := 0; i < h.params.t; i++ {
			input[i].Mul(&input[i], &h.params.diagInternalMatrices[i]).
				Add(&input[i], &sum)
		}
	}
}

// addRoundKeyInPlace adds the round-th key to the buffer
func (h *Hash) addRoundKeyInPlace(round int, input []fr.Element) {
	for i := 0; i < len(h.params.roundKeys[round]); i++ {
		input[i].Add(&input[i], &h.params.roundKeys[round][i])
	}
}

func (h *Hash) permutationInPlace(input []fr.Element) error {
	if len(input) != h.params.t {
		return ErrInvalidSizebuffer
	}

	// external matrix multiplication, cf https://eprint.iacr.org/2023/323.pdf page 14 (part 6)
	h.matMulExternalInPlace(input)

	rf := h.params.rF / 2
	for i := 0; i < rf; i++ {
		// one round = matMulExternal(sBox_Full(addRoundKey))
		h.addRoundKeyInPlace(i, input)
		for j := 0; j < h.params.t; j++ {
			h.sBox(j, input)
		}
		h.matMulExternalInPlace(input)
	}

	for i := rf; i < rf+h.params.rP; i++ {
		// one round = matMulInternal(sBox_sparse(addRoundKey))
		h.addRoundKeyInPlace(i, input)
		h.sBox(0, input)
		h.matMulInternalInPlace(input)
	}
	for i := rf + h.params.rP; i < h.params.rF+h.params.rP; i++ {
		// one round = matMulExternal(sBox_Full(addRoundKey))
		h.addRoundKeyInPlace(i, input)
		for j := 0; j < h.params.t; j++ {
			h.sBox(j, input)
		}
		h.matMulExternalInPlace(input)
	}

	return nil
}
