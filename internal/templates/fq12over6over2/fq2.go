package fq12over6over2

const Fq2FallBack = `

func addE2(z, x, y *E2) {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
}

func subE2(z, x, y *E2) {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
}

func doubleE2(z, x *E2) {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
}

func negE2(z, x *E2) {
	z.A0.Neg(&x.A0)
	z.A1.Neg(&x.A1)
}

func squareAdxE2(z, x *E2) {
	panic("not implemented")
}

func mulAdxE2(z, x, y *E2) {
	panic("not implemented")
}

`

const Fq2Common = `

import (
	"math/big"
	"github.com/consensys/gurvy/{{toLower .CurveName}}/fp"
)

// E2 is a degree two finite field extension of fp.Element
type E2 struct {
	A0, A1 fp.Element
}

// Equal returns true if z equals x, fasle otherwise
func (z *E2) Equal(x *E2) bool {
	return z.A0.Equal(&x.A0) && z.A1.Equal(&x.A1)
}

// SetString sets a E2 element from strings
func (z *E2) SetString(s1, s2 string) *E2 {
	z.A0.SetString(s1)
	z.A1.SetString(s2)
	return z
}

// SetZero sets an e2 elmt to zero
func (z *E2) SetZero() *E2 {
	z.A0.SetZero()
	z.A1.SetZero()
	return z
}

// Set sets an E2 from x
func (z *E2) Set(x *E2) *E2 {
	z.A0 = x.A0
	z.A1 = x.A1
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E2) SetOne() *E2 {
	z.A0.SetOne()
	z.A1.SetZero()
	return z
}

// SetRandom sets a0 and a1 to random values
func (z *E2) SetRandom() *E2 {
	z.A0.SetRandom()
	z.A1.SetRandom()
	return z
}

// IsZero returns true if the two elements are equal, fasle otherwise
func (z *E2) IsZero() bool {
	return z.A0.IsZero() && z.A1.IsZero()
}

// Add adds two elements of E2
func (z *E2) Add(x, y *E2) *E2 {
	addE2(z, x, y)
	return z
}

// Sub two elements of E2
func (z *E2) Sub(x, y *E2) *E2 {
	subE2(z, x, y)
	return z
}


// Double doubles an E2 element
func (z *E2) Double(x *E2) *E2 {
	doubleE2(z, x)
	return z
}


// Neg negates an E2 element
func (z *E2) Neg(x *E2) *E2 {
	negE2(z, x)
	return z
}

// String implements Stringer interface for fancy printing
func (z *E2) String() string {
	return (z.A0.String() + "+" + z.A1.String() + "*u")
}

// ToMont converts to mont form
func (z *E2) ToMont() *E2 {
	z.A0.ToMont()
	z.A1.ToMont()
	return z
}

// FromMont converts from mont form
func (z *E2) FromMont() *E2 {
	z.A0.FromMont()
	z.A1.FromMont()
	return z
}

// MulByElement multiplies an element in E2 by an element in fp
func (z *E2) MulByElement(x *E2, y *fp.Element) *E2 {
	var yCopy fp.Element
	yCopy.Set(y)
	z.A0.Mul(&x.A0, &yCopy)
	z.A1.Mul(&x.A1, &yCopy)
	return z
}

// Conjugate conjugates an element in E2
func (z *E2) Conjugate(x *E2) *E2 {
	z.A0 = x.A0
	z.A1.Neg(&x.A1)
	return z
}

// Legendre returns the Legendre symbol of z
func (z *E2) Legendre() int {
	var n fp.Element
	z.norm(&n)
	return n.Legendre()
}

// Exp sets z=x**e and returns it
func (z *E2) Exp(x *E2, e big.Int) *E2 {
	var res E2
	res.SetOne()
	b := e.Bytes()
	for i := range b {
		w := b[i]
		mask := byte(0x80)
		for j := 7; j >= 0; j-- {
			res.Square(&res)
			if (w&mask)>>j != 0 {
				res.Mul(&res, x)
			}
			mask = mask >> 1
		}
	}
	z.Set(&res)
	return z
}

{{if eq .PMod4 3 }}
	// Sqrt sets z to the square root of and returns z
	// The function does not test wether the square root
	// exists or not, it's up to the caller to call
	// Legendre beforehand.
	// cf https://eprint.iacr.org/2012/685.pdf (algo 9)
	func (z *E2) Sqrt(x *E2) *E2 {

		var a1, alpha, b, x0, minusone E2
		var e big.Int

		minusone.SetOne().Neg(&minusone)

		q := fp.Modulus()
		tmp := big.NewInt(3)
		e.Set(q).Sub(&e, tmp).Rsh(&e, 2)
		a1.Exp(x, e)
		alpha.Square(&a1).
			Mul(&alpha, x)
		x0.Mul(x, &a1)
		if alpha.Equal(&minusone) {
			var c fp.Element
			c.Set(&x0.A0)
			z.A0.Neg(&x0.A1)
			z.A1.Set(&c)
			return z
		}
		a1.SetOne()
		b.Add(&a1, &alpha)
		tmp.SetUint64(1)
		e.Set(q).Sub(&e, tmp).Rsh(&e, 1)
		b.Exp(&b, e).Mul(&x0, &b)
		z.Set(&b)
		return z
	}
{{else }}
	// Sqrt sets z to the square root of and returns z
	// The function does not test wether the square root
	// exists or not, it's up to the caller to call
	// Legendre beforehand.
	// cf https://eprint.iacr.org/2012/685.pdf (algo 10)
	func (z *E2) Sqrt(x *E2) *E2 {

		// precomputation
		var b, c, d, e, f, x0 E2
		var _b, o fp.Element
		c.SetOne()
		for c.Legendre() == 1 {
			c.SetRandom()
		}
		q := fp.Modulus()
		var exp, one big.Int
		one.SetUint64(1)
		exp.Set(q).Sub(&exp, &one).Rsh(&exp, 1)
		d.Exp(&c, exp)
		e.Mul(&d, &c).Inverse(&e)
		f.Mul(&d, &c).Square(&f)

		// computation
		exp.Rsh(&exp, 1)
		b.Exp(x, exp)
		b.norm(&_b)
		o.SetOne()
		if _b.Equal(&o) {
			x0.Square(&b).Mul(&x0, x)
			_b.Set(&x0.A0).Sqrt(&_b)
			z.Conjugate(&b).MulByElement(z, &_b)
			return z
		}
		x0.Square(&b).Mul(&x0, x).Mul(&x0, &f)
		_b.Set(&x0.A0).Sqrt(&_b)
		z.Conjugate(&b).MulByElement(z, &_b).Mul(z, &e)

		return z
	}
{{end}}

`
