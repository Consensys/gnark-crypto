// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"errors"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
)

var bigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}

// E6 is a degree two finite field extension of fp3
type E6 struct {
	B0, B1 E3
}

// Equal returns true if z equals x, false otherwise
func (z *E6) Equal(x *E6) bool {
	return z.B0.Equal(&x.B0) && z.B1.Equal(&x.B1)
}

// String puts E6 in string form
func (z *E6) String() string {
	return (z.B0.String() + "+(" + z.B1.String() + ")*v")
}

// SetString sets a E6 from string
func (z *E6) SetString(s0, s1, s2, s3, s4, s5 string) *E6 {
	z.B0.SetString(s0, s1, s2)
	z.B1.SetString(s3, s4, s5)
	return z
}

// Set copies x into z and returns z
func (z *E6) Set(x *E6) *E6 {
	*z = *x
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E6) SetOne() *E6 {
	*z = E6{}
	z.B0.A0.SetOne()
	return z
}

// Add sets z=x+y in E6 and returns z
func (z *E6) Add(x, y *E6) *E6 {
	z.B0.Add(&x.B0, &y.B0)
	z.B1.Add(&x.B1, &y.B1)
	return z
}

// Sub sets z to x-y and returns z
func (z *E6) Sub(x, y *E6) *E6 {
	z.B0.Sub(&x.B0, &y.B0)
	z.B1.Sub(&x.B1, &y.B1)
	return z
}

// Double sets z=2*x and returns z
func (z *E6) Double(x *E6) *E6 {
	z.B0.Double(&x.B0)
	z.B1.Double(&x.B1)
	return z
}

// SetRandom used only in tests
func (z *E6) SetRandom() (*E6, error) {
	if _, err := z.B0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.B1.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// MustSetRandom sets z to a uniform random value.
// It panics if reading from crypto/rand fails.
func (z *E6) MustSetRandom() *E6 {
	if _, err := z.SetRandom(); err != nil {
		panic(err)
	}
	return z
}

// IsZero returns true if z is zero, false otherwise
func (z *E6) IsZero() bool {
	return z.B0.IsZero() && z.B1.IsZero()
}

// IsOne returns true if z is one, false otherwise
func (z *E6) IsOne() bool {
	return z.B0.IsOne() && z.B1.IsZero()
}

// Mul sets z=x*y in E6 and returns z
func (z *E6) Mul(x, y *E6) *E6 {
	var a, b, c E3
	a.Add(&x.B0, &x.B1)
	b.Add(&y.B0, &y.B1)
	a.Mul(&a, &b)
	b.Mul(&x.B0, &y.B0)
	c.Mul(&x.B1, &y.B1)
	z.B1.Sub(&a, &b).Sub(&z.B1, &c)
	z.B0.MulByNonResidue(&c).Add(&z.B0, &b)
	return z
}

// Square sets z=x*x in E6 and returns z
func (z *E6) Square(x *E6) *E6 {

	//Algorithm 22 from https://eprint.iacr.org/2010/354.pdf
	var c0, c2, c3 E3
	c0.Sub(&x.B0, &x.B1)
	c3.MulByNonResidue(&x.B1).Neg(&c3).Add(&x.B0, &c3)
	c2.Mul(&x.B0, &x.B1)
	c0.Mul(&c0, &c3).Add(&c0, &c2)
	z.B1.Double(&c2)
	c2.MulByNonResidue(&c2)
	z.B0.Add(&c0, &c2)

	return z
}

// Karabina's compressed cyclotomic square
// https://eprint.iacr.org/2010/542.pdf
// Th. 3.2 with minor modifications to fit our tower
func (z *E6) CyclotomicSquareCompressed(x *E6) *E6 {

	var t [7]fp.Element

	// t0 = g1²
	t[0].Square(&x.B0.A1)
	// t1 = g5²
	t[1].Square(&x.B1.A2)
	// t5 = g1 + g5
	t[5].Add(&x.B0.A1, &x.B1.A2)
	// t2 = (g1 + g5)²
	t[2].Square(&t[5])

	// t3 = g1² + g5²
	t[3].Add(&t[0], &t[1])
	// t5 = 2 * g1 * g5
	t[5].Sub(&t[2], &t[3])

	// t6 = g3 + g2
	t[6].Add(&x.B1.A0, &x.B0.A2)
	// t3 = (g3 + g2)²
	t[3].Square(&t[6])
	// t2 = g3²
	t[2].Square(&x.B1.A0)

	// t6 = 2 * nr * g1 * g5
	t[6].MulByNonResidue(&t[5])
	// t5 = 4 * nr * g1 * g5 + 2 * g3
	t[5].Add(&t[6], &x.B1.A0).
		Double(&t[5])
	// z3 = 6 * nr * g1 * g5 + 2 * g3
	z.B1.A0.Add(&t[5], &t[6])

	// t4 = nr * g5²
	t[4].MulByNonResidue(&t[1])
	// t5 = nr * g5² + g1²
	t[5].Add(&t[0], &t[4])
	// t6 = nr * g5² + g1² - g2
	t[6].Sub(&t[5], &x.B0.A2)

	// t1 = g2²
	t[1].Square(&x.B0.A2)

	// t6 = 2 * nr * g5² + 2 * g1² - 2*g2
	t[6].Double(&t[6])
	// z2 = 3 * nr * g5² + 3 * g1² - 2*g2
	z.B0.A2.Add(&t[6], &t[5])

	// t4 = nr * g2²
	t[4].MulByNonResidue(&t[1])
	// t5 = g3² + nr * g2²
	t[5].Add(&t[2], &t[4])
	// t6 = g3² + nr * g2² - g1
	t[6].Sub(&t[5], &x.B0.A1)
	// t6 = 2 * g3² + 2 * nr * g2² - 2 * g1
	t[6].Double(&t[6])
	// z1 = 3 * g3² + 3 * nr * g2² - 2 * g1
	z.B0.A1.Add(&t[6], &t[5])

	// t0 = g2² + g3²
	t[0].Add(&t[2], &t[1])
	// t5 = 2 * g3 * g2
	t[5].Sub(&t[3], &t[0])
	// t6 = 2 * g3 * g2 + g5
	t[6].Add(&t[5], &x.B1.A2)
	// t6 = 4 * g3 * g2 + 2 * g5
	t[6].Double(&t[6])
	// z5 = 6 * g3 * g2 + 2 * g5
	z.B1.A2.Add(&t[5], &t[6])

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
func (z *E6) DecompressKarabina(x *E6) *E6 {

	var t [3]fp.Element
	var one fp.Element
	one.SetOne()

	if x.B1.A0.IsZero() /* g3 == 0 */ {
		t[0].Mul(&x.B0.A1, &x.B1.A2).
			Double(&t[0])
		// t1 = g2
		t[1].Set(&x.B0.A2)

		// g3 != 0

		if t[1].IsZero() /* g2 == g3 == 0 */ {
			return z.SetOne()
		}
	} else /* g3 != 0 */ {
		// t0 = g1^2
		t[0].Square(&x.B0.A1)
		// t1 = 3 * g1^2 - 2 * g2
		t[1].Sub(&t[0], &x.B0.A2).
			Double(&t[1]).
			Add(&t[1], &t[0])
		// t0 = E * g5^2 + t1
		t[2].Square(&x.B1.A2)
		t[0].MulByNonResidue(&t[2]).
			Add(&t[0], &t[1])
		// t1 = 1/(4 * g3)
		t[1].Double(&x.B1.A0).
			Double(&t[1])
	}

	// z4 = g4
	z.B1.A1.Div(&t[0], &t[1]) // costly

	// t1 = g2 * g1
	t[1].Mul(&x.B0.A2, &x.B0.A1)
	// t2 = 2 * g4² - 3 * g2 * g1
	t[2].Square(&x.B1.A1).
		Sub(&t[2], &t[1]).
		Double(&t[2]).
		Sub(&t[2], &t[1])
	// t1 = g3 * g5 (g3 can be 0)
	t[1].Mul(&x.B1.A0, &x.B1.A2)
	// c₀ = E * (2 * g4² + g3 * g5 - 3 * g2 * g1) + 1
	t[2].Add(&t[2], &t[1])
	z.B0.A0.MulByNonResidue(&t[2]).
		Add(&z.B0.A0, &one)

	z.B0.A1.Set(&x.B0.A1)
	z.B0.A2.Set(&x.B0.A2)
	z.B1.A0.Set(&x.B1.A0)
	z.B1.A2.Set(&x.B1.A2)

	return z
}

// Granger-Scott's cyclotomic square
// https://eprint.iacr.org/2009/565.pdf, 3.2
func (z *E6) CyclotomicSquare(x *E6) *E6 {
	// x=(x0,x1,x2,x3,x4,x5,x6,x7) in E3⁶
	// cyclosquare(x)=(3*x4²*u + 3*x0² - 2*x0,
	//					3*x2²*u + 3*x3² - 2*x1,
	//					3*x5²*u + 3*x1² - 2*x2,
	//					6*x1*x5*u + 2*x3,
	//					6*x0*x4 + 2*x4,
	//					6*x2*x3 + 2*x5)

	var t [9]fp.Element

	t[0].Square(&x.B1.A1)
	t[1].Square(&x.B0.A0)
	t[6].Add(&x.B1.A1, &x.B0.A0).Square(&t[6]).Sub(&t[6], &t[0]).Sub(&t[6], &t[1]) // 2*x4*x0
	t[2].Square(&x.B0.A2)
	t[3].Square(&x.B1.A0)
	t[7].Add(&x.B0.A2, &x.B1.A0).Square(&t[7]).Sub(&t[7], &t[2]).Sub(&t[7], &t[3]) // 2*x2*x3
	t[4].Square(&x.B1.A2)
	t[5].Square(&x.B0.A1)
	t[8].Add(&x.B1.A2, &x.B0.A1).Square(&t[8]).Sub(&t[8], &t[4]).Sub(&t[8], &t[5]).MulByNonResidue(&t[8]) // 2*x5*x1*u

	t[0].MulByNonResidue(&t[0]).Add(&t[0], &t[1]) // x4²*u + x0²
	t[2].MulByNonResidue(&t[2]).Add(&t[2], &t[3]) // x2²*u + x3²
	t[4].MulByNonResidue(&t[4]).Add(&t[4], &t[5]) // x5²*u + x1²

	z.B0.A0.Sub(&t[0], &x.B0.A0).Double(&z.B0.A0).Add(&z.B0.A0, &t[0])
	z.B0.A1.Sub(&t[2], &x.B0.A1).Double(&z.B0.A1).Add(&z.B0.A1, &t[2])
	z.B0.A2.Sub(&t[4], &x.B0.A2).Double(&z.B0.A2).Add(&z.B0.A2, &t[4])

	z.B1.A0.Add(&t[8], &x.B1.A0).Double(&z.B1.A0).Add(&z.B1.A0, &t[8])
	z.B1.A1.Add(&t[6], &x.B1.A1).Double(&z.B1.A1).Add(&z.B1.A1, &t[6])
	z.B1.A2.Add(&t[7], &x.B1.A2).Double(&z.B1.A2).Add(&z.B1.A2, &t[7])

	return z
}

// Inverse sets z to the inverse of x in E6 and returns z
//
// if x == 0, sets and returns z = x
func (z *E6) Inverse(x *E6) *E6 {
	// Algorithm 23 from https://eprint.iacr.org/2010/354.pdf

	var t0, t1, tmp E3
	t0.Square(&x.B0)
	t1.Square(&x.B1)
	tmp.MulByNonResidue(&t1)
	t0.Sub(&t0, &tmp)
	t1.Inverse(&t0)
	z.B0.Mul(&x.B0, &t1)
	z.B1.Mul(&x.B1, &t1).Neg(&z.B1)

	return z
}

// BatchInvertE6 returns a new slice with every element in a inverted.
// It uses Montgomery batch inversion trick.
//
// if a[i] == 0, returns result[i] = a[i]
func BatchInvertE6(a []E6) []E6 {
	res := make([]E6, len(a))
	if len(a) == 0 {
		return res
	}

	zeroes := make([]bool, len(a))
	var accumulator E6
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

// Exp sets z=xᵏ (mod q⁶) and returns it
// uses 2-bits windowed method
func (z *E6) Exp(x E6, k *big.Int) *E6 {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q⁶) == (x⁻¹)ᵏ (mod q⁶)
		x.Inverse(&x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = bigIntPool.Get().(*big.Int)
		defer bigIntPool.Put(e)
		e.Neg(k)
	}

	var res E6
	var ops [3]E6

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

// CyclotomicExp sets z=xᵏ (mod q⁶) and returns it
// uses 2-NAF decomposition
// x must be in the cyclotomic subgroup
// TODO: use a windowed method
func (z *E6) CyclotomicExp(x E6, k *big.Int) *E6 {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert (=conjugate)
		// if k < 0: xᵏ (mod q⁶) == (x⁻¹)ᵏ (mod q⁶)
		x.Conjugate(&x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = bigIntPool.Get().(*big.Int)
		defer bigIntPool.Put(e)
		e.Neg(k)
	}

	var res, xInv E6
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

// ExpGLV sets z=xᵏ (q⁶) and returns it
// uses 2-dimensional GLV with 2-bits windowed method
// x must be in GT
// TODO: use 2-NAF
// TODO: use higher dimensional decomposition
func (z *E6) ExpGLV(x E6, k *big.Int) *E6 {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q⁶) == (x⁻¹)ᵏ (mod q⁶)
		x.Conjugate(&x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = bigIntPool.Get().(*big.Int)
		defer bigIntPool.Put(e)
		e.Neg(k)
	}

	var table [15]E6
	var res E6
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

// InverseUnitary inverses a unitary element
func (z *E6) InverseUnitary(x *E6) *E6 {
	return z.Conjugate(x)
}

// Conjugate sets z to x conjugated and returns z
func (z *E6) Conjugate(x *E6) *E6 {
	*z = *x
	z.B1.Neg(&z.B1)
	return z
}

// SizeOfGT represents the size in bytes that a GT element need in binary form
const SizeOfGT = fp.Bytes * 6

// Bytes returns the regular (non montgomery) value
// of z as a big-endian byte array.
// z.C1.B2.A1 | z.C1.B2.A0 | z.C1.B1.A1 | ...
func (z *E6) Bytes() (r [SizeOfGT]byte) {

	offset := 0
	var buf [fp.Bytes]byte

	buf = z.B1.A2.Bytes()
	copy(r[offset:offset+fp.Bytes], buf[:])
	offset += fp.Bytes

	buf = z.B1.A1.Bytes()
	copy(r[offset:offset+fp.Bytes], buf[:])
	offset += fp.Bytes

	buf = z.B1.A0.Bytes()
	copy(r[offset:offset+fp.Bytes], buf[:])
	offset += fp.Bytes

	buf = z.B0.A2.Bytes()
	copy(r[offset:offset+fp.Bytes], buf[:])
	offset += fp.Bytes

	buf = z.B0.A1.Bytes()
	copy(r[offset:offset+fp.Bytes], buf[:])
	offset += fp.Bytes

	buf = z.B0.A0.Bytes()
	copy(r[offset:offset+fp.Bytes], buf[:])

	return
}

// SetBytes interprets e as the bytes of a big-endian GT
// sets z to that value (in Montgomery form), and returns z.
// z.C1.B2.A1 | z.C1.B2.A0 | z.C1.B1.A1 | ...
func (z *E6) SetBytes(e []byte) error {
	if len(e) != SizeOfGT {
		return errors.New("invalid buffer size")
	}
	offset := 0
	z.B1.A2.SetBytes(e[offset : offset+fp.Bytes])
	offset += fp.Bytes
	z.B1.A1.SetBytes(e[offset : offset+fp.Bytes])
	offset += fp.Bytes
	z.B1.A0.SetBytes(e[offset : offset+fp.Bytes])
	offset += fp.Bytes
	z.B0.A2.SetBytes(e[offset : offset+fp.Bytes])
	offset += fp.Bytes
	z.B0.A1.SetBytes(e[offset : offset+fp.Bytes])
	offset += fp.Bytes
	z.B0.A0.SetBytes(e[offset : offset+fp.Bytes])

	return nil
}

// IsInSubGroup ensures GT/E6 is in correct subgroup
func (z *E6) IsInSubGroup() bool {
	var tmp, a, _a, b E6
	var t [6]E6

	// check z^(phi_k(p)) == 1
	a.Frobenius(z)
	b.Frobenius(&a).Mul(&b, z)

	if !a.Equal(&b) {
		return false
	}

	// check z^(p+1-t) == 1
	_a.Frobenius(z)
	a.CyclotomicSquare(&_a).Mul(&a, &_a) // z^(3p)

	// t(x)-1 = (13x⁶ − 23x⁵ − 9x⁴ + 35x³ + 10x + 19)/3
	t[0].CyclotomicSquare(z) // z²
	t[1].CyclotomicSquare(&t[0]).
		CyclotomicSquare(&t[1]) // z⁸
	t[2].CyclotomicSquare(&t[1]).
		Mul(&t[2], &t[0]).
		Mul(&t[2], z) // z¹⁹*
	t[3].Mul(&t[0], &t[1]).
		Expt(&t[3]) // z^(10u)*
	t[4].CyclotomicSquare(&t[3]).
		Mul(&t[4], &t[3]) // z^(30u)
	t[0].CyclotomicSquare(&t[0]).
		Mul(&t[0], z).
		Expt(&t[0]) // z^(5u)
	t[4].Mul(&t[4], &t[0]).
		Expt(&t[4]).
		Expt(&t[4]) // z^(35u³)*
	t[1].Mul(&t[1], z).
		Expt(&t[1]).
		Expt(&t[1]).
		Expt(&t[1]).
		Expt(&t[1]).
		Conjugate(&t[1]) // z^(-9u⁴)*
	t[0].Expt(&t[0]).
		Expt(&t[0]).
		Expt(&t[0]).
		Conjugate(&t[0]) // z^(-5u⁴)
	t[5].CyclotomicSquare(&t[1]).
		Mul(&t[5], &t[0]).
		Expt(&t[5]) // z^(-23u⁵)*
	tmp.CyclotomicSquare(&t[1]).
		Conjugate(&tmp) // z^(18u⁴)
	t[0].Mul(&t[0], &tmp).
		Expt(&t[0]).
		Expt(&t[0]) // z^(13u⁶)*

	b.Mul(&t[2], &t[3]).
		Mul(&b, &t[4]).
		Mul(&b, &t[1]).
		Mul(&b, &t[5]).
		Mul(&b, &t[0]) // z^(3(t-1))

	return a.Equal(&b)
}

// CompressTorus GT/E6 element to half its size
// z must be in the cyclotomic subgroup
// i.e. z^(p⁴-p²+1)=1
// e.g. GT
// "COMPRESSION IN FINITE FIELDS AND TORUS-BASED CRYPTOGRAPHY", K. RUBIN AND A. SILVERBERG
// z.B1 == 0 only when z ∈ {-1,1}
func (z *E6) CompressTorus() (E3, error) {

	if z.B1.IsZero() {
		return E3{}, errors.New("invalid input")
	}

	var res, tmp, one E3
	one.SetOne()
	tmp.Inverse(&z.B1)
	res.Add(&z.B0, &one).
		Mul(&res, &tmp)

	return res, nil
}

// BatchCompressTorus GT/E6 elements to half their size
// using a batch inversion
func BatchCompressTorus(x []E6) ([]E3, error) {

	n := len(x)
	if n == 0 {
		return []E3{}, errors.New("invalid input size")
	}

	var one E3
	one.SetOne()
	res := make([]E3, n)

	for i := 0; i < n; i++ {
		res[i].Set(&x[i].B1)
		//  throw an error if any of the x[i].C1 is 0
		if res[i].IsZero() {
			return []E3{}, errors.New("invalid input")
		}
	}

	t := BatchInvertE3(res) // costs 1 inverse

	for i := 0; i < n; i++ {
		res[i].Add(&x[i].B0, &one).
			Mul(&res[i], &t[i])
	}

	return res, nil
}

// DecompressTorus GT/E6 a compressed element
// element must be in the cyclotomic subgroup
// "COMPRESSION IN FINITE FIELDS AND TORUS-BASED CRYPTOGRAPHY", K. RUBIN AND A. SILVERBERG
func (z *E3) DecompressTorus() E6 {

	var res, num, denum E6
	num.B0.Set(z)
	num.B1.SetOne()
	denum.B0.Set(z)
	denum.B1.SetOne().Neg(&denum.B1)
	res.Inverse(&denum).
		Mul(&res, &num)

	return res
}

// BatchDecompressTorus GT/E6 compressed elements
// using a batch inversion
func BatchDecompressTorus(x []E3) ([]E6, error) {

	n := len(x)
	if n == 0 {
		return []E6{}, errors.New("invalid input size")
	}

	res := make([]E6, n)
	num := make([]E6, n)
	denum := make([]E6, n)

	for i := 0; i < n; i++ {
		num[i].B0.Set(&x[i])
		num[i].B1.SetOne()
		denum[i].B0.Set(&x[i])
		denum[i].B1.SetOne().Neg(&denum[i].B1)
	}

	denum = BatchInvertE6(denum) // costs 1 inverse

	for i := 0; i < n; i++ {
		res[i].Mul(&num[i], &denum[i])
	}

	return res, nil
}
