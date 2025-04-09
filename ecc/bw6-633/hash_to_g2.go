// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package bw6633

import (
	"github.com/consensys/gnark-crypto/ecc/bw6-633/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-633/hash_to_curve"
)

// MapToG2 invokes the SSWU map, and guarantees that the result is in G2.
func MapToG2(u fp.Element) G2Affine {
	res := MapToCurve2(&u)
	//this is in an isogenous curve
	hash_to_curve.G2Isogeny(&res.X, &res.Y)
	res.ClearCofactor(&res)
	return res
}

// EncodeToG2 hashes a message to a point on the G2 curve using the SSWU map.
// It is faster than [HashToG2], but the result is not uniformly distributed. Unsuitable as a random oracle.
// dst stands for "domain separation tag", a string unique to the construction using the hash function
//
// See: https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#roadmap
func EncodeToG2(msg, dst []byte) (G2Affine, error) {

	var res G2Affine
	u, err := fp.Hash(msg, dst, 1)
	if err != nil {
		return res, err
	}

	res = MapToCurve2(&u[0])

	//this is in an isogenous curve
	hash_to_curve.G2Isogeny(&res.X, &res.Y)
	res.ClearCofactor(&res)
	return res, nil
}

// HashToG2 hashes a message to a point on the G2 curve using the SSWU map.
// Slower than [EncodeToG2], but usable as a random oracle.
// dst stands for "domain separation tag", a string unique to the construction using the hash function.
//
// https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#roadmap
func HashToG2(msg, dst []byte) (G2Affine, error) {
	u, err := fp.Hash(msg, dst, 2*1)
	if err != nil {
		return G2Affine{}, err
	}

	Q0 := MapToCurve2(&u[0])
	Q1 := MapToCurve2(&u[1])

	//TODO (perf): Add in E' first, then apply isogeny
	hash_to_curve.G2Isogeny(&Q0.X, &Q0.Y)
	hash_to_curve.G2Isogeny(&Q1.X, &Q1.Y)

	var _Q0, _Q1 G2Jac
	_Q0.FromAffine(&Q0)
	_Q1.FromAffine(&Q1).AddAssign(&_Q0)

	_Q1.ClearCofactor(&_Q1)

	Q1.FromJacobian(&_Q1)
	return Q1, nil
}

// MapToCurve2 implements the SSWU map. It does not perform cofactor clearing nor isogeny. For map to group, use [MapToG2].
//
// See: https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#name-simplified-swu-method
func MapToCurve2(u *fp.Element) G2Affine {
	g2sswuCurveACoeff, g2sswuCurveBCoeff := hash_to_curve.G2SSWUIsogenyCurveCoefficients()

	var tv1 fp.Element
	tv1.Square(u) // 1.  tv1 = u²

	//mul tv1 by Z
	hash_to_curve.G2MulByZ(&tv1, &tv1) // 2.  tv1 = Z * tv1

	var tv2 fp.Element
	tv2.Square(&tv1)    // 3.  tv2 = tv1²
	tv2.Add(&tv2, &tv1) // 4.  tv2 = tv2 + tv1

	var tv3 fp.Element
	var tv4 fp.Element
	tv4.SetOne()
	tv3.Add(&tv2, &tv4)               // 5.  tv3 = tv2 + 1
	tv3.Mul(&tv3, &g2sswuCurveBCoeff) // 6.  tv3 = B * tv3

	tv2NZero := hash_to_curve.G2NotZero(&tv2)

	// tv4 = Z
	tv4 = hash_to_curve.G2SSWUIsogenyZ()

	tv2.Neg(&tv2)
	tv4.Select(int(tv2NZero), &tv4, &tv2) // 7.  tv4 = CMOV(Z, -tv2, tv2 != 0)
	tv4.Mul(&tv4, &g2sswuCurveACoeff)     // 8.  tv4 = A * tv4

	tv2.Square(&tv3) // 9.  tv2 = tv3²

	var tv6 fp.Element
	tv6.Square(&tv4) // 10. tv6 = tv4²

	var tv5 fp.Element
	tv5.Mul(&tv6, &g2sswuCurveACoeff) // 11. tv5 = A * tv6

	tv2.Add(&tv2, &tv5) // 12. tv2 = tv2 + tv5
	tv2.Mul(&tv2, &tv3) // 13. tv2 = tv2 * tv3
	tv6.Mul(&tv6, &tv4) // 14. tv6 = tv6 * tv4

	tv5.Mul(&tv6, &g2sswuCurveBCoeff) // 15. tv5 = B * tv6
	tv2.Add(&tv2, &tv5)               // 16. tv2 = tv2 + tv5

	var x fp.Element
	x.Mul(&tv1, &tv3) // 17.   x = tv1 * tv3

	var y1 fp.Element
	gx1NSquare := hash_to_curve.G2SqrtRatio(&y1, &tv2, &tv6) // 18. (is_gx1_square, y1) = sqrt_ratio(tv2, tv6)

	var y fp.Element
	y.Mul(&tv1, u) // 19.   y = tv1 * u

	y.Mul(&y, &y1) // 20.   y = y * y1

	x.Select(int(gx1NSquare), &tv3, &x) // 21.   x = CMOV(x, tv3, is_gx1_square)
	y.Select(int(gx1NSquare), &y1, &y)  // 22.   y = CMOV(y, y1, is_gx1_square)

	y1.Neg(&y)
	y.Select(int(hash_to_curve.G2Sgn0(u)^hash_to_curve.G2Sgn0(&y)), &y, &y1)

	// 23.  e1 = sgn0(u) == sgn0(y)
	// 24.   y = CMOV(-y, y, e1)

	x.Div(&x, &tv4) // 25.   x = x / tv4

	return G2Affine{x, y}
}
