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
// The conjugate of an Eisenstein integer x₀ + x₁ω is defined as:
// (x₀ - x₁) - x₁ω
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
//
// We use Karatsuba multiplication to compute the product efficiently.
func (z *ComplexNumber) Mul(x, y *ComplexNumber) *ComplexNumber {
	z.t0.Mul(&x.A0, &y.A0) // t0 = x₀y₀
	z.t1.Mul(&x.A1, &y.A1) // t1 = x₁y₁
	z.t2.Add(&x.A0, &x.A1) // t2 = x₀ + x₁
	z.t3.Add(&y.A0, &y.A1) // t3 = y₀ + y₁
	z.t2.Mul(&z.t2, &z.t3) // t2 = (x₀ + x₁)(y₀ + y₁)

	z.A0.Sub(&z.t0, &z.t1) // A0 = x₀y₀ - x₁y₁
	z.t3.Add(&z.t1, &z.t1)
	z.t3.Add(&z.t3, &z.t0)

	z.A1.Sub(&z.t2, &z.t3) // A1 = (x₀ + x₁)(y₀ + y₁) - x₀y₀ - x₁y₁

	return z
}

// MulByConjugate sets z to the product of x and the conjugate of y
//
//	x * ȳ = (x₀ + x₁ω)((y₀ - y₁) - y₁ω) = (x₀(y₀-y₁) + x₁y₁) + (-x₀y₁ + x₁(y₀-y₁) + x₁y₁)ω
//								        = (x₀y₀ + x₁y₁ - x₀y₁) + (x₁y₀ - x₀y₁)ω
func (z *ComplexNumber) MulByConjugate(x, y *ComplexNumber) *ComplexNumber {
	z.t0.Mul(&x.A1, &y.A0) // t0 = x₁y₀
	z.t1.Mul(&x.A0, &y.A1) // t1 = x₀y₁
	z.t2.Add(&x.A0, &x.A1) // t2 = x₀ + x₁
	z.t3.Add(&y.A0, &y.A1) // t3 = y₀ + y₁
	z.t2.Mul(&z.t2, &z.t3) // t2 = (x₀ + x₁)(y₀ + y₁) = x₀y₀ + x₁y₁ + x₀y₁ + x₁y₀

	z.t3.Add(&z.t1, &z.t1)
	z.t3.Add(&z.t3, &z.t0)
	z.A0.Sub(&z.t2, &z.t3) // A0 = x₀y₀ + x₁y₁ - x₀y₁ = t₂ - t₀ - 2t₁

	z.A1.Sub(&z.t0, &z.t1) // A1 = x₁y₀ - x₀y₁ = t₀ - t₁

	return z
}

// Norm returns the norm of z.
//
// The explicit formula is:
//
//	N(x0+x1ω) = x₀² + x₁² - x₀x₁
//
// We rearrange into it (x₀-x₁)² + x₀x₁
func (z *ComplexNumber) Norm(norm *big.Int) *big.Int {
	z.t1.Sub(&z.A0, &z.A1).Mul(&z.t1, &z.t1)
	z.t2.Mul(&z.A0, &z.A1)
	norm.Add(&z.t1, &z.t2)
	return norm
}

func (z *ComplexNumber) roundNearest(num *ComplexNumber, d *big.Int) {
	z.t1.Abs(d)
	dBitLen := z.t1.BitLen()

	// Helper function for rounding one component
	roundComp := func(result, comp *big.Int) {
		isNegativeResult := (comp.Sign() < 0) != (d.Sign() < 0)
		z.t0.Abs(comp)

		// Bit length shortcut before full comparison
		t0BitLen := z.t0.BitLen()
		if t0BitLen < dBitLen || (t0BitLen == dBitLen && z.t0.Cmp(&z.t1) < 0) {
			// |a| < |b|
			z.t2.Lsh(&z.t0, 1) // t2 = 2 * |a|
			if z.t2.BitLen() > dBitLen || (z.t2.BitLen() == dBitLen && z.t2.Cmp(&z.t1) >= 0) {
				if isNegativeResult {
					result.SetInt64(-1)
				} else {
					result.SetInt64(1)
				}
			} else {
				result.SetInt64(0)
			}
		} else {
			// division and rounding
			z.t2.Set(&z.t0) // remainder = |a|
			k := t0BitLen - dBitLen
			z.t3.Lsh(&z.t1, uint(k))
			if z.t3.Cmp(&z.t0) > 0 {
				k--
			}
			result.SetInt64(0)
			for i := k; i >= 0; i-- {
				z.t3.Lsh(&z.t1, uint(i))
				if z.t2.Cmp(&z.t3) >= 0 {
					z.t2.Sub(&z.t2, &z.t3)
					result.SetBit(result, i, 1)
				}
			}
			z.t3.Lsh(&z.t2, 1)
			if z.t3.Cmp(&z.t1) >= 0 {
				z.t4.SetUint64(1)
				result.Add(result, &z.t4)
			}
			if isNegativeResult {
				result.Neg(result)
			}
		}
	}

	// Round both components
	roundComp(&z.A0, &num.A0)
	roundComp(&z.A1, &num.A1)
}

// Quo sets z to the Euclidean quotient of x / y
// and guarantees ‖r‖ < ‖y‖ (true Euclidean division in ℤ[ω]).
func (z *ComplexNumber) Quo(x, y *ComplexNumber) *ComplexNumber {

	// y.t0 = Norm(y)
	y.Norm(&y.t0)
	if y.t0.Sign() == 0 {
		panic("division by zero")
	}

	// z = x * ȳ
	z.MulByConjugate(x, y)

	// rounding of both coordinates
	z.roundNearest(z, &y.t0)

	return z
}

// QuoRem sets z to the Euclidean quotient of x / y, r to the remainder,
// and guarantees ‖r‖ < ‖y‖ (true Euclidean division in ℤ[ω]).
func (z *ComplexNumber) QuoRem(x, y, r *ComplexNumber) (*ComplexNumber, *ComplexNumber) {

	y.Norm(&y.t0) // > 0  (Eisenstein norm is always non-neg)
	if y.t0.Sign() == 0 {
		panic("division by zero")
	}

	// z = x * ȳ
	z.MulByConjugate(x, y)

	// first guess by *symmetric* rounding of both coordinates
	z.roundNearest(z, &y.t0)

	// r = x – q*y
	r.Mul(y, z)
	r.Sub(x, r)

	// If Euclidean inequality already holds we're done.
	if r.Norm(&x.t1).Cmp(&y.t0) < 0 {
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

	bestNorm.Set(&x.t1) // bestNorm = N(r)

	// six axial directions of the hexagonal lattice
	// {1, 0}, {0, 1}, {-1, 1}, {-1, 0}, {0, -1}, {1, -1}
	var candR ComplexNumber
	for _, dir := range neighbours {
		z.A0.Add(a0, dir[0])
		z.A1.Add(a1, dir[1])

		candR.Mul(y, z)
		candR.Sub(x, &candR)

		if candR.Norm(&x.t1).Cmp(bestNorm) < 0 {
			bestQ0.Set(&z.A0)
			bestQ1.Set(&z.A1)
			r.Set(&candR)
			bestNorm.Set(&x.t1)
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

	aRun.Set(a)
	bRun.Set(b)
	u.SetOne()
	v.SetZero()
	u_.SetZero()
	v_.SetOne()

	// Eisenstein integers form an Euclidean domain for the norm = a.t0
	a.t0.Sqrt(a.Norm(&a.t1))
	for bRun.Norm(&a.t1).Cmp(&a.t0) >= 0 {
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
