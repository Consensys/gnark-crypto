// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

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
	d = 7
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
	return fmt.Sprintf("Poseidon2-babybear[t=%d,rF=%d,rP=%d,d=%d]", p.Width, p.NbFullRounds, p.NbPartialRounds, d)
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
	var tmp1, tmp2 fr.Element
	tmp1.Set(&input[index])
	tmp2.Square(&input[index])
	// sbox degree is 7
	input[index].Square(&tmp2).
		Mul(&input[index], &tmp1).
		Mul(&input[index], &tmp2)

}

// matMulM4 computes
// s <- M4*s
// where M4=
// (2 3 1 1)
// (1 2 3 1)
// (1 1 2 3)
// (3 1 1 2)
// on chunks of 4 elemts on each part of the buffer
// for the addition chain, see:
// https://github.com/Plonky3/Plonky3/blob/f91c76545cf5c4ae9182897bcc557715817bcbdc/poseidon2/src/external.rs#L43
// this MDS matrix is more efficient than
// https://eprint.iacr.org/2023/323.pdf appendix Bb
func (h *Permutation) matMulM4InPlace(s []fr.Element) {
	c := len(s) / 4
	for i := 0; i < c; i++ {
		var t01, t23, t0123, t01123, t01233 fr.Element
		t01.Add(&s[4*i], &s[4*i+1])
		t23.Add(&s[4*i+2], &s[4*i+3])
		t0123.Add(&t01, &t23)
		t01123.Add(&t0123, &s[4*i+1])
		t01233.Add(&t0123, &s[4*i+3])
		// The order here is important. Need to overwrite x[0] and x[2] after x[1] and x[3].
		s[4*i+3].Double(&s[4*i]).Add(&s[4*i+3], &t01233)
		s[4*i+1].Double(&s[4*i+2]).Add(&s[4*i+1], &t01123)
		s[4*i].Add(&t01, &t01123)
		s[4*i+2].Add(&t23, &t01233)
	}
}

// when Width = 0 mod 4, the buffer is multiplied by circ(2M4,M4,..,M4)
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
		input[4*i+1].Add(&input[4*i+1], &tmp[1])
		input[4*i+2].Add(&input[4*i+2], &tmp[2])
		input[4*i+3].Add(&input[4*i+3], &tmp[3])
	}
}

// when Width = 0 mod 4 the matrix is filled with ones except on the diagonal
func (h *Permutation) matMulInternalInPlace(input []fr.Element) {
	switch h.params.Width {
	case 16:
		var sum fr.Element
		sum.Set(&input[0])
		for i := 1; i < h.params.Width; i++ {
			sum.Add(&sum, &input[i])
		}
		// mul by diag16:
		// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/2^27, -1/2^8, -1/16, -1/2^27]
		var temp fr.Element
		input[0].Sub(&sum, temp.Double(&input[0]))
		input[1].Add(&sum, &input[1])
		input[2].Add(&sum, temp.Double(&input[2]))
		temp.Set(&input[3]).Halve()
		input[3].Add(&sum, &temp)
		input[4].Add(&sum, temp.Double(&input[4]).Add(&temp, &input[4]))
		input[5].Add(&sum, temp.Double(&input[5]).Double(&temp))
		temp.Set(&input[6]).Halve()
		input[6].Sub(&sum, &temp)
		input[7].Sub(&sum, temp.Double(&input[7]).Add(&temp, &input[7]))
		input[8].Sub(&sum, temp.Double(&input[8]).Double(&temp))
		input[9].Add(&sum, temp.Mul2ExpNegN(&input[9], 8))
		input[10].Add(&sum, temp.Mul2ExpNegN(&input[10], 2))
		input[11].Add(&sum, temp.Mul2ExpNegN(&input[11], 3))
		input[12].Add(&sum, temp.Mul2ExpNegN(&input[12], 27))
		input[13].Sub(&sum, temp.Mul2ExpNegN(&input[13], 8))
		input[14].Sub(&sum, temp.Mul2ExpNegN(&input[14], 4))
		input[15].Sub(&sum, temp.Mul2ExpNegN(&input[15], 27))
	case 24:
		var sum fr.Element
		sum.Set(&input[0])
		for i := 1; i < h.params.Width; i++ {
			sum.Add(&sum, &input[i])
		}
		// mul by diag24:
		// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/16, 1/2^7, 1/2^9, 1/2^27, -1/2^8, -1/4, -1/8, -1/16, -1/32, -1/64, -1/2^7, -1/2^27]
		var temp fr.Element
		input[0].Sub(&sum, temp.Double(&input[0]))
		input[1].Add(&sum, &input[1])
		input[2].Add(&sum, temp.Double(&input[2]))
		temp.Set(&input[3]).Halve()
		input[3].Add(&sum, &temp)
		input[4].Add(&sum, temp.Double(&input[4]).Add(&temp, &input[4]))
		input[5].Add(&sum, temp.Double(&input[5]).Double(&temp))
		temp.Set(&input[6]).Halve()
		input[6].Sub(&sum, &temp)
		input[7].Sub(&sum, temp.Double(&input[7]).Add(&temp, &input[7]))
		input[8].Sub(&sum, temp.Double(&input[8]).Double(&temp))
		input[9].Add(&sum, temp.Mul2ExpNegN(&input[9], 8))
		input[10].Add(&sum, temp.Mul2ExpNegN(&input[10], 2))
		input[11].Add(&sum, temp.Mul2ExpNegN(&input[11], 3))
		input[12].Add(&sum, temp.Mul2ExpNegN(&input[12], 4))
		input[13].Add(&sum, temp.Mul2ExpNegN(&input[13], 7))
		input[14].Add(&sum, temp.Mul2ExpNegN(&input[14], 9))
		input[15].Add(&sum, temp.Mul2ExpNegN(&input[15], 27)) //nolint: gosec // incorrectly flagged by gosec as out of bounds read (G602)
		input[16].Sub(&sum, temp.Mul2ExpNegN(&input[16], 8))  //nolint: gosec // incorrectly flagged by gosec as out of bounds read (G602)
		input[17].Sub(&sum, temp.Mul2ExpNegN(&input[17], 2))  //nolint: gosec // incorrectly flagged by gosec as out of bounds read (G602)
		input[18].Sub(&sum, temp.Mul2ExpNegN(&input[18], 3))  //nolint: gosec // incorrectly flagged by gosec as out of bounds read (G602)
		input[19].Sub(&sum, temp.Mul2ExpNegN(&input[19], 4))  //nolint: gosec // incorrectly flagged by gosec as out of bounds read (G602)
		input[20].Sub(&sum, temp.Mul2ExpNegN(&input[20], 5))  //nolint: gosec // incorrectly flagged by gosec as out of bounds read (G602)
		input[21].Sub(&sum, temp.Mul2ExpNegN(&input[21], 6))  //nolint: gosec // incorrectly flagged by gosec as out of bounds read (G602)
		input[22].Sub(&sum, temp.Mul2ExpNegN(&input[22], 7))  //nolint: gosec // incorrectly flagged by gosec as out of bounds read (G602)
		input[23].Sub(&sum, temp.Mul2ExpNegN(&input[23], 27)) //nolint: gosec // incorrectly flagged by gosec as out of bounds read (G602)
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
