// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package poseidon2

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/sha3"

	fr "github.com/consensys/gnark-crypto/field/babybear"
)

var (
	ErrInvalidSizebuffer = errors.New("the size of the input should match the size of the hash buffer")
)

const (
	// d is the degree of the sBox
	d = 3
)

// DegreeSBox returns the degree of the sBox function used in the Poseidon2
// permutation.
func DegreeSBox() int {
	return d
}

// Parameters describing the Poseidon2 implementation. Use [NewParameters] or
// [NewParametersWithSeed] to initialize a new set of parameters to
// deterministically precompute the round keys.
type Parameters struct {
	// len(preimage)+len(digest)=len(preimage)+ceil(log(2*<security_level>/r))
	Width int

	// number of full rounds (even number)
	NbFullRounds int

	// number of partial rounds
	NbPartialRounds int

	// derived round keys from the parameter seed and curve ID
	RoundKeys [][]fr.Element
}

// NewParameters returns a new set of parameters for the Poseidon2 permutation.
// After creating the parameters, the round keys are initialized deterministically
// from the seed which is a digest of the parameters and curve ID.
func NewParameters(width, nbFullRounds, nbPartialRounds int) *Parameters {
	p := Parameters{Width: width, NbFullRounds: nbFullRounds, NbPartialRounds: nbPartialRounds}
	seed := p.String()
	p.initRC(seed)
	return &p
}

// NewParametersWithSeed returns a new set of parameters for the Poseidon2 permutation.
// After creating the parameters, the round keys are initialized deterministically
// from the given seed.
func NewParametersWithSeed(width, nbFullRounds, nbPartialRounds int, seed string) *Parameters {
	p := Parameters{Width: width, NbFullRounds: nbFullRounds, NbPartialRounds: nbPartialRounds}
	p.initRC(seed)
	return &p
}

// String returns a string representation of the parameters. It is unique for
// specific parameters and curve.
func (p *Parameters) String() string {
	return fmt.Sprintf("Poseidon2-BABYBEAR[t=%d,rF=%d,rP=%d,d=%d]", p.Width, p.NbFullRounds, p.NbPartialRounds, d)
}

// initRC initiate round keys. Only one entry is non zero for the internal
// rounds, cf https://eprint.iacr.org/2023/323.pdf page 9
func (p *Parameters) initRC(seed string) {

	bseed := ([]byte)(seed)
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write(bseed)
	rnd := hash.Sum(nil) // pre hash before use
	hash.Reset()
	_, _ = hash.Write(rnd)

	roundKeys := make([][]fr.Element, p.NbFullRounds+p.NbPartialRounds)
	for i := 0; i < p.NbFullRounds/2; i++ {
		roundKeys[i] = make([]fr.Element, p.Width)
		for j := 0; j < p.Width; j++ {
			rnd = hash.Sum(nil)
			roundKeys[i][j].SetBytes(rnd)
			hash.Reset()
			_, _ = hash.Write(rnd)
		}
	}
	for i := p.NbFullRounds / 2; i < p.NbPartialRounds+p.NbFullRounds/2; i++ {
		roundKeys[i] = make([]fr.Element, 1)
		rnd = hash.Sum(nil)
		roundKeys[i][0].SetBytes(rnd)
		hash.Reset()
		_, _ = hash.Write(rnd)
	}
	for i := p.NbPartialRounds + p.NbFullRounds/2; i < p.NbPartialRounds+p.NbFullRounds; i++ {
		roundKeys[i] = make([]fr.Element, p.Width)
		for j := 0; j < p.Width; j++ {
			rnd = hash.Sum(nil)
			roundKeys[i][j].SetBytes(rnd)
			hash.Reset()
			_, _ = hash.Write(rnd)
		}
	}
	p.RoundKeys = roundKeys
}

// Permutation stores the buffer of the Poseidon2 permutation and provides
// Poseidon2 permutation methods on the buffer
type Permutation struct {
	// parameters describing the instance
	params *Parameters
}

// NewPermutation returns a new Poseidon2 permutation instance.
func NewPermutation(t, rf, rp int) *Permutation {
	if t != 16 && t != 24 {
		panic("only Width=16,24 are supported")
	}
	params := NewParameters(t, rf, rp)
	res := &Permutation{params: params}
	return res
}

// NewPermutationWithSeed returns a new Poseidon2 permutation instance with a
// given seed.
func NewPermutationWithSeed(t, rf, rp int, seed string) *Permutation {
	if t != 16 && t != 24 {
		panic("only Width=16,24 are supported")
	}
	params := NewParametersWithSeed(t, rf, rp, seed)
	res := &Permutation{params: params}
	return res
}

// sBox applies the sBox on buffer[index]
func (h *Permutation) sBox(index int, input []fr.Element) {
	var tmp fr.Element
	tmp.Set(&input[index])

	// sbox degree is 3
	input[index].Square(&input[index]).
		Square(&input[index]).
		Square(&input[index]).
		Square(&input[index]).
		Mul(&input[index], &tmp)

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
func (h *Permutation) matMulM4InPlace(s []fr.Element) {
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

// when T=2,3 the buffer is multiplied by circ(2,1) and circ(2,1,1)
// see https://eprint.iacr.org/2023/323.pdf page 15, case T=2,3
//
// when T=0[4], the buffer is multiplied by circ(2M4,M4,..,M4)
// see https://eprint.iacr.org/2023/323.pdf
func (h *Permutation) matMulExternalInPlace(input []fr.Element) {

	// at this stage t is supposed to be a multiple of 4
	// the MDS matrix is circ(2M4,M4,..,M4)
	h.matMulM4InPlace(input)
	tmp := make([]fr.Element, 4)
	for i := 0; i < h.params.Width/4; i++ {
		tmp[0].Add(&tmp[0], &input[4*i])
		tmp[1].Add(&tmp[1], &input[4*i+1])
		tmp[2].Add(&tmp[2], &input[4*i+2])
		tmp[3].Add(&tmp[3], &input[4*i+3])
	}
	for i := 0; i < h.params.Width/4; i++ {
		input[4*i].Add(&input[4*i], &tmp[0])
		input[4*i+1].Add(&input[4*i], &tmp[1])
		input[4*i+2].Add(&input[4*i], &tmp[2])
		input[4*i+3].Add(&input[4*i], &tmp[3])
	}
}

// when T=2,3 the matrix are respectibely [[2,1][1,3]] and [[2,1,1][1,2,1][1,1,3]]
// otherwise the matrix is filled with ones except on the diagonal,
func (h *Permutation) matMulInternalInPlace(input []fr.Element) {
	switch h.params.Width {
	case 16:
		// TODO: optimize multiplication by diag16
		// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/2^27, -1/2^8, -1/16, -1/2^27]
		var sum fr.Element
		sum.Set(&input[0])
		for i := 1; i < h.params.Width; i++ {
			sum.Add(&sum, &input[i])
		}
		for i := 0; i < h.params.Width; i++ {
			input[i].Mul(&input[i], &diag16[i]).
				Add(&input[i], &sum)
		}
	case 24:
		// TODO: optimize multiplication by diag24
		// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/16, 1/2^7, 1/2^9, 1/2^27, -1/2^8, -1/4, -1/8, -1/16, -1/32, -1/64, -1/2^7, -1/2^27]
		var sum fr.Element
		sum.Set(&input[0])
		for i := 1; i < h.params.Width; i++ {
			sum.Add(&sum, &input[i])
		}
		for i := 0; i < h.params.Width; i++ {
			input[i].Mul(&input[i], &diag24[i]).
				Add(&input[i], &sum)
		}
	default:
		panic("only Width=16,24 are supported")
	}
}

// addRoundKeyInPlace adds the round-th key to the buffer
func (h *Permutation) addRoundKeyInPlace(round int, input []fr.Element) {
	for i := 0; i < len(h.params.RoundKeys[round]); i++ {
		input[i].Add(&input[i], &h.params.RoundKeys[round][i])
	}
}

func (h *Permutation) BlockSize() int {
	return fr.Bytes
}

// Permutation applies the permutation on input, and stores the result in input.
func (h *Permutation) Permutation(input []fr.Element) error {
	if len(input) != h.params.Width {
		return ErrInvalidSizebuffer
	}

	// external matrix multiplication, cf https://eprint.iacr.org/2023/323.pdf page 14 (part 6)
	h.matMulExternalInPlace(input)

	rf := h.params.NbFullRounds / 2
	for i := 0; i < rf; i++ {
		// one round = matMulExternal(sBox_Full(addRoundKey))
		h.addRoundKeyInPlace(i, input)
		for j := 0; j < h.params.Width; j++ {
			h.sBox(j, input)
		}
		h.matMulExternalInPlace(input)
	}

	for i := rf; i < rf+h.params.NbPartialRounds; i++ {
		// one round = matMulInternal(sBox_sparse(addRoundKey))
		h.addRoundKeyInPlace(i, input)
		h.sBox(0, input)
		h.matMulInternalInPlace(input)
	}
	for i := rf + h.params.NbPartialRounds; i < h.params.NbFullRounds+h.params.NbPartialRounds; i++ {
		// one round = matMulExternal(sBox_Full(addRoundKey))
		h.addRoundKeyInPlace(i, input)
		for j := 0; j < h.params.Width; j++ {
			h.sBox(j, input)
		}
		h.matMulExternalInPlace(input)
	}

	return nil
}

// Compress applies the permutation on left and right and returns the right lane
// of the result. Panics if the permutation instance is not initialized with a
// width of 2.
func (h *Permutation) Compress(left []byte, right []byte) ([]byte, error) {
	if h.params.Width != 2 {
		return nil, errors.New("need a 2-1 function")
	}
	var x [2]fr.Element

	if err := x[0].SetBytesCanonical(left); err != nil {
		return nil, err
	}
	if err := x[1].SetBytesCanonical(right); err != nil {
		return nil, err
	}
	if err := h.Permutation(x[:]); err != nil {
		return nil, err
	}
	res := x[1].Bytes()
	return res[:], nil
}
