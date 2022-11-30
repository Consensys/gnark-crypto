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

// the numeration corresponds to the following formatting:
// num = int(p.Info.Basis)*4 + int(p.Info.Layout)*2 + int(p.Info.Status)
func toLagrange0(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	_p.Info.Layout = BitReverse
	_p.Info.Status = Unlocked
	d.FFT(_p.Coefficients, fft.DIF)
	return &_p
}

func toLagrange1(p *Polynomial, d *fft.Domain) *Polynomial {
	p.Info.Basis = Lagrange
	p.Info.Layout = BitReverse
	d.FFT(p.Coefficients, fft.DIF)
	return p
}

func toLagrange2(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	_p.Info.Layout = Regular
	_p.Info.Status = Unlocked
	d.FFT(_p.Coefficients, fft.DIT)
	return &_p
}

func toLagrange3(p *Polynomial, d *fft.Domain) *Polynomial {
	p.Info.Basis = Lagrange
	p.Info.Layout = Regular
	d.FFT(p.Coefficients, fft.DIT)
	return p
}

func toLagrange4(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

func toLagrange5(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

func toLagrange6(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

func toLagrange7(p *Polynomial, d *fft.Domain) *Polynomial {
	return p
}

func toLagrange8(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	_p.Info.Layout = Regular
	_p.Info.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF, true)
	d.FFT(_p.Coefficients, fft.DIT)
	return &_p
}

func toLagrange9(p *Polynomial, d *fft.Domain) *Polynomial {
	p.Info.Basis = Lagrange
	d.FFTInverse(p.Coefficients, fft.DIF, true)
	d.FFT(p.Coefficients, fft.DIT)
	return p
}

func toLagrange10(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	d.FFTInverse(_p.Coefficients, fft.DIT, true)
	d.FFT(_p.Coefficients, fft.DIF)
	return &_p
}

func toLagrange11(p *Polynomial, d *fft.Domain) *Polynomial {
	p.Info.Basis = Lagrange
	d.FFTInverse(p.Coefficients, fft.DIT, true)
	d.FFT(p.Coefficients, fft.DIF)
	return p
}

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

// toLagrange changes or returns a copy of p (according to its
// status, Locked or Unlocked), or modifies p to put it in Lagrange
// basis. The result is not bit reversed.
func toLagrange(p *Polynomial, d *fft.Domain) *Polynomial {
	return _toLagrange[getShapeID(*p)](p, d)
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
