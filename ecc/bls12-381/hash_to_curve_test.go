package bls12381

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"math/big"
	"testing"
)

func TestComputeC2(t *testing.T) {
	var z fp.Element
	z.SetUint64(Z)

	var zP fp.Element
	zP.Neg(&z)
	zP.Sqrt(&zP)

	//[14304544101977590919 3350176034073442437 17582609757678985529 1309042698909992113 4737065203462589718 1706412243078167948]
	fmt.Println(zP)

	zP.Square(&zP)
	zP.Add(&zP, &z)

	if !zP.IsZero() {
		t.Fail()
	}
}

func TestComputeC2Int(t *testing.T) {
	z := big.NewInt(-Z)
	z.ModSqrt(z, fp.Modulus())
	fmt.Println(z)

	z.Mul(z, z)
	z.Add(z, big.NewInt(Z))
	z.Mod(z, fp.Modulus())
	if z.BitLen() != 0 {
		t.Fail()
	}
}

func TestComputeC1Int(t *testing.T) {
	c1 := fp.Modulus()
	c1.Rsh(c1, 2)
	fmt.Println(c1)

	c1.Lsh(c1, 2)
	c1.Add(c1, big.NewInt(3))

	if c1.Cmp(fp.Modulus()) != 0 {
		t.Fail()
	}
}

func TestSqrtRatio(t *testing.T) {
	testSqrtRatio(&fp.Element{3752852834233450803, 10015304229637369378, 6482406239105581310, 1802624635905610022, 11716583840524549243, 1670704604553607051}, &fp.Element{16538149341274582162, 2654217574689430748, 4191868356445146499, 16611300210497698397, 10619697645702806389, 130786230622822284}, t)
	testSqrtRatio(&fp.Element{0}, &fp.Element{1}, t)
	testSqrtRatio(&fp.Element{1}, &fp.Element{1}, t)

	for i := 0; i < 1000; i++ {
		var u fp.Element
		var v fp.Element
		u.SetRandom()
		v.SetRandom()
		testSqrtRatio(&u, &v, t)
	}
}

func testSqrtRatio(u *fp.Element, v *fp.Element, t *testing.T) {
	var ref fp.Element
	ref.Div(u, v)
	var qrRef bool
	if ref.Legendre() == -1 {
		ref.MulByConstant(Z)
		qrRef = false
	} else {
		qrRef = true
	}
	ref.Sqrt(&ref)

	var seen fp.Element
	qr := sqrtRatio(&seen, u, v)

	if qr != qrRef || seen != ref {
		seen.Div(&ref, &seen)
		fmt.Println(seen)
		t.Error(*u, *v)
	}
}

func TestMulByConstant(t *testing.T) {

	for test := 0; test < 100; test++ {
		var x fp.Element
		x.SetRandom()

		y := x

		var yP fp.Element

		y.MulByConstant(11)

		for i := 0; i < 11; i++ {
			yP.Add(&yP, &x)
		}

		if y != yP {
			t.Fail()
		}

	}
}
