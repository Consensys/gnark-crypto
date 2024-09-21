// The Eisenstein integers form a commutative ring of algebraic integers in the
// algebraic number field Q(ω) – the third cyclotomic field.  These are of the
// form z = a + bω, where a and b are integers and ω is a primitive third root
// of unity i.e. ω²+ω+1 = 0.
package ecc

import (
	"math/big"
)

// A ComplexNumber represents an arbitrary-precision Eisenstein integer.
type ComplexNumber struct {
	A0, A1 big.Int
}

// String implements Stringer interface for fancy printing
func (z *ComplexNumber) String() string {
	return z.A0.String() + "+(" + z.A1.String() + "*ω)"
}

// Equal returns true if z equals x, false otherwise
func (z *ComplexNumber) Equal(x *ComplexNumber) bool {
	return z.A0.Cmp(&x.A0) == 0 && z.A1.Cmp(&x.A1) == 0
}

// Set sets z to x, and returns z.
func (z *ComplexNumber) Set(x *ComplexNumber) *ComplexNumber {
	z.A0.Set(&x.A0)
	z.A1.Set(&x.A1)
	return z
}

// Set sets z to 0, and returns z.
func (z *ComplexNumber) SetZero() *ComplexNumber {
	z.A0 = *big.NewInt(0)
	z.A1 = *big.NewInt(0)
	return z
}

// Set sets z to 1, and returns z.
func (z *ComplexNumber) SetOne() *ComplexNumber {
	z.A0 = *big.NewInt(1)
	z.A1 = *big.NewInt(0)
	return z
}

// Neg sets z to the negative of x, and returns z.
func (z *ComplexNumber) Neg(x *ComplexNumber) *ComplexNumber {
	z.A0.Neg(&x.A0)
	z.A1.Neg(&x.A1)
	return z
}

// Conjugate sets z to the conjugate of x, and returns z.
func (z *ComplexNumber) Conjugate(x *ComplexNumber) *ComplexNumber {
	z.A0.Sub(&x.A0, &x.A1)
	z.A1.Neg(&x.A1)
	return z
}

// Add sets z to the sum of x and y, and returns z.
func (z *ComplexNumber) Add(x, y *ComplexNumber) *ComplexNumber {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	return z
}

// Sub sets z to the difference of x and y, and returns z.
func (z *ComplexNumber) Sub(x, y *ComplexNumber) *ComplexNumber {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
	return z
}

// Mul sets z to the product of x and y, and returns z.
//
// Given that ω²+ω+1=0, the explicit formula is:
//
//	(x0+x1ω)(y0+y1ω) = (x0y0-x1y1) + (x0y1+x1y0-x1y1)ω
func (z *ComplexNumber) Mul(x, y *ComplexNumber) *ComplexNumber {
	var t [3]big.Int
	var z0, z1 big.Int
	t[0].Mul(&x.A0, &y.A0)
	t[1].Mul(&x.A1, &y.A1)
	z0.Sub(&t[0], &t[1])
	t[0].Mul(&x.A0, &y.A1)
	t[2].Mul(&x.A1, &y.A0)
	t[0].Add(&t[0], &t[2])
	z1.Sub(&t[0], &t[1])
	z.A0.Set(&z0)
	z.A1.Set(&z1)
	return z
}

// Norm returns the norm of z.
//
// The explicit formula is:
//
//	N(x0+x1ω) = x0² + x1² - x0*x1
func (z *ComplexNumber) Norm() *big.Int {
	norm := new(big.Int)
	temp := new(big.Int)
	norm.Add(
		norm.Mul(&z.A0, &z.A0),
		temp.Mul(&z.A1, &z.A1),
	)
	norm.Sub(
		norm,
		temp.Mul(&z.A0, &z.A1),
	)
	return norm
}

// Quo sets z to the quotient of x and y, and returns z.
func (z *ComplexNumber) Quo(x, y *ComplexNumber) *ComplexNumber {
	norm := y.Norm()
	z.Conjugate(y)
	z.Mul(x, z)
	z.A0.Div(&z.A0, norm)
	z.A1.Div(&z.A1, norm)
	return z
}

// HalfGCD returns the rational reconstruction of l mod r.
// This outputs u, v s.t. l = u/v mod r
func HalfGCD(a, b *ComplexNumber) [3]*ComplexNumber {

	var aRun, bRun, u, v, u_, v_, quotient, remainder, t, t1, t2 ComplexNumber

	aRun.Set(a)
	bRun.Set(b)
	u.SetOne()
	v.SetZero()
	u_.SetZero()
	v_.SetOne()
	nbits := a.Norm().BitLen() / 2

	for aRun.Norm().BitLen() > nbits {
		quotient.Quo(&aRun, &bRun)
		t.Mul(&bRun, &quotient)
		remainder.Sub(&aRun, &t)
		t.Mul(&u_, &quotient)
		t1.Sub(&u, &t)
		t.Mul(&v_, &quotient)
		t2.Sub(&v, &t)
		aRun.Set(&bRun)
		u.Set(&u_)
		v.Set(&v_)
		bRun.Set(&remainder)
		u_.Set(&t1)
		v_.Set(&t2)
	}

	return [3]*ComplexNumber{&aRun, &v, &u}
}
