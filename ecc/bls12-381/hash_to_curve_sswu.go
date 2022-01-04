package bls12381

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"math/big"
)

// From https://eprint.iacr.org/2019/403 by Wahby and Boneh
// compatible with https://datatracker.ietf.org/doc/draft-irtf-cfrg-hash-to-curve/ TODO: Is it?

// Using the isogenous curve E1': y² = x³ + a₁ x + b₁
//TODO: Montgomery form

const a10 uint64 = 0x5cf428082d584c1d
const a11 uint64 = 0x98936f8da0e0f97f
const a12 uint64 = 0xd8e8981aefd881ac
const a13 uint64 = 0xb0ea985383ee66a8
const a14 uint64 = 0x3d693a02c96d4982
const a15 uint64 = 0x144698a3b8e943

const b10 uint64 = 0xd1cc48e98e172be0
const b11 uint64 = 0x5a23215a316ceaa5
const b12 uint64 = 0xa0b9c14fcef35ef5
const b13 uint64 = 0x2016c1f0f24f4070
const b14 uint64 = 0x018b12e8753eee3b
const b15 uint64 = 0x12e2908d11688030

//EFFECTIVE h?
const hEff uint64 = 0xd201000000010001

// From https://datatracker.ietf.org/doc/draft-irtf-cfrg-hash-to-curve/ Pg 80
func sswuMapG1(u *fp.Element) G1Jac {

	var tv1 fp.Element
	tv1.Square(u)

	//mul tv1 by Z
	tv1.MulByConstant(Z)

	var tv2 fp.Element
	tv2.Square(&tv1)
	tv2.Add(&tv2, &tv1)

	var tv3 fp.Element
	//Standard doc line 5
	tv3.Add(&tv2, &fp.Element{1}) //TODO: Optimize? I think not
	tv3.Mul(&tv3, &fp.Element{b10, b11, b12, b13, b14, b15})

	tv4 := fp.Element{a10, a11, a12, a13, a14, a15}
	//TODO: Std doc uses conditional move. If-then-else good enough here?
	if tv2.IsZero() {
		tv4.MulByConstant(Z) //WARNING: this takes less time
	} else {
		tv4.Mul(&tv4, &tv2)
		tv4.Neg(&tv4)
	}
	tv2.Square(&tv3)

	var tv6 fp.Element
	//Standard doc line 10
	tv6.Square(&tv4)

	var tv5 fp.Element
	tv5.Mul(&tv6, &fp.Element{a10, a11, a12, a13, a14, a15})

	tv2.Add(&tv2, &tv5)
	tv2.Add(&tv2, &tv3)
	tv6.Mul(&tv6, &tv4)

	//Standards doc line 15
	tv5.Mul(&tv6, &fp.Element{b10, b11, b12, b13, b14, b15})
	tv2.Add(&tv2, &tv5)

	var x fp.Element
	x.Mul(&tv1, &tv3)

	//var y1 fp.Element
	return G1Jac{}
}

const Z = 11

// sqrtRatio computes the square root of u/v and returns true if u/v was indeed a quadratic residue
// if not, we get sqrt(Z * u / v). Recall that Z is non-residue
// Taken from https://datatracker.ietf.org/doc/draft-irtf-cfrg-hash-to-curve/ F.2.1.2. q = 3 mod 4
// The main idea is that since the computation of the square root involves taking large powers of u/v, the inversion of v can be avoided
func sqrtRatio(z *fp.Element, u *fp.Element, v *fp.Element) bool {
	var tv1 fp.Element
	tv1.Square(v)
	var tv2 fp.Element
	tv2.Mul(u, v)
	tv1.Mul(&tv1, &tv2)

	var y1 fp.Element
	expByC2(&y1, &tv1)
	y1.Mul(&y1, &tv2)

	var y2 fp.Element
	// y2 = y1 * c2
	// TODO: c2 value worked out experimentally. Derive it correctly using bigInt ops
	y2.Mul(&y1, &fp.Element{14304544101977590919, 3350176034073442437, 17582609757678985529, 1309042698909992113, 4737065203462589718, 1706412243078167948})

	var tv3 fp.Element
	tv3.Square(&y1)
	tv3.Mul(&tv3, v)
	isQr := tv3 == *u //TODO: == or .Equals ?

	if isQr {
		*z = y1
	} else {
		*z = y2
	}

	return isQr
}

//TODO: Use https://github.com/mmcloughlin/addchain for addition chain
func expByC2(z *fp.Element, x *fp.Element) {
	var c2 big.Int
	c2.SetString("1000602388805416848354447456433976039139220704984751971333014534031007912622709466110671907282253916009473568139946", 10)
	if x == z {
		panic("Writing to z will overwrite x")
	}
	z.SetOne()
	for i := c2.BitLen() - 1; i >= 0; i-- {
		z.Square(z)
		if c2.Bit(i) != 0 {
			z.Mul(z, x)
		}
	}
}
