// Copyright 2020 ConsenSys AG
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
	"github.com/consensys/gnark-crypto/ecc/bls12-39/fp"
)

// Frobenius set z to Frobenius(x), return z
func (z *E12) Frobenius(x *E12) *E12 {
	// Algorithm 28 from https://eprint.iacr.org/2010/354.pdf (beware typos!)
	var t [6]E2

	// Frobenius acts on fp2 by conjugation
	t[0].Conjugate(&x.C0.B0)
	t[1].Conjugate(&x.C0.B1)
	t[2].Conjugate(&x.C0.B2)
	t[3].Conjugate(&x.C1.B0)
	t[4].Conjugate(&x.C1.B1)
	t[5].Conjugate(&x.C1.B2)

	t[1].MulByNonResidue1Power2(&t[1])
	t[2].MulByNonResidue1Power4(&t[2])
	t[3].MulByNonResidue1Power1(&t[3])
	t[4].MulByNonResidue1Power3(&t[4])
	t[5].MulByNonResidue1Power5(&t[5])

	z.C0.B0 = t[0]
	z.C0.B1 = t[1]
	z.C0.B2 = t[2]
	z.C1.B0 = t[3]
	z.C1.B1 = t[4]
	z.C1.B2 = t[5]

	return z
}

// FrobeniusSquare set z to Frobenius^2(x), and return z
func (z *E12) FrobeniusSquare(x *E12) *E12 {
	// Algorithm 29 from https://eprint.iacr.org/2010/354.pdf (beware typos!)
	var t [6]E2

	t[1].MulByNonResidue2Power2(&x.C0.B1)
	t[2].MulByNonResidue2Power4(&x.C0.B2)
	t[3].MulByNonResidue2Power1(&x.C1.B0)
	t[4].MulByNonResidue2Power3(&x.C1.B1)
	t[5].MulByNonResidue2Power5(&x.C1.B2)

	z.C0.B0 = x.C0.B0
	z.C0.B1 = t[1]
	z.C0.B2 = t[2]
	z.C1.B0 = t[3]
	z.C1.B1 = t[4]
	z.C1.B2 = t[5]

	return z
}

// FrobeniusCube set z to Frobenius^3(x), return z
func (z *E12) FrobeniusCube(x *E12) *E12 {
	// Algorithm 30 from https://eprint.iacr.org/2010/354.pdf (beware typos!)
	var t [6]E2

	// Frobenius^3 acts on fp2 by conjugation
	t[0].Conjugate(&x.C0.B0)
	t[1].Conjugate(&x.C0.B1)
	t[2].Conjugate(&x.C0.B2)
	t[3].Conjugate(&x.C1.B0)
	t[4].Conjugate(&x.C1.B1)
	t[5].Conjugate(&x.C1.B2)

	t[1].MulByNonResidue3Power2(&t[1])
	t[2].MulByNonResidue3Power4(&t[2])
	t[3].MulByNonResidue3Power1(&t[3])
	t[4].MulByNonResidue3Power3(&t[4])
	t[5].MulByNonResidue3Power5(&t[5])

	z.C0.B0 = t[0]
	z.C0.B1 = t[1]
	z.C0.B2 = t[2]
	z.C1.B0 = t[3]
	z.C1.B1 = t[4]
	z.C1.B2 = t[5]

	return z
}

// MulByNonResidue1Power1 set z=x*(1,1)^(1*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power1(x *E2) *E2 {
	// 171574396630*u + 199624070845
	var b E2
	b.A0 = fp.Element{186886073064}
	b.A1 = fp.Element{75254894234}
	z.Mul(x, &b)
	return z
}

// MulByNonResidue1Power2 set z=x*(1,1)^(2*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power2(x *E2) *E2 {
	// 182009477101*u + 159768345029
	var b E2
	b.A0 = fp.Element{56732333561}
	b.A1 = fp.Element{75375024703}
	z.Mul(x, &b)
	return z
}

// MulByNonResidue1Power3 set z=x*(1,1)^(3*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power3(x *E2) *E2 {
	// 228828781692*u + 293515655025
	var b E2
	b.A0 = fp.Element{320901648360}
	b.A1 = fp.Element{219700117247}
	z.Mul(x, &b)
	return z
}

// MulByNonResidue1Power4 set z=x*(1,1)^(4*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power4(x *E2) *E2 {
	// 66088819674*u + 34127110868
	var b E2
	b.A0 = fp.Element{228884275100}
	b.A1 = fp.Element{278145123361}
	z.Mul(x, &b)
	return z
}

// MulByNonResidue1Power5 set z=x*(1,1)^(5*(p^1-1)/6) and return z
func (z *E2) MulByNonResidue1Power5(x *E2) *E2 {
	// 214835958777*u + 36448585468
	var b E2
	b.A0 = fp.Element{195402511428}
	b.A1 = fp.Element{227508494027}
	z.Mul(x, &b)
	return z
}

// MulByNonResidue2Power1 set z=x*(1,1)^(1*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power1(x *E2) *E2 {
	// 9702999902
	b := fp.Element{93717443168}
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)
	return z
}

// MulByNonResidue2Power2 set z=x*(1,1)^(2*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power2(x *E2) *E2 {
	// 9702999901
	b := fp.Element{268249031722}
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)

	return z
}

// MulByNonResidue2Power3 set z=x*(1,1)^(3*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power3(x *E2) *E2 {
	// 326667333366
	b := fp.Element{174531588554}
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)

	return z
}

// MulByNonResidue2Power4 set z=x*(1,1)^(4*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power4(x *E2) *E2 {
	// 316964333465
	b := fp.Element{232949890199}
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)

	return z
}

// MulByNonResidue2Power5 set z=x*(1,1)^(5*(p^2-1)/6) and return z
func (z *E2) MulByNonResidue2Power5(x *E2) *E2 {
	// 316964333466
	b := fp.Element{58418301645}
	z.A0.Mul(&x.A0, &b)
	z.A1.Mul(&x.A1, &b)

	return z
}

// MulByNonResidue3Power1 set z=x*(1,1)^(1*(p^3-1)/6) and return z
func (z *E2) MulByNonResidue3Power1(x *E2) *E2 {
	// 32121226975*u + 263694451378
	var b E2
	b.A0.SetString("263694451378")
	b.A1.SetString("32121226975")
	z.Mul(x, &b)
	return z
}

// MulByNonResidue3Power2 set z=x*(1,1)^(2*(p^3-1)/6) and return z
func (z *E2) MulByNonResidue3Power2(x *E2) *E2 {
	// 305020439417*u + 106752335729
	var b E2
	b.A0.SetString("106752335729")
	b.A1.SetString("305020439417")
	z.Mul(x, &b)
	return z
}

// MulByNonResidue3Power3 set z=x*(1,1)^(3*(p^3-1)/6) and return z
func (z *E2) MulByNonResidue3Power3(x *E2) *E2 {
	// 97838551675*u + 33151678342
	var b E2
	b.A0.SetString("33151678342")
	b.A1.SetString("97838551675")
	z.Mul(x, &b)
	return z
}

// MulByNonResidue3Power4 set z=x*(1,1)^(4*(p^3-1)/6) and return z
func (z *E2) MulByNonResidue3Power4(x *E2) *E2 {
	// 63458547829*u + 178103343759
	var b E2
	b.A0.SetString("178103343759")
	b.A1.SetString("63458547829")
	z.Mul(x, &b)
	return z
}

// MulByNonResidue3Power5 set z=x*(1,1)^(5*(p^3-1)/6) and return z
func (z *E2) MulByNonResidue3Power5(x *E2) *E2 {
	// 325397761406*u + 297085250314
	var b E2
	b.A0.SetString("297085250314")
	b.A1.SetString("325397761406")
	z.Mul(x, &b)
	return z
}
