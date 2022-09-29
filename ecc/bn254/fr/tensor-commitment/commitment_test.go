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
	"math/big"
	"strconv"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/sis"
)

type DummyHash uint

func (d DummyHash) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (d DummyHash) Sum(b []byte) []byte {
	return b
}

func (d DummyHash) Reset() {
	return
}

func (d DummyHash) Size() int {
	return 0
}

func (d DummyHash) BlockSize() int {
	return 0
}

func getRandomVector(size int) []fr.Element {
	a := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		a[i].SetRandom()
	}
	return a
}

func TestLinearCombination(t *testing.T) {

	var rho, size, sqrtSize int
	rho = 4
	size = 64
	sqrtSize = 8
	capacity := 1

	var h DummyHash
	tc, err := NewTensorCommitment(rho, size, capacity, h)
	if err != nil {
		t.Fatal(err)
	}

	p := make([]fr.Element, size)
	for i := 0; i < 64; i++ {
		p[i].SetRandom()
	}

	// we select all the entries for the test
	entryList := make([]int, rho*sqrtSize)
	for i := 0; i < rho*sqrtSize; i++ {
		entryList[i] = i
	}

	// at each trial, it's the i-th line which is selected
	for i := 0; i < sqrtSize; i++ {

		// used for the random linear combination.
		// it will act as a selector for the test: it selects the i-th
		// row of p, when p is written as a matrix M_ij, where M_ij=p[i*m+j].
		// The i-th entry of l is 1, the others are 0.
		l := make([]fr.Element, sqrtSize)
		l[i].SetInt64(1)

		proof, err := tc.BuildProof(p, l, entryList)
		if err != nil {
			t.Fatal(err)
		}

		// the i-th line of p is the one that is supposed to be selected
		// (corresponding to the linear combination)
		expected := make([]fr.Element, rho*sqrtSize)
		copy(expected, p[i*sqrtSize:(i+1)*sqrtSize])

		for j := 0; j < sqrtSize; j++ {
			if !expected[j].Equal(&proof.LinearCombination[j]) {
				t.Fatal("expected linear combination is incorrect")
			}
		}

	}
}

// Test the verification of a correct proof using a mock hash
func TestCommitmentDummyHash(t *testing.T) {

	var rho, size, sqrtSize int
	rho = 4
	size = 64
	sqrtSize = 8
	capacity := 1

	var h DummyHash
	tc, err := NewTensorCommitment(rho, size, capacity, h)
	if err != nil {
		t.Fatal(err)
	}

	// random polynomial
	p := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		p[i].SetRandom()
	}

	// coefficients for the linear combination
	l := make([]fr.Element, sqrtSize)
	for i := 0; i < sqrtSize; i++ {
		l[i].SetRandom()
	}

	// we select all the entries for the test
	entryList := make([]int, rho*sqrtSize)
	for i := 0; i < rho*sqrtSize; i++ {
		entryList[i] = i
	}

	// compute the digest...
	tc.Append(p)
	digest, err := tc.Commit()
	if err != nil {
		t.Fatal(err)
	}

	// build the proof...
	proof, err := tc.BuildProof(p, l, entryList)
	if err != nil {
		t.Fatal(err)
	}

	// verfiy that the proof is correct
	err = Verify(proof, digest, l, h)
	if err != nil {
		t.Fatal(err)
	}

}

// Test the opening using a dummy hash
func TestOpeningDummyHash(t *testing.T) {

	var rho, size, sqrtSize int
	rho = 4
	size = 64
	sqrtSize = 8
	capacity := 1

	var h DummyHash
	tc, err := NewTensorCommitment(rho, size, capacity, h)
	if err != nil {
		t.Fatal(err)
	}

	// random polynomial
	p := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		p[i].SetRandom()
	}

	// the coefficients are (1,x,x^2,..,x^{n-1}) where x is the point
	// at which the opening is done
	var xm, x fr.Element
	x.SetRandom()
	hi := make([]fr.Element, sqrtSize) // stores [1,x^{m},..,x^{m^{2}-1}]
	lo := make([]fr.Element, sqrtSize) // stores [1,x,..,x^{m-1}]
	lo[0].SetInt64(1)
	hi[0].SetInt64(1)
	xm.Exp(x, big.NewInt(int64(sqrtSize)))
	for i := 1; i < sqrtSize; i++ {
		lo[i].Mul(&lo[i-1], &x)
		hi[i].Mul(&hi[i-1], &xm)
	}

	// build the proof
	entryList := make([]int, rho*sqrtSize)
	for i := 0; i < rho*sqrtSize; i++ {
		entryList[i] = i
	}
	proof, err := tc.BuildProof(p, hi, entryList)
	if err != nil {
		t.Fatal(err)
	}

	// finish the evalutation by computing
	// [linearCombination] * [lo]^t
	var eval, tmp fr.Element
	for i := 0; i < sqrtSize; i++ {
		tmp.Mul(&proof.LinearCombination[i], &lo[i])
		eval.Add(&eval, &tmp)
	}

	// compute the real evaluation of p at x manually
	var expectedEval fr.Element
	for i := 0; i < size; i++ {
		expectedEval.Mul(&expectedEval, &x)
		expectedEval.Add(&expectedEval, &p[len(p)-i-1])
	}

	// the results coincide
	if !expectedEval.Equal(&eval) {
		t.Fatal("p(x) != [ hi ] x M x [ lo ]^t")
	}

}

// Test the verification of a correct proof using SIS as hash
func TestCommitmentSis(t *testing.T) {

	var rho, size, sqrtSize int
	rho = 4
	size = 64
	sqrtSize = 8
	capacity := 1

	logTwoDegree := 1
	logTwoBound := 4
	keySize := 256
	h, err := sis.NewRSis(5, logTwoDegree, logTwoBound, keySize)
	if err != nil {
		t.Fatal(err)
	}
	tc, err := NewTensorCommitment(rho, size, capacity, h)
	if err != nil {
		t.Fatal(err)
	}

	// random polynomial
	p := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		p[i].SetRandom()
	}

	// coefficients for the linear combination
	l := make([]fr.Element, sqrtSize)
	for i := 0; i < sqrtSize; i++ {
		l[i].SetRandom()
	}

	// test 1: we select all the entries
	{
		entryList := make([]int, rho*sqrtSize)
		for i := 0; i < rho*sqrtSize; i++ {
			entryList[i] = i
		}
		// compute the digest...
		tc.Append(p)
		digest, err := tc.Commit()
		if err != nil {
			t.Fatal(err)
		}

		// build the proof...
		proof, err := tc.BuildProof(p, l, entryList)
		if err != nil {
			t.Fatal(err)
		}

		// verfiy that the proof is correct
		err = Verify(proof, digest, l, h)
		if err != nil {
			t.Fatal(err)
		}
	}
	// test 2: we select a subset of the entries
	{

		entryList := make([]int, 2)
		entryList[0] = 1
		entryList[1] = 4

		// compute the digest...
		tc.Append(p)
		digest, err := tc.Commit()
		if err != nil {
			t.Fatal(err)
		}

		// build the proof...
		proof, err := tc.BuildProof(p, l, entryList)
		if err != nil {
			t.Fatal(err)
		}

		// verfiy that the proof is correct
		err = Verify(proof, digest, l, h)
		if err != nil {
			t.Fatal(err)
		}
	}

}

// benches
func BenchmarkTensorCommitment(b *testing.B) {

	// prepare the tensor commitment
	sizeFr := 256
	logTwoDegree := 4
	logTwoBound := 4
	rho := 4
	capacity := 1

	for i := 0; i < 4; i++ {

		sqrtSizePoly := (1 << (5 + i))
		sizePoly := sqrtSizePoly * sqrtSizePoly

		// (sqrtSizePoly * sizeFr) = nbBitsToHash
		// nbBitsToHash / (logTwoBound * degree) = nb coeffs to pack
		sizeKey := (sqrtSizePoly * sizeFr) / (logTwoBound * (1 << logTwoDegree))

		h, _ := sis.NewRSis(5, logTwoDegree, logTwoBound, sizeKey)
		tc, _ := NewTensorCommitment(rho, sizePoly, capacity, h)

		// random polynomial
		p := make([]fr.Element, sizePoly)
		for i := 0; i < sizePoly; i++ {
			p[i].SetRandom()
		}

		// run the benchmark
		b.Run("size poly"+strconv.Itoa(sizePoly), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tc.Append(p)
				tc.Commit()
			}
		})

	}

}
