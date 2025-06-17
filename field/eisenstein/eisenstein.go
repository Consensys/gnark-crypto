package eisenstein

import (
	"math/big"
	"sync"
)

// A ComplexNumber represents an arbitrary-precision Eisenstein integer.
type ComplexNumber struct {
	A0, A1             big.Int
	t0, t1, t2, t3, t4 big.Int    // temporary variables
	_                  sync.Mutex // to ensure there is no accidental value copy
}

// ──────────────────────────────────────────────────────────────────────────────
// helpers – hex-lattice geometry & symmetric rounding
// ──────────────────────────────────────────────────────────────────────────────

// six axial directions of the hexagonal lattice
var neighbours = [6][2]*big.Int{
	{big.NewInt(1), big.NewInt(0)},
	{big.NewInt(0), big.NewInt(1)},
	{big.NewInt(-1), big.NewInt(1)},
	{big.NewInt(-1), big.NewInt(0)},
	{big.NewInt(0), big.NewInt(-1)},
	{big.NewInt(1), big.NewInt(-1)},
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

// SetZero sets z to 0, and returns z.
func (z *ComplexNumber) SetZero() *ComplexNumber {
	z.A0.SetUint64(0)
	z.A1.SetUint64(0)
	return z
}

// SetOne sets z to 1, and returns z.
func (z *ComplexNumber) SetOne() *ComplexNumber {
	z.A0.SetUint64(1)
	z.A1.SetUint64(0)
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
//	(x₀ + x₁ω)(y₀ + y₁ω) = (x₀y₀ - x₁y₁) + (x₀y₁ + x₁y₀ - x₁y₁)ω
func (z *ComplexNumber) Mul(x, y *ComplexNumber) *ComplexNumber {
	z.t0.Mul(&x.A0, &y.A0) // t0 = x₀y₀
	z.t1.Mul(&x.A1, &y.A1) // t1 = x₁y₁
	z.t3.Sub(&z.t0, &z.t1) // t3 = x₀y₀ - x₁y₁  (real part)
	z.t0.Mul(&x.A0, &y.A1) // t0 = x₀y₁
	z.t2.Mul(&x.A1, &y.A0) // t2 = x₁y₀
	z.t0.Add(&z.t0, &z.t2) // t0 = x₀y₁ + x₁y₀
	z.t4.Sub(&z.t0, &z.t1) // t4 = x₀y₁ + x₁y₀ - x₁y₁  (imaginary part)
	z.A0.Set(&z.t3)        // z.A0 = real part
	z.A1.Set(&z.t4)        // z.A1 = imaginary part
	return z
}

// Norm returns the norm of z.
//
// The explicit formula is:
//
//	N(x0+x1ω) = x0² + x1² - x0*x1
func (z *ComplexNumber) Norm(norm *big.Int) *big.Int {
	norm.Add(
		z.t1.Mul(&z.A0, &z.A0),
		z.t0.Mul(&z.A1, &z.A1),
	)
	norm.Sub(
		norm,
		z.t1.Mul(&z.A0, &z.A1),
	)
	return norm
}

// roundNearest sets z to the coordinate-wise nearest integer division of num/d,
// using symmetric rounding (round half away from zero).
func (z *ComplexNumber) roundNearest(num *ComplexNumber, d *big.Int) {
	half := z.t0.Rsh(d, 1) // half = d / 2

	// Round A0 coordinate
	if num.A0.Sign() >= 0 {
		z.t1.Add(&num.A0, half)
		z.A0.Div(&z.t1, d)
	} else {
		z.t1.Neg(&num.A0)
		z.t1.Add(&z.t1, half)
		z.t1.Div(&z.t1, d)
		z.A0.Neg(&z.t1)
	}

	// Round A1 coordinate
	if num.A1.Sign() >= 0 {
		z.t2.Add(&num.A1, half)
		z.A1.Div(&z.t2, d)
	} else {
		z.t2.Neg(&num.A1)
		z.t2.Add(&z.t2, half)
		z.t2.Div(&z.t2, d)
		z.A1.Neg(&z.t2)
	}
}

// QuoRem sets z to the Euclidean quotient of x / y, r to the remainder,
// and guarantees ‖r‖ < ‖y‖ (true Euclidean division in ℤ[ω]).
func (z *ComplexNumber) QuoRem(x, y, r *ComplexNumber) (*ComplexNumber, *ComplexNumber) {

	norm, rNorm := new(big.Int), new(big.Int)
	y.Norm(norm) // > 0  (Eisenstein norm is always non-neg)
	if norm.Sign() == 0 {
		panic("division by zero")
	}

	// num = x * ȳ   (ȳ computed in a fresh variable → y unchanged)
	var yConj ComplexNumber
	yConj.Conjugate(y)
	yConj.Mul(x, &yConj)

	// first guess by *symmetric* rounding of both coordinates
	z.roundNearest(&yConj, norm)

	// r = x – q*y
	r.Mul(y, z)
	r.Sub(x, r)

	// If Euclidean inequality already holds we're done.
	if r.Norm(rNorm).Cmp(norm) < 0 {
		return z, r
	}

	// Otherwise walk ≤2 unit steps in the hex lattice until N(r) < N(y).
	bestNorm := &z.t0
	bestQ0, bestQ1 := &z.t1, &z.t2
	a0, a1 := &z.t3, &z.t4
	a0.Set(&z.A0)
	a1.Set(&z.A1)
	bestQ0.Set(a0)
	bestQ1.Set(a1)

	bestNorm.Set(rNorm) // bestNorm = N(r)

	// six axial directions of the hexagonal lattice
	// var neighbours = [6][2]int64{
	// 	{1, 0}, {0, 1}, {-1, 1}, {-1, 0}, {0, -1}, {1, -1},
	// }
	var candR ComplexNumber
	for _, dir := range neighbours {
		z.A0.Add(a0, dir[0])
		z.A1.Add(a1, dir[1])

		candR.Mul(y, z)
		candR.Sub(x, &candR)

		if candR.Norm(rNorm).Cmp(bestNorm) < 0 {
			bestQ0.Set(&z.A0)
			bestQ1.Set(&z.A1)
			r.Set(&candR)
			bestNorm.Set(rNorm)
		}
	}
	z.A0.Set(bestQ0)
	z.A1.Set(bestQ1)

	return z, r
}

// HalfGCD returns the rational reconstruction of a, b.
// This outputs w, v, u s.t. w = a*u + b*v.
func HalfGCD(a, b *ComplexNumber) [3]*ComplexNumber {

	var aRun, bRun, u, v, u_, v_, quotient, remainder, t, t1, t2 ComplexNumber
	var sqrt big.Int

	aRun.Set(a)
	bRun.Set(b)
	u.SetOne()
	v.SetZero()
	u_.SetZero()
	v_.SetOne()

	// Eisenstein integers form an Euclidean domain for the norm
	norm := new(big.Int)
	sqrt.Sqrt(a.Norm(norm))
	for bRun.Norm(norm).Cmp(&sqrt) >= 0 {
		quotient.QuoRem(&aRun, &bRun, &remainder)
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

	return [3]*ComplexNumber{&bRun, &v_, &u_}
}
