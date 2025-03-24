// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"errors"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
)

var bigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}

// E24 is a degree two finite field extension of fp6
type E24 struct {
	D0, D1 E12
}

// Equal returns true if z equals x, false otherwise
func (z *E24) Equal(x *E24) bool {
	return z.D0.Equal(&x.D0) && z.D1.Equal(&x.D1)
}

// String puts E24 in string form
func (z *E24) String() string {
	return (z.D0.String() + "+(" + z.D1.String() + ")*i")
}

// SetString sets a E24 from string
func (z *E24) SetString(s0, s1, s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12, s13, s14, s15, s16, s17, s18, s19, s20, s21, s22, s23 string) *E24 {
	z.D0.SetString(s0, s1, s2, s3, s4, s5, s6, s7, s8, s9, s10, s11)
	z.D1.SetString(s12, s13, s14, s15, s16, s17, s18, s19, s20, s21, s22, s23)
	return z
}

// Set copies x into z and returns z
func (z *E24) Set(x *E24) *E24 {
	z.D0 = x.D0
	z.D1 = x.D1
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E24) SetOne() *E24 {
	*z = E24{}
	z.D0.C0.B0.A0.SetOne()
	return z
}

// Add set z=x+y in E24 and return z
func (z *E24) Add(x, y *E24) *E24 {
	z.D0.Add(&x.D0, &y.D0)
	z.D1.Add(&x.D1, &y.D1)
	return z
}

// Sub sets z to x sub y and return z
func (z *E24) Sub(x, y *E24) *E24 {
	z.D0.Sub(&x.D0, &y.D0)
	z.D1.Sub(&x.D1, &y.D1)
	return z
}

// Double sets z=2*x and returns z
func (z *E24) Double(x *E24) *E24 {
	z.D0.Double(&x.D0)
	z.D1.Double(&x.D1)
	return z
}

// SetRandom used only in tests
func (z *E24) SetRandom() (*E24, error) {
	if _, err := z.D0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.D1.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// MustSetRandom sets z to a uniform random value.
// It panics if reading from crypto/rand fails.
func (z *E24) MustSetRandom() *E24 {
	if _, err := z.SetRandom(); err != nil {
		panic(err)
	}
	return z
}

// IsZero returns true if z is zero, false otherwise
func (z *E24) IsZero() bool {
	return z.D0.IsZero() && z.D1.IsZero()
}

// IsOne returns true if z is one, false otherwise
func (z *E24) IsOne() bool {
	return z.D0.IsOne() && z.D1.IsZero()
}

// Mul set z=x*y in E24 and return z
func (z *E24) Mul(x, y *E24) *E24 {
	var a, b, c E12
	a.Add(&x.D0, &x.D1)
	b.Add(&y.D0, &y.D1)
	a.Mul(&a, &b)
	b.Mul(&x.D0, &y.D0)
	c.Mul(&x.D1, &y.D1)
	z.D1.Sub(&a, &b).Sub(&z.D1, &c)
	z.D0.MulByNonResidue(&c).Add(&z.D0, &b)
	return z
}

// Square set z=x*x in E24 and return z
func (z *E24) Square(x *E24) *E24 {

	//Algorithm 22 from https://eprint.iacr.org/2010/354.pdf
	var c0, c2, c3 E12
	c0.Sub(&x.D0, &x.D1)
	c3.MulByNonResidue(&x.D1).Neg(&c3).Add(&x.D0, &c3)
	c2.Mul(&x.D0, &x.D1)
	c0.Mul(&c0, &c3).Add(&c0, &c2)
	z.D1.Double(&c2)
	c2.MulByNonResidue(&c2)
	z.D0.Add(&c0, &c2)

	return z
}

// Karabina's compressed cyclotomic square
// https://eprint.iacr.org/2010/542.pdf
// Th. 3.2 with minor modifications to fit our tower
func (z *E24) CyclotomicSquareCompressed(x *E24) *E24 {

	var t [7]E4

	// t0 = g1²
	t[0].Square(&x.D0.C1)
	// t1 = g5²
	t[1].Square(&x.D1.C2)
	// t5 = g1 + g5
	t[5].Add(&x.D0.C1, &x.D1.C2)
	// t2 = (g1 + g5)²
	t[2].Square(&t[5])

	// t3 = g1² + g5²
	t[3].Add(&t[0], &t[1])
	// t5 = 2 * g1 * g5
	t[5].Sub(&t[2], &t[3])

	// t6 = g3 + g2
	t[6].Add(&x.D1.C0, &x.D0.C2)
	// t3 = (g3 + g2)²
	t[3].Square(&t[6])
	// t2 = g3²
	t[2].Square(&x.D1.C0)

	// t6 = 2 * nr * g1 * g5
	t[6].MulByNonResidue(&t[5])
	// t5 = 4 * nr * g1 * g5 + 2 * g3
	t[5].Add(&t[6], &x.D1.C0).
		Double(&t[5])
	// z3 = 6 * nr * g1 * g5 + 2 * g3
	z.D1.C0.Add(&t[5], &t[6])

	// t4 = nr * g5²
	t[4].MulByNonResidue(&t[1])
	// t5 = nr * g5² + g1²
	t[5].Add(&t[0], &t[4])
	// t6 = nr * g5² + g1² - g2
	t[6].Sub(&t[5], &x.D0.C2)

	// t1 = g2²
	t[1].Square(&x.D0.C2)

	// t6 = 2 * nr * g5² + 2 * g1² - 2*g2
	t[6].Double(&t[6])
	// z2 = 3 * nr * g5² + 3 * g1² - 2*g2
	z.D0.C2.Add(&t[6], &t[5])

	// t4 = nr * g2²
	t[4].MulByNonResidue(&t[1])
	// t5 = g3² + nr * g2²
	t[5].Add(&t[2], &t[4])
	// t6 = g3² + nr * g2² - g1
	t[6].Sub(&t[5], &x.D0.C1)
	// t6 = 2 * g3² + 2 * nr * g2² - 2 * g1
	t[6].Double(&t[6])
	// z1 = 3 * g3² + 3 * nr * g2² - 2 * g1
	z.D0.C1.Add(&t[6], &t[5])

	// t0 = g2² + g3²
	t[0].Add(&t[2], &t[1])
	// t5 = 2 * g3 * g2
	t[5].Sub(&t[3], &t[0])
	// t6 = 2 * g3 * g2 + g5
	t[6].Add(&t[5], &x.D1.C2)
	// t6 = 4 * g3 * g2 + 2 * g5
	t[6].Double(&t[6])
	// z5 = 6 * g3 * g2 + 2 * g5
	z.D1.C2.Add(&t[5], &t[6])

	return z
}

// DecompressKarabina Karabina's cyclotomic square result
// if g3 != 0
//
//	g4 = (E * g5^2 + 3 * g1^2 - 2 * g2)/4g3
//
// if g3 == 0
//
//	g4 = 2g1g5/g2
//
// if g3=g2=0 then g4=g5=g1=0 and g0=1 (x=1)
// Theorem 3.1 is well-defined for all x in Gϕₙ\{1}
func (z *E24) DecompressKarabina(x *E24) *E24 {

	var t [3]E4
	var one E4
	one.SetOne()

	if x.D1.C0.IsZero() /* g3 == 0 */ {
		t[0].Mul(&x.D0.C1, &x.D1.C2).
			Double(&t[0])
		// t1 = g2
		t[1].Set(&x.D0.C2)

		if t[1].IsZero() /* g2 == g3 == 0 */ {
			return z.SetOne()
		}
	} else /* g3 != 0 */ {
		// t0 = g1^2
		t[0].Square(&x.D0.C1)
		// t1 = 3 * g1^2 - 2 * g2
		t[1].Sub(&t[0], &x.D0.C2).
			Double(&t[1]).
			Add(&t[1], &t[0])
		// t0 = E * g5^2 + t1
		t[2].Square(&x.D1.C2)
		t[0].MulByNonResidue(&t[2]).
			Add(&t[0], &t[1])
		// t1 = 1/(4 * g3)
		t[1].Double(&x.D1.C0).
			Double(&t[1])
	}

	// z4 = g4
	z.D1.C1.Div(&t[0], &t[1]) // costly

	// t1 = g2 * g1
	t[1].Mul(&x.D0.C2, &x.D0.C1)
	// t2 = 2 * g4² - 3 * g2 * g1
	t[2].Square(&x.D1.C1).
		Sub(&t[2], &t[1]).
		Double(&t[2]).
		Sub(&t[2], &t[1])
	// t1 = g3 * g5 (g3 can be 0)
	t[1].Mul(&x.D1.C0, &x.D1.C2)
	// c₀ = E * (2 * g4² + g3 * g5 - 3 * g2 * g1) + 1
	t[2].Add(&t[2], &t[1])
	z.D0.C0.MulByNonResidue(&t[2]).
		Add(&z.D0.C0, &one)

	z.D0.C1.Set(&x.D0.C1)
	z.D0.C2.Set(&x.D0.C2)
	z.D1.C0.Set(&x.D1.C0)
	z.D1.C2.Set(&x.D1.C2)

	return z
}

// BatchDecompressKarabina multiple Karabina's cyclotomic square results
// if g3 != 0
//
//	g4 = (E * g5^2 + 3 * g1^2 - 2 * g2)/4g3
//
// if g3 == 0
//
//	g4 = 2g1g5/g2
//
// if g3=g2=0 then g4=g5=g1=0 and g0=1 (x=1)
// Theorem 3.1 is well-defined for all x in Gϕₙ\{1}
//
// Divisions by 4g3 or g2 is batched using Montgomery batch inverse
func BatchDecompressKarabina(x []E24) []E24 {

	n := len(x)
	if n == 0 {
		return x
	}

	t0 := make([]E4, n)
	t1 := make([]E4, n)
	t2 := make([]E4, n)

	var one E4
	one.SetOne()

	for i := 0; i < n; i++ {
		// g3 == 0
		if x[i].D1.C0.IsZero() {
			t0[i].Mul(&x[i].D0.C1, &x[i].D1.C2).
				Double(&t0[i])
			// t1 = g2
			t1[i].Set(&x[i].D0.C2)

			// g3 != 0
		} else {
			// t0 = g1^2
			t0[i].Square(&x[i].D0.C1)
			// t1 = 3 * g1^2 - 2 * g2
			t1[i].Sub(&t0[i], &x[i].D0.C2).
				Double(&t1[i]).
				Add(&t1[i], &t0[i])
			// t0 = E * g5^2 + t1
			t2[i].Square(&x[i].D1.C2)
			t0[i].MulByNonResidue(&t2[i]).
				Add(&t0[i], &t1[i])
			// t1 = 4 * g3
			t1[i].Double(&x[i].D1.C0).
				Double(&t1[i])
		}
	}

	t1 = BatchInvertE4(t1) // costs 1 inverse

	for i := 0; i < n; i++ {
		// z4 = g4
		x[i].D1.C1.Mul(&t0[i], &t1[i])

		// t1 = g2 * g1
		t1[i].Mul(&x[i].D0.C2, &x[i].D0.C1)
		// t2 = 2 * g4² - 3 * g2 * g1
		t2[i].Square(&x[i].D1.C1)
		t2[i].Sub(&t2[i], &t1[i])
		t2[i].Double(&t2[i])
		t2[i].Sub(&t2[i], &t1[i])

		// t1 = g3 * g5 (g3s can be 0s)
		t1[i].Mul(&x[i].D1.C0, &x[i].D1.C2)
		// z0 = E * (2 * g4² + g3 * g5 - 3 * g2 * g1) + 1
		t2[i].Add(&t2[i], &t1[i])
		x[i].D0.C0.MulByNonResidue(&t2[i]).
			Add(&x[i].D0.C0, &one)
	}

	return x
}

// Granger-Scott's cyclotomic square
// https://eprint.iacr.org/2009/565.pdf, 3.2
func (z *E24) CyclotomicSquare(x *E24) *E24 {

	// x=(x0,x1,x2,x3,x4,x5,x6,x7) in E4⁶
	// cyclosquare(x)=(3*x4²*u + 3*x0² - 2*x0,
	//					3*x2²*u + 3*x3² - 2*x1,
	//					3*x5²*u + 3*x1² - 2*x2,
	//					6*x1*x5*u + 2*x3,
	//					6*x0*x4 + 2*x4,
	//					6*x2*x3 + 2*x5)

	var t [9]E4

	t[0].Square(&x.D1.C1)
	t[1].Square(&x.D0.C0)
	t[6].Add(&x.D1.C1, &x.D0.C0).Square(&t[6]).Sub(&t[6], &t[0]).Sub(&t[6], &t[1]) // 2*x4*x0
	t[2].Square(&x.D0.C2)
	t[3].Square(&x.D1.C0)
	t[7].Add(&x.D0.C2, &x.D1.C0).Square(&t[7]).Sub(&t[7], &t[2]).Sub(&t[7], &t[3]) // 2*x2*x3
	t[4].Square(&x.D1.C2)
	t[5].Square(&x.D0.C1)
	t[8].Add(&x.D1.C2, &x.D0.C1).Square(&t[8]).Sub(&t[8], &t[4]).Sub(&t[8], &t[5]).MulByNonResidue(&t[8]) // 2*x5*x1*u

	t[0].MulByNonResidue(&t[0]).Add(&t[0], &t[1]) // x4²*u + x0²
	t[2].MulByNonResidue(&t[2]).Add(&t[2], &t[3]) // x2²*u + x3²
	t[4].MulByNonResidue(&t[4]).Add(&t[4], &t[5]) // x5²*u + x1²

	z.D0.C0.Sub(&t[0], &x.D0.C0).Double(&z.D0.C0).Add(&z.D0.C0, &t[0])
	z.D0.C1.Sub(&t[2], &x.D0.C1).Double(&z.D0.C1).Add(&z.D0.C1, &t[2])
	z.D0.C2.Sub(&t[4], &x.D0.C2).Double(&z.D0.C2).Add(&z.D0.C2, &t[4])

	z.D1.C0.Add(&t[8], &x.D1.C0).Double(&z.D1.C0).Add(&z.D1.C0, &t[8])
	z.D1.C1.Add(&t[6], &x.D1.C1).Double(&z.D1.C1).Add(&z.D1.C1, &t[6])
	z.D1.C2.Add(&t[7], &x.D1.C2).Double(&z.D1.C2).Add(&z.D1.C2, &t[7])

	return z
}

// Inverse set z to the inverse of x in E24 and return z
//
// if x == 0, sets and returns z = x
func (z *E24) Inverse(x *E24) *E24 {
	// Algorithm 23 from https://eprint.iacr.org/2010/354.pdf

	var t0, t1, tmp E12
	t0.Square(&x.D0)
	t1.Square(&x.D1)
	tmp.MulByNonResidue(&t1)
	t0.Sub(&t0, &tmp)
	t1.Inverse(&t0)
	z.D0.Mul(&x.D0, &t1)
	z.D1.Mul(&x.D1, &t1).Neg(&z.D1)

	return z
}

// BatchInvertE24 returns a new slice with every element inverted.
// Uses Montgomery batch inversion trick
//
// if a[i] == 0, returns result[i] = a[i]
func BatchInvertE24(a []E24) []E24 {
	res := make([]E24, len(a))
	if len(a) == 0 {
		return res
	}

	zeroes := make([]bool, len(a))
	var accumulator E24
	accumulator.SetOne()

	for i := 0; i < len(a); i++ {
		if a[i].IsZero() {
			zeroes[i] = true
			continue
		}
		res[i].Set(&accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	accumulator.Inverse(&accumulator)

	for i := len(a) - 1; i >= 0; i-- {
		if zeroes[i] {
			continue
		}
		res[i].Mul(&res[i], &accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	return res
}

// Exp sets z=xᵏ (mod q²⁴) and returns it
// uses 2-bits windowed method
func (z *E24) Exp(x E24, k *big.Int) *E24 {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q²⁴) == (x⁻¹)ᵏ (mod q²⁴)
		x.Inverse(&x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = bigIntPool.Get().(*big.Int)
		defer bigIntPool.Put(e)
		e.Neg(k)
	}

	var res E24
	var ops [3]E24

	res.SetOne()
	ops[0].Set(&x)
	ops[1].Square(&ops[0])
	ops[2].Set(&ops[0]).Mul(&ops[2], &ops[1])

	b := e.Bytes()
	for i := range b {
		w := b[i]
		mask := byte(0xc0)
		for j := 0; j < 4; j++ {
			res.Square(&res).Square(&res)
			c := (w & mask) >> (6 - 2*j)
			if c != 0 {
				res.Mul(&res, &ops[c-1])
			}
			mask = mask >> 2
		}
	}
	z.Set(&res)

	return z
}

// CyclotomicExp sets z=xᵏ (mod q²⁴) and returns it
// uses 2-NAF decomposition
// x must be in the cyclotomic subgroup
// TODO: use a windowed method
func (z *E24) CyclotomicExp(x E24, k *big.Int) *E24 {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert (=conjugate)
		// if k < 0: xᵏ (mod q²⁴) == (x⁻¹)ᵏ (mod q²⁴)
		x.Conjugate(&x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = bigIntPool.Get().(*big.Int)
		defer bigIntPool.Put(e)
		e.Neg(k)
	}

	var res, xInv E24
	xInv.InverseUnitary(&x)
	res.SetOne()
	eNAF := make([]int8, e.BitLen()+3)
	n := ecc.NafDecomposition(e, eNAF[:])
	for i := n - 1; i >= 0; i-- {
		res.CyclotomicSquare(&res)
		if eNAF[i] == 1 {
			res.Mul(&res, &x)
		} else if eNAF[i] == -1 {
			res.Mul(&res, &xInv)
		}
	}
	z.Set(&res)
	return z
}

// ExpGLV sets z=xᵏ (q²⁴) and returns it
// uses 2-dimensional GLV with 2-bits windowed method
// x must be in GT
// TODO: use 2-NAF
// TODO: use higher dimensional decomposition
func (z *E24) ExpGLV(x E24, k *big.Int) *E24 {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q²⁴) == (x⁻¹)ᵏ (mod q²⁴)
		x.Conjugate(&x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = bigIntPool.Get().(*big.Int)
		defer bigIntPool.Put(e)
		e.Neg(k)
	}

	var table [15]E24
	var res E24
	var s1, s2 fr.Element

	res.SetOne()

	// table[b3b2b1b0-1] = b3b2*Frobinius(x) + b1b0*x
	table[0].Set(&x)
	table[3].Frobenius(&x)

	// split the scalar, modifies ±x, Frob(x) accordingly
	s := ecc.SplitScalar(e, &glvBasis)

	if s[0].Sign() == -1 {
		s[0].Neg(&s[0])
		table[0].InverseUnitary(&table[0])
	}
	if s[1].Sign() == -1 {
		s[1].Neg(&s[1])
		table[3].InverseUnitary(&table[3])
	}

	// precompute table (2 bits sliding window)
	// table[b3b2b1b0-1] = b3b2*Frobenius(x) + b1b0*x if b3b2b1b0 != 0
	table[1].CyclotomicSquare(&table[0])
	table[2].Mul(&table[1], &table[0])
	table[4].Mul(&table[3], &table[0])
	table[5].Mul(&table[3], &table[1])
	table[6].Mul(&table[3], &table[2])
	table[7].CyclotomicSquare(&table[3])
	table[8].Mul(&table[7], &table[0])
	table[9].Mul(&table[7], &table[1])
	table[10].Mul(&table[7], &table[2])
	table[11].Mul(&table[7], &table[3])
	table[12].Mul(&table[11], &table[0])
	table[13].Mul(&table[11], &table[1])
	table[14].Mul(&table[11], &table[2])

	// bounds on the lattice base vectors guarantee that s1, s2 are len(r)/2 bits long max
	s1 = s1.SetBigInt(&s[0]).Bits()
	s2 = s2.SetBigInt(&s[1]).Bits()

	maxBit := s1.BitLen()
	if s2.BitLen() > maxBit {
		maxBit = s2.BitLen()
	}
	hiWordIndex := (maxBit - 1) / 64

	// loop starts from len(s1)/2 due to the bounds
	for i := hiWordIndex; i >= 0; i-- {
		mask := uint64(3) << 62
		for j := 0; j < 32; j++ {
			res.CyclotomicSquare(&res).CyclotomicSquare(&res)
			b1 := (s1[i] & mask) >> (62 - 2*j)
			b2 := (s2[i] & mask) >> (62 - 2*j)
			if b1|b2 != 0 {
				s := (b2<<2 | b1)
				res.Mul(&res, &table[s-1])
			}
			mask = mask >> 2
		}
	}

	z.Set(&res)
	return z
}

// InverseUnitary inverse a unitary element
func (z *E24) InverseUnitary(x *E24) *E24 {
	return z.Conjugate(x)
}

// Conjugate set z to x conjugated and return z
func (z *E24) Conjugate(x *E24) *E24 {
	*z = *x
	z.D1.Neg(&z.D1)
	return z
}

// SizeOfGT represents the size in bytes that a GT element need in binary form
const sizeOfFp = 40
const SizeOfGT = sizeOfFp * 24

// Marshal converts z to a byte slice
func (z *E24) Marshal() []byte {
	b := z.Bytes()
	return b[:]
}

// Unmarshal is an alias to SetBytes()
func (z *E24) Unmarshal(buf []byte) error {
	return z.SetBytes(buf)
}

func (z *E24) Bytes() (r [SizeOfGT]byte) {

	offset := 0
	var buf [sizeOfFp]byte

	buf = z.D0.C0.B0.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C0.B0.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C0.B1.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C0.B1.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C1.B0.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C1.B0.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C1.B1.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C1.B1.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C2.B0.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C2.B0.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C2.B1.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D0.C2.B1.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C0.B0.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C0.B0.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C0.B1.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C0.B1.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C1.B0.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C1.B0.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C1.B1.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C1.B1.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C2.B0.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C2.B0.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C2.B1.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D1.C2.B1.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])

	return
}

// SetBytes interprets e as the bytes of a big-endian GT
// sets z to that value (in Montgomery form), and returns z.
func (z *E24) SetBytes(e []byte) error {
	if len(e) != SizeOfGT {
		return errors.New("invalid buffer size")
	}
	offset := 0
	z.D0.C0.B0.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C0.B0.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C0.B1.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C0.B1.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C1.B0.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C1.B0.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C1.B1.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C1.B1.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C2.B0.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C2.B0.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C2.B1.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D0.C2.B1.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C0.B0.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C0.B0.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C0.B1.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C0.B1.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C1.B0.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C1.B0.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C1.B1.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C1.B1.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C2.B0.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C2.B0.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C2.B1.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D1.C2.B1.A1.SetBytes(e[offset : offset+sizeOfFp])

	return nil
}

// IsInSubGroup ensures GT/E24 is in correct subgroup
func (z *E24) IsInSubGroup() bool {
	var a, b E24

	// check z^(phi_k(p)) == 1
	a.FrobeniusQuad(z)
	b.FrobeniusQuad(&a).Mul(&b, z)

	if !a.Equal(&b) {
		return false
	}

	// check z^(p+1-t) == 1
	a.Frobenius(z)
	b.Expt(z)

	return a.Equal(&b)
}

// CompressTorus GT/E24 element to half its size
// z must be in the cyclotomic subgroup
// i.e. z^(p⁴-p²+1)=1
// e.g. GT
// "COMPRESSION IN FINITE FIELDS AND TORUS-BASED CRYPTOGRAPHY", K. RUBIN AND A. SILVERBERG
// z.C1 == 0 only when z ∈ {-1,1}
func (z *E24) CompressTorus() (E12, error) {

	if z.D1.IsZero() {
		return E12{}, errors.New("invalid input")
	}

	var res, tmp, one E12
	one.SetOne()
	tmp.Inverse(&z.D1)
	res.Add(&z.D0, &one).
		Mul(&res, &tmp)

	return res, nil
}

// BatchCompressTorus GT/E24 elements to half their size
// using a batch inversion
func BatchCompressTorus(x []E24) ([]E12, error) {

	n := len(x)
	if n == 0 {
		return []E12{}, errors.New("invalid input size")
	}

	var one E12
	one.SetOne()
	res := make([]E12, n)

	for i := 0; i < n; i++ {
		res[i].Set(&x[i].D1)
		//  throw an error if any of the x[i].C1 is 0
		if res[i].IsZero() {
			return []E12{}, errors.New("invalid input")
		}
	}

	t := BatchInvertE12(res) // costs 1 inverse

	for i := 0; i < n; i++ {
		res[i].Add(&x[i].D0, &one).
			Mul(&res[i], &t[i])
	}

	return res, nil
}

// DecompressTorus GT/E24 a compressed element
// element must be in the cyclotomic subgroup
// "COMPRESSION IN FINITE FIELDS AND TORUS-BASED CRYPTOGRAPHY", K. RUBIN AND A. SILVERBERG
func (z *E12) DecompressTorus() E24 {

	var res, num, denum E24
	num.D0.Set(z)
	num.D1.SetOne()
	denum.D0.Set(z)
	denum.D1.SetOne().Neg(&denum.D1)
	res.Inverse(&denum).
		Mul(&res, &num)

	return res
}

// BatchDecompressTorus GT/E24 compressed elements
// using a batch inversion
func BatchDecompressTorus(x []E12) ([]E24, error) {

	n := len(x)
	if n == 0 {
		return []E24{}, errors.New("invalid input size")
	}

	res := make([]E24, n)
	num := make([]E24, n)
	denum := make([]E24, n)

	for i := 0; i < n; i++ {
		num[i].D0.Set(&x[i])
		num[i].D1.SetOne()
		denum[i].D0.Set(&x[i])
		denum[i].D1.SetOne().Neg(&denum[i].D1)
	}

	denum = BatchInvertE24(denum) // costs 1 inverse

	for i := 0; i < n; i++ {
		res[i].Mul(&num[i], &denum[i])
	}

	return res, nil
}
