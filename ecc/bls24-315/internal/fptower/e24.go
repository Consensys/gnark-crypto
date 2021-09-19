// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fptower

import (
	"errors"
	"math/big"
)

// E24 is a degree three finite field extension of fp8
type E24 struct {
	D0, D1, D2 E8
}

// Equal returns true if z equals x, fasle otherwise
func (z *E24) Equal(x *E24) bool {
	return z.D0.Equal(&x.D0) && z.D1.Equal(&x.D1) && z.D2.Equal(&x.D2)
}

// String puts E24 elmt in string form
func (z *E24) String() string {
	return (z.D0.String() + "+(" + z.D1.String() + ")*i+(" + z.D2.String() + ")*i**2")
}

// SetString sets a E24 elmt from stringf
func (z *E24) SetString(s0, s1, s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12, s13, s14, s15, s16, s17, s18, s19, s20, s21, s22, s23 string) *E24 {
	z.D0.SetString(s0, s1, s2, s3, s4, s5, s6, s7)
	z.D1.SetString(s8, s9, s10, s11, s12, s13, s14, s15)
	z.D2.SetString(s16, s17, s18, s19, s20, s21, s22, s23)
	return z
}

// Set Sets a E24 elmt form another E24 elmt
func (z *E24) Set(x *E24) *E24 {
	z.D0 = x.D0
	z.D1 = x.D1
	z.D2 = x.D2
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E24) SetOne() *E24 {
	*z = E24{}
	z.D0.C0.B0.A0.SetOne()
	return z
}

// SetRandom set z to a random elmt
func (z *E24) SetRandom() (*E24, error) {
	if _, err := z.D0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.D1.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.D2.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// ToMont converts to Mont form
func (z *E24) ToMont() *E24 {
	z.D0.ToMont()
	z.D1.ToMont()
	z.D2.ToMont()
	return z
}

// FromMont converts from Mont form
func (z *E24) FromMont() *E24 {
	z.D0.FromMont()
	z.D1.FromMont()
	z.D2.FromMont()
	return z
}

// Add adds two elements of E24
func (z *E24) Add(x, y *E24) *E24 {
	z.D0.Add(&x.D0, &y.D0)
	z.D1.Add(&x.D1, &y.D1)
	z.D2.Add(&x.D2, &y.D2)
	return z
}

// Neg negates the E24 number
func (z *E24) Neg(x *E24) *E24 {
	z.D0.Neg(&x.D0)
	z.D1.Neg(&x.D1)
	z.D2.Neg(&x.D2)
	return z
}

// Sub two elements of E24
func (z *E24) Sub(x, y *E24) *E24 {
	z.D0.Sub(&x.D0, &y.D0)
	z.D1.Sub(&x.D1, &y.D1)
	z.D2.Sub(&x.D2, &y.D2)
	return z
}

// Double doubles an element in E24
func (z *E24) Double(x *E24) *E24 {
	z.D0.Double(&x.D0)
	z.D1.Double(&x.D1)
	z.D2.Double(&x.D2)
	return z
}

// Mul sets z to the E24 product of x,y, returns z
func (z *E24) Mul(x, y *E24) *E24 {
	// Algorithm 13 from https://eprint.iacr.org/2010/354.pdf
	var t0, t1, t2, c0, c1, c2, tmp E8
	t0.Mul(&x.D0, &y.D0)
	t1.Mul(&x.D1, &y.D1)
	t2.Mul(&x.D2, &y.D2)

	c0.Add(&x.D1, &x.D2)
	tmp.Add(&y.D1, &y.D2)
	c0.Mul(&c0, &tmp).Sub(&c0, &t1).Sub(&c0, &t2).MulByNonResidue(&c0).Add(&c0, &t0)

	c1.Add(&x.D0, &x.D1)
	tmp.Add(&y.D0, &y.D1)
	c1.Mul(&c1, &tmp).Sub(&c1, &t0).Sub(&c1, &t1)
	tmp.MulByNonResidue(&t2)
	c1.Add(&c1, &tmp)

	tmp.Add(&x.D0, &x.D2)
	c2.Add(&y.D0, &y.D2).Mul(&c2, &tmp).Sub(&c2, &t0).Sub(&c2, &t2).Add(&c2, &t1)

	z.D0.Set(&c0)
	z.D1.Set(&c1)
	z.D2.Set(&c2)

	return z
}

// Square sets z to the E24 product of x,x, returns z
func (z *E24) Square(x *E24) *E24 {

	// Algorithm 16 from https://eprint.iacr.org/2010/354.pdf
	var c4, c5, c1, c2, c3, c0 E8
	c4.Mul(&x.D0, &x.D1).Double(&c4)
	c5.Square(&x.D2)
	c1.MulByNonResidue(&c5).Add(&c1, &c4)
	c2.Sub(&c4, &c5)
	c3.Square(&x.D0)
	c4.Sub(&x.D0, &x.D1).Add(&c4, &x.D2)
	c5.Mul(&x.D1, &x.D2).Double(&c5)
	c4.Square(&c4)
	c0.MulByNonResidue(&c5).Add(&c0, &c3)
	z.D2.Add(&c2, &c4).Add(&z.D2, &c5).Sub(&z.D2, &c3)
	z.D0.Set(&c0)
	z.D1.Set(&c1)

	return z
}

// Karabina's compressed cyclotomic square
// https://eprint.iacr.org/2010/542.pdf
// Th. 3.2 with minor modifications to fit our tower
func (z *E24) CyclotomicSquareCompressed(x *E24) *E24 {

	var t [7]E4

	// t0 = g4^2
	t[0].Square(&x.D2.C0)
	// t1 = g5^2
	t[1].Square(&x.D2.C1)
	// t5 = g4 + g5
	t[5].Add(&x.D2.C0, &x.D2.C1)
	// t2 = (g4 + g5)^2
	t[2].Square(&t[5])

	// t3 = g4^2 + g5^2
	t[3].Add(&t[0], &t[1])
	// t5 = 2 * g4 * g5
	t[5].Sub(&t[2], &t[3])

	// t6 = g3 + g2
	t[6].Add(&x.D1.C1, &x.D1.C0)
	// t3 = (g3 + g2)^2
	t[3].Square(&t[6])
	// t2 = g2^2
	t[2].Square(&x.D1.C0)

	// t6 = 2 * nr * g4 * g5
	t[6].MulByNonResidue(&t[5])
	// t5 = 4 * nr * g4 * g5 + 2 * g2
	t[5].Add(&t[6], &x.D1.C0).
		Double(&t[5])
	// z2 = 6 * nr * g4 * g5 + 2 * g2
	z.D1.C0.Add(&t[5], &t[6])

	// t4 = nr * g5^2
	t[4].MulByNonResidue(&t[1])
	// t5 = nr * g5^2 + g4^2
	t[5].Add(&t[0], &t[4])
	// t6 = nr * g5^2 + g1^2 - g3
	t[6].Sub(&t[5], &x.D1.C1)

	// t1 = g3^2
	t[1].Square(&x.D1.C1)

	// t6 = 2 * nr * g5^2 + 2 * g4^2 - 2*g3
	t[6].Double(&t[6])
	// z3 = 3 * nr * g5^2 + 3 * g4^2 - 2*g3
	z.D1.C1.Add(&t[6], &t[5])

	// t4 = nr * g3^2
	t[4].MulByNonResidue(&t[1])
	// t5 = g2^2 + nr * g3^2
	t[5].Add(&t[2], &t[4])
	// t6 = g2^2 + nr * g3^2 - g4
	t[6].Sub(&t[5], &x.D2.C0)
	// t6 = 2 * g2^2 + 2 * nr * g3^2 - 2 * g4
	t[6].Double(&t[6])
	// z4 = 3 * g2^2 + 3 * nr * g3^2 - 2 * g4
	z.D2.C0.Add(&t[6], &t[5])

	// t0 = g3^2 + g2^2
	t[0].Add(&t[2], &t[1])
	// t5 = 2 * g2 * g3
	t[5].Sub(&t[3], &t[0])
	// t6 = 2 * g2 * g3 + g5
	t[6].Add(&t[5], &x.D2.C1)
	// t6 = 4 * g2 * g3 + 2 * g5
	t[6].Double(&t[6])
	// z5 = 6 * g2 * g3 + 2 * g5
	z.D2.C1.Add(&t[5], &t[6])

	return z
}

// Decompress Karabina's cyclotomic square result
func (z *E24) Decompress(x *E24) *E24 {

	var t [3]E4
	var one E4
	one.SetOne()

	// t0 = g4^2
	t[0].Square(&x.D2.C0)
	// t1 = 3 * g4^2 - 2 * g3
	t[1].Sub(&t[0], &x.D1.C1).
		Double(&t[1]).
		Add(&t[1], &t[0])
		// t0 = E * g5^2 + t1
	t[2].Square(&x.D2.C1)
	t[0].MulByNonResidue(&t[2]).
		Add(&t[0], &t[1])
	// t1 = 1/(4 * g2)
	t[1].Double(&x.D1.C0).
		Double(&t[1]).
		Inverse(&t[1]) // costly
	// z1 = g4
	z.D0.C1.Mul(&t[0], &t[1])

	// t1 = g3 * g4
	t[1].Mul(&x.D1.C1, &x.D2.C0)
	// t2 = 2 * g1^2 - 3 * g3 * g4
	t[2].Square(&x.D0.C1).
		Sub(&t[2], &t[1]).
		Double(&t[2]).
		Sub(&t[2], &t[1])
	// t1 = g2 * g5
	t[1].Mul(&x.D1.C0, &x.D2.C1)
	// z0 = E * (2 * g1^2 + g2 * g5 - 3 * g3 * g4) + 1
	t[2].Add(&t[2], &t[1])
	z.D0.C0.MulByNonResidue(&t[2]).
		Add(&z.D0.C0, &one)

	z.D1.C0.Set(&x.D1.C0)
	z.D1.C1.Set(&x.D1.C1)
	z.D2.C0.Set(&x.D2.C0)
	z.D2.C1.Set(&x.D2.C1)

	return z
}

// BatchDecompress multiple Karabina's cyclotomic square results
func BatchDecompress(x []E24) []E24 {

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
		// t0 = g4^2
		t0[i].Square(&x[i].D2.C0)
		// t1 = 3 * g4^2 - 2 * g3
		t1[i].Sub(&t0[i], &x[i].D1.C1).
			Double(&t1[i]).
			Add(&t1[i], &t0[i])
			// t0 = E * g5^2 + t1
		t2[i].Square(&x[i].D2.C1)
		t0[i].MulByNonResidue(&t2[i]).
			Add(&t0[i], &t1[i])
		// t1 = 4 * g2
		t1[i].Double(&x[i].D1.C0).
			Double(&t1[i])
	}

	t1 = BatchInvert(t1) // costs 1 inverse

	for i := 0; i < n; i++ {
		// z4 = g1
		x[i].D0.C1.Mul(&t0[i], &t1[i])

		// t1 = g3 * g1
		t1[i].Mul(&x[i].D1.C1, &x[i].D2.C0)
		// t2 = 2 * g4^2 - 3 * g2 * g1
		t2[i].Square(&x[i].D0.C1).
			Sub(&t2[i], &t1[i]).
			Double(&t2[i]).
			Sub(&t2[i], &t1[i])

		// t1 = g2 * g5
		t1[i].Mul(&x[i].D1.C0, &x[i].D2.C1)
		// z0 = E * (2 * g1^2 + g2 * g5 - 3 * g3 * g4) + 1
		t2[i].Add(&t2[i], &t1[i])
		x[i].D0.C0.MulByNonResidue(&t2[i]).
			Add(&x[i].D0.C0, &one)
	}

	return x
}

// Granger-Scott's cyclotomic square
// https://eprint.iacr.org/2009/565.pdf, 3.2
func (z *E24) CyclotomicSquare(x *E24) *E24 {

	var A, B, C, D E8

	z.Set(x)
	A.Set(&z.D0)
	B.Set(&z.D2)
	C.Set(&z.D1)
	z.D0.Square(&z.D0)
	D.Double(&z.D0)
	z.D0.Add(&z.D0, &D)
	A.Conjugate(&A).Neg(&A)
	A.Double(&A)
	z.D0.Add(&z.D0, &A)
	B.Square(&B)
	B.MulByNonResidue(&B)
	D.Double(&B)
	B.Add(&B, &D)
	C.Square(&C)
	D.Double(&C)
	C.Add(&C, &D)
	z.D1.Conjugate(&z.D1)
	z.D1.Double(&z.D1)
	z.D2.Conjugate(&z.D2).Neg(&z.D2)
	z.D2.Double(&z.D2)
	z.D1.Add(&z.D1, &B)
	z.D2.Add(&z.D2, &C)

	return z
}

// Inverse an element in E24
func (z *E24) Inverse(x *E24) *E24 {
	// Algorithm 17 from https://eprint.iacr.org/2010/354.pdf
	// step 9 is wrong in the paper it's t1-t4
	var t0, t1, t2, t3, t4, t5, t6, c0, c1, c2, d1, d2 E8
	t0.Square(&x.D0)
	t1.Square(&x.D1)
	t2.Square(&x.D2)
	t3.Mul(&x.D0, &x.D1)
	t4.Mul(&x.D0, &x.D2)
	t5.Mul(&x.D1, &x.D2)
	c0.MulByNonResidue(&t5).Sub(&t0, &c0)
	c1.MulByNonResidue(&t2).Sub(&c1, &t3)
	c2.Sub(&t1, &t4)
	t6.Mul(&x.D0, &c0)
	d1.Mul(&x.D2, &c1)
	d2.Mul(&x.D1, &c2)
	d1.Add(&d1, &d2).MulByNonResidue(&d1)
	t6.Add(&t6, &d1)
	t6.Inverse(&t6)
	z.D0.Mul(&c0, &t6)
	z.D1.Mul(&c1, &t6)
	z.D2.Mul(&c2, &t6)

	return z
}

// Exp sets z=x**e and returns it
func (z *E24) Exp(x *E24, e big.Int) *E24 {
	var res E24
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

// InverseUnitary inverse a unitary element
func (z *E24) InverseUnitary(x *E24) *E24 {
	return z.Conjugate(x)
}

// Conjugate set z to x conjugated and return z
func (z *E24) Conjugate(x *E24) *E24 {
	z.D0.Conjugate(&x.D0)
	z.D1.Conjugate(&x.D1).Neg(&z.D1)
	z.D2.Conjugate(&x.D2)
	return z
}

// SizeOfGT represents the size in bytes that a GT element need in binary form
const SizeOfGT = sizeOfFp * 24
const sizeOfFp = 40

// Bytes returns the regular (non montgomery) value
// of z as a big-endian byte array.
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

	buf = z.D2.C0.B0.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D2.C0.B0.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D2.C0.B1.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D2.C0.B1.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D2.C1.B0.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D2.C1.B0.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D2.C1.B1.A0.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

	buf = z.D2.C1.B1.A1.Bytes()
	copy(r[offset:offset+sizeOfFp], buf[:])
	offset += sizeOfFp

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
	z.D2.C0.B0.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D2.C0.B0.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D2.C0.B1.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D2.C0.B1.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D2.C1.B0.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D2.C1.B0.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D2.C1.B1.A0.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp
	z.D2.C1.B1.A1.SetBytes(e[offset : offset+sizeOfFp])
	offset += sizeOfFp

	return nil
}

// Marshal converts z to a byte slice
func (z *E24) Marshal() []byte {
	b := z.Bytes()
	return b[:]
}

// Unmarshal is an allias to SetBytes()
func (z *E24) Unmarshal(buf []byte) error {
	return z.SetBytes(buf)
}

// IsInSubGroup ensures GT/E24 is in correct sugroup
func (z *E24) IsInSubGroup() bool {
	var a, b E24

	// check z^(Phi_k(p)) == 1
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
