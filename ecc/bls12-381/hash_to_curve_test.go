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

	IsogenyG1(&p)

	if ref != p {
		t.Fail()
	}
}

func TestToMont(t *testing.T) {
	s := []string{
		"0 = {math/big.Word} 1\n1 = {math/big.Word} 0\n2 = {math/big.Word} 0\n3 = {math/big.Word} 0\n4 = {math/big.Word} 0\n5 = {math/big.Word} 0", "0 = {math/big.Word} 13610436265939526458\n1 = {math/big.Word} 14561699186956717708\n2 = {math/big.Word} 14865066093874793548\n3 = {math/big.Word} 14231665274560599601\n4 = {math/big.Word} 7953345632490515027\n5 = {math/big.Word} 615153679158255179", "0 = {math/big.Word} 3043931179675399695\n1 = {math/big.Word} 4878853097464208641\n2 = {math/big.Word} 11144953449459701748\n3 = {math/big.Word} 12799317866228993986\n4 = {math/big.Word} 6944570897389024363\n5 = {math/big.Word} 667881774812630462", "0 = {math/big.Word} 3865851310648910679\n1 = {math/big.Word} 12501907331811675599\n2 = {math/big.Word} 12238389446615508914\n3 = {math/big.Word} 13704265888286126455\n4 = {math/big.Word} 6553643744354805799\n5 = {math/big.Word} 1213699486410940798", "0 = {math/big.Word} 3840949932698093167\n1 = {math/big.Word} 4009333284470593689\n2 = {math/big.Word} 17357686029068793032\n3 = {math/big.Word} 1646084630507526823\n4 = {math/big.Word} 8867372532604559818\n5 = {math/big.Word} 223781729648735329", "0 = {math/big.Word} 14813126532147369080\n1 = {math/big.Word} 5868165255598339622\n2 = {math/big.Word} 1121207778909989363\n3 = {math/big.Word} 1916248906037787718\n4 = {math/big.Word} 10795077326714211317\n5 = {math/big.Word} 695812133980925453", "0 = {math/big.Word} 18387959409564224282\n1 = {math/big.Word} 14840689404063775002\n2 = {math/big.Word} 6937378136204647022\n3 = {math/big.Word} 16124506650899193496\n4 = {math/big.Word} 10018794535072793499\n5 = {math/big.Word} 256990648642259228", "0 = {math/big.Word} 9391008319020907154\n1 = {math/big.Word} 9995698222462547125\n2 = {math/big.Word} 12452769355443485242\n3 = {math/big.Word} 4007096256474708190\n4 = {math/big.Word} 8046360065487399253\n5 = {math/big.Word} 1333658353946334127", "0 = {math/big.Word} 1326156799764198676\n1 = {math/big.Word} 2227120228414737348\n2 = {math/big.Word} 2997801608460978904\n3 = {math/big.Word} 3293753599947951320\n4 = {math/big.Word} 15434944276066759143\n5 = {math/big.Word} 1342656425708745322", "0 = {math/big.Word} 5564714553045267272\n1 = {math/big.Word} 10184335806762559071\n2 = {math/big.Word} 11908592643195788647\n3 = {math/big.Word} 1408397412514387242\n4 = {math/big.Word} 12526651621772595867\n5 = {math/big.Word} 614894643076346113", "0 = {math/big.Word} 1981405961318803243\n1 = {math/big.Word} 8721076847999834889\n2 = {math/big.Word} 8023158070544658029\n3 = {math/big.Word} 9384048753164702677\n4 = {math/big.Word} 13846846214150601562\n5 = {math/big.Word} 820924745816190625", "0 = {math/big.Word} 5378519619368490326\n1 = {math/big.Word} 2542298885948939279\n2 = {math/big.Word} 1035494106593601068\n3 = {math/big.Word} 4123540199000133648\n4 = {math/big.Word} 11599759719410769113\n5 = {math/big.Word} 521334752535042167", "0 = {math/big.Word} 11958932834865370824\n1 = {math/big.Word} 11554582496232138914\n2 = {math/big.Word} 1146065373692080960\n3 = {math/big.Word} 16012727719879505514\n4 = {math/big.Word} 12016547727149151797\n5 = {math/big.Word} 1126043410829816789", "0 = {math/big.Word} 8136075262792960945\n1 = {math/big.Word} 16842212521062443814\n2 = {math/big.Word} 14930771960835430251\n3 = {math/big.Word} 6578314546905007417\n4 = {math/big.Word} 9965336544922366599\n5 = {math/big.Word} 853639996198389885", "0 = {math/big.Word} 11132792215270049017\n1 = {math/big.Word} 16993842496309820931\n2 = {math/big.Word} 3671995413309298753\n3 = {math/big.Word} 17747748751803759251\n4 = {math/big.Word} 15314888350765616635\n5 = {math/big.Word} 1590750455491226264", "0 = {math/big.Word} 3522027620651039279\n1 = {math/big.Word} 11421432582115914769\n2 = {math/big.Word} 11664931178663920988\n3 = {math/big.Word} 11509243117701410857\n4 = {math/big.Word} 9094218744705950154\n5 = {math/big.Word} 187403685399302949",
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
