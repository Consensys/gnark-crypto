// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package polynomial

import (
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPolynomialEval(t *testing.T) {

	// build polynomial
	f := make(Polynomial, 20)
	for i := 0; i < 20; i++ {
		f[i].SetOne()
	}

	// random value
	var point fr.Element
	point.SetRandom()

	// compute manually f(val)
	var expectedEval, one, den fr.Element
	var expo big.Int
	one.SetOne()
	expo.SetUint64(20)
	expectedEval.Exp(point, &expo).
		Sub(&expectedEval, &one)
	den.Sub(&point, &one)
	expectedEval.Div(&expectedEval, &den)

	// compute purported evaluation
	purportedEval := f.Eval(&point)

	// check
	if !purportedEval.Equal(&expectedEval) {
		t.Fatal("polynomial evaluation failed")
	}
}

func TestPolynomialAddConstantInPlace(t *testing.T) {

	// build polynomial
	f := make(Polynomial, 20)
	for i := 0; i < 20; i++ {
		f[i].SetOne()
	}

	// constant to add
	var c fr.Element
	c.SetRandom()

	// add constant
	f.AddConstantInPlace(&c)

	// check
	var expectedCoeffs, one fr.Element
	one.SetOne()
	expectedCoeffs.Add(&one, &c)
	for i := 0; i < 20; i++ {
		if !f[i].Equal(&expectedCoeffs) {
			t.Fatal("AddConstantInPlace failed")
		}
	}
}

func TestPolynomialSubConstantInPlace(t *testing.T) {

	// build polynomial
	f := make(Polynomial, 20)
	for i := 0; i < 20; i++ {
		f[i].SetOne()
	}

	// constant to sub
	var c fr.Element
	c.SetRandom()

	// sub constant
	f.SubConstantInPlace(&c)

	// check
	var expectedCoeffs, one fr.Element
	one.SetOne()
	expectedCoeffs.Sub(&one, &c)
	for i := 0; i < 20; i++ {
		if !f[i].Equal(&expectedCoeffs) {
			t.Fatal("SubConstantInPlace failed")
		}
	}
}

func TestPolynomialScaleInPlace(t *testing.T) {

	// build polynomial
	f := make(Polynomial, 20)
	for i := 0; i < 20; i++ {
		f[i].SetOne()
	}

	// constant to scale by
	var c fr.Element
	c.SetRandom()

	// scale by constant
	f.ScaleInPlace(&c)

	// check
	for i := 0; i < 20; i++ {
		if !f[i].Equal(&c) {
			t.Fatal("ScaleInPlace failed")
		}
	}

}

func TestPolynomialAdd(t *testing.T) {

	// build unbalanced polynomials
	f1 := make(Polynomial, 20)
	f1Backup := make(Polynomial, 20)
	for i := 0; i < 20; i++ {
		f1[i].SetOne()
		f1Backup[i].SetOne()
	}
	f2 := make(Polynomial, 10)
	f2Backup := make(Polynomial, 10)
	for i := 0; i < 10; i++ {
		f2[i].SetOne()
		f2Backup[i].SetOne()
	}

	// expected result
	var one, two fr.Element
	one.SetOne()
	two.Double(&one)
	expectedSum := make(Polynomial, 20)
	for i := 0; i < 10; i++ {
		expectedSum[i].Set(&two)
	}
	for i := 10; i < 20; i++ {
		expectedSum[i].Set(&one)
	}

	// caller is empty
	var g Polynomial
	g.Add(f1, f2)
	if !g.Equal(expectedSum) {
		t.Fatal("add polynomials fails")
	}
	if !f1.Equal(f1Backup) {
		t.Fatal("side effect, f1 should not have been modified")
	}
	if !f2.Equal(f2Backup) {
		t.Fatal("side effect, f2 should not have been modified")
	}

	// all operands are distinct
	_f1 := f1.Clone()
	_f1.Add(f1, f2)
	if !_f1.Equal(expectedSum) {
		t.Fatal("add polynomials fails")
	}
	if !f1.Equal(f1Backup) {
		t.Fatal("side effect, f1 should not have been modified")
	}
	if !f2.Equal(f2Backup) {
		t.Fatal("side effect, f2 should not have been modified")
	}

	// first operand = caller
	_f1 = f1.Clone()
	_f2 := f2.Clone()
	_f1.Add(_f1, _f2)
	if !_f1.Equal(expectedSum) {
		t.Fatal("add polynomials fails")
	}
	if !_f2.Equal(f2Backup) {
		t.Fatal("side effect, _f2 should not have been modified")
	}

	// second operand = caller
	_f1 = f1.Clone()
	_f2 = f2.Clone()
	_f1.Add(_f2, _f1)
	if !_f1.Equal(expectedSum) {
		t.Fatal("add polynomials fails")
	}
	if !_f2.Equal(f2Backup) {
		t.Fatal("side effect, _f2 should not have been modified")
	}
}

func TestPolynomialText(t *testing.T) {
	var one, negTwo fr.Element
	one.SetOne()
	negTwo.SetInt64(-2)

	p := Polynomial{one, negTwo, one}

	assert.Equal(t, "X² - 2X + 1", p.Text(10))
}

func TestPrecomputeLagrange(t *testing.T) {

	testForDomainSize := func(domainSize uint8) bool {
		polys := computeLagrangeBasis(domainSize)

		for l := uint8(0); l < domainSize; l++ {
			for i := uint8(0); i < domainSize; i++ {
				var I fr.Element
				I.SetUint64(uint64(i))
				y := polys[l].Eval(&I)

				if i == l && !y.IsOne() || i != l && !y.IsZero() {
					t.Errorf("domainSize = %d: p_%d(%d) = %s", domainSize, l, i, y.Text(10))
					return false
				}
			}
		}
		return true
	}

	t.Parallel()
	parameters := gopter.DefaultTestParameters()

	const maxLagrangeDomainSize = 12

	parameters.MinSuccessfulTests = maxLagrangeDomainSize

	properties := gopter.NewProperties(parameters)

	properties.Property("l'th lagrange polynomials must evaluate to 1 on l and 0 on other values in the domain", prop.ForAll(
		testForDomainSize,
		gen.UInt8Range(2, maxLagrangeDomainSize),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestLagrangeCache(t *testing.T) {
	for _, i := range []int{5, 2, 8, 4, 6, 3, 0} {
		b := getLagrangeBasis(uint8(i))
		assert.Equal(t, b, getLagrangeBasis(uint8(i))) // second call must yield the same result
	}
}
