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
	ErrMaxCapacity         = errors.New("the state is full")
	ErrCommitmentNotDone   = errors.New("the proof cannot be built before the computation of the digest")
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

	// Capacity number of polynomials batched by the tensor commitment.
	// The size of the matrix storing the commitment is
	// SqrtSizePolynomial x (SqrtSizePolynomial * Capacity).
	// If SqrtSizePolynomial * Capacity is not a power of 2, for RS encoding
	// we take the smallest power of 2 'b' bounding above SqrtSizePolynomial * Capacity,
	// and we interpret the lines of the matrix as a polynomials of size b.
	Capacity int

	// Size of the the polynomials to be committed (so the degree of p is SizePolynomial-1)
	SizePolynomial int

	// √{d+1} where d = SizePolynomial (equal to the number of rows of the matrix storing the
	// polynomial)
	SqrtSizePolynomial int

	// Domains[1] used for the Reed Solomon encoding
	Domains [2]*fft.Domain

	// State contains the polynomials that have been appended so far.
	// The i-th entry is the i-th polynomial. Each polynomial is interpreted
	// as a square matrix M_ij=p[i*m+j] where m = \sqrt{len(p)}.
	state [][]fr.Element

	// same content as state, but the polynomials are displayed as a matrix
	// and the rows are encoded.
	// encodedState = encodeRows(M_0 || .. || M_n)
	// where M_i is the i-th polynomial layed out as a matrix, that is
	// M_i_jk = p_i[i*m+j] where m = \sqrt(len(p)).
	encodedState [][]fr.Element

	// boolean telling if the commitment has already been done.
	// The method BuildProof cannot be called before Commit(),
	// because it would allow to build a proof before giving the commitment
	// to a verifier, making the worklow not secure.
	isCommitted bool

	// number of polynomials stored so far
	nbPolynomialsStored int

	// number of columns which have already been hashed
	nbColumnsHashed int

	// Rho⁻¹, rate of the RS code ( > 1)
	Rho int

	// Hash function for hashing the columns
	Hash hash.Hash
}

// NewTensorCommitment retunrs a new TensorCommitment
// * ρ rate of the code ( > 1)
// * size size of the polynomial to be committed. The size of the commitment is
// then ρ * √(m) where m² = size
func NewTensorCommitment(codeRate, sizePolynomial, capacity int, h hash.Hash) (TensorCommitment, error) {

	res := TensorCommitment{
		SizePolynomial: sizePolynomial,
		Rho:            codeRate,
		Hash:           h,
	}

	sqrt := math.Floor(math.Sqrt(float64(sizePolynomial)))

	if sqrt*sqrt != float64(sizePolynomial) {
		return res, ErrNotSquare
	}

	res.SqrtSizePolynomial = int(sqrt)

	res.Capacity = capacity

	// create the state
	res.state = make([][]fr.Element, capacity)
	res.nbPolynomialsStored = 0

	// nothing has been committed...
	res.isCommitted = false

	// domain[0]: domain to perform the FFT^-1, of size capacity * sqrt
	// domain[1]: domain to perform FFT, of size rho * capacity * sqrt
	res.Domains[0] = fft.NewDomain(uint64(res.Capacity * res.SqrtSizePolynomial))
	res.Domains[1] = fft.NewDomain(uint64(res.Rho * res.Capacity * res.SqrtSizePolynomial))

	return res, nil

}

// Append appends p to the state.
// p is interpreted as the evaluation of a polynomial of degree len(p)
// on domain[0] (the domain of size SqrtSizePolynomial).
func (tc *TensorCommitment) Append(p []fr.Element) error {

	if tc.nbPolynomialsStored == tc.Capacity {
		return ErrMaxCapacity
	}
	if len(p) > tc.SizePolynomial {
		return ErrWrongSize
	}

	tc.state[tc.nbPolynomialsStored] = make([]fr.Element, tc.SizePolynomial)
	copy(tc.state[tc.nbPolynomialsStored], p)

	tc.nbPolynomialsStored++

	return nil
}

// HashState hashed the columns of the state. If the first k columns
// have already been hashed, then we hash the remaining columns only.
func (tc *TensorCommitment) HashState() []byte {
	return nil
}

// Commit to p. The commitment procedure is the following:
// for each polynomial p_i in the state, write p_i as a square matrix
// M_i, where M_i_jk=p_i[j*m+k], m = \sqrt(size poly). Then we
// build M = M_0 || ... || M_n. We then encode the rows
// of M, and then we hash the columns of M. If the capacity of the
// tensorCommitment is not reached, we padd M with zeros.
//
// p is a polynomial expressed in canonical form.
// For committing to P, we interpret P as a matrix M
// of size m x m where m² = len(p).
// The ij-th entry of M is p[m*i + j] (it's the transpose of the more
// logical order p[j*m + i], but it's more practical memory wise, it avoids
// rearranging the coeffs for the fft)
func (tc *TensorCommitment) Commit() (Digest, error) {

	// check the capacity
	if tc.Capacity > tc.nbPolynomialsStored {
		for i := tc.nbPolynomialsStored; i < tc.Capacity; i++ {
			tc.state[i] = make([]fr.Element, tc.SizePolynomial)
		}
	}

	// we encode the rows of p using Reed Solomon
	// encodedState[i][:] = i-th line of M. It is of size domain[1].Cardinality
	tc.encodedState = make([][]fr.Element, tc.SqrtSizePolynomial)
	for i := 0; i < tc.SqrtSizePolynomial; i++ { // we fill encodedState line by line
		tc.encodedState[i] = make([]fr.Element, tc.Domains[1].Cardinality) // size = SqrtSizePolynomial*rho*capacity
		for j := 0; j < tc.Capacity; j++ {                                 // for each polynomial
			offset := i * tc.SqrtSizePolynomial
			copy(tc.encodedState[i][j*tc.SqrtSizePolynomial:], tc.state[j][offset:offset+tc.SqrtSizePolynomial])
		}
		tc.Domains[0].FFTInverse(tc.encodedState[i][:tc.Domains[0].Cardinality], fft.DIF)
		fft.BitReverse(tc.encodedState[i][:tc.Domains[0].Cardinality])
		tc.Domains[1].FFT(tc.encodedState[i], fft.DIF)
		fft.BitReverse(tc.encodedState[i])
	}

	// now we hash each columns of _p
	res := make([][]byte, tc.Domains[1].Cardinality)
	for i := 0; i < int(tc.Domains[1].Cardinality); i++ {
		tc.Hash.Reset()
		for j := 0; j < tc.SqrtSizePolynomial; j++ {
			tc.Hash.Write(tc.encodedState[j][i].Marshal())
		}
		res[i] = tc.Hash.Sum(nil)
		tc.Hash.Reset()
	}

	// records that the ccommitment has been built
	tc.isCommitted = true

	return res, nil

}

func printVector(v []fr.Element) {
	fmt.Printf("[")
	for i := 0; i < len(v); i++ {
		fmt.Printf("%s,", v[i].String())
	}
	fmt.Printf("]\n")
}

// BuildProof builds a proof to be tested against a previous commitment of a list of
// polynomials.
// * l the random linear coefficients used for the linear combination of size SqrtSizePolynomial
// * entryList list of columns to hash
// l and entryList are supposed to be precomputed using Fiat Shamir
//
// The proof is the linear combination (using l) of the encoded rows of p written
// as a matrix. Only the entries contained in entryList are kept.
func (tc *TensorCommitment) BuildProof(l []fr.Element, entryList []int) (Proof, error) {

	// check the capacity
	if tc.Capacity > tc.nbPolynomialsStored {
		for i := tc.nbPolynomialsStored; i < tc.Capacity; i++ {
			tc.state[i] = make([]fr.Element, tc.SizePolynomial)
		}
	}

	var res Proof

	// check that the digest has been computed
	if !tc.isCommitted {
		return res, ErrCommitmentNotDone
	}

	// since the digest has been computed, the encodedState is already stored.
	// We use it to build the proof, without recomputing the ffts.

	// linear combination of the rows of the state, written as a matrix
	// M = M_0 || ... || M_n where M_i is the i-th polynomial, M_i_jk=p[j*m+k]
	// where m = \sqrt(len(p)).
	res.LinearCombination = make([]fr.Element, tc.Capacity*tc.SqrtSizePolynomial)
	for i := 0; i < tc.Capacity; i++ {
		linearCombination(
			res.LinearCombination[i*tc.SqrtSizePolynomial:(i+1)*tc.SqrtSizePolynomial],
			l,
			tc.state[i],
		)
	}

	// columns of the state whose rows have been encoded, written as a matrix,
	// corresponding to the indices in entryList (we will select the columns
	// entryList[0], entryList[1], etc.
	res.Columns = make([][]fr.Element, len(entryList))
	for i := 0; i < len(entryList); i++ { // for each column (corresponding to an elmt in entryList)
		res.Columns[i] = make([]fr.Element, tc.SqrtSizePolynomial)
		for j := 0; j < tc.SqrtSizePolynomial; j++ {
			res.Columns[i][j] = tc.encodedState[j][entryList[i]]
		}
	}

	return res, nil
}

// linearCombination writes p as a matrix
// M = (M_ij), where M_ij = p[i*m + j] and m² = len(p).
// Then it computes ∑_i r[i]*M[i:]
// It is assumed that len(r)² = len(p)
//
// The result is stored in res.
func linearCombination(res, r, p []fr.Element) {

	m := len(r)
	var tmp fr.Element

	for i := 0; i < m; i++ {
		for j := 0; j < m; j++ {
			tmp.Mul(&p[j*m+i], &r[j])
			res[i].Add(&res[i], &tmp)
		}
	}
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
			h.Write(proof.Columns[i][j].Marshal())
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
