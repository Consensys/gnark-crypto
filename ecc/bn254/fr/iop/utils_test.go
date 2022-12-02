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
	"reflect"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
)

func randomVector(size int) []fr.Element {

	r := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		r[i].SetRandom()
	}
	return r
}

// list of functions to turn a polynomial in Lagrange-regular form
// to all different forms in ordered using this encoding:
// int(p.Info.Basis)*4 + int(p.Info.Layout)*2 + int(p.Info.Status)
// p is in Lagrange/Regular here. This function is for testing purpose
// only.
type TransfoTest func(p Polynomial, d *fft.Domain) Polynomial

// CANONICAL REGULAR LOCKED
func fromLagrange0(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[0]
	d.FFTInverse(r.Coefficients, fft.DIF)
	fft.BitReverse(r.Coefficients)
	return &r
}

// CANONICAL REGULAR UNLOCKED
func fromLagrange1(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[1]
	d.FFTInverse(r.Coefficients, fft.DIF)
	fft.BitReverse(r.Coefficients)
	return &r
}

// CANONICAL BITREVERSE LOCKED
func fromLagrange2(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[2]
	d.FFTInverse(r.Coefficients, fft.DIF)
	return &r
}

// CANONICAL BITREVERSE UNLOCKED
func fromLagrange3(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[3]
	d.FFTInverse(r.Coefficients, fft.DIF)
	return &r
}

// LAGRANGE REGULAR LOCKED
func fromLagrange4(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[4]
	return &r
}

// LAGRANGE REGULAR UNLOCKED
func fromLagrange5(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[5]

	return &r
}

// LAGRANGE BITREVERSE LOCKED
func fromLagrange6(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[6]
	fft.BitReverse(r.Coefficients)
	return &r
}

// LAGRANGE BITREVERSE UNLOCKED
func fromLagrange7(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[7]
	fft.BitReverse(r.Coefficients)
	return &r
}

// LAGRANGE_COSET REGULAR LOCKED
func fromLagrange8(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[8]
	d.FFTInverse(r.Coefficients, fft.DIF)
	d.FFT(r.Coefficients, fft.DIT, true)
	return &r
}

// LAGRANGE_COSET REGULAR UNLOCKED
func fromLagrange9(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[9]
	d.FFTInverse(r.Coefficients, fft.DIF)
	d.FFT(r.Coefficients, fft.DIT, true)
	return &r
}

// LAGRANGE_COSET BITREVERSE LOCKED
func fromLagrange10(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[10]
	d.FFTInverse(r.Coefficients, fft.DIF)
	d.FFT(r.Coefficients, fft.DIT, true)
	fft.BitReverse(r.Coefficients)
	return &r
}

// LAGRANGE_COSET BITREVERSE UNLOCKED
func fromLagrange11(p *Polynomial, d *fft.Domain) *Polynomial {
	info := getListInfo()
	r := copyPoly(*p)
	r.Info = info[11]
	d.FFTInverse(r.Coefficients, fft.DIF)
	d.FFT(r.Coefficients, fft.DIT, true)
	fft.BitReverse(r.Coefficients)
	return &r
}

var fromLagrange [12]modifier = [12]modifier{
	fromLagrange0,
	fromLagrange1,
	fromLagrange2,
	fromLagrange3,
	fromLagrange4,
	fromLagrange5,
	fromLagrange6,
	fromLagrange7,
	fromLagrange8,
	fromLagrange9,
	fromLagrange10,
	fromLagrange11,
}

// return all the possible form combination, in the same order
// a generateTestPolynomials
func getListInfo() []Form {

	res := make([]Form, 12)

	res[0].Basis = Canonical
	res[0].Layout = Regular
	res[0].Status = Locked

	res[1].Basis = Canonical
	res[1].Layout = Regular
	res[1].Status = Unlocked

	res[2].Basis = Canonical
	res[2].Layout = BitReverse
	res[2].Status = Locked

	res[3].Basis = Canonical
	res[3].Layout = BitReverse
	res[3].Status = Unlocked

	res[4].Basis = Lagrange
	res[4].Layout = Regular
	res[4].Status = Locked

	res[5].Basis = Lagrange
	res[5].Layout = Regular
	res[5].Status = Unlocked

	res[6].Basis = Lagrange
	res[6].Layout = BitReverse
	res[6].Status = Locked

	res[7].Basis = Lagrange
	res[7].Layout = BitReverse
	res[7].Status = Unlocked

	res[8].Basis = LagrangeCoset
	res[8].Layout = Regular
	res[8].Status = Locked

	res[9].Basis = LagrangeCoset
	res[9].Layout = Regular
	res[9].Status = Unlocked

	res[10].Basis = LagrangeCoset
	res[10].Layout = BitReverse
	res[10].Status = Locked

	res[11].Basis = LagrangeCoset
	res[11].Layout = BitReverse
	res[11].Status = Unlocked

	return res
}

func getCopy(l []Polynomial) []Polynomial {
	r := make([]Polynomial, len(l))
	for i := 0; i < len(l); i++ {
		r[i].Coefficients = make([]fr.Element, len(l[i].Coefficients))
		copy(r[i].Coefficients, l[i].Coefficients)
		r[i].Info = l[i].Info
	}
	return r
}

func cmpCoefficents(p, q []fr.Element) bool {
	if len(p) != len(q) {
		return false
	}
	res := true
	for i := 0; i < len(p); i++ {
		res = res && (p[i].Equal(&q[i]))
	}
	return res
}

func TestPutInLagrangeForm(t *testing.T) {

	size := 64
	domain := fft.NewDomain(uint64(size))

	// reference vector in Lagrange-regular form
	c := randomVector(size)
	var regular, bitreversed Polynomial
	regular.Coefficients = c
	regular.Info.Basis = Lagrange
	regular.Info.Layout = Regular
	regular.Info.Status = Locked
	bitreversed.Info = regular.Info
	bitreversed.Coefficients = make([]fr.Element, size)
	copy(bitreversed.Coefficients, regular.Coefficients)
	polynomials := make([]Polynomial, 12)
	for i := 0; i < 12; i++ {
		polynomials[i] = *fromLagrange[i](&regular, domain)
	}

	// check that the info field is filled correctly
	info := getListInfo()
	for i := 0; i < 12; i++ {
		if polynomials[i].Info != info[i] {
			t.Fatal("info field is not filled correcly")
		}
	}

	// check that r has not been modified
	if !reflect.DeepEqual(&regular, &bitreversed) {
		t.Fatal("reference polynomial should not be modified")
	}

	// bit the reference vector for the bit reversed case
	bitreversed.Info.Layout = BitReverse
	fft.BitReverse(bitreversed.Coefficients)

	// get a backup
	backupPolynomials := getCopy(polynomials)

	// create the Lagrange form...
	lagrangePolynomials := make([]Polynomial, 12)
	for i := 0; i < 12; i++ {
		lagrangePolynomials[i] = *toLagrange(&polynomials[i], domain)
	}

	// compare the results that should be in regular form
	if lagrangePolynomials[2].Info.Layout != Regular {
		t.Fatal("expected layout is Regular")
	}
	if !cmpCoefficents(lagrangePolynomials[2].Coefficients, regular.Coefficients) {
		t.Fatal("Lagrange form is not as expected")
	}
	if lagrangePolynomials[3].Info.Layout != Regular {
		t.Fatal("expected layout is Regular")
	}
	if !cmpCoefficents(lagrangePolynomials[3].Coefficients, regular.Coefficients) {
		t.Fatal("Lagrange form is not as expected")
	}
	if lagrangePolynomials[4].Info.Layout != Regular {
		t.Fatal("expected layout is Regular")
	}
	if !cmpCoefficents(lagrangePolynomials[4].Coefficients, regular.Coefficients) {
		t.Fatal("Lagrange form is not as expected")
	}
	if lagrangePolynomials[5].Info.Layout != Regular {
		t.Fatal("expected layout is Regular")
	}
	if !cmpCoefficents(lagrangePolynomials[5].Coefficients, regular.Coefficients) {
		t.Fatal("Lagrange form is not as expected")
	}
	if lagrangePolynomials[8].Info.Layout != Regular {
		t.Fatal("expected layout is Regular")
	}
	if !cmpCoefficents(lagrangePolynomials[8].Coefficients, regular.Coefficients) {
		t.Fatal("Lagrange form is not as expected")
	}
	if lagrangePolynomials[9].Info.Layout != Regular {
		t.Fatal("expected layout is Regular")
	}
	if !cmpCoefficents(lagrangePolynomials[9].Coefficients, regular.Coefficients) {
		t.Fatal("Lagrange form is not as expected")
	}

	// compare the results that should be in bit reversed form
	if lagrangePolynomials[0].Info.Layout != BitReverse {
		t.Fatal("expected layout is bit reversed")
	}
	if !cmpCoefficents(lagrangePolynomials[0].Coefficients, bitreversed.Coefficients) {
		t.Fatal("bit reversed form is not as expected")
	}
	if lagrangePolynomials[1].Info.Layout != BitReverse {
		t.Fatal("expected layout is bit reversed")
	}
	if !cmpCoefficents(lagrangePolynomials[1].Coefficients, bitreversed.Coefficients) {
		t.Fatal("bit reversed form is not as expected")
	}
	if lagrangePolynomials[6].Info.Layout != BitReverse {
		t.Fatal("expected layout is bit reversed")
	}
	if !cmpCoefficents(lagrangePolynomials[6].Coefficients, bitreversed.Coefficients) {
		t.Fatal("bit reversed form is not as expected")
	}
	if lagrangePolynomials[7].Info.Layout != BitReverse {
		t.Fatal("expected layout is bit reversed")
	}
	if !cmpCoefficents(lagrangePolynomials[7].Coefficients, bitreversed.Coefficients) {
		t.Fatal("bit reversed form is not as expected")
	}
	if lagrangePolynomials[10].Info.Layout != BitReverse {
		t.Fatal("expected layout is bit reversed")
	}
	if !cmpCoefficents(lagrangePolynomials[10].Coefficients, bitreversed.Coefficients) {
		t.Fatal("bit reversed form is not as expected")
	}
	if lagrangePolynomials[11].Info.Layout != BitReverse {
		t.Fatal("expected layout is bit reversed")
	}
	if !cmpCoefficents(lagrangePolynomials[11].Coefficients, bitreversed.Coefficients) {
		t.Fatal("bit reversed form is not as expected")
	}

	// compare the result that shouldn't be modified
	for i := 0; i < 6; i++ {
		if !reflect.DeepEqual(&polynomials[2*i], &backupPolynomials[2*i]) {
			t.Fatal("locked polynomials should not be modified")
		}
	}

	// compare the result that should be modified
	if reflect.DeepEqual(&polynomials[1], &backupPolynomials[1]) {
		t.Fatal("unlocked polynomial should be modified")
	}
	if reflect.DeepEqual(&polynomials[3], &backupPolynomials[3]) {
		t.Fatal("unlocked polynomial should be modified")
	}
	if reflect.DeepEqual(&polynomials[9], &backupPolynomials[9]) {
		t.Fatal("unlocked polynomial should be modified")
	}
	if reflect.DeepEqual(&polynomials[11], &backupPolynomials[11]) {
		t.Fatal("unlocked polynomial should be modified")
	}
}

// CANONICAL REGULAR LOCKED
func fromCanonical0(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Canonical
	_p.Info.Layout = Regular
	_p.Info.Status = Locked
	return &_p
}

// CANONICAL REGULAR UNLOCKED
func fromCanonical1(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Canonical
	_p.Info.Layout = Regular
	_p.Info.Status = Unlocked
	return &_p
}

// CANONICAL BITREVERSE LOCKED
func fromCanonical2(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Canonical
	_p.Info.Layout = BitReverse
	_p.Info.Status = Locked
	return &_p
}

// CANONICAL BITREVERSE UNLOCKED
func fromCanonical3(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Canonical
	_p.Info.Layout = BitReverse
	_p.Info.Status = Unlocked
	return &_p
}

// LAGRANGE REGULAR LOCKED
func fromCanonical4(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	_p.Info.Layout = Regular
	_p.Info.Status = Locked
	d.FFTInverse(_p.Coefficients, fft.DIF)
	fft.BitReverse(_p.Coefficients)
	return &_p
}

// LAGRANGE REGULAR UNLOCKED
func fromCanonical5(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	_p.Info.Layout = Regular
	_p.Info.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF)
	fft.BitReverse(_p.Coefficients)
	return &_p
}

// LAGRANGE BITREVERSE LOCKED
func fromCanonical6(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	_p.Info.Layout = BitReverse
	_p.Info.Status = Locked
	d.FFTInverse(_p.Coefficients, fft.DIF)
	return &_p
}

// LAGRANGE BITREVERSE UNLOCKED
func fromCanonical7(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = Lagrange
	_p.Info.Layout = BitReverse
	_p.Info.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF)
	return &_p
}

// LAGRANGE_COSET REGULAR LOCKED
func fromCanonical8(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = LagrangeCoset
	_p.Info.Layout = Regular
	_p.Info.Status = Locked
	d.FFTInverse(_p.Coefficients, fft.DIF, true)
	fft.BitReverse(_p.Coefficients)
	return &_p
}

// LAGRANGE_COSET REGULAR UNLOCKED
func fromCanonical9(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = LagrangeCoset
	_p.Info.Layout = Regular
	_p.Info.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF, true)
	fft.BitReverse(_p.Coefficients)
	return &_p
}

// LAGRANGE_COSET BITREVERSE LOCKED
func fromCanonical10(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = LagrangeCoset
	_p.Info.Layout = BitReverse
	_p.Info.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF, true)
	return &_p
}

// LAGRANGE_COSET BITREVERSE UNLOCKED
func fromCanonical11(p *Polynomial, d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Info.Basis = LagrangeCoset
	_p.Info.Layout = BitReverse
	_p.Info.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF, true)
	return &_p
}

func TestPutInCanonicalForm(t *testing.T) {

	size := 64
	domain := fft.NewDomain(uint64(size))

	// reference vector in Lagrange-regular form
	c := randomVector(size)
	var p Polynomial
	p.Coefficients = c
	p.Info.Basis = Canonical
	p.Info.Layout = Regular
	p.Info.Status = Locked

	// CANONICAL REGULAR LOCKED
	{
		_p := fromCanonical0(&p, domain)
		// backup := copyPoly(*_p)
		q := toCanonical0(_p, domain)
		// if !reflect.DeepEqual(_p, backup) {
		// 	t.Fatal("locked polynomial should not be modified")
		// }
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != Regular {
			t.Fatal("epxected layout is regular")
		}
		if q.Info.Status != Locked {
			t.Fatal("expected status is locked")
		}
		if !cmpCoefficents(q.Coefficients, p.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

	// CANONICAL REGULAR UNLOCKED
	{
		_p := fromCanonical1(&p, domain)
		q := toCanonical1(_p, domain)
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != Regular {
			t.Fatal("epxected layout is regular")
		}
		if q.Info.Status != Unlocked {
			t.Fatal("expected status is locked")
		}
		if !cmpCoefficents(q.Coefficients, p.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

	// CANONICAL BITREVERSE LOCKED
	{
		_p := fromCanonical2(&p, domain)
		// backup := copyPoly(*_p)
		q := toCanonical1(_p, domain)
		// if !reflect.DeepEqual(_p, backup) {
		// 	t.Fatal("locked polynomial should not be modified")
		// }
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != BitReverse {
			t.Fatal("epxected layout is bitReverse")
		}
		if q.Info.Status != Locked {
			t.Fatal("expected status is locked")
		}
		if !cmpCoefficents(q.Coefficients, p.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

	// CANONICAL BITREVERSE UNLOCKED
	{
		_p := fromCanonical3(&p, domain)
		q := toCanonical3(_p, domain)
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != BitReverse {
			t.Fatal("epxected layout is bitReverse")
		}
		if q.Info.Status != Unlocked {
			t.Fatal("expected status is unLocked")
		}
		if !cmpCoefficents(q.Coefficients, p.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

	// LAGRANGE REGULAR LOCKED
	{
		_p := fromCanonical4(&p, domain)
		// backup := copyPoly(*_p)
		q := toCanonical4(_p, domain)
		// if !reflect.DeepEqual(_p, backup) {
		// 	t.Fatal("locked polynomial should not be modified")
		// }
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != BitReverse {
			t.Fatal("epxected layout is bitReverse")
		}
		if q.Info.Status != Unlocked {
			t.Fatal("expected status is unlocked")
		}
		fft.BitReverse(q.Coefficients)
		if !cmpCoefficents(p.Coefficients, q.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

	// LAGRANGE REGULAR UNLOCKED
	{
		_p := fromCanonical5(&p, domain)
		q := toCanonical5(_p, domain)
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != BitReverse {
			t.Fatal("epxected layout is bitReverse")
		}
		if q.Info.Status != Unlocked {
			t.Fatal("expected status is unlocked")
		}
		fft.BitReverse(q.Coefficients)
		if !cmpCoefficents(q.Coefficients, p.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

	// LAGRANGE BITREVERSE LOCKED
	{
		_p := fromCanonical6(&p, domain)
		// backup := copyPoly(*_p)
		q := toCanonical6(_p, domain)
		// if !reflect.DeepEqual(backup, *_p){
		// 	t.Fatal("")
		// }
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != Regular {
			t.Fatal("epxected layout is regular")
		}
		if q.Info.Status != Unlocked {
			t.Fatal("expected status is unlocked")
		}
		if !cmpCoefficents(q.Coefficients, p.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

	// LAGRANGE BITREVERSE UNLOCKED
	{
		_p := fromCanonical7(&p, domain)
		// backup := copyPoly(*_p)
		q := toCanonical7(_p, domain)
		// if !reflect.DeepEqual(backup, *_p){
		// 	t.Fatal("")
		// }
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != Regular {
			t.Fatal("epxected layout is regular")
		}
		if q.Info.Status != Unlocked {
			t.Fatal("expected status is unlocked")
		}
		if !cmpCoefficents(q.Coefficients, p.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

	// LAGRANGE_COSET REGULAR LOCKED
	{
		_p := fromCanonical8(&p, domain)
		// backup := copyPoly(*_p)
		q := toCanonical8(_p, domain)
		// if !reflect.DeepEqual(backup, *_p){
		// 	t.Fatal("")
		// }
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != BitReverse {
			t.Fatal("epxected layout is bitreverse")
		}
		if q.Info.Status != Unlocked {
			t.Fatal("expected status is unlocked")
		}
		// fft.BitReverse(q.Coefficients)
		// if !cmpCoefficents(q.Coefficients, p.Coefficients) {
		// 	t.Fatal("wrong coefficients")
		// }
	}

	// LAGRANGE_COSET REGULAR UNLOCKED
	{
		_p := fromCanonical9(&p, domain)
		// backup := copyPoly(*_p)
		q := toCanonical9(_p, domain)
		// if !reflect.DeepEqual(backup, *_p){
		// 	t.Fatal("")
		// }
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != BitReverse {
			t.Fatal("epxected layout is bitreverse")
		}
		if q.Info.Status != Unlocked {
			t.Fatal("expected status is unlocked")
		}
		fft.BitReverse(q.Coefficients)
		if !cmpCoefficents(q.Coefficients, p.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

	// LAGRANGE_COSET BITREVERSE LOCKED
	{
		_p := fromCanonical10(&p, domain)
		// backup := copyPoly(*_p)
		q := toCanonical10(_p, domain)
		// if !reflect.DeepEqual(backup, *_p){
		// 	t.Fatal("")
		// }
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != Regular {
			t.Fatal("epxected layout is regular")
		}
		if q.Info.Status != Unlocked {
			t.Fatal("expected status is unlocked")
		}
		if !cmpCoefficents(q.Coefficients, p.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

	// LAGRANGE_COSET BITREVERSE UNLOCKED
	{
		_p := fromCanonical11(&p, domain)
		// backup := copyPoly(*_p)
		q := toCanonical11(_p, domain)
		// if !reflect.DeepEqual(backup, *_p){
		// 	t.Fatal("")
		// }
		if q.Info.Basis != Canonical {
			t.Fatal("expected basis is canonical")
		}
		if q.Info.Layout != Regular {
			t.Fatal("epxected layout is regular")
		}
		if q.Info.Status != Unlocked {
			t.Fatal("expected status is unlocked")
		}
		if !cmpCoefficents(q.Coefficients, p.Coefficients) {
			t.Fatal("wrong coefficients")
		}
	}

}
