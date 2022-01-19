package bls12381

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func TestMulByConstant(t *testing.T) {

	for test := 0; test < 100; test++ {
		var x fp.Element
		x.SetRandom()

		y := x

		var yP fp.Element

		fp.MulBy11(&y)

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

func BenchmarkG1EncodeToCurveSSWU(b *testing.B) {
	const size = 54
	bytes := make([]byte, size)
	dst := []byte("QUUX-V01-CS02-with-BLS12381G1_XMD:SHA-256_SSWU_NU_")

	for i := 0; i < 100000; i++ {

		bytes[rand.Int()%size] = byte(rand.Int())

		if _, err := EncodeToCurveG1SSWU(bytes, dst); err != nil {
			b.Fail()
		}
	}
}

func BenchmarkG1HashToCurveSSWU(b *testing.B) {
	const size = 54
	bytes := make([]byte, size)
	dst := []byte("QUUX-V01-CS02-with-BLS12381G1_XMD:SHA-256_SSWU_RO_")

	for i := 0; i < 100000; i++ {

		bytes[rand.Int()%size] = byte(rand.Int())

		if _, err := HashToCurveG1SSWU(bytes, dst); err != nil {
			b.Fail()
		}
	}
}
