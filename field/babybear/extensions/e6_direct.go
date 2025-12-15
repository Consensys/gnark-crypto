// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import (
	fr "github.com/consensys/gnark-crypto/field/babybear"
)

// E6D is a degree 6 finite field extension of fr
type E6D struct {
	A0, A1, A2, A3, A4, A5 fr.Element
}

// Equal returns true if z equals x, false otherwise
func (z *E6D) Equal(x *E6D) bool {
	return z.A0.Equal(&x.A0) &&
		z.A1.Equal(&x.A1) &&
		z.A2.Equal(&x.A2) &&
		z.A3.Equal(&x.A3) &&
		z.A4.Equal(&x.A4) &&
		z.A5.Equal(&x.A5)
}

// String puts E6D elmt in string form
func (z *E6D) String() string {
	return (z.A0.String() + "+(" + z.A1.String() + ")*u+(" + z.A2.String() + ")*u**2+(" + z.A3.String() + ")*u**3+(" + z.A4.String() + ")*u**4+(" + z.A5.String() + ")*u**5")
}

// SetString sets a E6D elmt from string
func (z *E6D) SetString(s1, s2, s3, s4, s5, s6 string) *E6D {
	z.A0.SetString(s1)
	z.A1.SetString(s2)
	z.A2.SetString(s3)
	z.A3.SetString(s4)
	z.A4.SetString(s5)
	z.A5.SetString(s6)
	return z
}

// Set copies x into z and returns z
func (z *E6D) Set(x *E6D) *E6D {
	*z = *x
	return z
}

// SetOne sets z to 1 and returns z
func (z *E6D) SetOne() *E6D {
	z.A0.SetOne()
	z.A1.SetZero()
	z.A2.SetZero()
	z.A3.SetZero()
	z.A4.SetZero()
	z.A5.SetZero()
	return z
}

// Add sets z=x+y in E6D and returns z
func (z *E6D) Add(x, y *E6D) *E6D {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	z.A2.Add(&x.A2, &y.A2)
	z.A3.Add(&x.A3, &y.A3)
	z.A4.Add(&x.A4, &y.A4)
	z.A5.Add(&x.A5, &y.A5)
	return z
}

// Sub sets z to x-y and returns z
func (z *E6D) Sub(x, y *E6D) *E6D {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
	z.A2.Sub(&x.A2, &y.A2)
	z.A3.Sub(&x.A3, &y.A3)
	z.A4.Sub(&x.A4, &y.A4)
	z.A5.Sub(&x.A5, &y.A5)
	return z
}

// Double sets z=2*x and returns z
func (z *E6D) Double(x *E6D) *E6D {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
	z.A2.Double(&x.A2)
	z.A3.Double(&x.A3)
	z.A4.Double(&x.A4)
	z.A5.Double(&x.A5)
	return z
}

// SetRandom used only in tests
func (z *E6D) SetRandom() (*E6D, error) {
	if _, err := z.A0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A1.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A2.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A3.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A4.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A5.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// MustSetRandom sets z to a random value.
// It panics if reading from crypto/rand fails.
func (z *E6D) MustSetRandom() *E6D {
	if _, err := z.SetRandom(); err != nil {
		panic(err)
	}
	return z
}

// IsZero returns true if z is zero, false otherwise
func (z *E6D) IsZero() bool {
	return z.A0.IsZero() && z.A1.IsZero() && z.A2.IsZero() && z.A3.IsZero() && z.A4.IsZero() && z.A5.IsZero()
}

// IsOne returns true if z is one, false otherwise
func (z *E6D) IsOne() bool {
	return z.A0.IsOne() && z.A1.IsZero() && z.A2.IsZero() && z.A3.IsZero() && z.A4.IsZero() && z.A5.IsZero()
}

// Mul sets z=x*y in E6D and returns z
func (z *E6D) Mul(x, y *E6D) *E6D {
	return z.mulMontgomery6(x, y)
}

func (z *E6D) mulMontgomery6(a, b *E6D) *E6D {
	// Ref.: Peter L. Montgomery. Five, six, and seven-term Karatsuba-like formulae. IEEE
	// Transactions on Computers, 54(3):362–369, 2005.
	//
	// The product of two degree-5 polynomials a(X) and b(X):
	// a(X) = a0 + a1*X + a2*X^2 + a3*X^3 + a4*X^4 + a5*X^5
	// b(X) = b0 + b1*X + b2*X^2 + b3*X^3 + b4*X^4 + b5*X^5
	//
	// The result c(X) = a(X) * b(X) according to the ref. is:
	//
	// c(X) =
	//   	  (a0 + a1 + a2 + a3 + a4 + a5)(b0 + b1 + b2 + b3 + b4 + b5) * C
	// 		+ (a1 + a2 + a4 + a5)(b1 + b2 + b4 + b5) * (-C + X^6)
	// 		+ (a0 + a1 + a3 + a4)(b0 + b1 + b3 + b4) * (-C + X^4)
	// 		+ (a0 - a2 - a3 + a5)(b0 - b2 - b3 + b5) * (C - X^7 + X^6 - X^5 + X^4 - X^3)
	// 		+ (a0 - a2 - a5)(b0 - b2 - b5) * (C - X^5 + X^4 - X^3)
	// 		+ (a0 + a3 - a5)(b0 + b3 - b5) * (C - X^7 + X^6 - X^5)
	// 		+ (a0 + a1 + a2)(b0 + b1 + b2) * (C - X^7 + 6*X^6 - 2*X^5 + 2*X^4 - 2*X^3 + X^2)
	// 		+ (a3 + a4 + a5)(b3 + b4 + b5) * (C + X^8 - 2*X^7 + 2*X^6 - 2*X^5 + X^4 - X^3)
	// 		+ (a2 + a3)(b2 + b3) * (-2*C + X^7 - X^6 + 2*X^5 - X^4 + X^3)
	// 		+ (a1 - a4)(b1 - b4) * (-C + X^4 - X^5 + X^6)
	// 		+ (a1 + a2)(b1 + b2) * (-C + X^7 - 2*X^6 + 2*X^5 - 2*X^4 + 3*X^3 - X^2)
	// 		+ (a3 + a4)(b3 + b4) * (-C - X^8 + 3*X^7 - 2*X^6 + 2*X^5 - 2*X^4 + X^3)
	// 		+ (a0 + a1)(b0 + b1) * (-C + X^7 - X^6 + 2*X^5 - 3*X^4 + 2*X^3 - X^2 + X)
	// 		+ (a4 + a5)(b4 + b5) * (-C + X^9 - X^8 + 2*X^7 - 3*X^6 + 2*X^5 - X^4 + X^3)
	// 		+ a0*b0 * (3*C + 2*X^7 - 2*X^6 - 3*X^5 + 2*X^4 + 2*X^3 - X + 1)
	// 		+ a1*b1 * (3*C - X^7 - X^5 - X^4 - 3*X^3 + 2*X^2 - X)
	// 		+ a4*b4 * (3*C - X^9 + 2*X^8 - 3*X^7 + X^6 - X^5 - X^3)
	// 		+ a5*b5 * (-3*C + X^10 - X^9 + 2*X^7 - 2*X^6 + 3*X^5 - 2*X^4 + 2*X^3)
	//
	// We fix the parameter C to X^6 so that the second term disappears. We then compute the interpolation points
	// vi = a(Xi)*b(Xi) at Xi={0, ±1, ±2, ±3, ±4, 5,∞}:
	//
	//		v0 = (a0 + a1 + a2 + a3 + a4 + a5)(b0 + b1 + b2 + b3 + b4 + b5)
	//		v2 = (a0 + a1 + a3 + a4)(b0 + b1 + b3 + b4)
	//		v3 = (a0 − a2 − a3 + a5)(b0 − b2 − b3 + b5)
	//		v4 = (a0 − a2 − a5)(b0 − b2 − b5)
	//		v5 = (a0 + a3 − a5)(b0 + b3 − b5)
	//		v6 = (a0 + a1 + a2)(b0 + b1 + b2)
	//		v7 = (a3 + a4 + a5)(b3 + b4 + b5)
	//		v8 = (a2 + a3)(b2 + b3)
	//		v9 = (a1 − a4)(b1 − b4)
	//		v10 = (a1 + a2)(b1 + b2)
	//		v11 = (a3 + a4)(b3 + b4)
	//		v12 = (a0 + a1)(b0 + b1)
	//		v13 = (a4 + a5)(b4 + b5)
	//		v14 = a0b0
	//		v15 = a1b1
	//		v16 = a4b4
	//		v17 = a5b5
	//
	// 		We do this optimally in 17 multiplications and 30 additions/subtractions in Fr.

	var v [18]fr.Element
	var t [14]fr.Element

	// -------------------------------------------------------------------------
	// Phase 1: Evaluation of a
	// -------------------------------------------------------------------------
	v[12].Add(&a.A0, &a.A1)
	v[13].Add(&a.A4, &a.A5)
	v[10].Add(&a.A1, &a.A2)
	v[11].Add(&a.A3, &a.A4)
	v[8].Add(&a.A2, &a.A3)
	v[9].Sub(&a.A1, &a.A4)

	v[6].Add(&v[12], &a.A2)
	v[7].Add(&v[11], &a.A5)
	v[2].Add(&v[12], &v[11])
	v[0].Add(&v[6], &v[7])

	t[0].Sub(&a.A0, &a.A5)
	v[4].Sub(&t[0], &a.A2)
	v[5].Add(&v[4], &v[8])
	t[0].Add(&a.A0, &a.A5)
	v[3].Sub(&t[0], &v[8])

	// -------------------------------------------------------------------------
	// Phase 2: Evaluation of b
	// -------------------------------------------------------------------------
	t[12].Add(&b.A0, &b.A1)
	t[13].Add(&b.A4, &b.A5)
	t[10].Add(&b.A1, &b.A2)
	t[11].Add(&b.A3, &b.A4)
	t[8].Add(&b.A2, &b.A3)
	t[9].Sub(&b.A1, &b.A4)

	t[6].Add(&t[12], &b.A2)
	t[7].Add(&t[11], &b.A5)
	t[2].Add(&t[12], &t[11])
	t[0].Add(&t[6], &t[7])

	t[1].Sub(&b.A0, &b.A5)
	t[4].Sub(&t[1], &b.A2)
	t[5].Add(&t[4], &t[8])
	t[1].Add(&b.A0, &b.A5)
	t[3].Sub(&t[1], &t[8])

	// -------------------------------------------------------------------------
	// Phase 3: Pointwise Multiplication
	// -------------------------------------------------------------------------
	v[0].Mul(&v[0], &t[0])
	v[2].Mul(&v[2], &t[2])
	v[3].Mul(&v[3], &t[3])
	v[4].Mul(&v[4], &t[4])
	v[5].Mul(&v[5], &t[5])
	v[6].Mul(&v[6], &t[6])
	v[7].Mul(&v[7], &t[7])
	v[8].Mul(&v[8], &t[8])
	v[9].Mul(&v[9], &t[9])
	v[10].Mul(&v[10], &t[10])
	v[11].Mul(&v[11], &t[11])
	v[12].Mul(&v[12], &t[12])
	v[13].Mul(&v[13], &t[13])
	v[14].Mul(&a.A0, &b.A0)
	v[15].Mul(&a.A1, &b.A1)
	v[16].Mul(&a.A4, &b.A4)
	v[17].Mul(&a.A5, &b.A5)

	// We then we re-arrange the terms in function of the degree of X and use
	// the fact that X^6=2(X^3+5), because we construct 𝔽r⁶[w] as 𝔽r/w⁶-2w³-10. The
	// resulting coefficients c0,c1,c3,c4 and c5 are:
	//
	// c5 = -(v3+v4+v5+2v6) + 2(v8+v10+v12) - v9 + 3v14 - v15 + 3v16 + 3v17
	//
	// c4 =  v2 - v3 + v4 - 2v5 - 3v7 + v8 + v9 + 4v11 - v12 + 3v13
	//       + 2v14 - v15 - 6v16 + 16v17
	//
	// c3 =  2(v0 - v2) + 3v3 + v4 + 4v5 + 2v6 + 5v7
	//       - 5v8 - 3v10 - 5v11 - 2v12 + 7v13
	//       - 8v14 + 3v15 - 7v16 - 22v17
	//
	// c2 =  v6 + 10v7 - v10 - 10v11 - v12 - 10v13 + 2v15 + 20v16
	//
	// c1 =  -10v3 -10v5 -10v6 -20v7 +10v8 +10v10 +30v11 +11v12 +20v13
	//       +19v14 -11v15 -30v16 +40v17
	//
	// c0 =  10v0 -10v2 +20v3 +10v4 +20v5 +20v6 +30v7
	//       -30v8 -30v10 -30v11 -20v12 -20v13
	//       -49v14 +30v15 +20v16 -70v17

	// -------------------------------------------------------------------------
	// Phase 4: Reconstruction (optimized adds/subs)
	// -------------------------------------------------------------------------

	t[0].Add(&v[14], &v[14]) // 2v14
	t[1].Add(&t[0], &v[14])  // 3v14

	t[2].Add(&v[16], &v[16]) // 2v16
	t[3].Add(&t[2], &v[16])  // 3v16
	t[4].Add(&t[3], &t[3])   // 6v16

	t[5].Add(&v[17], &v[17]) // 2v17
	t[7].Add(&t[5], &t[5])   // 4v17
	t[8].Add(&t[7], &t[7])   // 8v17
	t[9].Add(&t[8], &t[8])   // 16v17
	t[10].Add(&t[9], &t[9])  // 32v17
	t[11].Add(&t[10], &t[8]) // 40v17
	t[6].Add(&t[5], &v[17])  // 3v17

	// -------------------------------------------------------------------------
	// c5 = -(v3+v4+v5+2v6) + 2(v8+v10+v12) - v9 + 3v14 - v15 + 3v16 + 3v17
	// -------------------------------------------------------------------------
	t[12].Add(&v[3], &v[4])
	t[12].Add(&t[12], &v[5])
	t[13].Add(&v[6], &v[6])   // 2v6
	t[12].Add(&t[12], &t[13]) // v3+v4+v5+2v6

	t[13].Add(&v[8], &v[10])
	t[13].Add(&t[13], &v[12])
	t[13].Add(&t[13], &t[13]) // 2(v8+v10+v12)

	z.A5.Sub(&t[13], &t[12])
	z.A5.Sub(&z.A5, &v[9])
	z.A5.Add(&z.A5, &t[1])  // +3v14
	z.A5.Sub(&z.A5, &v[15]) // -v15
	z.A5.Add(&z.A5, &t[3])  // +3v16
	z.A5.Add(&z.A5, &t[6])  // +3v17

	// -------------------------------------------------------------------------
	// c4 = (v2 - v3 + v4) - 2v5 - 3v7 + v8 + v9 + 4v11 - v12 + 3v13
	//   - 2v14 - v15 - 6v16 + 16v17
	//
	// -------------------------------------------------------------------------
	z.A4.Sub(&v[2], &v[3])
	z.A4.Add(&z.A4, &v[4])

	t[12].Add(&v[5], &v[5]) // 2v5
	z.A4.Sub(&z.A4, &t[12])

	t[13].Add(&v[7], &v[7])
	t[13].Add(&t[13], &v[7]) // 3v7
	z.A4.Sub(&z.A4, &t[13])

	z.A4.Add(&z.A4, &v[8])
	z.A4.Add(&z.A4, &v[9])

	t[12].Add(&v[11], &v[11])
	t[12].Add(&t[12], &t[12]) // 4v11
	z.A4.Add(&z.A4, &t[12])

	z.A4.Sub(&z.A4, &v[12])

	t[13].Add(&v[13], &v[13])
	t[13].Add(&t[13], &v[13]) // 3v13
	z.A4.Add(&z.A4, &t[13])

	z.A4.Add(&z.A4, &t[0])  // +2v14
	z.A4.Sub(&z.A4, &v[15]) // -v15
	z.A4.Sub(&z.A4, &t[4])  // -6v16
	z.A4.Add(&z.A4, &t[9])  // +16v17

	// -------------------------------------------------------------------------
	// c3 = 2(v0 - v2) + 3v3 + v4 + 4v5 + 2v6 + 5v7
	//   - 5v8 - 3v10 - 5v11 - 2v12 + 7v13
	//   - 8v14 + 3v15 - 7v16 - 22v17
	//
	// -------------------------------------------------------------------------
	t[12].Sub(&v[0], &v[2])
	t[12].Add(&t[12], &t[12]) // 2(v0-v2)
	z.A3.Set(&t[12])

	t[12].Add(&v[3], &v[3])
	t[12].Add(&t[12], &v[3]) // 3v3
	z.A3.Add(&z.A3, &t[12])
	z.A3.Add(&z.A3, &v[4])

	t[12].Add(&v[5], &v[5])
	t[12].Add(&t[12], &t[12]) // 4v5
	z.A3.Add(&z.A3, &t[12])

	t[12].Add(&v[6], &v[6]) // 2v6
	z.A3.Add(&z.A3, &t[12])

	t[12].Add(&v[7], &v[7])   // 2v7
	t[13].Add(&t[12], &t[12]) // 4v7
	t[13].Add(&t[13], &v[7])  // 5v7
	z.A3.Add(&z.A3, &t[13])

	t[12].Add(&v[8], &v[8])   // 2v8
	t[13].Add(&t[12], &t[12]) // 4v8
	t[13].Add(&t[13], &v[8])  // 5v8
	z.A3.Sub(&z.A3, &t[13])

	t[12].Add(&v[10], &v[10])
	t[12].Add(&t[12], &v[10]) // 3v10
	z.A3.Sub(&z.A3, &t[12])

	t[12].Add(&v[11], &v[11]) // 2v11
	t[13].Add(&t[12], &t[12]) // 4v11
	t[13].Add(&t[13], &v[11]) // 5v11
	z.A3.Sub(&z.A3, &t[13])

	t[12].Add(&v[12], &v[12]) // 2v12
	z.A3.Sub(&z.A3, &t[12])

	t[12].Add(&v[13], &v[13]) // 2v13
	t[13].Add(&t[12], &t[12]) // 4v13
	t[12].Add(&t[12], &t[13]) // 6v13
	t[12].Add(&t[12], &v[13]) // 7v13
	z.A3.Add(&z.A3, &t[12])

	t[12].Add(&t[0], &t[0])   // 4v14
	t[12].Add(&t[12], &t[12]) // 8v14
	z.A3.Sub(&z.A3, &t[12])

	t[12].Add(&v[15], &v[15])
	t[12].Add(&t[12], &v[15]) // 3v15
	z.A3.Add(&z.A3, &t[12])

	t[12].Add(&t[2], &t[2])   // 4v16
	t[12].Add(&t[12], &t[12]) // 8v16
	t[12].Sub(&t[12], &v[16]) // 7v16
	z.A3.Sub(&z.A3, &t[12])

	t[12].Add(&t[9], &t[7])  // 16v17 + 4v17
	t[12].Add(&t[12], &t[5]) // +2v17 = 22v17
	z.A3.Sub(&z.A3, &t[12])

	// -------------------------------------------------------------------------
	// c2 = v6 + 10v7 - v10 - 10v11 - v12 - 10v13 + 2v15 + 20v16
	// -------------------------------------------------------------------------
	z.A2.Set(&v[6])
	z.A2.Sub(&z.A2, &v[10])
	z.A2.Sub(&z.A2, &v[12])

	t[12].Add(&v[15], &v[15]) // 2v15
	z.A2.Add(&z.A2, &t[12])

	// +10v7  (8+2)
	t[12].Add(&v[7], &v[7])   // 2v7
	t[13].Add(&t[12], &t[12]) // 4v7
	t[13].Add(&t[13], &t[13]) // 8v7
	t[13].Add(&t[13], &t[12]) // 10v7
	z.A2.Add(&z.A2, &t[13])

	// -10v11
	t[12].Add(&v[11], &v[11]) // 2v11
	t[13].Add(&t[12], &t[12]) // 4v11
	t[13].Add(&t[13], &t[13]) // 8v11
	t[13].Add(&t[13], &t[12]) // 10v11
	z.A2.Sub(&z.A2, &t[13])

	// -10v13
	t[12].Add(&v[13], &v[13]) // 2v13
	t[13].Add(&t[12], &t[12]) // 4v13
	t[13].Add(&t[13], &t[13]) // 8v13
	t[13].Add(&t[13], &t[12]) // 10v13
	z.A2.Sub(&z.A2, &t[13])

	// +20v16  (10 then double)
	t[12].Set(&t[2])          // 2v16
	t[13].Add(&t[12], &t[12]) // 4v16
	t[13].Add(&t[13], &t[13]) // 8v16
	t[13].Add(&t[13], &t[12]) // 10v16
	t[13].Add(&t[13], &t[13]) // 20v16
	z.A2.Add(&z.A2, &t[13])

	// -------------------------------------------------------------------------
	// c1 = -10v3 -10v5 -10v6 -20v7 +10v8 +10v10 +30v11 +11v12 +20v13
	//
	//	+19v14 -11v15 -30v16 +40v17
	//
	// -------------------------------------------------------------------------
	z.A1.Set(&t[11]) // 40v17

	// +10v10
	t[12].Add(&v[10], &v[10]) // 2
	t[13].Add(&t[12], &t[12]) // 4
	t[13].Add(&t[13], &t[13]) // 8
	t[13].Add(&t[13], &t[12]) // 10
	z.A1.Add(&z.A1, &t[13])

	// +30v11 = 10v11 + 20v11
	t[12].Add(&v[11], &v[11]) // 2
	t[13].Add(&t[12], &t[12]) // 4
	t[13].Add(&t[13], &t[13]) // 8
	t[13].Add(&t[13], &t[12]) // 10
	t[12].Add(&t[13], &t[13]) // 20
	t[12].Add(&t[12], &t[13]) // 30
	z.A1.Add(&z.A1, &t[12])

	// +11v12 = 10v12 + v12
	t[12].Add(&v[12], &v[12]) // 2
	t[13].Add(&t[12], &t[12]) // 4
	t[13].Add(&t[13], &t[13]) // 8
	t[13].Add(&t[13], &t[12]) // 10
	t[13].Add(&t[13], &v[12]) // 11
	z.A1.Add(&z.A1, &t[13])

	// +20v13
	t[12].Add(&v[13], &v[13]) // 2
	t[13].Add(&t[12], &t[12]) // 4
	t[13].Add(&t[13], &t[13]) // 8
	t[13].Add(&t[13], &t[12]) // 10
	t[13].Add(&t[13], &t[13]) // 20
	z.A1.Add(&z.A1, &t[13])

	// +19v14 = 16+2+1
	t[12].Add(&t[0], &t[0])   // 4v14
	t[12].Add(&t[12], &t[12]) // 8v14
	t[12].Add(&t[12], &t[12]) // 16v14
	t[12].Add(&t[12], &t[0])  // 18v14
	t[12].Add(&t[12], &v[14]) // 19v14
	z.A1.Add(&z.A1, &t[12])

	// -11v15
	t[12].Add(&v[15], &v[15]) // 2
	t[13].Add(&t[12], &t[12]) // 4
	t[13].Add(&t[13], &t[13]) // 8
	t[13].Add(&t[13], &t[12]) // 10
	t[13].Add(&t[13], &v[15]) // 11
	z.A1.Sub(&z.A1, &t[13])

	// -30v16
	t[12].Set(&t[2])          // 2
	t[13].Add(&t[12], &t[12]) // 4
	t[13].Add(&t[13], &t[13]) // 8
	t[13].Add(&t[13], &t[12]) // 10
	t[12].Add(&t[13], &t[13]) // 20
	t[12].Add(&t[12], &t[13]) // 30
	z.A1.Sub(&z.A1, &t[12])

	// -10v3
	t[12].Add(&v[3], &v[3])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12]) // 10
	z.A1.Sub(&z.A1, &t[13])

	// -10v5
	t[12].Add(&v[5], &v[5])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12])
	z.A1.Sub(&z.A1, &t[13])

	// -10v6
	t[12].Add(&v[6], &v[6])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12])
	z.A1.Sub(&z.A1, &t[13])

	// -20v7
	t[12].Add(&v[7], &v[7])   // 2
	t[13].Add(&t[12], &t[12]) // 4
	t[13].Add(&t[13], &t[13]) // 8
	t[13].Add(&t[13], &t[12]) // 10
	t[13].Add(&t[13], &t[13]) // 20
	z.A1.Sub(&z.A1, &t[13])

	// +10v8
	t[12].Add(&v[8], &v[8])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12])
	z.A1.Add(&z.A1, &t[13])

	// -------------------------------------------------------------------------
	// c0 = 10v0 -10v2 +20v3 +10v4 +20v5 +20v6 +30v7
	//      -30v8 -30v10 -30v11 -20v12 -20v13
	//      -49v14 +30v15 +20v16 -70v17
	// -------------------------------------------------------------------------

	// z.A0 = 10*(v0 - v2)
	t[12].Sub(&v[0], &v[2])   // s = v0 - v2
	t[13].Add(&t[12], &t[12]) // 2s
	t[11].Add(&t[13], &t[13]) // 4s
	t[11].Add(&t[11], &t[11]) // 8s
	t[11].Add(&t[11], &t[13]) // 10s
	z.A0.Set(&t[11])

	// +20v3
	t[12].Add(&v[3], &v[3])   // 2
	t[13].Add(&t[12], &t[12]) // 4
	t[13].Add(&t[13], &t[13]) // 8
	t[13].Add(&t[13], &t[12]) // 10
	t[13].Add(&t[13], &t[13]) // 20
	z.A0.Add(&z.A0, &t[13])

	// +10v4
	t[12].Add(&v[4], &v[4])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12]) // 10
	z.A0.Add(&z.A0, &t[13])

	// +20v5
	t[12].Add(&v[5], &v[5])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12]) // 10
	t[13].Add(&t[13], &t[13]) // 20
	z.A0.Add(&z.A0, &t[13])

	// +20v6
	t[12].Add(&v[6], &v[6])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12]) // 10
	t[13].Add(&t[13], &t[13]) // 20
	z.A0.Add(&z.A0, &t[13])

	// +30v7
	t[12].Add(&v[7], &v[7])   // 2
	t[13].Add(&t[12], &t[12]) // 4
	t[13].Add(&t[13], &t[13]) // 8
	t[13].Add(&t[13], &t[12]) // 10
	t[12].Add(&t[13], &t[13]) // 20
	t[12].Add(&t[12], &t[13]) // 30
	z.A0.Add(&z.A0, &t[12])

	// -30v8
	t[12].Add(&v[8], &v[8])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12]) // 10
	t[12].Add(&t[13], &t[13]) // 20
	t[12].Add(&t[12], &t[13]) // 30
	z.A0.Sub(&z.A0, &t[12])

	// -30v10
	t[12].Add(&v[10], &v[10])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12]) // 10
	t[12].Add(&t[13], &t[13]) // 20
	t[12].Add(&t[12], &t[13]) // 30
	z.A0.Sub(&z.A0, &t[12])

	// -30v11
	t[12].Add(&v[11], &v[11])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12]) // 10
	t[12].Add(&t[13], &t[13]) // 20
	t[12].Add(&t[12], &t[13]) // 30
	z.A0.Sub(&z.A0, &t[12])

	// -20v12
	t[12].Add(&v[12], &v[12])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12]) // 10
	t[13].Add(&t[13], &t[13]) // 20
	z.A0.Sub(&z.A0, &t[13])

	// -20v13
	t[12].Add(&v[13], &v[13])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12]) // 10
	t[13].Add(&t[13], &t[13]) // 20
	z.A0.Sub(&z.A0, &t[13])

	// -49v14 = -(32+16+1)
	t[12].Add(&t[0], &t[0])   // 4v14
	t[12].Add(&t[12], &t[12]) // 8v14
	t[12].Add(&t[12], &t[12]) // 16v14
	t[13].Add(&t[12], &t[12]) // 32v14
	t[13].Add(&t[13], &t[12]) // 48v14
	t[13].Add(&t[13], &v[14]) // 49v14
	z.A0.Sub(&z.A0, &t[13])

	// +30v15
	t[12].Add(&v[15], &v[15])
	t[13].Add(&t[12], &t[12])
	t[13].Add(&t[13], &t[13])
	t[13].Add(&t[13], &t[12]) // 10
	t[12].Add(&t[13], &t[13]) // 20
	t[12].Add(&t[12], &t[13]) // 30
	z.A0.Add(&z.A0, &t[12])

	// +20v16  (reuse: 10 then double)
	t[12].Set(&t[2])          // 2
	t[13].Add(&t[12], &t[12]) // 4
	t[13].Add(&t[13], &t[13]) // 8
	t[13].Add(&t[13], &t[12]) // 10
	t[13].Add(&t[13], &t[13]) // 20
	z.A0.Add(&z.A0, &t[13])

	// -70v17 = -(64+4+2)
	t[12].Add(&t[10], &t[10]) // 64v17
	t[12].Add(&t[12], &t[7])  // +4v17
	t[12].Add(&t[12], &t[5])  // +2v17 = 70v17
	z.A0.Sub(&z.A0, &t[12])

	return z

}

// Square sets z=x*x in E6D and returns z
func (z *E6D) Square(x *E6D) *E6D {
	_x := ToTower(x)
	_x.Square(_x)
	_z := FromTower(_x)
	return z.Set(_z)
}

// Inverse sets z to the inverse of x in E6D and returns z
//
// if x == 0, sets and returns z = x
func (z *E6D) Inverse(x *E6D) *E6D {
	_x := ToTower(x)
	_x.Inverse(_x)
	_z := FromTower(_x)
	return z.Set(_z)
}

// InverseUnitary inverses a unitary element
func (z *E6D) InverseUnitary(x *E6D) *E6D {
	return z.Conjugate(x)
}

// Conjugate sets z to x conjugated and returns z
func (z *E6D) Conjugate(x *E6D) *E6D {
	_x := ToTower(x)
	_x.Conjugate(_x)
	_z := FromTower(_x)
	return z.Set(_z)
}

// FromTower
func FromTower(x *E6) *E6D {
	// The 2-3 tower and direct extensions are isomorphic and the coefficients
	// are permuted as follows:
	// 		a00-a01 a10-a11 a20-a21 a01 a11 a21
	// 		A0      A1      A2      A3  A4  A5
	var z E6D
	z.A0.Sub(&x.B0.A0, &x.B0.A1)
	z.A1.Sub(&x.B1.A0, &x.B1.A1)
	z.A2.Sub(&x.B2.A0, &x.B2.A1)
	z.A3.Set(&x.B0.A1)
	z.A4.Set(&x.B1.A1)
	z.A5.Set(&x.B2.A1)
	return &z
}

// ToTower
func ToTower(x *E6D) *E6 {
	// The 2-3 tower and direct extensions are isomorphic and the coefficients
	// are permuted as follows:
	// 		a00    a01 a10    a11 a20    a21
	// 		A0+A3  A3  A1+A4  A4  A2+A5  A5
	var z E6
	z.B0.A0.Add(&x.A0, &x.A3)
	z.B0.A1.Set(&x.A3)
	z.B1.A0.Add(&x.A1, &x.A4)
	z.B1.A1.Set(&x.A4)
	z.B2.A0.Add(&x.A2, &x.A5)
	z.B2.A1.Set(&x.A5)
	return &z
}
