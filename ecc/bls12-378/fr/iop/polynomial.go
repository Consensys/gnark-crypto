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

// Code generated by consensys/gnark-crypto DO NOT EDIT

package iop

import (
	"errors"
	"math/big"
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc/bls12-378/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-378/fr/fft"
)

//-----------------------------------------------------
// univariate polynomials

// Enum to tell in which basis a polynomial is represented.
type Basis int64

const (
	Canonical Basis = iota
	Lagrange
	LagrangeCoset
)

// Enum to tell if a polynomial is in bit reverse form or
// in the regular form.
type Layout int64

const (
	Regular Layout = iota
	BitReverse
)

// Form describes the form of a polynomial.
type Form struct {
	Basis  Basis
	Layout Layout
}

// Polynomial represents a polynomial, the vector of coefficients
// along with the basis and the layout.
type Polynomial struct {
	Coefficients []fr.Element
	Form
}

// NewPolynomial creates a new polynomial. The slice coeff NOT copied
// but directly assigned to the new polynomial.
func NewPolynomial(coeffs []fr.Element, form Form) *Polynomial {
	return &Polynomial{Coefficients: coeffs, Form: form}
}

// return a copy of p
func (p *Polynomial) Copy() *Polynomial {
	size := len(p.Coefficients)
	var r Polynomial
	r.Coefficients = make([]fr.Element, size)
	copy(r.Coefficients, p.Coefficients)
	r.Form = p.Form
	return &r
}

// return an ID corresponding to the polynomial extra data
func getShapeID(p Polynomial) int {
	return int(p.Basis)*2 + int(p.Layout)
}

// WrappedPolynomial wrapps a polynomial so that it is
// interpreted as P'(X)=P(\omega^{s}X).
// Size is the real size of the polynomial (seen as a vector).
// For instance if len(P)=32 but P.Size=8, it means that P has been
// extended (e.g. it is evaluated on a larger set) but P is a polynomial
// of degree 7.
// BlindedSize is the size of the polynomial when it is blinded. By
// default BlindedSize=Size, until the polynomial is blinded.
type WrappedPolynomial struct {
	P           *Polynomial
	Shift       int
	Size        int
	BlindedSize int
}

//----------------------------------------------------
// Blind a polynomial

// blindPoly blinds a polynomial q by adding Q(X)*(X^{n}-1),
// where deg Q = blindingOrder and Q is random, and n is the
// size of q. Sets the result to p and returns it.
//
// * bo blinding order,  it's the degree of Q, where the blinding is Q(X)*(X^{n}-1)
// where n is the size of wp. The size of wp is modified since the underlying
// polynomial is of bigger degree now. The new size is wp.Size+1+blindingOrder.
//
// /!\ The code panics if wq is not in canonical, regular layout
func (wp *WrappedPolynomial) Blind(wq *WrappedPolynomial, blindingOrder int) *WrappedPolynomial {

	// check that q is in canonical basis
	if wq.P.Basis != Canonical || wq.P.Layout != Regular {
		panic("the input must be in canonical basis, regular layout")
	}

	// take care of who is modified
	if wp != wq {
		wp.P = wq.P.Copy()
		wp.Shift = wq.Shift
		wp.Size = wq.Size
	}

	// we add Q*(x^{n}-1) so the new size is deg(Q)+n+1
	// where n is the size of wq.
	newSize := wp.Size + blindingOrder + 1

	// Resize wp. The size of wq might has already been increased
	// (e.g. whent the polynomial is evaluated on a larger domain),
	// if that's the case we don't resize the polynomial.
	offset := newSize - len(wp.P.Coefficients)
	if offset > 0 {
		z := make([]fr.Element, offset)
		wp.P.Coefficients = append(wp.P.Coefficients, z...)
	}

	// blinding: we add Q(X)(X^{n}-1) to P, where deg(Q)=blindingOrder
	var r fr.Element

	for i := 0; i <= blindingOrder; i++ {
		r.SetRandom()
		wp.P.Coefficients[i].Sub(&wp.P.Coefficients[i], &r)
		wp.P.Coefficients[i+wp.Size].Add(&wp.P.Coefficients[i+wp.Size], &r)
	}
	wp.BlindedSize = newSize

	return wp
}

//----------------------------------------------------
// Evaluation

// Evaluate evaluates p at x.
// The code panics if the function is not in canonical form.
func (p *Polynomial) Evaluate(x fr.Element) fr.Element {

	var r fr.Element
	if p.Basis != Canonical {
		panic("p must be in canonical basis")
	}

	if p.Layout == Regular {
		for i := len(p.Coefficients) - 1; i >= 0; i-- {
			r.Mul(&r, &x).Add(&r, &p.Coefficients[i])
		}
	} else {
		nn := uint64(64 - bits.TrailingZeros(uint(len(p.Coefficients))))
		for i := len(p.Coefficients) - 1; i >= 0; i-- {
			iRev := bits.Reverse64(uint64(i)) >> nn
			r.Mul(&r, &x).Add(&r, &p.Coefficients[iRev])
		}
	}

	return r

}

// Evaluate evaluates p at x.
// The code panics if the function is not in canonical form.
func (wp *WrappedPolynomial) Evaluate(x fr.Element) fr.Element {

	if wp.Shift == 0 {
		return wp.P.Evaluate(x)
	}

	// TODO find a way to retrieve the root properly instead of re generating the fft domain
	d := fft.NewDomain(uint64(wp.Size))
	var g fr.Element
	if wp.Shift <= 5 {
		g = smallExp(d.Generator, wp.Shift)
		x.Mul(&x, &g)
		return wp.P.Evaluate(x)
	}

	bs := big.NewInt(int64(wp.Shift))
	g = *g.Exp(g, bs)
	x.Mul(&x, &g)
	return wp.P.Evaluate(x)
}

// Copy returns a copy of wp. The underlying polynomial is copied, that
// it it's a new pointer to a newly alloacted polynomial. In particular
// the slice representing the coefficients of the polynomial is reallocated
// and its content is copied from wp's coefficients.
func (wp *WrappedPolynomial) Copy() *WrappedPolynomial {
	var res WrappedPolynomial
	res.P = wp.P.Copy()
	res.Shift = wp.Shift
	res.Size = wp.Size
	return &res
}

// WrapMe same as Copy, but the underlying polynomial is a pointer to
// wp's polynomial.
func (wp *WrappedPolynomial) WrapMe(shift int) *WrappedPolynomial {
	var res WrappedPolynomial
	res.P = wp.P
	res.Shift = shift
	res.Size = wp.Size
	res.BlindedSize = wp.Size
	return &res
}

// GetCoeff returns the i-th entry of wp, taking the layout in account.
func (wp *WrappedPolynomial) GetCoeff(i int) fr.Element {

	n := len(wp.P.Coefficients)
	rho := n / wp.Size
	if wp.P.Form.Layout == Regular {
		return wp.P.Coefficients[(i+rho*wp.Shift)%n]
	} else {
		nn := uint64(64 - bits.TrailingZeros(uint(n)))
		iRev := bits.Reverse64(uint64((i+rho*wp.Shift)%n)) >> nn
		return wp.P.Coefficients[iRev]
	}

}

//----------------------------------------------------
// ToRegular

func (p *Polynomial) ToRegular(q *Polynomial) *Polynomial {

	if p != q {
		*p = *q.Copy()
	}
	if p.Layout == Regular {
		return p
	}
	fft.BitReverse(p.Coefficients)
	p.Layout = Regular
	return p
}

func (wp *WrappedPolynomial) ToRegular(wq *WrappedPolynomial) *WrappedPolynomial {
	if wp != wq {
		*wp = *wq.Copy() // --> former content of wp is now a danlging pointer...
	}
	wp.P.ToRegular(wp.P)
	return wp
}

//----------------------------------------------------
// ToBitreverse

func (p *Polynomial) ToBitreverse(q *Polynomial) *Polynomial {

	if p != q {
		*p = *q.Copy()
	}
	if p.Layout == BitReverse {
		return p
	}
	fft.BitReverse(p.Coefficients)
	p.Layout = BitReverse
	return p
}

func (wp *WrappedPolynomial) ToBitreverse(wq *WrappedPolynomial) *WrappedPolynomial {
	if wp != wq {
		*wp = *wq.Copy() // --> former content of wp is now a danlging pointer...
	}
	wp.P.ToBitreverse(wp.P)
	return wp
}

//----------------------------------------------------
// Wrap a polynomial

// WrapMe returned a WrappedPolynomial from p.
// * shift integer meaning that the result should be interpreted as p(\omega^shift X)
// * size optional parameter telling the size of p (as a vector). If not provided,
// len(p) is the default size.
func (p *Polynomial) WrapMe(shift int, size ...int) *WrappedPolynomial {
	res := WrappedPolynomial{
		P:           p,
		Shift:       shift,
		Size:        len(p.Coefficients),
		BlindedSize: len(p.Coefficients),
	}
	if len(size) > 0 {
		res.Size = size[0]
	}
	return &res
}

//----------------------------------------------------
// toLagrange

// the numeration corresponds to the following formatting:
// num = int(p.Basis)*2 + int(p.Layout)

// CANONICAL REGULAR
func (p *Polynomial) toLagrange0(d *fft.Domain) *Polynomial {
	p.Basis = Lagrange
	p.Layout = BitReverse
	d.FFT(p.Coefficients, fft.DIF)
	return p
}

// CANONICAL BITREVERSE
func (p *Polynomial) toLagrange1(d *fft.Domain) *Polynomial {
	p.Basis = Lagrange
	p.Layout = Regular
	d.FFT(p.Coefficients, fft.DIT)
	return p
}

// LAGRANGE REGULAR
func (p *Polynomial) toLagrange2(d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE BITREVERSE
func (p *Polynomial) toLagrange3(d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE_COSET REGULAR
func (p *Polynomial) toLagrange4(d *fft.Domain) *Polynomial {
	p.Basis = Lagrange
	p.Layout = Regular
	d.FFTInverse(p.Coefficients, fft.DIF, true)
	d.FFT(p.Coefficients, fft.DIT)
	return p
}

// LAGRANGE_COSET BITREVERSE
func (p *Polynomial) toLagrange5(d *fft.Domain) *Polynomial {
	p.Basis = Lagrange
	p.Layout = BitReverse
	d.FFTInverse(p.Coefficients, fft.DIT, true)
	d.FFT(p.Coefficients, fft.DIF)
	return p
}

// Set p to q in Lagrange form and returns it.
func (p *Polynomial) ToLagrange(q *Polynomial, d *fft.Domain) *Polynomial {
	id := getShapeID(*q)
	if q != p {
		*p = *q.Copy()
	}
	resize(p, d.Cardinality)
	switch id {
	case 0:
		return p.toLagrange0(d)
	case 1:
		return p.toLagrange1(d)
	case 2:
		return p.toLagrange2(d)
	case 3:
		return p.toLagrange3(d)
	case 4:
		return p.toLagrange4(d)
	case 5:
		return p.toLagrange5(d)
	default:
		panic("unknown ID")
	}
}

// ToLagrange Sets wp to wq, in ToLagrange form and returns it.
func (wp *WrappedPolynomial) ToLagrange(wq *WrappedPolynomial, d *fft.Domain) *WrappedPolynomial {
	if wp != wq {
		*wp = *wq.Copy() // --> former content of wp is now a danlging pointer...
	}
	wp.P.ToLagrange(wp.P, d)
	return wp
}

//----------------------------------------------------
// toCanonical

// CANONICAL REGULAR
func (p *Polynomial) toCanonical0(d *fft.Domain) *Polynomial {
	return p
}

// CANONICAL BITREVERSE
func (p *Polynomial) toCanonical1(d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE REGULAR
func (p *Polynomial) toCanonical2(d *fft.Domain) *Polynomial {
	p.Basis = Canonical
	p.Layout = BitReverse
	d.FFTInverse(p.Coefficients, fft.DIF)
	return p
}

// LAGRANGE BITREVERSE
func (p *Polynomial) toCanonical3(d *fft.Domain) *Polynomial {
	p.Basis = Canonical
	p.Layout = Regular
	d.FFTInverse(p.Coefficients, fft.DIT)
	return p
}

// LAGRANGE_COSET REGULAR
func (p *Polynomial) toCanonical4(d *fft.Domain) *Polynomial {
	p.Basis = Canonical
	p.Layout = BitReverse
	d.FFTInverse(p.Coefficients, fft.DIF, true)
	return p
}

// LAGRANGE_COSET BITREVERSE
func (p *Polynomial) toCanonical5(d *fft.Domain) *Polynomial {
	p.Basis = Canonical
	p.Layout = Regular
	d.FFTInverse(p.Coefficients, fft.DIT, true)
	return p
}

// ToCanonical Sets p to q, in canonical form and returns it.
func (p *Polynomial) ToCanonical(q *Polynomial, d *fft.Domain) *Polynomial {
	id := getShapeID(*q)
	if q != p {
		*p = *q.Copy()
	}
	resize(p, d.Cardinality)
	switch id {
	case 0:
		return p.toCanonical0(d)
	case 1:
		return p.toCanonical1(d)
	case 2:
		return p.toCanonical2(d)
	case 3:
		return p.toCanonical3(d)
	case 4:
		return p.toCanonical4(d)
	case 5:
		return p.toCanonical5(d)
	default:
		panic("unknown ID")
	}
}

// ToCanonical Sets wp to wq, in canonical form and returns it.
func (wp *WrappedPolynomial) ToCanonical(wq *WrappedPolynomial, d *fft.Domain) *WrappedPolynomial {
	if wp != wq {
		*wp = *wq.Copy() // --> former content of wp is now a danlging pointer...
	}
	wp.P.ToCanonical(wp.P, d)
	return wp
}

//-----------------------------------------------------
// ToLagrangeCoset

func resize(p *Polynomial, newSize uint64) {
	z := make([]fr.Element, int(newSize)-len(p.Coefficients))
	p.Coefficients = append(p.Coefficients, z...)
}

// CANONICAL REGULAR
func (p *Polynomial) toLagrangeCoset0(d *fft.Domain) *Polynomial {
	p.Basis = LagrangeCoset
	p.Layout = BitReverse
	d.FFT(p.Coefficients, fft.DIF, true)
	return p
}

// CANONICAL BITREVERSE
func (p *Polynomial) toLagrangeCoset1(d *fft.Domain) *Polynomial {
	p.Basis = LagrangeCoset
	p.Layout = Regular
	d.FFT(p.Coefficients, fft.DIT, true)
	return p
}

// LAGRANGE REGULAR
func (p *Polynomial) toLagrangeCoset2(d *fft.Domain) *Polynomial {
	p.Basis = LagrangeCoset
	p.Layout = Regular
	d.FFTInverse(p.Coefficients, fft.DIF)
	d.FFT(p.Coefficients, fft.DIT, true)
	return p
}

// LAGRANGE BITREVERSE
func (p *Polynomial) toLagrangeCoset3(d *fft.Domain) *Polynomial {
	p.Basis = LagrangeCoset
	p.Layout = BitReverse
	d.FFTInverse(p.Coefficients, fft.DIT)
	d.FFT(p.Coefficients, fft.DIF, true)
	return p
}

// LAGRANGE_COSET REGULAR
func (p *Polynomial) toLagrangeCoset4(d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE_COSET BITREVERSE
func (p *Polynomial) toLagrangeCoset5(d *fft.Domain) *Polynomial {
	return p
}

// ToLagrangeCoset Sets p to q, in LagrangeCoset form and returns it.
func (p *Polynomial) ToLagrangeCoset(q *Polynomial, d *fft.Domain) *Polynomial {
	id := getShapeID(*q)
	if q != p {
		*p = *q.Copy()
	}
	resize(p, d.Cardinality)
	switch id {
	case 0:
		return p.toLagrangeCoset0(d)
	case 1:
		return p.toLagrangeCoset1(d)
	case 2:
		return p.toLagrangeCoset2(d)
	case 3:
		return p.toLagrangeCoset3(d)
	case 4:
		return p.toLagrangeCoset4(d)
	case 5:
		return p.toLagrangeCoset5(d)
	default:
		panic("unknown ID")
	}
}

// ToLagrangeCoset Sets wp to wq, in LagrangeCoset form and returns it.
func (wp *WrappedPolynomial) ToLagrangeCoset(wq *WrappedPolynomial, d *fft.Domain) *WrappedPolynomial {
	if wp != wq {
		*wp = *wq.Copy() // --> former content of wp is now a danlging pointer...
	}
	wp.P.ToLagrangeCoset(wp.P, d)
	return wp
}

//-----------------------------------------------------
// multivariate polynomials

// errors related to the polynomials.
var ErrInconsistentNumberOfVariable = errors.New("the number of variables is not consistent")

// Monomial represents a Monomial encoded as
// coeff*X₁^{i₁}*..*X_n^{i_n} if exponents = [i₁,..iₙ]
type Monomial struct {
	coeff     fr.Element
	exponents []int
}

// it is supposed that the number of variables matches
func (m Monomial) evaluate(x []fr.Element) fr.Element {

	var res, tmp fr.Element

	nbVars := len(x)
	res.SetOne()
	for i := 0; i < nbVars; i++ {
		if m.exponents[i] <= 5 {
			tmp = smallExp(x[i], m.exponents[i])
			res.Mul(&res, &tmp)
			continue
		}
		bi := big.NewInt(int64(i))
		tmp.Exp(x[i], bi)
		res.Mul(&res, &tmp)
	}
	res.Mul(&res, &m.coeff)

	return res

}

// reprensents a multivariate polynomial as a list of Monomial,
// the multivariate polynomial being the sum of the Monomials.
type MultivariatePolynomial struct {
	M []Monomial
	C fr.Element
}

// degree returns the total degree
func (m *MultivariatePolynomial) Degree() uint64 {
	r := 0
	for i := 0; i < len(m.M); i++ {
		t := 0
		for j := 0; j < len(m.M[i].exponents); j++ {
			t += m.M[i].exponents[j]
		}
		if t > r {
			r = t
		}
	}
	return uint64(r)
}

// AddMonomial adds a Monomial to m. If m is empty, the Monomial is
// added no matter what. But if m is already populated, an error is
// returned if len(e)\neq size of the previous list of exponents. This
// ensure that the number of variables is given by the size of any of
// the slices of exponent in any Monomial.
func (m *MultivariatePolynomial) AddMonomial(c fr.Element, e []int) error {

	// if m is empty, we add the first Monomial.
	if len(m.M) == 0 {
		r := Monomial{c, e}
		m.M = append(m.M, r)
		return nil
	}

	// at this stage all of exponennt in m are supposed to be of
	// the same size.
	if len(m.M[0].exponents) != len(e) {
		return ErrInconsistentNumberOfVariable
	}
	r := Monomial{c, e}
	m.M = append(m.M, r)
	return nil

}

// EvaluateSinglePoint a multivariate polynomial in x
// /!\ It is assumed that the multivariate polynomial has been
// built correctly, that is the sizes of the slices in exponents
// are the same /!\
func (m *MultivariatePolynomial) EvaluateSinglePoint(x []fr.Element) fr.Element {

	var res fr.Element

	for i := 0; i < len(m.M); i++ {
		tmp := m.M[i].evaluate(x)
		res.Add(&res, &tmp)
	}
	res.Add(&res, &m.C)
	return res
}

// EvaluatePolynomials evaluate h on x, interpreted as vectors.
// No transformations are made on the polynomials.
// The basis of the returned polynomial is the same as x[0]'s, and
// the layout is Regular.
func (m *MultivariatePolynomial) EvaluatePolynomials(x []WrappedPolynomial) (Polynomial, error) {

	var res Polynomial

	// check that the sizes are consistent
	nbPolynomials := len(x)
	nbElmts := len(x[0].P.Coefficients)
	for i := 0; i < nbPolynomials; i++ {
		if len(x[i].P.Coefficients) != nbElmts {
			return res, ErrInconsistentSize
		}
	}

	// compute \rho for all polynomials
	rho := make([]int, nbPolynomials)
	for i := 0; i < nbPolynomials; i++ {
		rho[i] = len(x[i].P.Coefficients) / x[i].Size
	}

	res.Coefficients = make([]fr.Element, nbElmts)

	v := make([]fr.Element, nbPolynomials)
	nn := uint64(64 - bits.TrailingZeros(uint(nbElmts)))

	for i := 0; i < nbElmts; i++ {

		for j := 0; j < nbPolynomials; j++ {

			if x[j].P.Form.Layout == Regular {

				v[j].Set(&x[j].P.Coefficients[(i+x[j].Shift*rho[j])%nbElmts])

			} else {

				// take in account the fact that the polynomial mght be shifted...
				iRev := bits.Reverse64(uint64((i+x[j].Shift*rho[j]))%uint64(nbElmts)) >> nn
				v[j].Set(&x[j].P.Coefficients[iRev])
			}

		}

		// evaluate h on x
		res.Coefficients[i] = m.EvaluateSinglePoint(v)

	}
	res.Form.Basis = x[0].P.Form.Basis
	res.Form.Layout = Regular

	return res, nil

}
