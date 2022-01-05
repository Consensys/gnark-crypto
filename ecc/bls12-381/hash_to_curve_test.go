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
			3660217524291093078,
			10096673235325531916,
			228883846699980880,
			13273309082988818590,
			5645112663858216297,
			1475745906155504807,
		},
		fp.Element{
			7179819451626801451,
			8122998708501415251,
			10493900036512999567,
			8666325578439571587,
			1547096619901497872,
			644447436619416978,
		},
	}
	p.X.ToMont()
	p.Y.ToMont()

	ref := G1Affine{
		fp.Element{
			15068149172194637577,
			9957346779704953421,
			14194629579302688285,
			14905041577284894537,
			12723787027614029596,
			1241178457703452833,
		},
		fp.Element{
			8713071345859776370,
			18097455281831542002,
			18193395493462724643,
			6332597957331977118,
			3845332352253397392,
			1815350252291127063,
		},
	}

	ref.X.ToMont()
	ref.Y.ToMont()

	after1Iteration := fp.Element{
		12234233096825917993,
		14000314106239596831,
		2576269112800734056,
		3591512392926246844,
		13627494601717575229,
		495550047642324592,
	}
	after1Iteration.ToMont()

	after2Iterations := fp.Element{
		5111946664532277196,
		14549320335678844073,
		10185424646563936348,
		10364927776233079091,
		575272455112125070,
		2461664960513112322,
	}
	after2Iterations.ToMont()

	IsogenyG1(&p)

	if ref != p {
		t.Fail()
	}
}

func TestToMont(t *testing.T) {
	s := []string{
		"0 = {math/big.Word} 3055001050381004142\n1 = {math/big.Word} 17980914612811311061\n2 = {math/big.Word} 8932722884053879865\n3 = {math/big.Word} 16243047670605951396\n4 = {math/big.Word} 10578260746861642877\n5 = {math/big.Word} 19218141351574462",
		"0 = {math/big.Word} 1178783500747518065\n1 = {math/big.Word} 8802523577013696646\n2 = {math/big.Word} 5666989223581262097\n3 = {math/big.Word} 16755304854723248611\n4 = {math/big.Word} 8411140328324188223\n5 = {math/big.Word} 1868411116419964690",
		"0 = {math/big.Word} 13368360334501362571\n1 = {math/big.Word} 3466838909926296255\n2 = {math/big.Word} 3143139544335844302\n3 = {math/big.Word} 17561966084202659185\n4 = {math/big.Word} 4305537054398371723\n5 = {math/big.Word} 1373031427352947672",
		"0 = {math/big.Word} 417417237046423156\n1 = {math/big.Word} 323280268401769046\n2 = {math/big.Word} 742273599641433786\n3 = {math/big.Word} 17920511521503291334\n4 = {math/big.Word} 15975955140407537332\n5 = {math/big.Word} 1115541194573543015",
		"0 = {math/big.Word} 17283442855767433160\n1 = {math/big.Word} 5116804784925790422\n2 = {math/big.Word} 14651117517718964762\n3 = {math/big.Word} 10793220149269005514\n4 = {math/big.Word} 14674073047365171490\n5 = {math/big.Word} 384489558977295275",
		"0 = {math/big.Word} 13987681225706524916\n1 = {math/big.Word} 3881103106504103101\n2 = {math/big.Word} 2177276969163821573\n3 = {math/big.Word} 95320324557271057\n4 = {math/big.Word} 12253028289903615055\n5 = {math/big.Word} 258573728199572940",
		"0 = {math/big.Word} 2146020182223758245\n1 = {math/big.Word} 11293268164564994731\n2 = {math/big.Word} 17831260093212941686\n3 = {math/big.Word} 15409755650773875361\n4 = {math/big.Word} 992681825351972855\n5 = {math/big.Word} 885629365910324847",
		"0 = {math/big.Word} 11293569701532279294\n1 = {math/big.Word} 9207835367879847341\n2 = {math/big.Word} 11420920133902317100\n3 = {math/big.Word} 15610793020886052853\n4 = {math/big.Word} 13359797796107525129\n5 = {math/big.Word} 760587961584440567",
		"0 = {math/big.Word} 4865350408643357178\n1 = {math/big.Word} 6958865257497711200\n2 = {math/big.Word} 5000823708216313329\n3 = {math/big.Word} 6086643723040764324\n4 = {math/big.Word} 2182111210109514540\n5 = {math/big.Word} 1411529777535521911",
		"0 = {math/big.Word} 9438903658083346336\n1 = {math/big.Word} 8359155498453587497\n2 = {math/big.Word} 6766915389702346032\n3 = {math/big.Word} 18161690170262915953\n4 = {math/big.Word} 13732282459177049689\n5 = {math/big.Word} 1046646928620882810",
	}

	for _, e := range s {
		textToMont(e)
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
