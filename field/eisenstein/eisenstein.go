package eisenstein

import (
	"math/big"
)

// A ComplexNumber represents an arbitrary-precision Eisenstein integer.
type ComplexNumber struct {
	A0, A1 *big.Int
}

// ──────────────────────────────────────────────────────────────────────────────
// helpers – hex-lattice geometry & symmetric rounding
// ──────────────────────────────────────────────────────────────────────────────

// six axial directions of the hexagonal lattice
var neighbours = [][2]int64{
	{1, 0}, {0, 1}, {-1, 1}, {-1, 0}, {0, -1}, {1, -1},
}

// roundNearest returns ⌊(z + d/2) / d⌋  for *any* sign of z, d>0
func roundNearest(z, d *big.Int) *big.Int {
	half := new(big.Int).Rsh(d, 1) // d / 2
	if z.Sign() >= 0 {
		return new(big.Int).Div(new(big.Int).Add(z, half), d)
	}
	tmp := new(big.Int).Neg(z)
	tmp.Add(tmp, half)
	tmp.Div(tmp, d)
	return tmp.Neg(tmp)
}

func (z *ComplexNumber) init() {
	if z.A0 == nil {
		z.A0 = new(big.Int)

	}
	if z.A1 == nil {
		z.A1 = new(big.Int)

	}
}

// String implements Stringer interface for fancy printing
func (z *ComplexNumber) String() string {
	return z.A0.String() + "+(" + z.A1.String() + "*ω)"
}

// Equal returns true if z equals x, false otherwise
func (z *ComplexNumber) Equal(x *ComplexNumber) bool {
	return z.A0.Cmp(x.A0) == 0 && z.A1.Cmp(x.A1) == 0
}

// Set sets z to x, and returns z.
func (z *ComplexNumber) Set(x *ComplexNumber) *ComplexNumber {
	z.init()
	z.A0.Set(x.A0)
	z.A1.Set(x.A1)
	return z
}

// SetZero sets z to 0, and returns z.
func (z *ComplexNumber) SetZero() *ComplexNumber {
	z.A0 = big.NewInt(0)
	z.A1 = big.NewInt(0)
	return z
}

// SetOne sets z to 1, and returns z.
func (z *ComplexNumber) SetOne() *ComplexNumber {
	z.A0 = big.NewInt(1)
	z.A1 = big.NewInt(0)
	return z
}

// Neg sets z to the negative of x, and returns z.
func (z *ComplexNumber) Neg(x *ComplexNumber) *ComplexNumber {
	z.init()
	z.A0.Neg(x.A0)
	z.A1.Neg(x.A1)
	return z
}

// Conjugate sets z to the conjugate of x, and returns z.
func (z *ComplexNumber) Conjugate(x *ComplexNumber) *ComplexNumber {
	z.init()
	z.A0.Sub(x.A0, x.A1)
	z.A1.Neg(x.A1)
	return z
}

// Add sets z to the sum of x and y, and returns z.
func (z *ComplexNumber) Add(x, y *ComplexNumber) *ComplexNumber {
	z.init()
	z.A0.Add(x.A0, y.A0)
	z.A1.Add(x.A1, y.A1)
	return z
}

// Sub sets z to the difference of x and y, and returns z.
func (z *ComplexNumber) Sub(x, y *ComplexNumber) *ComplexNumber {
	z.init()
	z.A0.Sub(x.A0, y.A0)
	z.A1.Sub(x.A1, y.A1)
	return z
}

// Mul sets z to the product of x and y, and returns z.
//
// Given that ω²+ω+1=0, the explicit formula is:
//
//	(x0+x1ω)(y0+y1ω) = (x0y0-x1y1) + (x0y1+x1y0-x1y1)ω
func (z *ComplexNumber) Mul(x, y *ComplexNumber) *ComplexNumber {
	z.init()
	var t [3]big.Int
	var z0, z1 big.Int
	t[0].Mul(x.A0, y.A0)
	t[1].Mul(x.A1, y.A1)
	z0.Sub(&t[0], &t[1])
	t[0].Mul(x.A0, y.A1)
	t[2].Mul(x.A1, y.A0)
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
		norm.Mul(z.A0, z.A0),
		temp.Mul(z.A1, z.A1),
	)
	norm.Sub(
		norm,
		temp.Mul(z.A0, z.A1),
	)
	return norm
}

// QuoRem sets z to the Euclidean quotient of x / y, r to the remainder,
// and guarantees ‖r‖ < ‖y‖ (true Euclidean division in ℤ[ω]).
func (z *ComplexNumber) QuoRem(x, y, r *ComplexNumber) (*ComplexNumber, *ComplexNumber) {

	norm := y.Norm() // > 0  (Eisenstein norm is always non-neg)
	if norm.Sign() == 0 {
		panic("division by zero")
	}

	// num = x * ȳ   (ȳ computed in a fresh variable → y unchanged)
	var yConj, num ComplexNumber
	yConj.Conjugate(y)
	num.Mul(x, &yConj)

	// first guess by *symmetric* rounding of both coordinates
	q0 := roundNearest(num.A0, norm)
	q1 := roundNearest(num.A1, norm)
	z.A0, z.A1 = q0, q1

	// r = x – q*y
	r.Mul(y, z)
	r.Sub(x, r)

	// If Euclidean inequality already holds we're done.
	// Otherwise walk ≤2 unit steps in the hex lattice until N(r) < N(y).
	if r.Norm().Cmp(norm) >= 0 {
		bestQ0, bestQ1 := new(big.Int).Set(z.A0), new(big.Int).Set(z.A1)
		bestR := new(ComplexNumber).Set(r)
		bestN2 := bestR.Norm()

		for _, dir := range neighbours {
			candQ0 := new(big.Int).Add(z.A0, big.NewInt(dir[0]))
			candQ1 := new(big.Int).Add(z.A1, big.NewInt(dir[1]))
			var candQ ComplexNumber
			candQ.A0, candQ.A1 = candQ0, candQ1

			var candR ComplexNumber
			candR.Mul(y, &candQ)
			candR.Sub(x, &candR)

			if candR.Norm().Cmp(bestN2) < 0 {
				bestQ0, bestQ1 = candQ0, candQ1
				bestR.Set(&candR)
				bestN2 = bestR.Norm()
			}
		}
		z.A0, z.A1 = bestQ0, bestQ1
		r.Set(bestR) // update remainder and retry; Euclidean property ⇒ ≤ 2 loops
	}
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
	sqrt.Sqrt(a.Norm())
	for bRun.Norm().Cmp(&sqrt) >= 0 {
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
