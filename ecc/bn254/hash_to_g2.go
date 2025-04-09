// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package bn254

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/hash_to_curve"
	"github.com/consensys/gnark-crypto/ecc/bn254/internal/fptower"
)

// MapToG2 invokes the SVDW map, and guarantees that the result is in G2.
func MapToG2(u fptower.E2) G2Affine {
	res := MapToCurve2(&u)
	res.ClearCofactor(&res)
	return res
}

// EncodeToG2 hashes a message to a point on the G2 curve using the SVDW map.
// It is faster than [HashToG2], but the result is not uniformly distributed. Unsuitable as a random oracle.
// dst stands for "domain separation tag", a string unique to the construction using the hash function
//
// See: https://www.rfc-editor.org/rfc/rfc9380.html#roadmap
func EncodeToG2(msg, dst []byte) (G2Affine, error) {

	var res G2Affine
	u, err := fp.Hash(msg, dst, 2)
	if err != nil {
		return res, err
	}

	res = MapToCurve2(&fptower.E2{
		A0: u[0],
		A1: u[1],
	})

	res.ClearCofactor(&res)
	return res, nil
}

// HashToG2 hashes a message to a point on the G2 curve using the SVDW map.
// Slower than [EncodeToG2], but usable as a random oracle.
// dst stands for "domain separation tag", a string unique to the construction using the hash function.
//
// See https://www.rfc-editor.org/rfc/rfc9380.html#roadmap
func HashToG2(msg, dst []byte) (G2Affine, error) {
	u, err := fp.Hash(msg, dst, 2*2)
	if err != nil {
		return G2Affine{}, err
	}

	Q0 := MapToCurve2(&fptower.E2{
		A0: u[0],
		A1: u[1],
	})
	Q1 := MapToCurve2(&fptower.E2{
		A0: u[2+0],
		A1: u[2+1],
	})

	var _Q0, _Q1 G2Jac
	_Q0.FromAffine(&Q0)
	_Q1.FromAffine(&Q1).AddAssign(&_Q0)

	_Q1.ClearCofactor(&_Q1)

	Q1.FromJacobian(&_Q1)
	return Q1, nil
}

// MapToCurve2 implements the Shallue and van de Woestijne method, applicable to any elliptic curve in Weierstrass form.
// It does not perform cofactor clearing nor isogeny. Use [MapToG2] for mapping to group.
//
// See: https://www.rfc-editor.org/rfc/rfc9380.html#straightline-svdw
func MapToCurve2(u *fptower.E2) G2Affine {
	var tv1, tv2, tv3, tv4 fptower.E2
	var x1, x2, x3, gx1, gx2, gx, x, y fptower.E2
	var one fptower.E2
	var gx1NotSquare, gx1SquareOrGx2Not int

	//constants
	//c1 = g(Z)
	//c2 = -Z / 2
	//c3 = sqrt(-g(Z) * (3 * Z² + 4 * A))     # sgn0(c3) MUST equal 0
	//c4 = -4 * g(Z) / (3 * Z² + 4 * A)

	Z := fptower.E2{
		A0: fp.Element{15230403791020821917, 754611498739239741, 7381016538464732716, 1011752739694698287},
		A1: fp.Element{0},
	}
	c1 := fptower.E2{
		A0: fp.Element{15219334786797146878, 8431472696017589261, 15336528771359260718, 196732871012706162},
		A1: fp.Element{4100506350182530919, 7345568344173317438, 15513160039642431658, 90557763186888013},
	}
	c2 := fptower.E2{
		A0: fp.Element{12997850613838968789, 14304628359724097447, 2950087706404981016, 1237622763554136189},
		A1: fp.Element{0},
	}
	c3 := fptower.E2{
		A0: fp.Element{12298500088583694207, 17447120171744064890, 14097510924717921191, 2278398337453771183},
		A1: fp.Element{4693446565795584099, 18320164443970680666, 6792758484113206563, 2989688171181581768},
	}
	c4 := fptower.E2{
		A0: fp.Element{7191623630069643826, 8333948550768170742, 13001081703983517696, 2062355016518372226},
		A1: fp.Element{11163104453509316115, 7271947710149976975, 4894807947557820282, 3366254582553786647},
	}

	one.SetOne()

	tv1.Square(u)       //    1.  tv1 = u²
	tv1.Mul(&tv1, &c1)  //    2.  tv1 = tv1 * c1
	tv2.Add(&one, &tv1) //    3.  tv2 = 1 + tv1
	tv1.Sub(&one, &tv1) //    4.  tv1 = 1 - tv1
	tv3.Mul(&tv1, &tv2) //    5.  tv3 = tv1 * tv2

	tv3.Inverse(&tv3)   //    6.  tv3 = inv0(tv3)
	tv4.Mul(u, &tv1)    //    7.  tv4 = u * tv1
	tv4.Mul(&tv4, &tv3) //    8.  tv4 = tv4 * tv3
	tv4.Mul(&tv4, &c3)  //    9.  tv4 = tv4 * c3
	x1.Sub(&c2, &tv4)   //    10.  x1 = c2 - tv4

	gx1.Square(&x1) //    11. gx1 = x1²
	//12. gx1 = gx1 + A     All curves in gnark-crypto have A=0 (j-invariant=0). It is crucial to include this step if the curve has nonzero A coefficient.
	gx1.Mul(&gx1, &x1)                 //    13. gx1 = gx1 * x1
	gx1.Add(&gx1, &bTwistCurveCoeff)   //    14. gx1 = gx1 + B
	gx1NotSquare = gx1.Legendre() >> 1 //    15.  e1 = is_square(gx1)
	// gx1NotSquare = 0 if gx1 is a square, -1 otherwise

	x2.Add(&c2, &tv4) //    16.  x2 = c2 + tv4
	gx2.Square(&x2)   //    17. gx2 = x2²
	//    18. gx2 = gx2 + A     See line 12
	gx2.Mul(&gx2, &x2)               //    19. gx2 = gx2 * x2
	gx2.Add(&gx2, &bTwistCurveCoeff) //    20. gx2 = gx2 + B

	{
		gx2NotSquare := gx2.Legendre() >> 1              // gx2Square = 0 if gx2 is a square, -1 otherwise
		gx1SquareOrGx2Not = gx2NotSquare | ^gx1NotSquare //    21.  e2 = is_square(gx2) AND NOT e1   # Avoid short-circuit logic ops
	}

	x3.Square(&tv2)   //    22.  x3 = tv2²
	x3.Mul(&x3, &tv3) //    23.  x3 = x3 * tv3
	x3.Square(&x3)    //    24.  x3 = x3²
	x3.Mul(&x3, &c4)  //    25.  x3 = x3 * c4

	x3.Add(&x3, &Z)                  //    26.  x3 = x3 + Z
	x.Select(gx1NotSquare, &x1, &x3) //    27.   x = CMOV(x3, x1, e1)   # x = x1 if gx1 is square, else x = x3
	// Select x1 iff gx1 is square iff gx1NotSquare = 0
	x.Select(gx1SquareOrGx2Not, &x2, &x) //    28.   x = CMOV(x, x2, e2)    # x = x2 if gx2 is square and gx1 is not
	// Select x2 iff gx2 is square and gx1 is not, iff gx1SquareOrGx2Not = 0
	gx.Square(&x) //    29.  gx = x²
	//    30.  gx = gx + A

	gx.Mul(&gx, &x)                //    31.  gx = gx * x
	gx.Add(&gx, &bTwistCurveCoeff) //    32.  gx = gx + B

	y.Sqrt(&gx)                                                         //    33.   y = sqrt(gx)
	signsNotEqual := hash_to_curve.G2Sgn0(u) ^ hash_to_curve.G2Sgn0(&y) //    34.  e3 = sgn0(u) == sgn0(y)

	tv1.Neg(&y)
	y.Select(int(signsNotEqual), &y, &tv1) //    35.   y = CMOV(-y, y, e3)       # Select correct sign of y
	return G2Affine{x, y}
}
