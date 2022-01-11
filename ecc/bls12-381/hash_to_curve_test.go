package bls12381

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"math/big"
	"strconv"
	"strings"
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

func TestIsogenyG1(t *testing.T) {
	p := G1Affine{
		fp.Element{
			3660217524291093078, 10096673235325531916, 228883846699980880, 13273309082988818590, 5645112663858216297, 1475745906155504807,
		},
		fp.Element{
			7179819451626801451, 8122998708501415251, 10493900036512999567, 8666325578439571587, 1547096619901497872, 644447436619416978,
		},
	}
	p.X.ToMont()
	p.Y.ToMont()

	ref := G1Affine{
		fp.Element{
			15068149172194637577, 9957346779704953421, 14194629579302688285, 14905041577284894537, 12723787027614029596, 1241178457703452833,
		},
		fp.Element{
			8713071345859776370, 18097455281831542002, 18193395493462724643, 6332597957331977118, 3845332352253397392, 1815350252291127063,
		},
	}

	ref.X.ToMont()
	ref.Y.ToMont()

	isogenyG1(&p)

	if ref != p {
		t.Fail()
	}
}

func textToMont(s string) {
	sLines := strings.Split(s, "\n")

	var elem fp.Element

	for lineIndex, sLine := range sLines {
		if sLine == "" {
			continue
		}
		lineSplit := strings.Split(sLine, " = {math/big.Word} ")
		numString := lineSplit[1]
		var err error
		elem[lineIndex], err = strconv.ParseUint(numString, 10, 64)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println(elem)
	elem.ToMont()
	fmt.Println(elem)
}

func TestToMont(t *testing.T) {
	s := []string{
		"0 = {math/big.Word} 7509098555544196687\n1 = {math/big.Word} 4243872485483722269\n2 = {math/big.Word} 14878500680061908427\n3 = {math/big.Word} 16926531971033154030\n4 = {math/big.Word} 1596876708006491832\n5 = {math/big.Word} 838034401176413344",
		"    0 = {math/big.Word} 6891547135157885641\n1 = {math/big.Word} 6138139758682421950\n2 = {math/big.Word} 6936729421058986545\n3 = {math/big.Word} 10572248604636930284\n4 = {math/big.Word} 8752427191448882401\n5 = {math/big.Word} 115857565692138085",
	}

	for _, e := range s {
		textToMont(e)
	}
}

/*func TestMapToCurveG1SSWU(t *testing.T) {
	MapToCurveSSWU
}*/

/*func TestMapToCurveG1SSWU(t *testing.T) {
	Q := sswuMapG1(&fp.Element{941031641141724048, 7593419090796165139, 13447299832369701844, 7664570780628181207, 16833839340160123079, 332469494419187881})
	expected := G1Affine{
		fp.Element{4563475290293962576, 3982361128921378550, 16152256253200838243, 12773063786225987449, 2858682674850780732, 785746258097921522},
		fp.Element{10945627526752281529, 13120484463621343084, 500907696078610998, 17841918537664625985, 667297683540361872, 773042732898677554},
	}

	if Q != expected {
		t.Fail()
	}
}*/

func TestEncodeToCurveG1SSWU(t *testing.T) {
	dst := "QUUX-V01-CS02-with-BLS12381G1_XMD:SHA-256_SSWU_NU_"
	seen, err := EncodeToCurveG1SSWU([]byte{}, []byte(dst))
	if err != nil {
		t.Fatal(err)
	}

	expectedP := G1Affine{
		fp.Element{4508701981465676087, 16014981725829343206, 1429121664596480851, 16754737785772897928, 14176845108067946534, 575224408015977794},
		fp.Element{4886454369921712624, 10597955738183899813, 11346608665277124313, 7940767554533245898, 16448266045496148945, 285064240491496456},
	}

	if seen != expectedP {
		t.Fail()
	}
}

func TestHashToCurveG1SSWU(t *testing.T) {
	dst := "QUUX-V01-CS02-with-BLS12381G1_XMD:SHA-256_SSWU_RO_"
	seen, err := HashToCurveG1SSWU([]byte{}, []byte(dst))
	if err != nil {
		t.Fatal(err)
	}

	var x fp.Element
	var y fp.Element

	x.SetHex("052926add2207b76ca4fa57a8734416c8dc95e24501772c814278700eed6d1e4e8cf62d9c09db0fac349612b759e79a1")
	y.SetHex("08ba738453bfed09cb546dbb0783dbb3a5f1f566ed67bb6be0e8c67e2e81a4cc68ee29813bb7994998f3eae0c9c6a265")

	expectedP := G1Affine{x, y}

	if seen != expectedP {
		t.Fail()
	}
}
