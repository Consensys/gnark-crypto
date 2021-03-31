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

package bls12377

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/internal/fptower"
)

// hashToFp hashes msg to count prime field elements.
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-5.2
func hashToFp(msg, dst []byte, count int) ([]fp.Element, error) {

	// 128 bits of security
	// L = ceil((ceil(log2(p)) + k) / 8), where k is the security parameter = 128
	L := 64

	lenInBytes := count * L
	pseudoRandomBytes, err := ecc.ExpandMsgXmd(msg, dst, lenInBytes)
	if err != nil {
		return nil, err
	}

	res := make([]fp.Element, count)
	for i := 0; i < count; i++ {
		res[i].SetBytes(pseudoRandomBytes[i*L : (i+1)*L])
	}
	return res, nil
}

// returns false if u>-u when seen as a bigInt
func sign0(u fp.Element) bool {
	var a, b big.Int
	u.ToBigIntRegular(&a)
	u.Neg(&u)
	u.ToBigIntRegular(&b)
	return a.Cmp(&b) <= 0
}

// ----------------------------------------------------------------------------------------
// G1Affine

// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-4.1
// Shallue and van de Woestijne method, works for any elliptic curve in Weierstrass curve
func svdwMapG1(u fp.Element) G1Affine {

	var res G1Affine

	// constants
	// sage script to find z: https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#appendix-E.1
	var z, c1, c2, c3, c4 fp.Element
	z.SetOne()
	c1.SetString("2")
	c2.SetString("129332213006484547005326366847446766768196756377457330269942131333360234174170411387484444069786680062220160729088")
	c3.SetString("97648839010665214827241242728596775338087731732850880761532715038339062821120154619091300503722809961039397351015")
	c4.SetString("172442950675312729340435155796595689024262341836609773693256175111146978898893881849979258759715573416293547638782")

	var tv1, tv2, tv3, tv4, one, x1, gx1, x2, gx2, x3, x, gx, y fp.Element
	one.SetOne()
	tv1.Square(&u).Mul(&tv1, &c1)
	tv2.Add(&one, &tv1)
	tv1.Sub(&one, &tv1)
	tv3.Mul(&tv2, &tv1).Inverse(&tv3)
	tv4.Mul(&u, &tv1)
	tv4.Mul(&tv4, &tv3)
	tv4.Mul(&tv4, &c3)
	x1.Sub(&c2, &tv4)
	gx1.Square(&x1)
	// 12. gx1 = gx1 + A
	gx1.Mul(&gx1, &x1)
	gx1.Add(&gx1, &bCurveCoeff)
	e1 := gx1.Legendre()
	x2.Add(&c2, &tv4)
	gx2.Square(&x2)
	// 18. gx2 = gx2 + A
	gx2.Mul(&gx2, &x2)
	gx2.Add(&gx2, &bCurveCoeff)
	e2 := gx2.Legendre() - e1 // 2 if is_square(gx2) AND NOT e1
	x3.Square(&tv2)
	x3.Mul(&x3, &tv3)
	x3.Square(&x3)
	x3.Mul(&x3, &c4)
	x3.Add(&x3, &z)
	if e1 == 1 {
		x.Set(&x1)
	} else {
		x.Set(&x3)
	}
	if e2 == 2 {
		x.Set(&x2)
	}
	gx.Square(&x)
	// gx = gx + A
	gx.Mul(&gx, &x)
	gx.Add(&gx, &bCurveCoeff)
	y.Sqrt(&gx)
	e3 := sign0(u) && sign0(y)
	if !e3 {
		y.Neg(&y)
	}
	res.X.Set(&x)
	res.Y.Set(&y)

	return res
}

// MapToCurveG1Svdw maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-2.2.1
func MapToCurveG1Svdw(t fp.Element) G1Affine {
	res := svdwMapG1(t)
	return res
}

// EncodeToCurveG1Svdw maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-2.2.2
func EncodeToCurveG1Svdw(msg, dst []byte) (G1Affine, error) {
	var res G1Affine
	t, err := hashToFp(msg, dst, 1)
	if err != nil {
		return res, err
	}
	res = MapToCurveG1Svdw(t[0])
	return res, nil
}

// HashToCurveG1Svdw maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-3
func HashToCurveG1Svdw(msg, dst []byte) (G1Affine, error) {
	var res G1Affine
	u, err := hashToFp(msg, dst, 2)
	if err != nil {
		return res, err
	}
	Q0 := MapToCurveG1Svdw(u[0])
	Q1 := MapToCurveG1Svdw(u[1])
	var _Q0, _Q1, _res G1Jac
	_Q0.FromAffine(&Q0)
	_Q1.FromAffine(&Q1)
	_res.Set(&_Q1).AddAssign(&_Q0)
	res.FromJacobian(&_res)
	return res, nil
}

// ----------------------------------------------------------------------------------------
// G2Affine

// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-4.1
// Shallue and van de Woestijne method, works for any elliptic curve in Weierstrass curve
func svdwMapG2(u fptower.E2) G2Affine {

	var res G2Affine

	// constants
	// sage script to find z: https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#appendix-E.1
	var z, c1, c2, c3, c4 fptower.E2
	z.A0.SetString("2")
	z.A1.SetString("1")
	c1.A0.SetString("38")
	c1.A1.SetString("103465770405187637604261093477957413414557405101965864215953705066688187339336329109987555255829344049776128583288")
	c2.A0.SetString("258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458176")
	c2.A1.SetString("129332213006484547005326366847446766768196756377457330269942131333360234174170411387484444069786680062220160729088")
	c3.A0.SetString("141354814213091513024678201993701334402983894315682828752154607295963278662565530559432544731234911957814229457027")
	c3.A1.SetString("105105533309732059517333177844331590042891296673681156116453121436104838422613314245519046656808355505493854902221")
	c4.A0.SetString("86221475337656364670217577898297844512131170918304886846628087555573489449446940924989629379857786708146773819395")
	c4.A1.SetString("137954360540250183472348124637276551219409873469287818954604940088917583119115105479983407007772458733034838111024")

	var tv1, tv2, tv3, tv4, one, x1, gx1, x2, gx2, x3, x, gx, y fptower.E2
	one.SetOne()
	tv1.Square(&u).Mul(&tv1, &c1)
	tv2.Add(&one, &tv1)
	tv1.Sub(&one, &tv1)
	tv3.Mul(&tv2, &tv1).Inverse(&tv3)
	tv4.Mul(&u, &tv1)
	tv4.Mul(&tv4, &tv3)
	tv4.Mul(&tv4, &c3)
	x1.Sub(&c2, &tv4)
	gx1.Square(&x1)
	// 12. gx1 = gx1 + A
	gx1.Mul(&gx1, &x1)
	gx1.Add(&gx1, &bTwistCurveCoeff)
	e1 := gx1.Legendre()
	x2.Add(&c2, &tv4)
	gx2.Square(&x2)
	// 18. gx2 = gx2 + A
	gx2.Mul(&gx2, &x2)
	gx2.Add(&gx2, &bTwistCurveCoeff)
	e2 := gx2.Legendre() - e1 // 2 if is_square(gx2) AND NOT e1
	x3.Square(&tv2)
	x3.Mul(&x3, &tv3)
	x3.Square(&x3)
	x3.Mul(&x3, &c4)
	x3.Add(&x3, &z)
	if e1 == 1 {
		x.Set(&x1)
	} else {
		x.Set(&x3)
	}
	if e2 == 2 {
		x.Set(&x2)
	}
	gx.Square(&x)
	// gx = gx + A
	gx.Mul(&gx, &x)
	gx.Add(&gx, &bTwistCurveCoeff)
	y.Sqrt(&gx)
	e3 := sign0(u.A0) && sign0(y.A0)
	if !e3 {
		y.Neg(&y)
	}
	res.X.Set(&x)
	res.Y.Set(&y)

	return res
}

// MapToCurveG2Svdw maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-2.2.1
func MapToCurveG2Svdw(t fptower.E2) G2Affine {
	res := svdwMapG2(t)
	res.ClearCofactor(&res)
	return res
}

// EncodeToCurveG2Svdw maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-2.2.2
func EncodeToCurveG2Svdw(msg, dst []byte) (G2Affine, error) {
	var res G2Affine
	_t, err := hashToFp(msg, dst, 2)
	if err != nil {
		return res, err
	}
	var t fptower.E2
	t.A0.Set(&_t[0])
	t.A1.Set(&_t[1])
	res = MapToCurveG2Svdw(t)
	return res, nil
}

// HashToCurveG2Svdw maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-3
func HashToCurveG2Svdw(msg, dst []byte) (G2Affine, error) {
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
	Q0 := MapToCurveG2Svdw(u0)
	Q1 := MapToCurveG2Svdw(u1)
	var _Q0, _Q1, _res G2Jac
	_Q0.FromAffine(&Q0)
	_Q1.FromAffine(&Q1)
	_res.Set(&_Q1).AddAssign(&_Q0)
	res.FromJacobian(&_res)
	return res, nil
}
