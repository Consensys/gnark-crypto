// Copyright 2020 Consensys Software Inc.
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
	"encoding/binary"
	"io"
	"math/big"
	"math/bits"
	"runtime"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
)

// Basis indicates the basis in which a polynomial is represented.
type Basis uint32

const (
	Canonical Basis = 1 << iota
	Lagrange
	LagrangeCoset
)

// Layout indicates if a polynomial has a BitReverse or a Regular layout
type Layout uint32

const (
	Regular Layout = 8 << iota
	BitReverse
)

// Form describes the form of a polynomial.
// TODO should be a regular enum?
type Form struct {
	Basis  Basis
	Layout Layout
}

// enum of the possible Form values for type-safe switches
// in this package
var (
	canonicalRegular        = Form{Canonical, Regular}
	canonicalBitReverse     = Form{Canonical, BitReverse}
	lagrangeRegular         = Form{Lagrange, Regular}
	lagrangeBitReverse      = Form{Lagrange, BitReverse}
	lagrangeCosetRegular    = Form{LagrangeCoset, Regular}
	lagrangeCosetBitReverse = Form{LagrangeCoset, BitReverse}
)

// Polynomial wraps a polynomial so that it is
// interpreted as P'(X)=P(\omega^{s}X).
// Size is the real size of the polynomial (seen as a vector).
// For instance if len(P)=32 but P.Size=8, it means that P has been
// extended (e.g. it is evaluated on a larger set) but P is a polynomial
// of degree 7.
// blindedSize is the size of the polynomial when it is blinded. By
// default blindedSize=Size, until the polynomial is blinded.
type Polynomial struct {
	*polynomial
	shift int
	size  int
}

// NewPolynomial returned a Polynomial from the provided coefficients in the given form.
// A Polynomial can be seen as a "shared pointer" on a list of coefficients.
// It is the responsibility of the user to call the Clone method if the coefficients
// shouldn't be mutated.
func NewPolynomial(coeffs *[]fr.Element, form Form) *Polynomial {
	return &Polynomial{
		polynomial: newPolynomial(coeffs, form),
		size:       len(*coeffs),
	}
}

// Shift the wrapped polynomial; it doesn't modify the underlying data structure,
// but flag the Polynomial such that it will be interpreted as p(\omega^shift X)
func (p *Polynomial) Shift(shift int) *Polynomial {
	p.shift = shift
	return p
}

// Size returns the real size of the polynomial (seen as a vector).
// For instance if len(P)=32 but P.Size=8, it means that P has been
// extended (e.g. it is evaluated on a larger set) but P is a polynomial
// of degree 7.
func (p *Polynomial) Size() int {
	return p.size
}

// SetSize sets the size of the polynomial.
// size is the real size of the polynomial (seen as a vector).
// For instance if len(P)=32 but P.size=8, it means that P has been
// extended (e.g. it is evaluated on a larger set) but P is a polynomial
// of degree 7.
func (p *Polynomial) SetSize(size int) {
	p.size = size
}

// Evaluate evaluates p at x.
// The code panics if the function is not in canonical form.
func (p *Polynomial) Evaluate(x fr.Element) fr.Element {

	if p.shift == 0 {
		return p.polynomial.evaluate(x)
	}

	var g fr.Element
	if p.shift <= 5 {
		gen, err := fft.Generator(uint64(p.size))
		if err != nil {
			panic(err)
		}
		g = smallExp(gen, p.shift)
		x.Mul(&x, &g)
		return p.polynomial.evaluate(x)
	}

	bs := big.NewInt(int64(p.shift))
	g = *g.Exp(g, bs)
	x.Mul(&x, &g)
	return p.polynomial.evaluate(x)
}

// Clone returns a deep copy of p. The underlying polynomial is cloned;
// see also ShallowClone to perform a ShallowClone on the underlying polynomial.
// If capacity is provided, the new coefficient slice capacity will be set accordingly.
func (p *Polynomial) Clone(capacity ...int) *Polynomial {
	res := p.ShallowClone()
	res.polynomial = p.polynomial.clone(capacity...)
	return res
}

// ShallowClone returns a shallow copy of p. The underlying polynomial coefficient
// is NOT cloned and both objects will point to the same coefficient vector.
func (p *Polynomial) ShallowClone() *Polynomial {
	res := *p
	return &res
}

// GetCoeff returns the i-th entry of p, taking the layout in account.
func (p *Polynomial) GetCoeff(i int) fr.Element {

	n := p.coefficients.Len()
	rho := n / p.size
	if p.polynomial.Form.Layout == Regular {
		return (*p.coefficients)[(i+rho*p.shift)%n]
	} else {
		nn := uint64(64 - bits.TrailingZeros(uint(n)))
		iRev := bits.Reverse64(uint64((i+rho*p.shift)%n)) >> nn
		return (*p.coefficients)[iRev]
	}

}

// polynomial represents a polynomial, the vector of coefficients
// along with the basis and the layout.
type polynomial struct {
	coefficients *fr.Vector
	Form
}

// Coefficients returns a slice on the underlying data structure.
func (p *polynomial) Coefficients() []fr.Element {
	return (*p.coefficients)
}

// newPolynomial creates a new polynomial. The slice coeff NOT copied
// but directly assigned to the new polynomial.
func newPolynomial(coeffs *[]fr.Element, form Form) *polynomial {
	return &polynomial{coefficients: (*fr.Vector)(coeffs), Form: form}
}

// clone returns a deep copy of the underlying data structure.
func (p *polynomial) clone(capacity ...int) *polynomial {
	c := p.coefficients.Len()
	if len(capacity) == 1 && capacity[0] > c {
		c = capacity[0]
	}
	newCoeffs := make(fr.Vector, p.coefficients.Len(), c)
	r := &polynomial{
		coefficients: &newCoeffs,
		Form:         p.Form,
	}
	copy((*r.coefficients), (*p.coefficients))
	return r
}

// evaluate evaluates p at x.
// The code panics if the function is not in canonical form.
func (p *polynomial) evaluate(x fr.Element) fr.Element {

	var r fr.Element
	// if p.Basis != Canonical {
	// 	panic("p must be in canonical basis")
	// }

	if p.Basis == Canonical {
		if p.Layout == Regular {
			for i := p.coefficients.Len() - 1; i >= 0; i-- {
				r.Mul(&r, &x).Add(&r, &(*p.coefficients)[i])
			}
		} else {
			nn := uint64(64 - bits.TrailingZeros(uint(p.coefficients.Len())))
			for i := p.coefficients.Len() - 1; i >= 0; i-- {
				iRev := bits.Reverse64(uint64(i)) >> nn
				r.Mul(&r, &x).Add(&r, &(*p.coefficients)[iRev])
			}
		}
	} else if p.Basis == Lagrange {
		sizeP := p.coefficients.Len()
		w, err := fft.Generator(uint64(sizeP))
		if err != nil {
			panic(err)
		}
		var accw fr.Element
		accw.SetOne()
		dens := make([]fr.Element, sizeP) // [x-1, x-ω, x-ω², ...]
		for i := 0; i < sizeP; i++ {
			dens[i].Sub(&x, &accw)
			accw.Mul(&accw, &w)
		}
		invdens := fr.BatchInvert(dens) // [1/(x-1), 1/(x-ω), 1/(x-ω²), ...]
		var tmp fr.Element
		var one fr.Element
		one.SetOne()
		tmp.Exp(x, big.NewInt(int64(sizeP))).Sub(&tmp, &one) // xⁿ-1
		var li fr.Element
		li.SetUint64(uint64(sizeP)).Inverse(&li).Mul(&li, &tmp) // 1/n * (xⁿ-1)
		if p.Layout == Regular {
			for i := 0; i < sizeP; i++ {
				li.Mul(&li, &invdens[i]) // li <- li*ω/(x-ωⁱ)
				tmp.Mul(&li, &(*p.coefficients)[i])
				r.Add(&r, &tmp)
				li.Mul(&li, &dens[i]).Mul(&li, &w) // li <- li*ω*(x-ωⁱ)
			}
		} else {
			nn := uint64(64 - bits.TrailingZeros(uint(p.coefficients.Len())))
			for i := 0; i < sizeP; i++ {
				iRev := bits.Reverse64(uint64(i)) >> nn
				li.Mul(&li, &invdens[i]) // li <- li*ω/(x-ωⁱ)
				tmp.Mul(&li, &(*p.coefficients)[iRev])
				r.Add(&r, &tmp)
				li.Mul(&li, &dens[i]).Mul(&li, &w) // li <- li*ω*(x-ωⁱ)
			}
		}
	} // else if p.Basis == LagrangeCoset {
	// 	if p.Layout==Regular {

	// 	} else {

	// 	}
	// }

	return r

}

// ToRegular changes the layout of p to Regular.
// Leaves p unchanged if p's layout was already Regular.
func (p *Polynomial) ToRegular() *Polynomial {
	if p.Layout == Regular {
		return p
	}
	fft.BitReverse((*p.coefficients))
	p.Layout = Regular
	return p
}

// ToBitReverse changes the layout of p to BitReverse.
// Leaves p unchanged if p's layout was already BitReverse.
func (p *Polynomial) ToBitReverse() *Polynomial {
	if p.Layout == BitReverse {
		return p
	}
	fft.BitReverse((*p.coefficients))
	p.Layout = BitReverse
	return p
}

// ToLagrange converts p to Lagrange form.
// Leaves p unchanged if p was already in Lagrange form.
func (p *Polynomial) ToLagrange(d *fft.Domain, nbTasks ...int) *Polynomial {
	id := p.Form
	p.grow(int(d.Cardinality))

	n := runtime.NumCPU()
	if len(nbTasks) > 0 {
		n = nbTasks[0]
	}

	switch id {
	case canonicalRegular:
		p.Layout = BitReverse
		d.FFT((*p.coefficients), fft.DIF, fft.WithNbTasks(n))
	case canonicalBitReverse:
		p.Layout = Regular
		d.FFT((*p.coefficients), fft.DIT, fft.WithNbTasks(n))
	case lagrangeRegular, lagrangeBitReverse:
		return p
	case lagrangeCosetRegular:
		p.Layout = Regular
		d.FFTInverse((*p.coefficients), fft.DIF, fft.OnCoset(), fft.WithNbTasks(n))
		d.FFT((*p.coefficients), fft.DIT)
	case lagrangeCosetBitReverse:
		p.Layout = BitReverse
		d.FFTInverse((*p.coefficients), fft.DIT, fft.OnCoset(), fft.WithNbTasks(n))
		d.FFT((*p.coefficients), fft.DIF)
	default:
		panic("unknown ID")
	}
	p.Basis = Lagrange
	return p
}

// ToCanonical converts p to canonical form.
// Leaves p unchanged if p was already in Canonical form.
func (p *Polynomial) ToCanonical(d *fft.Domain, nbTasks ...int) *Polynomial {
	id := p.Form
	p.grow(int(d.Cardinality))
	n := runtime.NumCPU()
	if len(nbTasks) > 0 {
		n = nbTasks[0]
	}
	switch id {
	case canonicalRegular, canonicalBitReverse:
		return p
	case lagrangeRegular:
		p.Layout = BitReverse
		d.FFTInverse((*p.coefficients), fft.DIF, fft.WithNbTasks(n))
	case lagrangeBitReverse:
		p.Layout = Regular
		d.FFTInverse((*p.coefficients), fft.DIT, fft.WithNbTasks(n))
	case lagrangeCosetRegular:
		p.Layout = BitReverse
		d.FFTInverse((*p.coefficients), fft.DIF, fft.OnCoset(), fft.WithNbTasks(n))
	case lagrangeCosetBitReverse:
		p.Layout = Regular
		d.FFTInverse((*p.coefficients), fft.DIT, fft.OnCoset(), fft.WithNbTasks(n))
	default:
		panic("unknown ID")
	}
	p.Basis = Canonical
	return p
}

func (p *polynomial) grow(newSize int) {
	offset := newSize - p.coefficients.Len()
	if offset > 0 {
		(*p.coefficients) = append((*p.coefficients), make(fr.Vector, offset)...)
	}
}

// ToLagrangeCoset Sets p to q, in LagrangeCoset form and returns it.
func (p *Polynomial) ToLagrangeCoset(d *fft.Domain) *Polynomial {
	id := p.Form
	p.grow(int(d.Cardinality))
	switch id {
	case canonicalRegular:
		p.Layout = BitReverse
		d.FFT((*p.coefficients), fft.DIF, fft.OnCoset())
	case canonicalBitReverse:
		p.Layout = Regular
		d.FFT((*p.coefficients), fft.DIT, fft.OnCoset())
	case lagrangeRegular:
		p.Layout = Regular
		d.FFTInverse((*p.coefficients), fft.DIF)
		d.FFT((*p.coefficients), fft.DIT, fft.OnCoset())
	case lagrangeBitReverse:
		p.Layout = BitReverse
		d.FFTInverse((*p.coefficients), fft.DIT)
		d.FFT((*p.coefficients), fft.DIF, fft.OnCoset())
	case lagrangeCosetRegular, lagrangeCosetBitReverse:
		return p
	default:
		panic("unknown ID")
	}

	p.Basis = LagrangeCoset
	return p
}

// WriteTo implements io.WriterTo
func (p *Polynomial) WriteTo(w io.Writer) (int64, error) {
	// encode coefficients
	n, err := p.polynomial.coefficients.WriteTo(w)
	if err != nil {
		return n, err
	}

	// encode Form.Basis, Form.Layout, shift, size & blindedSize as uint32
	var data = []uint32{
		uint32(p.Basis),
		uint32(p.Layout),
		uint32(p.shift),
		uint32(p.size),
	}
	for _, v := range data {
		err = binary.Write(w, binary.BigEndian, v)
		if err != nil {
			return n, err
		}
		n += 4
	}
	return n, nil
}

// ReadFrom implements io.ReaderFrom
func (p *Polynomial) ReadFrom(r io.Reader) (int64, error) {
	// decode coefficients
	if p.polynomial == nil {
		p.polynomial = new(polynomial)
	}
	if p.polynomial.coefficients == nil {
		v := make(fr.Vector, 0)
		p.polynomial.coefficients = &v
	}
	n, err := p.polynomial.coefficients.ReadFrom(r)
	if err != nil {
		return n, err
	}

	// decode Form.Basis, Form.Layout, shift as uint32
	var data [4]uint32
	var buf [4]byte
	for i := range data {
		read, err := io.ReadFull(r, buf[:4])
		n += int64(read)
		if err != nil {
			return n, err
		}
		data[i] = binary.BigEndian.Uint32(buf[:4])
	}

	p.Basis = Basis(data[0])
	p.Layout = Layout(data[1])
	p.shift = int(data[2])
	p.size = int(data[3])

	return n, nil
}
