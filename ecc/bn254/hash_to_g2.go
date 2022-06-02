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

package bn254

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/internal/fptower"
)

// https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-hash-to-curve-14#appendix-F.1
// Shallue and van de Woestijne method, works for any elliptic curve in Weierstrass curve
func mapToCurve2(u fptower.E2) G2Affine {
	var tv1, tv2, tv3, tv4 fptower.E2
	var x1, x2, x3, gx1, gx2, gx, x, y fptower.E2
	var one fptower.E2
	var gx1NotSquare, gx1SquareOrGx2Not int

	//constants
	//c1 = g(Z)
	//c2 = -Z / 2
	//c3 = sqrt(-g(Z) * (3 * Z² + 4 * A))     # sgn0(c3) MUST equal 0
	//c4 = -4 * g(Z) / (3 * Z² + 4 * A)

	var Z, c1, c2, c3, c4 fptower.E2
	Z.A1.SetString("1")
	c1.A0.SetString("19485874751759354771024239261021720505790618469301721065564631296452457478373")
	c1.A1.SetString("266929791119991161246907387137283842545076965332900288569378510910307636689")
	c2.A1.SetString("10944121435919637611123202872628637544348155578648911831344518947322613104291")

	c3.A0.SetString("8270257801618377462829664163334948115088143961679076698731296916415895764198")
	c3.A1.SetString("15403170217607925661891511707918230497750592932893890913125906786266381721360")

	c4.A0.SetString("18685085378399381287283517099609868978155387573303020199856495763721534568303")
	c4.A1.SetString("355906388159988214995876516183045123393435953777200384759171347880410182252")

	one.SetOne()

	tv1.Square(&u)      //    1.  tv1 = u²
	tv1.Mul(&tv1, &c1)  //    2.  tv1 = tv1 * c1
	tv2.Add(&one, &tv1) //    3.  tv2 = 1 + tv1
	tv1.Sub(&one, &tv1) //    4.  tv1 = 1 - tv1
	tv3.Mul(&tv1, &tv2) //    5.  tv3 = tv1 * tv2

	tv3.Inverse(&tv3)   //    6.  tv3 = inv0(tv3)
	tv4.Mul(&u, &tv1)   //    7.  tv4 = u * tv1
	tv4.Mul(&tv4, &tv3) //    8.  tv4 = tv4 * tv3
	tv4.Mul(&tv4, &c3)  //    9.  tv4 = tv4 * c3
	x1.Sub(&c2, &tv4)   //    10.  x1 = c2 - tv4

	gx1.Square(&x1) //    11. gx1 = x1²
	//12. gx1 = gx1 + A
	gx1.Mul(&gx1, &x1)                 //    13. gx1 = gx1 * x1
	gx1.Add(&gx1, &bTwistCurveCoeff)   //    14. gx1 = gx1 + B
	gx1NotSquare = gx1.Legendre() >> 1 //    15.  e1 = is_square(gx1)
	// gx1NotSquare = 0 if gx1 is a square, -1 otherwise

	x2.Add(&c2, &tv4) //    16.  x2 = c2 + tv4
	gx2.Square(&x2)   //    17. gx2 = x2²
	//    18. gx2 = gx2 + A
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

	y.Sqrt(&gx)                              //    33.   y = sqrt(gx)
	signsNotEqual := g2Sgn0(&u) ^ g2Sgn0(&y) //    34.  e3 = sgn0(u) == sgn0(y)

	tv1.Neg(&y)
	y.Select(int(signsNotEqual), &y, &tv1) //    35.   y = CMOV(-y, y, e3)       # Select correct sign of y
	return G2Affine{x, y}
}

// MapToG2 maps an Fp2 to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-2.2.1
func MapToG2(t fptower.E2) G2Affine {
	res := mapToCurve2(t)
	res.ClearCofactor(&res)
	return res
}

// EncodeToG2 maps an Fp2 to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-2.2.2
func EncodeToG2(msg, dst []byte) (G2Affine, error) {
	_t, err := hashToFp(msg, dst, 2)
	if err != nil {
		return G2Affine{}, err
	}
	var t fptower.E2
	t.A0.Set(&_t[0])
	t.A1.Set(&_t[1])
	return MapToG2(t), nil
}

// HashToG2 maps an Fp2 to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-3
func HashToG2(msg, dst []byte) (G2Affine, error) {
	var res G2Affine
	u, err := hashToFp(msg, dst, 4)
	if err != nil {
		return res, err
	}
	var u0, u1 fptower.E2
	u0.A0.Set(&u[0])
	u0.A1.Set(&u[1])
	u1.A0.Set(&u[2])
	u1.A1.Set(&u[3])
	Q0 := mapToCurve2(u0)
	Q1 := mapToCurve2(u1)
	var _Q0, _Q1, _res G2Jac
	_Q0.FromAffine(&Q0)
	_Q1.FromAffine(&Q1)
	_res.Set(&_Q1).AddAssign(&_Q0)
	_res.ClearCofactor(&_res)
	res.FromJacobian(&_res)
	return res, nil
}

func g1NotZero(x *fp.Element) uint64 {

	return x[0] | x[1] | x[2] | x[3]

}

// g2Sgn0 is an algebraic substitute for the notion of sign in ordered fields
// Namely, every non-zero quadratic residue in a finite field of characteristic =/= 2 has exactly two square roots, one of each sign
// Taken from https://datatracker.ietf.org/doc/draft-irtf-cfrg-hash-to-curve/ section 4.1
// The sign of an element is not obviously related to that of its Montgomery form
func g2Sgn0(z *fptower.E2) uint64 {

	nonMont := *z
	nonMont.FromMont()

	sign := uint64(0)
	zero := uint64(1)
	var signI uint64
	var zeroI uint64

	signI = nonMont.A0[0] % 2
	sign = sign | (zero & signI)

	zeroI = g1NotZero(&nonMont.A0)
	zeroI = 1 ^ (zeroI|-zeroI)>>63
	zero = zero & zeroI

	signI = nonMont.A1[0] % 2
	sign = sign | (zero & signI)

	return sign

}
