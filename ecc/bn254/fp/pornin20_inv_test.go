package fp

import (
	"fmt"
	"math/big"
	"testing"
)

func testBigNumMult(a *Element, c int64) {
	var aHi uint64
	var aTimes Element
	aHi = aTimes.bigNumMultiply(a, c)

	checkMult(a, c, &aTimes, aHi)
}

func TestBigNumNeg(t *testing.T) {
	var a = Element{0, 0, 0, 0}
	aHi := a.bigNumNeg(&a, 0)
	if !a.IsZero() || aHi != 0 {
		panic("not zero")
	}
}

func TestSparseMult(t *testing.T) {
	var a = Element{
		0,
		0,
		0,
		0,
	}

	testBigNumMult(&a, -1) //aka 4
}

func TestBigNumMultPositive(t *testing.T) {
	var a = Element{
		4332616871279656263,
		10917124144477883021,
		13281191951274694749,
		3486998266802970665,
	}
	testBigNumMult(&a, 1496528)
}

func TestVeryBigIntConversion(t *testing.T) {
	xHi := uint64(18446744073687286931)
	x := Element{
		15230403791020821917,
		754611498739239741,
		7381016538464732716,
		1011752739694698287,
	}
	var xInt big.Int
	x.ToVeryBigInt(&xInt, xHi)
	checkMatchBigInt(&x, xHi, &xInt)
}

func TestBigNumAddition(t *testing.T) {
	xHi := uint64(18446744073687286931)
	x := Element{
		15230403791020821917,
		754611498739239741,
		7381016538464732716,
		1011752739694698287,
	}
	var xInt big.Int
	x.ToVeryBigInt(&xInt, xHi)

	yHi := uint64(22264684)
	y := Element{
		4332616871279656263,
		10917124144477883021,
		13281191951274694749,
		3486998266802970665,
	}
	var yInt big.Int
	y.ToVeryBigInt(&yInt, yHi)

	var sumInt big.Int
	sumInt.Add(&xInt, &yInt)

	var sum Element
	sumHi := sum.bigNumAdd(&x, xHi, &y, yHi)

	checkMatchBigInt(&sum, sumHi, &sumInt)
}

func TestComputeCorrectiveFactor(t *testing.T) {
	//ComputeCorrectiveFactorPornin1()

	var c Element
	c.SetOne()
	c.Inverse(&c)
	c.InverseOld(&c)

	fmt.Println(c)

	var one Element
	one.SetOne()
	if !c.Equal(&one) {
		panic("Not one")
	}
}

/*func TestRshNegCommutation(t *testing.T) {
	b := Element {
		0xAEB6AECF80000000,
		8552640749534187906,
		9455549699771608402,
		17692725854587193526,
	}
	bHi := uint64(0xFFFFFFFFFFFFFFFF)

	var bNeg Element
	bNegHi := bNeg.bigNumNeg(&b, bHi)

	b.bigNumRshBy31(&b, bHi)
	bNeg.bigNumRshBy31(&bNeg, bNegHi)

	bHi = b.bigNumAdd(&bNeg, bNegHi, &b, bHi)
}*/

func TestLinearComb(t *testing.T) {
	f1 := int64(405940026)
	g1 := int64(-117783518)
	a := Element{
		15230403791020821917,
		754611498739239741,
		7381016538464732716,
		1011752739694698287,
	}
	b := Element{
		4332616871279656263,
		10917124144477883021,
		13281191951274694749,
		3486998266802970665,
	}
	bHi := b.bigNumLinearComb(&a, f1, &b, g1)
	print(bHi)
}

func TestElementApproximation(t *testing.T) {
	superCorrect := uint64(0b1101101100010011111011100011101010011100010101111101010001001)
	correct := approximateRef(&rSquare)
	observed := approximate(&rSquare)

	if correct != superCorrect {
		panic("oof")
	}

	if observed != correct || observed != superCorrect || correct != superCorrect {
		panic("oops")
	}
}

func TestMulWord(t *testing.T) {
	var prodFast Element
	var prodRef Element
	var x = Element{
		9999999999999999999,
		9999999999999999998,
		9999999999999999997,
		234254345,
	}

	var yWord int64 = 4999999999999999996

	var y = Element{
		uint64(yWord),
		0,
		0,
		0,
	}
	prodFast.MulWord(&x, yWord)
	prodRef.Mul(&x, &y)
	if prodFast.Equal(&prodRef) {
		print("Good\n")
	} else {
		panic("Oop")
	}
}

func TestRsh(t *testing.T) {
	a := Element{
		14577615541645606912,
		8737333314812511136,
		5915853752549640268,
		0b101110100101101110000001010110101110101001100111000110111001001,
	}
	aHi := uint64(0x111)

	var aInt big.Int
	a.ToVeryBigInt(&aInt, aHi)

	aInt.Rsh(&aInt, 31)
	a.bigNumRshBy31(&a, aHi)

	checkMatchBigInt(&a, 0, &aInt)
}

func TestRshSmall(t *testing.T) {
	a := Element{
		0,
		1 << 30,
		0,
		0,
	}
	aHi := uint64(0)

	a.bigNumRshBy31(&a, aHi)

	if a[0] != 1<<63 {
		panic("wrong")
	}
}

func TestMulWord2(t *testing.T) {
	var u = Element{
		15524365416767025468,
		8999220800619366266,
		17035922559114310297,
		834761565329023031,
	}

	var coeff int64 = 71931499032677
	var coeffElem = Element{
		uint64(coeff),
		0,
		0,
		0,
	}

	var quickRes Element
	var correctRes Element

	correctRes.Mul(&u, &coeffElem)
	quickRes.MulWord(&u, coeff)

	if !quickRes.Equal(&correctRes) {
		panic("Multiplication failed")
	}
}
