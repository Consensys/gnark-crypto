package extensions

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// MulBy11 x *= 11 (mod r)
func MulBy11(x *fr.Element) {
	y := fr.Element{
		1855201571499933546,
		8511318076631809892,
		6222514765367795509,
		1122129207579058019,
	}
	x.Mul(x, &y)
}

// Mul sets z to the E2-product of x,y, returns z
func (z *E2) Mul(x, y *E2) *E2 {
	var a, b, c fr.Element
	a.Add(&x.A0, &x.A1)
	b.Add(&y.A0, &y.A1)
	a.Mul(&a, &b)
	b.Mul(&x.A0, &y.A0)
	c.Mul(&x.A1, &y.A1)
	z.A1.Sub(&a, &b).Sub(&z.A1, &c)
	MulBy11(&c)
	z.A0.Sub(&b, &c)
	return z
}

// Square sets z to the E2-product of x,x returns z
func (z *E2) Square(x *E2) *E2 {
	var a, b, c fr.Element
	a.Mul(&x.A0, &x.A1).Double(&a)
	c.Square(&x.A0)
	b.Square(&x.A1)
	MulBy11(&b)
	z.A0.Sub(&c, &b)
	z.A1 = a
	return z
}

// MulByNonResidue multiplies a E2 by (0,1)
func (z *E2) MulByNonResidue(x *E2) *E2 {
	a := x.A0
	b := x.A1 // fetching x.A1 in the function below is slower
	MulBy11(&b)
	z.A0.Neg(&b)
	z.A1 = a
	return z
}

// MulByNonResidueInv multiplies a E2 by (0,1)^{-1}
func (z *E2) MulByNonResidueInv(x *E2) *E2 {
	a := x.A1
	// 1/11 mod r
	elevenInv := fr.Element{
		7989155441247042094,
		18276457113184108543,
		17999817914616464103,
		943187440870955565,
	}
	z.A1.Mul(&x.A0, &elevenInv).Neg(&z.A1)
	z.A0 = a
	return z
}

// Inverse sets z to the E2-inverse of x, returns z
func (z *E2) Inverse(x *E2) *E2 {
	// Algorithm 8 from https://eprint.iacr.org/2010/354.pdf
	var t0, t1, tmp fr.Element
	a := &x.A0 // creating the buffers a, b is faster than querying &x.A0, &x.A1 in the functions call below
	b := &x.A1
	t0.Square(a)
	t1.Square(b)
	tmp.Set(&t1)
	MulBy11(&tmp)
	t0.Add(&t0, &tmp)
	t1.Inverse(&t0)
	z.A0.Mul(a, &t1)
	z.A1.Mul(b, &t1).Neg(&z.A1)

	return z
}

// norm sets x to the norm of z
func (z *E2) norm(x *fr.Element) {
	var tmp fr.Element
	x.Square(&z.A1)
	tmp.Set(x)
	MulBy11(&tmp)
	x.Square(&z.A0).Add(x, &tmp)
}
