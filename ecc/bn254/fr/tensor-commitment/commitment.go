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

package tensorcommitment

import (
	"errors"
	"fmt"
	"hash"
	"math"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
)

var (
	ErrWrongSize           = errors.New("polynomial is too large")
	ErrNotSquare           = errors.New("the size of the polynomial must be a square")
	ErrProofFailedHash     = errors.New("hash of one of the columns is wrong")
	ErrProofFailedEncoding = errors.New("inconsistency with the code word")
	ErrProofFailedOob      = errors.New("the entry is out of bound")
)

// commitment (TODO Merkle tree for that...)
// The i-th entry is the hash of the i-th columns of P,
// where P is written as a matrix √(m) x √(m)
// (m = len(P)), and the ij-th entry of M is p[m*j + i].
type Digest [][]byte

// Proof that a commitment is correct
// cf https://eprint.iacr.org/2021/1043.pdf page 10
type Proof struct {

	// list of entries of ̂{u} to query (see https://eprint.iacr.org/2021/1043.pdf for notations)
	EntryList []int

	// columns on against which the linear combination is checked
	// (the i-th entry is the EntryList[i]-th column)
	Columns [][]fr.Element

	// Linear combination of the rows of the polynomial P written as a square matrix
	LinearCombination []fr.Element

	// root of unity
	Generator fr.Element
}

// TensorCommitment stores the data to use a tensor commitment
type TensorCommitment struct {

	// Size of the the polynomials to be committed (so the degree of p is MaxSize-1)
	MaxSize int

	// √{d+1} where d = MaxSize
	SqrtSize int

	// Domain used for the Reed Solomon encoding
	Domain *fft.Domain

	// Rho⁻¹, rate of the RS code
	Rho int

	// Hash function for hashing the columns
	Hash hash.Hash
}

// NewTensorCommitment retunrs a new TensorCommitment
// * ρ rate of the code ( > 1)
// * size size of the polynomial to be committed. The size of the commitment is
// then ρ * √(m) where m² = size
func NewTensorCommitment(rho, size int, h hash.Hash) (TensorCommitment, error) {

	res := TensorCommitment{
		MaxSize: size,
		Rho:     rho,
		Hash:    h,
	}

	sqrt := math.Floor(math.Sqrt(float64(size)))

	if sqrt*sqrt != float64(size) {
		return res, ErrNotSquare
	}

	res.SqrtSize = int(sqrt)

	sizeDomain := uint64(rho * res.SqrtSize)
	res.Domain = fft.NewDomain(sizeDomain)

	return res, nil

}

// Commit to p. The commitment procedure is the following:
// write p as a m x m matrix with m² = len(p)
//
// p is a polynomial expressed in canonical form.
// For committing to P, we interpret P as a matrix M
// of size m x m where m² = len(p).
// The ij-th entry of M is p[m*i + j] (it's the transpose of the more
// logical order p[j*m + i], but it's more practical memory wise, it avoids
// rearranging the coeffs for the fft)
func (tc *TensorCommitment) Commit(p []fr.Element) (Digest, error) {

	// first we adjus	t the size of p so it fits the fft domain
	if len(p) > tc.MaxSize {
		return nil, ErrWrongSize
	}

	// we encode the rows of p using Reed Solomon
	_p := make([][]fr.Element, tc.SqrtSize)
	for i := 0; i < tc.SqrtSize; i++ {
		_p[i] = make([]fr.Element, tc.Domain.Cardinality)
		copy(_p[i], p[i*tc.SqrtSize:(i+1)*tc.SqrtSize])
		tc.Domain.FFT(_p[i], fft.DIF)
		fft.BitReverse(_p[i])
	}

	// now we hash each columns of _p
	res := make([][]byte, tc.Domain.Cardinality)
	for i := 0; i < int(tc.Domain.Cardinality); i++ {
		tc.Hash.Reset()
		for j := 0; j < tc.SqrtSize; j++ {
			tc.Hash.Write(_p[j][i].Marshal())
		}
		res[i] = tc.Hash.Sum(nil)
		tc.Hash.Reset()
	}

	return res, nil

}

func printVector(v []fr.Element) {
	fmt.Printf("[")
	for i := 0; i < len(v); i++ {
		fmt.Printf("%s,", v[i].String())
	}
	fmt.Printf("]\n")
}

// BuildProof builds a proof to be tested against a previous commitment to p
// attesting that the commitment corresponds to p.
// * p the polynomial which has been committed (supposed to be of the correct size)
// * l the random linear coefficients used for the linear combination
// * entryList list of columns to hash
// l and entryList are supposed to be precomputed using Fiat Shamir
//
// The proof is the linear combination (using l) of the encoded rows of p written
// as a matrix. Only the entries contained in entryList are kept.
func (tc *TensorCommitment) BuildProof(p, l []fr.Element, entryList []int) (Proof, error) {

	var res Proof

	res.Generator.Set(&tc.Domain.Generator)
	res.EntryList = entryList

	// Linear combination of the line of p (written as a matrix
	// M = M_ij where M_ij = p[i*m + j], m² = len(p)))
	var tmp fr.Element
	res.LinearCombination = make([]fr.Element, tc.SqrtSize)
	for i := 0; i < tc.SqrtSize; i++ { // for each column of p
		for j := 0; j < tc.SqrtSize; j++ { // for each line of p
			tmp.Mul(&l[j], &p[j*tc.SqrtSize+i])
			res.LinearCombination[i].Add(&res.LinearCombination[i], &tmp)
		}
	}

	// Reed Solomon encoding of each rows of p (when p is interpreted as a matrix
	// M = M_ij where M_ij = p[i*m + j], m² = len(p)) corresponding to the indices
	// in entryList
	res.Columns = make([][]fr.Element, len(entryList))
	for i := 0; i < len(entryList); i++ { // for each column (corresponding to an elmt in entryList)
		res.Columns[i] = make([]fr.Element, len(l))
		for j := 0; j < tc.SqrtSize; j++ {
			res.Columns[i][j] = evalAtPower(p[j*tc.SqrtSize:(j+1)*tc.SqrtSize], tc.Domain.Generator, entryList[i])
		}
	}

	return res, nil

}

// linearCombination writes p as a matrix
// M = (M_ij), where M_ij = p[i*m + j] and m² = len(p).
// Then it computes ∑_i r[i]*M[i:]
// It is assmed that len(r)² = len(p)
func linearCombination(r, p []fr.Element) []fr.Element {

	m := len(r)
	res := make([]fr.Element, m)
	var tmp fr.Element

	for i := 0; i < m; i++ {
		for j := 0; j < m; j++ {
			tmp.Mul(&p[j*m+i], &r[j])
			res[i].Add(&res[i], &tmp)
		}
	}

	return res
}

// evalAtPower returns p(x**n) where p is interpreted as a polynomial
// p[0] + p[1]X + .. p[len(p)-1]xˡᵉⁿ⁽ᵖ⁾⁻¹
func evalAtPower(p []fr.Element, x fr.Element, n int) fr.Element {

	var xexp fr.Element
	xexp.Exp(x, big.NewInt(int64(n)))

	var res fr.Element
	for i := 0; i < len(p); i++ {
		res.Mul(&res, &xexp)
		res.Add(&p[len(p)-1-i], &res)
	}

	return res

}

func cmpBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	t := true
	for i := 0; i < len(a); i++ {
		t = t && (a[i] == b[i])
	}
	return t
}

// Verify a proof that digest is the hash of a  polynomial given a proof
// proof: proof that the commitment is correct
// digest: hash of the polynomial
// l: random coefficients for the linear combination, chosen by the verifier
// h: hash function that is used for hashing the columns of the polynomial
// TODO make this function private and add a Verify function that derives
// the randomness using Fiat Shamir
func Verify(proof Proof, digest Digest, l []fr.Element, h hash.Hash) error {

	// for each entry in the list -> it corresponds to the sampling
	// set on which we probabilistically check that
	// Encoded(linear_combination) = linear_combination(encoded)
	for i := 0; i < len(proof.EntryList); i++ {

		if proof.EntryList[i] >= len(digest) {
			return ErrProofFailedOob
		}

		// check that the hash of the columns correspond to what's in the digest
		h.Reset()
		for j := 0; j < len(proof.Columns[i]); j++ {
			h.Write(proof.Columns[proof.EntryList[i]][j].Marshal())
		}
		s := h.Sum(nil)
		if !cmpBytes(s, digest[proof.EntryList[i]]) {
			return ErrProofFailedHash
		}

		// linear combination of the i-th column, whose entries
		// are the entryList[i]-th entries of the encoded lines
		// of p
		var linCombEncoded, tmp fr.Element
		for j := 0; j < len(proof.Columns[i]); j++ {

			// linear combination of the encoded rows at column i
			tmp.Mul(&proof.Columns[i][j], &l[j])
			linCombEncoded.Add(&linCombEncoded, &tmp)
		}

		// entry i of the encoded linear combination
		var encodedLinComb fr.Element
		encodedLinComb = evalAtPower(proof.LinearCombination, proof.Generator, proof.EntryList[i])

		// compare both values
		if !encodedLinComb.Equal(&linCombEncoded) {
			return ErrProofFailedEncoding

		}
	}

	return nil

}
