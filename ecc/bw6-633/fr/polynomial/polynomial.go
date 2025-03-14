// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package polynomial

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bw6-633/fr"
	"github.com/consensys/gnark-crypto/utils"
	"strconv"
	"strings"
	"sync"
)

// Polynomial represented by coefficients in the field.
type Polynomial []fr.Element

// Degree returns the degree of the polynomial, which is the length of Data.
func (p *Polynomial) Degree() uint64 {
	return uint64(len(*p) - 1)
}

// Eval evaluates p at v
// returns a fr.Element
func (p *Polynomial) Eval(v *fr.Element) fr.Element {

	res := (*p)[len(*p)-1]
	for i := len(*p) - 2; i >= 0; i-- {
		res.Mul(&res, v)
		res.Add(&res, &(*p)[i])
	}

	return res
}

// Clone returns a copy of the polynomial
func (p *Polynomial) Clone() Polynomial {
	_p := make(Polynomial, len(*p))
	copy(_p, *p)
	return _p
}

// Set to another polynomial
func (p *Polynomial) Set(p1 Polynomial) {
	if len(*p) != len(p1) {
		*p = p1.Clone()
		return
	}

	for i := 0; i < len(p1); i++ {
		(*p)[i].Set(&p1[i])
	}
}

// AddConstantInPlace adds a constant to the polynomial, modifying p
func (p *Polynomial) AddConstantInPlace(c *fr.Element) {
	for i := 0; i < len(*p); i++ {
		(*p)[i].Add(&(*p)[i], c)
	}
}

// SubConstantInPlace subs a constant to the polynomial, modifying p
func (p *Polynomial) SubConstantInPlace(c *fr.Element) {
	for i := 0; i < len(*p); i++ {
		(*p)[i].Sub(&(*p)[i], c)
	}
}

// ScaleInPlace multiplies p by v, modifying p
func (p *Polynomial) ScaleInPlace(c *fr.Element) {
	for i := 0; i < len(*p); i++ {
		(*p)[i].Mul(&(*p)[i], c)
	}
}

// Scale multiplies p0 by v, storing the result in p
func (p *Polynomial) Scale(c *fr.Element, p0 Polynomial) {
	if len(*p) != len(p0) {
		*p = make(Polynomial, len(p0))
	}
	for i := 0; i < len(p0); i++ {
		(*p)[i].Mul(c, &p0[i])
	}
}

// Add adds p1 to p2
// This function allocates a new slice unless p == p1 or p == p2
func (p *Polynomial) Add(p1, p2 Polynomial) *Polynomial {

	bigger := p1
	smaller := p2
	if len(bigger) < len(smaller) {
		bigger, smaller = smaller, bigger
	}

	if len(*p) == len(bigger) && (&(*p)[0] == &bigger[0]) {
		for i := 0; i < len(smaller); i++ {
			(*p)[i].Add(&(*p)[i], &smaller[i])
		}
		return p
	}

	if len(*p) == len(smaller) && (&(*p)[0] == &smaller[0]) {
		for i := 0; i < len(smaller); i++ {
			(*p)[i].Add(&(*p)[i], &bigger[i])
		}
		*p = append(*p, bigger[len(smaller):]...)
		return p
	}

	res := make(Polynomial, len(bigger))
	copy(res, bigger)
	for i := 0; i < len(smaller); i++ {
		res[i].Add(&res[i], &smaller[i])
	}
	*p = res
	return p
}

// Sub subtracts p2 from p1
// TODO make interface more consistent with Add
func (p *Polynomial) Sub(p1, p2 Polynomial) *Polynomial {
	if len(p1) != len(p2) || len(p2) != len(*p) {
		return nil
	}
	for i := 0; i < len(*p); i++ {
		(*p)[i].Sub(&p1[i], &p2[i])
	}
	return p
}

// Equal checks equality between two polynomials
func (p *Polynomial) Equal(p1 Polynomial) bool {
	if (*p == nil) != (p1 == nil) {
		return false
	}

	if len(*p) != len(p1) {
		return false
	}

	for i := range p1 {
		if !(*p)[i].Equal(&p1[i]) {
			return false
		}
	}

	return true
}

func (p Polynomial) SetZero() {
	for i := 0; i < len(p); i++ {
		p[i].SetZero()
	}
}

func (p Polynomial) Text(base int) string {

	var builder strings.Builder

	first := true
	for d := len(p) - 1; d >= 0; d-- {
		if p[d].IsZero() {
			continue
		}

		pD := p[d]
		pDText := pD.Text(base)

		initialLen := builder.Len()

		if pDText[0] == '-' {
			pDText = pDText[1:]
			if first {
				builder.WriteString("-")
			} else {
				builder.WriteString(" - ")
			}
		} else if !first {
			builder.WriteString(" + ")
		}

		first = false

		if !pD.IsOne() || d == 0 {
			builder.WriteString(pDText)
		}

		if builder.Len()-initialLen > 10 {
			builder.WriteString("×")
		}

		if d != 0 {
			builder.WriteString("X")
		}
		if d > 1 {
			builder.WriteString(
				utils.ToSuperscript(strconv.Itoa(d)),
			)
		}

	}

	if first {
		return "0"
	}

	return builder.String()
}

// InterpolateOnRange maps vector v to polynomial f
// such that f(i) = v[i] for 0 ≤ i < len(v).
// len(f) = len(v) and deg(f) ≤ len(v) - 1
func InterpolateOnRange(v []fr.Element) Polynomial {
	nEvals := uint8(len(v))
	if int(nEvals) != len(v) {
		panic("interpolation method too inefficient for nEvals > 255")
	}
	lagrange := getLagrangeBasis(nEvals)

	var res Polynomial
	res.Scale(&v[0], lagrange[0])

	temp := make(Polynomial, nEvals)

	for i := uint8(1); i < nEvals; i++ {
		temp.Scale(&v[i], lagrange[i])
		res.Add(res, temp)
	}

	return res
}

// lagrange bases used by InterpolateOnRange
var lagrangeBasis sync.Map

func getLagrangeBasis(domainSize uint8) []Polynomial {
	if res, ok := lagrangeBasis.Load(domainSize); ok {
		return res.([]Polynomial)
	}

	// not found. compute
	var res []Polynomial
	if domainSize >= 2 {
		res = computeLagrangeBasis(domainSize)
	} else if domainSize == 1 {
		res = []Polynomial{make(Polynomial, 1)}
		res[0][0].SetOne()
	}
	lagrangeBasis.Store(domainSize, res)

	return res
}

// computeLagrangeBasis precomputes in explicit coefficient form for each 0 ≤ l < domainSize the polynomial
// pₗ := X (X-1) ... (X-l-1) (X-l+1) ... (X - domainSize + 1) / ( l (l-1) ... 2 (-1) ... (l - domainSize +1) )
// Note that pₗ(l) = 1 and pₗ(n) = 0 if 0 ≤ l < domainSize, n ≠ l
func computeLagrangeBasis(domainSize uint8) []Polynomial {

	constTerms := make([]fr.Element, domainSize)
	for i := uint8(0); i < domainSize; i++ {
		constTerms[i].SetInt64(-int64(i))
	}

	res := make([]Polynomial, domainSize)
	multScratch := make(Polynomial, domainSize-1)

	// compute pₗ
	for l := uint8(0); l < domainSize; l++ {

		// TODO @Tabaie Optimize this with some trees? O(log(domainSize)) polynomial mults instead of O(domainSize)? Then again it would be fewer big poly mults vs many small poly mults
		d := uint8(0) //d is the current degree of res
		for i := uint8(0); i < domainSize; i++ {
			if i == l {
				continue
			}
			if d == 0 {
				res[l] = make(Polynomial, domainSize)
				res[l][domainSize-2] = constTerms[i]
				res[l][domainSize-1].SetOne()
			} else {
				current := res[l][domainSize-d-2:]
				timesConst := multScratch[domainSize-d-2:]

				timesConst.Scale(&constTerms[i], current[1:]) //TODO: Directly double and add since constTerms are tiny? (even less than 4 bits)
				nonLeading := current[0 : d+1]

				nonLeading.Add(nonLeading, timesConst)

			}
			d++
		}

	}

	// We have pₗ(i≠l)=0. Now scale so that pₗ(l)=1
	// Replace the constTerms with norms
	for l := uint8(0); l < domainSize; l++ {
		constTerms[l].Neg(&constTerms[l])
		constTerms[l] = res[l].Eval(&constTerms[l])
	}
	constTerms = fr.BatchInvert(constTerms)
	for l := uint8(0); l < domainSize; l++ {
		res[l].ScaleInPlace(&constTerms[l])
	}

	return res
}

// Interpolate fits a polynomial of degree len(X) - 1 = len(Y) - 1 to the points (X[i], Y[i])
// Note that the runtime is O(len(X)³)
func Interpolate(X, Y []fr.Element) (Polynomial, error) {
	if len(X) != len(Y) {
		return nil, errors.New("X and Y must have the same length")
	}

	// solve the system of equations by Gaussian elimination
	augmentedRows := make([][]fr.Element, len(X)) // the last column is the Y values
	for i := range augmentedRows {
		augmentedRows[i] = make([]fr.Element, len(X)+1)
		augmentedRows[i][0].SetOne()
		augmentedRows[i][1].Set(&X[i])
		for j := 2; j < len(augmentedRows[i])-1; j++ {
			augmentedRows[i][j].Mul(&augmentedRows[i][j-1], &X[i])
		}
		augmentedRows[i][len(augmentedRows[i])-1].Set(&Y[i])
	}

	// make the upper triangle
	for i := range len(augmentedRows) - 1 {
		// use row i to eliminate the ith element in all rows below
		var negInv fr.Element
		if augmentedRows[i][i].IsZero() {
			return nil, errors.New("singular matrix")
		}
		negInv.Inverse(&augmentedRows[i][i])
		negInv.Neg(&negInv)
		for j := i + 1; j < len(augmentedRows); j++ {
			var c fr.Element
			c.Mul(&augmentedRows[j][i], &negInv)
			// augmentedRows[j][i].SetZero() omitted
			for k := i + 1; k < len(augmentedRows[i]); k++ {
				var t fr.Element
				t.Mul(&augmentedRows[i][k], &c)
				augmentedRows[j][k].Add(&augmentedRows[j][k], &t)
			}
		}
	}

	// back substitution
	res := make(Polynomial, len(X))
	for i := len(augmentedRows) - 1; i >= 0; i-- {
		res[i] = augmentedRows[i][len(augmentedRows[i])-1]
		for j := i + 1; j < len(augmentedRows[i])-1; j++ {
			var t fr.Element
			t.Mul(&res[j], &augmentedRows[i][j])
			res[i].Sub(&res[i], &t)
		}
		res[i].Div(&res[i], &augmentedRows[i][i])
	}

	return res, nil
}
