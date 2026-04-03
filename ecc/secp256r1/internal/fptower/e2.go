// Package fptower implements Fp2 = Fp[u]/(u²+1) arithmetic for the P-256 (secp256r1) base field.
// The non-residue is β = −1, valid since q ≡ 3 mod 4.
package fptower

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/secp256r1/fp"
)

// E2 is a degree-two extension of fp.Element: A0 + A1·u, u² = −1.
type E2 struct {
	A0, A1 fp.Element
}

// SetZero sets z to 0.
func (z *E2) SetZero() *E2 { z.A0.SetZero(); z.A1.SetZero(); return z }

// SetOne sets z to 1.
func (z *E2) SetOne() *E2 { z.A0.SetOne(); z.A1.SetZero(); return z }

// Set sets z to x.
func (z *E2) Set(x *E2) *E2 { z.A0 = x.A0; z.A1 = x.A1; return z }

// Equal returns true if z == x.
func (z *E2) Equal(x *E2) bool { return z.A0.Equal(&x.A0) && z.A1.Equal(&x.A1) }

// IsZero returns true if z == 0.
func (z *E2) IsZero() bool { return z.A0.IsZero() && z.A1.IsZero() }

// IsOne returns true if z == 1.
func (z *E2) IsOne() bool { return z.A0.IsOne() && z.A1.IsZero() }

// SetRandom sets z to a random element and returns z.
func (z *E2) SetRandom() (*E2, error) {
	if _, err := z.A0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A1.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// MustSetRandom sets z to a random element, panicking on error.
func (z *E2) MustSetRandom() {
	if _, err := z.SetRandom(); err != nil {
		panic(err)
	}
}

// String returns the string representation of z.
func (z *E2) String() string {
	return fmt.Sprintf("%s+(%s)*u", z.A0.String(), z.A1.String())
}

// SetString sets z from two decimal strings and returns z.
func (z *E2) SetString(s1, s2 string) *E2 {
	z.A0.SetString(s1)
	z.A1.SetString(s2)
	return z
}

// Legendre returns the Legendre symbol of z in Fp2.
// Returns 1 if z is a non-zero square, -1 if z is not a square, 0 if z is zero.
func (z *E2) Legendre() int {
	var n fp.Element
	z.Norm(&n)
	return n.Legendre()
}

// Neg sets z = −x.
func (z *E2) Neg(x *E2) *E2 {
	z.A0.Neg(&x.A0)
	z.A1.Neg(&x.A1)
	return z
}

// Conjugate sets z = x̄ = A0 − A1·u.
func (z *E2) Conjugate(x *E2) *E2 {
	z.A0 = x.A0
	z.A1.Neg(&x.A1)
	return z
}

// Add sets z = x + y.
func (z *E2) Add(x, y *E2) *E2 {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	return z
}

// Sub sets z = x − y.
func (z *E2) Sub(x, y *E2) *E2 {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
	return z
}

// Double sets z = 2·x.
func (z *E2) Double(x *E2) *E2 {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
	return z
}

// MulByElement sets z = x·y for y ∈ Fp.
func (z *E2) MulByElement(x *E2, y *fp.Element) *E2 {
	z.A0.Mul(&x.A0, y)
	z.A1.Mul(&x.A1, y)
	return z
}

// Mul sets z = x·y using Karatsuba (3 Fp muls).
// (a+bu)(c+du) = (ac−bd) + (ad+bc)u  with β = −1.
func (z *E2) Mul(x, y *E2) *E2 {
	var a, b, c fp.Element
	a.Add(&x.A0, &x.A1)
	b.Add(&y.A0, &y.A1)
	a.Mul(&a, &b)
	b.Mul(&x.A0, &y.A0)
	c.Mul(&x.A1, &y.A1)
	z.A1.Sub(&a, &b).Sub(&z.A1, &c)
	z.A0.Sub(&b, &c)
	return z
}

// Square sets z = x² using complex squaring (2 Fp muls + 1 add).
func (z *E2) Square(x *E2) *E2 {
	var a, b fp.Element
	a.Add(&x.A0, &x.A1)
	b.Sub(&x.A0, &x.A1)
	a.Mul(&a, &b)
	b.Mul(&x.A0, &x.A1).Double(&b)
	z.A0.Set(&a)
	z.A1.Set(&b)
	return z
}

// Inverse sets z = 1/x via norm: N(x) = x0² + x1².
func (z *E2) Inverse(x *E2) *E2 {
	var t0, t1 fp.Element
	t0.Square(&x.A0)
	t1.Square(&x.A1)
	t0.Add(&t0, &t1)
	t1.Inverse(&t0)
	z.A0.Mul(&x.A0, &t1)
	z.A1.Mul(&x.A1, &t1).Neg(&z.A1)
	return z
}

// Norm sets x = N(z) = z.A0² + z.A1².
func (z *E2) Norm(x *fp.Element) {
	var tmp fp.Element
	x.Square(&z.A0)
	tmp.Square(&z.A1)
	x.Add(x, &tmp)
}

// Exp sets z = x^k using square-and-multiply (big-endian).
func (z *E2) Exp(x E2, k *big.Int) *E2 {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}
	e := k
	if k.Sign() == -1 {
		x.Inverse(&x)
		e = new(big.Int).Neg(k)
	}
	z.SetOne()
	b := e.Bytes()
	for i := 0; i < len(b); i++ {
		w := b[i]
		for j := 0; j < 8; j++ {
			z.Square(z)
			if (w & (0b10000000 >> j)) != 0 {
				z.Mul(z, &x)
			}
		}
	}
	return z
}

// Sqrt sets z = √x in Fp2 using Scott §6.3, valid for q ≡ 3 mod 4.
func (z *E2) Sqrt(x *E2) *E2 {
	var a1, alpha, x0, minusOne E2
	minusOne.SetOne().Neg(&minusOne)

	a1.expBySqrtHelper(x)
	alpha.Square(&a1).Mul(&alpha, x)
	x0.Mul(x, &a1)

	if alpha.Equal(&minusOne) {
		c := x0.A0
		z.A0.Neg(&x0.A1)
		z.A1.Set(&c)
		return z
	}
	var b E2
	b.SetOne()
	b.A0.Add(&b.A0, &alpha.A0)
	b.A1.Add(&b.A1, &alpha.A1)
	b.Exp(b, &sqrtExp2).Mul(&b, &x0)
	return z.Set(&b)
}

var sqrtExp2 big.Int

func init() {
	q := fp.Modulus()
	sqrtExp2.Sub(q, big.NewInt(1))
	sqrtExp2.Rsh(&sqrtExp2, 1) // (q-1)/2
}

// expBySqrtHelper sets z = x^{(q-3)/4} in Fp2 using a short addition chain.
// (q-3)/4 = 0x3fffffffc00000004000000000000000000000003fffffffffffffffffffffff
// Addition chain: cost 264 = 253 sq + 11 mul.
func (z *E2) expBySqrtHelper(x *E2) *E2 {
	var t0, t1, t2, t3, t4, t5, t6, t7, t8 E2

	t0.Square(x)
	t1.Mul(x, &t0)
	t2.Square(&t1)
	t3.Mul(x, &t2)
	t4.Square(&t3)
	t4.Square(&t4)
	t4.Square(&t4)
	t5.Mul(&t3, &t4)
	t8.Square(&t5)
	for k := 0; k < 5; k++ {
		t8.Square(&t8)
	}
	t8.Mul(&t8, &t5)
	t6.Square(&t8)
	t6.Square(&t6)
	t6.Square(&t6)
	t6.Mul(&t6, &t3)
	t7.Square(&t6)
	t7.Mul(&t7, x)
	t8.Square(&t7)
	for k := 0; k < 15; k++ {
		t8.Square(&t8)
	}
	t8.Mul(&t8, &t7)
	for k := 0; k < 15; k++ {
		t8.Square(&t8)
	}
	t5.Mul(&t6, &t8)
	for k := 0; k < 17; k++ {
		t8.Square(&t8)
	}
	t8.Mul(&t8, x)
	for k := 0; k < 143; k++ {
		t8.Square(&t8)
	}
	t8.Mul(&t8, &t5)
	for k := 0; k < 47; k++ {
		t8.Square(&t8)
	}
	z.Mul(&t5, &t8)
	return z
}

// Cbrt sets z = ∛x in Fp2 using the algebraic torus T₂(Fp).
// For q ≡ 4 mod 9 and β = −1: v₃(q+1) = 0 ensures every element of T₂ has a
// unique cube root. Returns z, or nil if x is not a cubic residue in Fp2.
func (z *E2) Cbrt(x *E2) *E2 {
	if x.A1.IsZero() {
		if z.A0.Cbrt(&x.A0) == nil {
			return nil
		}
		z.A1.SetZero()
		return z
	}

	if x.A0.IsZero() {
		var negA1 fp.Element
		negA1.Neg(&x.A1)
		if z.A1.Cbrt(&negA1) == nil {
			return nil
		}
		z.A0.SetZero()
		return z.cbrtVerify(x)
	}

	var x0sq, x1sq, norm fp.Element
	x0sq.Square(&x.A0)
	x1sq.Square(&x.A1)
	norm.Add(&x0sq, &x1sq)

	m, normInv, ok := cbrtAndNormInverse(&norm)
	if !ok {
		return nil
	}

	// τ = 2·(A0² − A1²)/N
	var tau fp.Element
	tau.Sub(&x0sq, &x1sq)
	tau.Double(&tau)
	tau.Mul(&tau, &normInv)

	sigma := lucasV(&tau)

	// z₀ = A0/(m·(σ−1)), z₁ = A1/(m·(σ+1))
	var one, d0, d1, d0d1, d0d1Inv fp.Element
	one.SetOne()
	d0.Sub(&sigma, &one)
	d0.Mul(&m, &d0)
	d1.Add(&sigma, &one)
	d1.Mul(&m, &d1)

	d0d1.Mul(&d0, &d1)
	d0d1Inv.Inverse(&d0d1)

	z.A0.Mul(&d1, &d0d1Inv).Mul(&z.A0, &x.A0)
	z.A1.Mul(&d0, &d0d1Inv).Mul(&z.A1, &x.A1)

	return z.cbrtVerify(x)
}

func (z *E2) cbrtVerify(x *E2) *E2 {
	var c E2
	c.Square(z).Mul(&c, z)
	if !c.Equal(x) {
		return nil
	}
	return z
}

func cbrtAndNormInverse(norm *fp.Element) (m, normInv fp.Element, ok bool) {
	var t, t2, t4, t8, t9, n2, n3 fp.Element
	t.ExpByCbrtHelperQMinus4Div9(*norm)
	t2.Square(&t)
	t4.Square(&t2)
	t8.Square(&t4)
	t9.Mul(&t8, &t)
	n2.Square(norm)
	n3.Mul(&n2, norm)
	m.Mul(&t8, &n3)
	normInv.Mul(&t9, &n2)

	var c fp.Element
	c.Square(&m).Mul(&c, &m)
	if !c.Equal(norm) {
		return m, normInv, false
	}
	return m, normInv, true
}

// lucasExponent is e = 3⁻¹ mod (q+1) as little-endian uint64 limbs.
var lucasExponent = [4]uint64{
	12297829382473034411,
	6148914692668172970,
	6148914691236517205,
	6148914689804861440,
}

func lucasV(alpha *fp.Element) fp.Element {
	var v0, v1, two, prod fp.Element
	two.SetUint64(2)
	v0.Set(alpha)
	v1.Square(alpha).Sub(&v1, &two)

	for i := 253; i >= 1; i-- {
		bit := (lucasExponent[i/64] >> uint(i%64)) & 1
		prod.Mul(&v0, &v1).Sub(&prod, alpha)
		if bit == 0 {
			v1.Set(&prod)
			v0.Square(&v0).Sub(&v0, &two)
		} else {
			v0.Set(&prod)
			v1.Square(&v1).Sub(&v1, &two)
		}
	}
	v0.Mul(&v0, &v1).Sub(&v0, alpha)
	return v0
}
