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

// FOO

package starkcurve

import (
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// MapToCurve1 implements the Shallue and van de Woestijne method, applicable to any elliptic curve in Weierstrass form
// No cofactor clearing or isogeny
// https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#straightline-svdw
func MapToCurve1(u *fp.Element) G1Affine {
	var tv1, tv2, tv3, tv4 fp.Element
	var x1, x2, x3, gx1, gx2, gx, x, y fp.Element
	var one fp.Element
	var gx1NotSquare, gx1SquareOrGx2Not int

	//constants
	//c1 = g(Z)
	//c2 = -Z / 2
	//c3 = sqrt(-g(Z) * (3 * Z² + 4 * A))     # sgn0(c3) MUST equal 0
	//c4 = -4 * g(Z) / (3 * Z² + 4 * A)

	Z := fp.Element{18446744073709551585, 18446744073709551615, 18446744073709551615, 576460752303422960}
	c1 := fp.Element{3863487492851900810, 7432612994240712710, 12360725113329547591, 88155977965379647}
	c2 := fp.Element{16, 0, 0, 272}
	c3 := fp.Element{9918255022489886019, 17523995898334911653, 15291095870552318715, 510280297527296511}
	c4 := fp.Element{13603787781549958066, 11564287495042065550, 842475966830066354, 443734371708431777}

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

	gx1.Square(&x1)                    //    11. gx1 = x1²
	gx1.Add(&gx1, &one)                // 12. gx1 = gx1 + A (A=1)
	gx1.Mul(&gx1, &x1)                 //    13. gx1 = gx1 * x1
	gx1.Add(&gx1, &bCurveCoeff)        //    14. gx1 = gx1 + B
	gx1NotSquare = gx1.Legendre() >> 1 //    15.  e1 = is_square(gx1)
	// gx1NotSquare = 0 if gx1 is a square, -1 otherwise

	x2.Add(&c2, &tv4)           //    16.  x2 = c2 + tv4
	gx2.Square(&x2)             //    17. gx2 = x2²
	gx2.Add(&gx2, &one)         //    18. gx2 = gx2 + A (A=1)
	gx2.Mul(&gx2, &x2)          //    19. gx2 = gx2 * x2
	gx2.Add(&gx2, &bCurveCoeff) //    20. gx2 = gx2 + B

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
	gx.Square(&x)     //    29.  gx = x²
	gx.Add(&gx, &one) //    30.  gx = gx + A (A=1)

	gx.Mul(&gx, &x)           //    31.  gx = gx * x
	gx.Add(&gx, &bCurveCoeff) //    32.  gx = gx + B

	y.Sqrt(&gx)                             //    33.   y = sqrt(gx)
	signsNotEqual := g1Sgn0(u) ^ g1Sgn0(&y) //    34.  e3 = sgn0(u) == sgn0(y)

	tv1.Neg(&y)
	y.Select(int(signsNotEqual), &y, &tv1) //    35.   y = CMOV(-y, y, e3)       # Select correct sign of y
	return G1Affine{x, y}
}

// g1Sgn0 is an algebraic substitute for the notion of sign in ordered fields
// Namely, every non-zero quadratic residue in a finite field of characteristic =/= 2 has exactly two square roots, one of each sign
// https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#name-the-sgn0-function
// The sign of an element is not obviously related to that of its Montgomery form
func g1Sgn0(z *fp.Element) uint64 {

	nonMont := z.Bits()

	// m == 1
	return nonMont[0] % 2

}

// MapToG1 invokes the SVDW map, and guarantees that the result is in g1
func MapToG1(u fp.Element) G1Affine {
	res := MapToCurve1(&u)
	return res
}

// EncodeToG1 hashes a message to a point on the G1 curve using the SVDW map.
// It is faster than HashToG1, but the result is not uniformly distributed. Unsuitable as a random oracle.
// dst stands for "domain separation tag", a string unique to the construction using the hash function
// https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#roadmap
func EncodeToG1(msg, dst []byte) (G1Affine, error) {

	var res G1Affine
	u, err := fp.Hash(msg, dst, 1)
	if err != nil {
		return res, err
	}

	res = MapToCurve1(&u[0])

	return res, nil
}

// HashToG1 hashes a message to a point on the G1 curve using the SVDW map.
// Slower than EncodeToG1, but usable as a random oracle.
// dst stands for "domain separation tag", a string unique to the construction using the hash function
// https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#roadmap
func HashToG1(msg, dst []byte) (G1Affine, error) {
	u, err := fp.Hash(msg, dst, 2*1)
	if err != nil {
		return G1Affine{}, err
	}

	Q0 := MapToCurve1(&u[0])
	Q1 := MapToCurve1(&u[1])

	var _Q0, _Q1 G1Jac
	_Q0.FromAffine(&Q0)
	_Q1.FromAffine(&Q1).AddAssign(&_Q0)

	Q1.FromJacobian(&_Q1)
	return Q1, nil
}

func g1NotZero(x *fp.Element) uint64 {

	return x[0] | x[1] | x[2] | x[3]

}
