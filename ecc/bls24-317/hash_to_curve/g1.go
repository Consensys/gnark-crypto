// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package hash_to_curve

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls24-317/fp"
)

// Note: This only works for simple extensions

var (
	g1sswuCurveACoeff = fp.Element{2751493217506761890, 10508083672876982400, 9568653941102734201, 1934905759174260726, 590687129635764257}
	g1sswuCurveBCoeff = fp.Element{14477170886729819615, 1154054877908840441, 13400991584556574205, 3277375072715511934, 979998381373634863}
)

var g1sswuCurveZ = fp.Element{18400687542797871745, 809728271075671860, 17770696641280178537, 10361798156408411167, 334758614216279309}

// G1SSWUCurveCoefficients returns the coefficients of the SSWU curve.
func G1SSWUIsogenyCurveCoefficients() (A fp.Element, B fp.Element) {
	return g1sswuCurveACoeff, g1sswuCurveBCoeff
}

// G1SSWUIsogenyZ returns the recommended Z value of the SSWU curve.
//
// See https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#weierstrass
func G1SSWUIsogenyZ() fp.Element {
	return g1sswuCurveZ
}

var (
	g1IsogenyXNumeratorMap = []fp.Element{
		{13523513236317711848, 15327023349232218118, 8703648794266574884, 8264167271110563191, 40794431846902569},
		{8812074666074491586, 50960250954420133, 14056404179861272537, 929938412739573318, 947153270783672532},
		{15051608682446262522, 9488224519772198430, 11710444855428888956, 16398015671457218553, 1029622088557318610},
		{4296820476805851409, 10780602457143466946, 10247933845608112961, 6951059907314751932, 722213278859423782},
		{14764184048304945149, 5865289230433310091, 5581095008736809995, 9208735173835224741, 528727552546926153},
		{11398597359936714397, 1594057801015249474, 13954376621701424207, 16271868308895978452, 690753220876234821},
	}
	g1IsogenyXDenominatorMap = []fp.Element{
		{5399775903125704630, 4517816096475473808, 8510054034683086600, 15646083100922413141, 906999227924553668},
		{828013697853572132, 458878942468938987, 5757230941761973224, 5158770805028806783, 869290263606291835},
		{11118632362304015867, 6158437615457151578, 8167114226690349799, 18398210184903822958, 32908558142489459},
		{17284245259114832476, 13149059755030257718, 10930970338758309391, 1062425496339030960, 261139743832662079},
	}
	g1IsogenyYNumeratorMap = []fp.Element{
		{5736138424590314750, 6015908605773073009, 6156792889286183843, 17896612273365749807, 821435345686805089},
		{9373359301599115869, 655867965241119234, 3304264667834595975, 12237805962366901484, 297609776634465799},
		{3480981777823324659, 9475237666221295368, 11936228663660569620, 16004883291078000733, 694053280005543484},
		{4229115995671887337, 9233280055297188894, 1359384483422747035, 11273993240180143056, 469494085796341224},
		{18113844587232876680, 14242937351038565984, 777537960123335163, 6685524189684440232, 980736769871245076},
		{11922196649017768415, 7237889860522244398, 3155125612682980193, 3938240406780725187, 665921220498498902},
		{3446223578941560630, 13846992323172164671, 12292264306216531556, 7620005162288670125, 97432066185489249},
	}
	g1IsogenyYDenominatorMap = []fp.Element{
		{8602082813304143536, 14359122824402329793, 2469007073274644071, 4254406725226729972, 992519966230345268},
		{3085489453415801238, 15662911842127999867, 9714633693652399946, 9543599792786380558, 789455890382293440},
		{17898042109793411276, 8772407166446083546, 16320058043659241709, 18250219114565265632, 721227617678419637},
		{12665654738497754715, 10529888736786073619, 14298592531231225548, 714005056864991408, 1088730156414821854},
		{11181082342903713721, 9065467944505387329, 647327075925674801, 8268923912961120967, 264633289965085690},
		{7479623814962697098, 10500217595690610770, 16396455508137464087, 10817010281363322248, 391709615748993118},
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
	// https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#name-optimized-sqrt_ratio-for-q- (3 mod 4)
	var tv1 fp.Element
	tv1.Square(v) // 1. tv1 = v²
	var tv2 fp.Element
	tv2.Mul(u, v)       // 2. tv2 = u * v
	tv1.Mul(&tv1, &tv2) // 3. tv1 = tv1 * tv2

	var y1 fp.Element
	{
		var c1 big.Int
		// c1 = 34098267776073977878774941477068514265486278030354898494302534825976493299308006404506539182762
		c1.SetBytes([]byte{4, 22, 50, 136, 155, 216, 34, 75, 60, 163, 241, 104, 45, 254, 116, 14, 69, 166, 152, 121, 161, 49, 205, 17, 181, 188, 206, 121, 13, 9, 47, 223, 163, 84, 75, 149, 151, 106, 202, 170}) // c1 = (q - 3) / 4     # Integer arithmetic

		y1.Exp(tv1, &c1) // 4. y1 = tv1ᶜ¹
	}

	y1.Mul(&y1, &tv2) // 5. y1 = y1 * tv2

	var y2 fp.Element
	// c2 = sqrt(-Z)
	tv3 := fp.Element{10652859563586318787, 3643689439157831556, 9236201363192486412, 11781990169133948855, 1044489031832785863}
	y2.Mul(&y1, &tv3)              // 6. y2 = y1 * c2
	tv3.Square(&y1)                // 7. tv3 = y1²
	tv3.Mul(&tv3, v)               // 8. tv3 = tv3 * v
	isQNr := tv3.NotEqual(u)       // 9. isQR = tv3 == u
	z.Select(int(isQNr), &y1, &y2) // 10. y = CMOV(y2, y1, isQR)
	return isQNr
}

// G1MulByZ multiplies x by [8] and stores the result in z
func G1MulByZ(z *fp.Element, x *fp.Element) {

	res := *x

	res.Double(&res)
	res.Double(&res)
	res.Double(&res)

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
// See: https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#name-the-sgn0-function
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
