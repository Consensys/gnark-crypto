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

package kzg

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	bn254_pol "github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	"github.com/consensys/gnark-crypto/polynomial"
)

func TestDividePolyByXminusA(t *testing.T) {

	sizePol := 230

	domain := fft.NewDomain(uint64(sizePol), 0, false)

	// build random polynomial
	pol := make(bn254_pol.Polynomial, sizePol)
	for i := 0; i < sizePol; i++ {
		pol[i].SetRandom()
	}

	// evaluate the polynomial at a random point
	var point, evaluation fr.Element
	point.SetRandom()
	_evaluation := pol.Eval(&point)
	evaluation.Set(_evaluation.(*fr.Element))

	// compute f-f(a)/x-a
	h := dividePolyByXminusA(*domain, pol, evaluation, point)

	if len(h) != 229 {
		t.Fatal("inconsistant size of quotient")
	}

	// probabilistic test (using Schwartz Zippel lemma, evaluation at one point is enough)
	var randPoint, hRandPoint, polRandpoint, xminusa fr.Element
	randPoint.SetRandom()
	_r := pol.Eval(&randPoint)
	polRandpoint.Set(_r.(*fr.Element)).Sub(&polRandpoint, &evaluation) // f(rand)-f(point)
	_r = h.Eval(&randPoint)
	hRandPoint.Set(_r.(*fr.Element)) // h(rand)
	xminusa.Sub(&randPoint, &point)  // rand-point

	// f(rand)-f(point)	==? h(rand)*(rand-point)
	hRandPoint.Mul(&hRandPoint, &xminusa)

	if !hRandPoint.Equal(&polRandpoint) {
		t.Fatal("Error f-f(a)/x-a")
	}
}

func buildScheme() Scheme {

	var s Scheme
	d := fft.NewDomain(64, 0, true)
	s.Domain = *d
	s.Srs.G1 = make([]bn254.G1Affine, 64)

	// generate the SRS
	var alpha fr.Element
	//alpha.SetRandom()
	alpha.SetString("1234")
	var alphaBigInt big.Int
	alpha.ToBigIntRegular(&alphaBigInt)

	_, _, gen1Aff, gen2Aff := bn254.Generators()
	s.Srs.G1[0].Set(&gen1Aff)
	s.Srs.G2[0].Set(&gen2Aff)
	s.Srs.G2[1].ScalarMultiplication(&gen2Aff, &alphaBigInt)
	for i := 1; i < 64; i++ {
		s.Srs.G1[i].ScalarMultiplication(&s.Srs.G1[i-1], &alphaBigInt)
	}

	return s
}

func TestSerialization(t *testing.T) {

	// create a KZG scheme
	s := buildScheme()

	// serialize it...
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// reconstruct the scheme
	var _s Scheme
	_, err = _s.ReadFrom(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// compare
	if !reflect.DeepEqual(&s, &_s) {
		t.Fatal("scheme serialization failed")
	}

}

func TestCommit(t *testing.T) {

	// create a KZG scheme
	s := buildScheme()

	// create a polynomial
	f := make(bn254_pol.Polynomial, 60)
	for i := 0; i < 60; i++ {
		f[i].SetRandom()
	}

	// commit using the method from KZG
	_kzgCommit := s.Commit(&f)
	var kzgCommit bn254.G1Affine
	kzgCommit.Set(_kzgCommit.(*bn254.G1Affine))

	// check commitment using manual commit
	var x fr.Element
	x.SetString("1234")
	_fx := f.Eval(&x)
	fx := bn254_pol.FromInterface(_fx)
	var fxbi big.Int
	fx.ToBigIntRegular(&fxbi)
	var manualCommit bn254.G1Affine
	manualCommit.Set(&s.Srs.G1[0])
	manualCommit.ScalarMultiplication(&manualCommit, &fxbi)

	// compare both results
	if !kzgCommit.Equal(&manualCommit) {
		t.Fatal("error KZG commitment")
	}

}

func randomPolynomial() bn254_pol.Polynomial {
	f := make(bn254_pol.Polynomial, 60)
	for i := 0; i < 60; i++ {
		f[i].SetRandom()
	}
	return f
}

func TestVerifySinglePoint(t *testing.T) {

	// create a KZG scheme
	s := buildScheme()

	// create a polynomial
	f := randomPolynomial()

	// commit the polynomial
	digest := s.Commit(&f)

	// compute opening proof at a random point
	var point fr.Element
	point.SetString("4321")
	proof := s.Open(&point, &f)

	// verify correct proof
	err := s.Verify(digest, proof)
	if err != nil {
		t.Fatal(err)
	}

	// verify wrong proof
	_proof := proof.(*Proof)
	_proof.ClaimedValue.Double(&_proof.ClaimedValue)
	err = s.Verify(digest, _proof)
	if err == nil {
		t.Fatal("verifying wrong proof should have failed")
	}
}

func TestBatchVerifySinglePoint(t *testing.T) {

	// create a KZG scheme
	s := buildScheme()

	// create polynomials
	f := make([]polynomial.Polynomial, 10)
	for i := 0; i < 10; i++ {
		_f := randomPolynomial()
		f[i] = &_f
	}

	// commit the polynomials
	digests := make([]polynomial.Digest, 10)
	for i := 0; i < 10; i++ {
		digests[i] = s.Commit(f[i])
	}

	// compute opening proof at a random point
	var point fr.Element
	point.SetString("4321")
	proof := s.BatchOpenSinglePoint(&point, digests, f)

	// verify correct proof
	err := s.BatchVerifySinglePoint(digests, proof)
	if err != nil {
		t.Fatal(err)
	}

	// verify wrong proof
	_proof := proof.(*BatchProofsSinglePoint)
	_proof.ClaimedValues[0].Double(&_proof.ClaimedValues[0])
	err = s.BatchVerifySinglePoint(digests, _proof)
	if err == nil {
		t.Fatal("verifying wrong proof should have failed")
	}

}
