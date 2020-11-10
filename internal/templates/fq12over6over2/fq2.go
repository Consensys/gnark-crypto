package fq12over6over2

// Fq2FallBack ...
const Fq2FallBack = `

func addE2(z, x, y *e2) {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
}

func subE2(z, x, y *e2) {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
}

func doubleE2(z, x *e2) {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
}

func negE2(z, x *e2) {
	z.A0.Neg(&x.A0)
	z.A1.Neg(&x.A1)
}

func squareAdxE2(z, x *e2) {
	panic("not implemented")
}

func mulAdxE2(z, x, y *e2) {
	panic("not implemented")
}

`

// Fq2Common ...
const Fq2Common = `

import (
	"math/big"
	"github.com/consensys/gurvy/{{toLower .CurveName}}/fp"
)


// e2 is a degree two finite field extension of fp.Element
type e2 struct {
	A0, A1 fp.Element
}

// Equal returns true if z equals x, fasle otherwise
func (z *e2) Equal(x *e2) bool {
	return z.A0.Equal(&x.A0) && z.A1.Equal(&x.A1)
}

// Cmp compares (lexicographic order) z and x and returns:
//
//   -1 if z <  x
//    0 if z == x
//   +1 if z >  x
//
func (z *e2) Cmp(x *e2) int {
	if a1 := z.A1.Cmp(&x.A1); a1 != 0 {
		return a1
	} 
	return z.A0.Cmp(&x.A0)
}

// LexicographicallyLargest returns true if this element is strictly lexicographically
// larger than its negation, false otherwise
func (z *e2) LexicographicallyLargest() bool {
	// adapted from github.com/zkcrypto/bls12_381
	if z.A1.IsZero() {
		return z.A0.LexicographicallyLargest()
	}
	return z.A1.LexicographicallyLargest()
}

// SetString sets a e2 element from strings
func (z *e2) SetString(s1, s2 string) *e2 {
	z.A0.SetString(s1)
	z.A1.SetString(s2)
	return z
}

// SetZero sets an e2 elmt to zero
func (z *e2) SetZero() *e2 {
	z.A0.SetZero()
	z.A1.SetZero()
	return z
}

// Set sets an e2 from x
func (z *e2) Set(x *e2) *e2 {
	z.A0 = x.A0
	z.A1 = x.A1
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *e2) SetOne() *e2 {
	z.A0.SetOne()
	z.A1.SetZero()
	return z
}

// SetRandom sets a0 and a1 to random values
func (z *e2) SetRandom() *e2 {
	z.A0.SetRandom()
	z.A1.SetRandom()
	return z
}

// IsZero returns true if the two elements are equal, fasle otherwise
func (z *e2) IsZero() bool {
	return z.A0.IsZero() && z.A1.IsZero()
}

// Add adds two elements of e2
func (z *e2) Add(x, y *e2) *e2 {
	addE2(z, x, y)
	return z
}

// Sub two elements of e2
func (z *e2) Sub(x, y *e2) *e2 {
	subE2(z, x, y)
	return z
}


// Double doubles an e2 element
func (z *e2) Double(x *e2) *e2 {
	doubleE2(z, x)
	return z
}


// Neg negates an e2 element
func (z *e2) Neg(x *e2) *e2 {
	negE2(z, x)
	return z
}

// String implements Stringer interface for fancy printing
func (z *e2) String() string {
	return (z.A0.String() + "+" + z.A1.String() + "*u")
}

// ToMont converts to mont form
func (z *e2) ToMont() *e2 {
	z.A0.ToMont()
	z.A1.ToMont()
	return z
}

// FromMont converts from mont form
func (z *e2) FromMont() *e2 {
	z.A0.FromMont()
	z.A1.FromMont()
	return z
}

// MulByElement multiplies an element in e2 by an element in fp
func (z *e2) MulByElement(x *e2, y *fp.Element) *e2 {
	var yCopy fp.Element
	yCopy.Set(y)
	z.A0.Mul(&x.A0, &yCopy)
	z.A1.Mul(&x.A1, &yCopy)
	return z
}

// Conjugate conjugates an element in e2
func (z *e2) Conjugate(x *e2) *e2 {
	z.A0 = x.A0
	z.A1.Neg(&x.A1)
	return z
}

// Legendre returns the Legendre symbol of z
func (z *e2) Legendre() int {
	var n fp.Element
	z.norm(&n)
	return n.Legendre()
}

// Exp sets z=x**e and returns it
func (z *e2) Exp(x e2, exponent *big.Int) *e2 {
	z.SetOne()
    b := exponent.Bytes()
    for i :=0;i<len(b); i++ {
		w := b[i]
		for j := 0; j < 8; j++ {
			z.Square(z)
			if (w&(0b10000000 >> j)) != 0 {
				z.Mul(z, &x)
			}
		}
	}
	
	return z
}

{{if eq .PMod4 3 }}
	func init() {
		q := fp.Modulus()
		tmp := big.NewInt(3)
		sqrtExp1.Set(q).Sub(&sqrtExp1, tmp).Rsh(&sqrtExp1, 2)

		tmp.SetUint64(1)
		sqrtExp2.Set(q).Sub(&sqrtExp2, tmp).Rsh(&sqrtExp2, 1)
	}

	var sqrtExp1, sqrtExp2 big.Int

	// Sqrt sets z to the square root of and returns z
	// The function does not test wether the square root
	// exists or not, it's up to the caller to call
	// Legendre beforehand.
	// cf https://eprint.iacr.org/2012/685.pdf (algo 9)
	func (z *e2) Sqrt(x *e2) *e2 {

		var a1, alpha, b, x0, minusone e2

		minusone.SetOne().Neg(&minusone)
		
		a1.Exp(*x, &sqrtExp1)
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
		
		b.Exp(b, &sqrtExp2).Mul(&x0, &b)
		z.Set(&b)
		return z
	}
{{else }}
	// Sqrt sets z to the square root of and returns z
	// The function does not test wether the square root
	// exists or not, it's up to the caller to call
	// Legendre beforehand.
	// cf https://eprint.iacr.org/2012/685.pdf (algo 10)
	func (z *e2) Sqrt(x *e2) *e2 {

		// precomputation
		var b, c, d, e, f, x0 e2
		var _b, o fp.Element
		c.SetOne()
		for c.Legendre() == 1 {
			c.SetRandom()
		}
		q := fp.Modulus()
		var exp, one big.Int
		one.SetUint64(1)
		exp.Set(q).Sub(&exp, &one).Rsh(&exp, 1)
		d.Exp(c, &exp)
		e.Mul(&d, &c).Inverse(&e)
		f.Mul(&d, &c).Square(&f)

		// computation
		exp.Rsh(&exp, 1)
		b.Exp(*x, &exp)
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

// Fq2Amd64 ...
const Fq2Amd64 = `

import "golang.org/x/sys/cpu"

// supportAdx will be set only on amd64 that has MULX and ADDX instructions
var (
	supportAdx = cpu.X86.HasADX && cpu.X86.HasBMI2
	_          = supportAdx // used in asm
)

// q (modulus)
var qe2 = [{{.NbWords}}]uint64{
	{{- range $i := .NbWordsIndexesFull}}
	{{index $.Q $i}},{{end}}
}

// q'[0], see montgommery multiplication algorithm
var (
	qe2Inv0 uint64 = {{index $.QInverse 0}}
	_ = qe2Inv0 // used in asm
)


//go:noescape
func addE2(res,x,y *e2)

//go:noescape
func subE2(res,x,y *e2)

//go:noescape
func doubleE2(res,x *e2)

//go:noescape
func negE2(res,x *e2)

{{if eq .CurveName "bn256"}}

//go:noescape
func mulNonResE2(res, x *e2)

//go:noescape
func squareAdxE2(res, x *e2)

//go:noescape
func mulAdxE2(res, x, y *e2)

// MulByNonResidue multiplies a e2 by (9,1)
func (z *e2) MulByNonResidue(x *e2) *e2 {
	mulNonResE2(z, x)
	return z
}

// Mul sets z to the e2-product of x,y, returns z
func (z *e2) Mul(x, y *e2) *e2 {
	mulAdxE2(z, x, y)
	return z
}

// Square sets z to the e2-product of x,x, returns z
func (z *e2) Square(x *e2) *e2 {
	squareAdxE2(z, x)
	return z
}


{{end}}

`
