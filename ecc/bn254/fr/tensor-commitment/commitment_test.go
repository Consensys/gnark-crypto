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
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
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

func TestCommitmentLinearCombination(t *testing.T) {

	var rho, size, sqrtSize int
	rho = 4
	size = 64
	sqrtSize = 8

	var h DummyHash
	tc, err := NewTensorCommitment(rho, size, h)
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

		proof, err := tc.buildProof(p, l, entryList)
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

func TestCommitmentDummyHash(t *testing.T) {

	var rho, size, sqrtSize int
	rho = 4
	size = 64
	sqrtSize = 8

	var h DummyHash
	tc, err := NewTensorCommitment(rho, size, h)
	if err != nil {
		t.Fatal(err)
	}

	// random polynomial and random linear
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
	digest, err := tc.Commit(p)
	if err != nil {
		t.Fatal(err)
	}

	// build the proof...
	proof, err := tc.buildProof(p, l, entryList)
	if err != nil {
		t.Fatal(err)
	}

	// verfiy that the proof is correct
	err = Verify(proof, digest, l, h)
	if err != nil {
		t.Fatal(err)
	}

}
