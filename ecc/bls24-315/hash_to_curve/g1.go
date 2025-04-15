// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package hash_to_curve

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
)

// Note: This only works for simple extensions

var (
	g1sswuCurveACoeff = fp.Element{5402807948305211529, 9163880483319140034, 7646126700453841420, 11071466103913358468, 124200740526673728}
	g1sswuCurveBCoeff = fp.Element{16058189711238232929, 8302337653269510588, 11411933349841587630, 8954038365926617417, 177308873523699836}
)

var g1sswuCurveZ = fp.Element{8178485296672800069, 8476448362227282520, 14180928431697993131, 4308307642551989706, 120359802761433421}

// G1SSWUCurveCoefficients returns the coefficients of the SSWU curve.
func G1SSWUIsogenyCurveCoefficients() (A fp.Element, B fp.Element) {
	return g1sswuCurveACoeff, g1sswuCurveBCoeff
}

// G1SSWUIsogenyZ returns the recommended Z value of the SSWU curve.
//
// See https://www.rfc-editor.org/rfc/rfc9380.html#weierstrass
func G1SSWUIsogenyZ() fp.Element {
	return g1sswuCurveZ
}

var (
	g1IsogenyXNumeratorMap = []fp.Element{
		{11620002718874663739, 4984467296741409765, 9174718300976205935, 11374294140644765434, 331965326722599209},
		{12794915441326992831, 3515443655574390653, 6174257928039766159, 70148989344615692, 200953992158149919},
		{5852384876649512947, 11848499933279379168, 12693517207910261404, 4355336966086013201, 153982054162701797},
	}
	g1IsogenyXDenominatorMap = []fp.Element{
		{16605520835351066362, 4532778258980819953, 11041097066391022716, 6626569051763865297, 118015358745724890},
	}
	g1IsogenyYNumeratorMap = []fp.Element{
		{9734843649657667679, 9905469488516037607, 12244225131002460472, 12160927269755757379, 293726840634836990},
		{332611309977308143, 8673449249147179720, 7968180610051701274, 525286427825436485, 27337445552095458},
		{5937151911073875102, 12114288429387176123, 10459089249045026662, 1691716757613274170, 129980835765506182},
		{6958041652386594810, 8306499057468875249, 14372428283824529086, 591175209446655968, 248441179553069595},
	}
	g1IsogenyYDenominatorMap = []fp.Element{
		{7466136663908942255, 5910124112997814042, 598236406339551119, 15948603688162126360, 216078840945103380},
		{16886107642822413408, 14927238232380936652, 17792216571653695247, 7051181952824703829, 174959651533410931},
		{4859375930510419181, 8833836595284088531, 17071951839434271380, 4605949628774745540, 11145771293737278},
	}
)

// G1IsogenyMap returns the isogeny map for the curve.
// The isogeny map is a list of polynomial coefficients for the x and y coordinate computation.
// The order of the coefficients is as follows:
// - x numerator, x denominator, y numerator, y denominator.
func G1IsogenyMap() [4][]fp.Element {
	return [4][]fp.Element{
		g1IsogenyXNumeratorMap,
		g1IsogenyXDenominatorMap,
		g1IsogenyYNumeratorMap,
		g1IsogenyYDenominatorMap,
	}
}

func g1IsogenyXNumerator(dst *fp.Element, x *fp.Element) {
	g1EvalPolynomial(dst, false, g1IsogenyXNumeratorMap, x)
}

func g1IsogenyXDenominator(dst *fp.Element, x *fp.Element) {
	g1EvalPolynomial(dst, true, g1IsogenyXDenominatorMap, x)
}

func g1IsogenyYNumerator(dst *fp.Element, x *fp.Element, y *fp.Element) {
	var _dst fp.Element
	g1EvalPolynomial(&_dst, false, g1IsogenyYNumeratorMap, x)
	dst.Mul(&_dst, y)
}

func g1IsogenyYDenominator(dst *fp.Element, x *fp.Element) {
	g1EvalPolynomial(dst, true, g1IsogenyYDenominatorMap, x)
}

// G1 computes the isogeny map of the curve element, given by its coordinates pX and pY.
// It mutates the coordinates pX and pY to the new coordinates of the isogeny map.
func G1Isogeny(pX, pY *fp.Element) {

	den := make([]fp.Element, 2)

	g1IsogenyYDenominator(&den[1], pX)
	g1IsogenyXDenominator(&den[0], pX)

	g1IsogenyYNumerator(pY, pX, pY)
	g1IsogenyXNumerator(pX, pX)

	den = fp.BatchInvert(den)

	pX.Mul(pX, &den[0])
	pY.Mul(pY, &den[1])
}

// G1SqrtRatio computes the square root of u/v and returns 0 iff u/v was indeed a quadratic residue.
// If not, we get sqrt(Z * u / v). Recall that Z is non-residue.
// If v = 0, u/v is meaningless and the output is unspecified, without raising an error.
// The main idea is that since the computation of the square root involves taking large powers of u/v, the inversion of v can be avoided
func G1SqrtRatio(z *fp.Element, u *fp.Element, v *fp.Element) uint64 {

	// https://www.rfc-editor.org/rfc/rfc9380.html#name-sqrt_ratio-for-any-field

	tv1 := fp.Element{11195128742969911322, 1359304652430195240, 15267589139354181340, 10518360976114966361, 300769513466036652} //tv1 = c6

	var tv2, tv3, tv4, tv5 fp.Element
	var exp big.Int
	// c4 = 1048575 = 2²⁰ - 1
	// q is odd so c1 is at least 1.
	exp.SetBytes([]byte{15, 255, 255})

	tv2.Exp(*v, &exp) // 2. tv2 = vᶜ⁴
	tv3.Square(&tv2)  // 3. tv3 = tv2²
	tv3.Mul(&tv3, v)  // 4. tv3 = tv3 * v
	tv5.Mul(u, &tv3)  // 5. tv5 = u * tv3

	// c3 = 18932887415653914611351818986134037849871398170907377879650252106493894621432467626129921
	exp.SetBytes([]byte{38, 17, 208, 21, 172, 54, 178, 134, 159, 186, 76, 95, 75, 226, 245, 126, 246, 14, 128, 213, 19, 208, 215, 2, 16, 247, 46, 210, 149, 239, 40, 19, 127, 64, 23, 250, 1})

	tv5.Exp(tv5, &exp)  // 6. tv5 = tv5ᶜ³
	tv5.Mul(&tv5, &tv2) // 7. tv5 = tv5 * tv2
	tv2.Mul(&tv5, v)    // 8. tv2 = tv5 * v
	tv3.Mul(&tv5, u)    // 9. tv3 = tv5 * u
	tv4.Mul(&tv3, &tv2) // 10. tv4 = tv3 * tv2

	// c5 = 524288
	exp.SetBytes([]byte{8, 0, 0})
	tv5.Exp(tv4, &exp)      // 11. tv5 = tv4ᶜ⁵
	isQNr := g1NotOne(&tv5) // 12. isQR = tv5 == 1
	c7 := fp.Element{1141794007209116247, 256324699145650176, 2958838397954514392, 9976887947641032208, 153331829745922234}
	tv2.Mul(&tv3, &c7)                 // 13. tv2 = tv3 * c7
	tv5.Mul(&tv4, &tv1)                // 14. tv5 = tv4 * tv1
	tv3.Select(int(isQNr), &tv3, &tv2) // 15. tv3 = CMOV(tv2, tv3, isQR)
	tv4.Select(int(isQNr), &tv4, &tv5) // 16. tv4 = CMOV(tv5, tv4, isQR)
	exp.Lsh(big.NewInt(1), 20-2)       // 18, 19: tv5 = 2ⁱ⁻² for i = c1

	for i := 20; i >= 2; i-- { // 17. for i in (c1, c1 - 1, ..., 2):

		tv5.Exp(tv4, &exp)               // 20.    tv5 = tv4ᵗᵛ⁵
		nE1 := g1NotOne(&tv5)            // 21.    e1 = tv5 == 1
		tv2.Mul(&tv3, &tv1)              // 22.    tv2 = tv3 * tv1
		tv1.Mul(&tv1, &tv1)              // 23.    tv1 = tv1 * tv1    Why not write square?
		tv5.Mul(&tv4, &tv1)              // 24.    tv5 = tv4 * tv1
		tv3.Select(int(nE1), &tv3, &tv2) // 25.    tv3 = CMOV(tv2, tv3, e1)
		tv4.Select(int(nE1), &tv4, &tv5) // 26.    tv4 = CMOV(tv5, tv4, e1)

		if i > 2 {
			exp.Rsh(&exp, 1) // 18, 19. tv5 = 2ⁱ⁻²
		}
	}

	*z = tv3
	return isQNr
}

func g1NotOne(x *fp.Element) uint64 {

	var one fp.Element
	return one.SetOne().NotEqual(x)

}

// G1MulByZ multiplies x by [13] and stores the result in z
func G1MulByZ(z *fp.Element, x *fp.Element) {

	res := *x

	res.Double(&res)
	res.Add(&res, x)
	res.Double(&res)
	res.Double(&res)
	res.Add(&res, x)

	*z = res
}

func g1EvalPolynomial(z *fp.Element, monic bool, coefficients []fp.Element, x *fp.Element) {
	dst := coefficients[len(coefficients)-1]

	if monic {
		dst.Add(&dst, x)
	}

	for i := len(coefficients) - 2; i >= 0; i-- {
		dst.Mul(&dst, x)
		dst.Add(&dst, &coefficients[i])
	}

	z.Set(&dst)
}

// G1Sgn0 is an algebraic substitute for the notion of sign in ordered fields.
// Namely, every non-zero quadratic residue in a finite field of characteristic =/= 2 has exactly two square roots, one of each sign.
//
// See: https://www.rfc-editor.org/rfc/rfc9380.html#name-the-sgn0-function
//
// The sign of an element is not obviously related to that of its Montgomery form
func G1Sgn0(z *fp.Element) uint64 {

	nonMont := z.Bits()

	// m == 1
	return nonMont[0] % 2

}

func G1NotZero(x *fp.Element) uint64 {

	return x[0] | x[1] | x[2] | x[3] | x[4]

}
