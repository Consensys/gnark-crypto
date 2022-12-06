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

package iop

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
)

func printVector(v []fr.Element) {
	fmt.Printf("[")
	for i := 0; i < len(v); i++ {
		fmt.Printf("Fr(%s), ", v[i].String())
	}
	fmt.Printf("]\n")
}

func printPolynomials(p []*Polynomial) {
	fmt.Printf("[\n")
	for i := 0; i < len(p); i++ {
		printVector(p[i].Coefficients)
		fmt.Printf(",\n")
	}
	fmt.Printf("]\n")
}

func printLayout(f Form) {

	if f.Basis == Canonical {
		fmt.Printf("CANONICAL")
	} else if f.Basis == LagrangeCoset {
		fmt.Printf("LAGRANGE_COSET")
	} else {
		fmt.Printf("LAGRANGE")
	}
	fmt.Println("")

	if f.Layout == Regular {
		fmt.Printf("REGULAR")
	} else {
		fmt.Printf("BIT REVERSED")
	}
	fmt.Println("")

	if f.Status == Locked {
		fmt.Printf("LOCKED")
	} else {
		fmt.Printf("UNLOCKED")
	}
	fmt.Println("")
}

type modifier func(p *Polynomial, d *fft.Domain) *Polynomial

// return a copy of p
func copyPoly(p Polynomial) Polynomial {
	size := len(p.Coefficients)
	var r Polynomial
	r.Coefficients = make([]fr.Element, size)
	copy(r.Coefficients, p.Coefficients)
	r.Info = p.Info
	return r
}

// return an ID corresponding to the polynomial extra data
func getShapeID(p Polynomial) int {
	return int(p.Info.Basis)*4 + int(p.Info.Layout)*2 + int(p.Info.Status)
}

//----------------------------------------------------
// toLagrange

// the numeration corresponds to the following formatting:
// num = int(p.Info.Basis)*4 + int(p.Info.Layout)*2 + int(p.Info.Status)

// CANONICAL REGULAR LOCKED
func toLagrange0(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	_p.Info.Layout = BitReverse
	_p.Info.Status = Unlocked
	d.FFT(_p.Coefficients, fft.DIF)
	return &_p
}

// CANONICAL REGULAR UNLOCKED
func toLagrange1(p *Polynomial, d *fft.Domain) *Polynomial {
	p.Info.Basis = Lagrange
	p.Info.Layout = BitReverse
	d.FFT(p.Coefficients, fft.DIF)
	return p
}

// CANONICAL BITREVERSE LOCKED
func toLagrange2(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	_p.Info.Layout = Regular
	_p.Info.Status = Unlocked
	d.FFT(_p.Coefficients, fft.DIT)
	return &_p
}

// CANONICAL BITREVERSE UNLOCKED
func toLagrange3(p *Polynomial, d *fft.Domain) *Polynomial {
	p.Info.Basis = Lagrange
	p.Info.Layout = Regular
	d.FFT(p.Coefficients, fft.DIT)
	return p
}

// LAGRANGE REGULAR LOCKED
func toLagrange4(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE REGULAR UNLOCKED
func toLagrange5(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE BITREVERSE LOCKED
func toLagrange6(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE BITREVERSE UNLOCKED
func toLagrange7(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE_COSET REGULAR LOCKED
func toLagrange8(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	_p.Info.Layout = Regular
	_p.Info.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF, true)
	d.FFT(_p.Coefficients, fft.DIT)
	return &_p
}

// LAGRANGE_COSET REGULAR UNLOCKED
func toLagrange9(p *Polynomial, d *fft.Domain) *Polynomial {
	p.Info.Basis = Lagrange
	d.FFTInverse(p.Coefficients, fft.DIF, true)
	d.FFT(p.Coefficients, fft.DIT)
	return p
}

// LAGRANGE_COSET BITREVERSE LOCKED
func toLagrange10(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	d.FFTInverse(_p.Coefficients, fft.DIT, true)
	d.FFT(_p.Coefficients, fft.DIF)
	return &_p
}

// LAGRANGE_COSET BITREVERSE UNLOCKED
func toLagrange11(p *Polynomial, d *fft.Domain) *Polynomial {
	p.Info.Basis = Lagrange
	d.FFTInverse(p.Coefficients, fft.DIT, true)
	d.FFT(p.Coefficients, fft.DIF)
	return p
}

var _toLagrange [12]modifier = [12]modifier{
	toLagrange0,
	toLagrange1,
	toLagrange2,
	toLagrange3,
	toLagrange4,
	toLagrange5,
	toLagrange6,
	toLagrange7,
	toLagrange8,
	toLagrange9,
	toLagrange10,
	toLagrange11,
}

// toLagrange changes or returns a copy of p (according to its
// status, Locked or Unlocked), or modifies p to put it in Lagrange
// basis. The result is not bit reversed.
func toLagrange(p *Polynomial, d *fft.Domain) *Polynomial {
	return _toLagrange[getShapeID(*p)](p, d)
}

//----------------------------------------------------
// toCanonical

// CANONICAL REGULAR LOCKED
func toCanonical0(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

// CANONICAL REGULAR UNLOCKED
func toCanonical1(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

// CANONICAL BITREVERSE LOCKED
func toCanonical2(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

// CANONICAL BITREVERSE UNLOCKED
func toCanonical3(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE REGULAR LOCKED
func toCanonical4(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Canonical
	_p.Info.Layout = BitReverse
	_p.Info.Status = Unlocked
	d.FFT(_p.Coefficients, fft.DIF)
	return &_p
}

// LAGRANGE REGULAR UNLOCKED
func toCanonical5(p *Polynomial, d *fft.Domain) *Polynomial {
	d.FFT(p.Coefficients, fft.DIF)
	p.Info.Basis = Canonical
	p.Info.Layout = BitReverse
	return p
}

// LAGRANGE BITREVERSE LOCKED
func toCanonical6(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Canonical
	_p.Info.Layout = Regular
	_p.Info.Status = Unlocked
	d.FFT(_p.Coefficients, fft.DIT)
	return &_p
}

// LAGRANGE BITREVERSE UNLOCKED
func toCanonical7(p *Polynomial, d *fft.Domain) *Polynomial {
	d.FFT(p.Coefficients, fft.DIT)
	p.Info.Basis = Canonical
	p.Info.Layout = Regular
	return p
}

// LAGRANGE_COSET REGULAR LOCKED
func toCanonical8(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Canonical
	_p.Info.Layout = BitReverse
	_p.Info.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF, true)
	return &_p
}

// LAGRANGE_COSET REGULAR UNLOCKED
func toCanonical9(p *Polynomial, d *fft.Domain) *Polynomial {
	p.Info.Basis = Canonical
	p.Info.Layout = BitReverse
	d.FFT(p.Coefficients, fft.DIF, true)
	return p
}

// LAGRANGE_COSET BITREVERSE LOCKED
func toCanonical10(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Canonical
	_p.Info.Layout = Regular
	d.FFT(_p.Coefficients, fft.DIT, true)
	return &_p
}

// LAGRANGE_COSET BITREVERSE UNLOCKED
func toCanonical11(p *Polynomial, d *fft.Domain) *Polynomial {
	p.Info.Basis = Canonical
	p.Info.Layout = Regular
	d.FFT(p.Coefficients, fft.DIT, true)
	return p
}

var _toCanonical [12]modifier = [12]modifier{
	toCanonical0,
	toCanonical1,
	toCanonical2,
	toCanonical3,
	toCanonical4,
	toCanonical5,
	toCanonical6,
	toCanonical7,
	toCanonical8,
	toCanonical9,
	toCanonical10,
	toCanonical11,
}

// toCanonical changes or returns a copy of p (according to its
// status, Locked or Unlocked), or modifies p to put it in Lagrange
// basis. The result is not bit reversed.
func toCanonical(p *Polynomial, d *fft.Domain) *Polynomial {
	return _toCanonical[getShapeID(*p)](p, d)
}

//----------------------------------------------------
// exp functions until 5

func exp0(x fr.Element) fr.Element {
	var res fr.Element
	res.SetOne()
	return res
}

func exp1(x fr.Element) fr.Element {
	return x
}

func exp2(x fr.Element) fr.Element {
	return *x.Square(&x)
}

func exp3(x fr.Element) fr.Element {
	var res fr.Element
	res.Square(&x).Mul(&res, &x)
	return res
}

func exp4(x fr.Element) fr.Element {
	x.Square(&x).Square(&x)
	return x
}

func exp5(x fr.Element) fr.Element {
	var res fr.Element
	res.Square(&x).Square(&res).Mul(&res, &x)
	return res
}

// doesn't return any errors, it is a private method, that
// is assumed to be called with correct arguments.
func smallExp(x fr.Element, n int) fr.Element {
	if n == 2 {
		return exp2(x)
	}
	if n == 3 {
		return exp3(x)
	}
	if n == 4 {
		return exp4(x)
	}
	if n == 5 {
		return exp5(x)
	}
	return fr.Element{}
}
