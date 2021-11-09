package fp

import (
	"fmt"
	"math/big"
	"testing"
)

func TestComputeCorrectiveFactor(t *testing.T) {

	var c Element
	computeCorrectiveFactor(&c)

	fmt.Println(c)

	var one Element
	one.SetOne()
	if !c.Equal(&one) {
		i := c.log(&rSquare, 8)
		fmt.Println("Result is rSq^", i)
		panic("Not one")
	}
}

func TestCorrectiveFactorConsistency(t *testing.T) {
	var correctiveFactor Element
	computeCorrectiveFactor(&correctiveFactor)

	a := Element{239472382928373468, 3242934823798534, 345984723476857987, 23239348948234376} //TODO: randomization by banging on keyboard, replace with something better

	var aInv Element
	aInv.Inverse(&a)
	aInv.Mul(&aInv, &correctiveFactor)
	aInv.Mul(&aInv, &a)

	var one Element
	one.SetOne()
	if !aInv.Equal(&one) {
		panic("Not one")
	}
}

func TestComputeAllCorrectionFactors(t *testing.T) {
	var b int64 = 0x4000000000000000
	factor := Element{5743661648749932980, 12551916556084744593, 23273105902916091, 802172129993363311}
	for i := 0; i < 8; i++ {
		factor.MulWord(&factor, b)
		fmt.Println(factor)
	}
}

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

func computeCorrectiveFactor(c *Element) {
	c.SetOne()
	c.Inverse(c)
	c.InverseOld(c)
}

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
	superCorrect := uint64(0b1101101100010011111011100011100101010011100010101111101010001001)

	correct := approximateRef(&rSquare)
	if correct != superCorrect {
		panic("oof")
	}

	var xInt big.Int
	for rSquare.ToBigInt(&xInt); xInt.BitLen() != 0; xInt.Rsh(&xInt, 1) {

		var x Element
		x.SetBigInt(&xInt)
		observed := approximate(&x, x.BitLen())
		correct = approximateRef(&x)

		if observed != correct {
			fmt.Println("At bit length ", xInt.BitLen())
			panic("oops")
		}
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

/*func TestFindCorrectiveFactorDlog(t *testing.T) {
	i := inversionCorrectionFactorP20Full.log(&rSquare, 40)
	fmt.Println(i)
}*/

func approximateRef(x *Element) uint64 {

	var asInt big.Int
	x.ToBigInt(&asInt)
	n := x.BitLen()

	if n <= 64 {
		return asInt.Uint64()
	}

	modulus := big.NewInt(1 << 31)
	var lo big.Int
	lo.Mod(&asInt, modulus)

	modulus.Lsh(modulus, uint(n-64))
	var hi big.Int
	hi.Div(&asInt, modulus)
	hi.Lsh(&hi, 31)

	hi.Add(&hi, &lo)
	return hi.Uint64()
}

//------

func checkMult(x *Element, c int64, result *Element, resultHi uint64) big.Int {
	var xInt big.Int
	x.ToBigInt(&xInt)

	xInt.Mul(&xInt, big.NewInt(c))

	checkMatchBigInt(result, resultHi, &xInt)
	return xInt
}

func checkMatchBigInt(a *Element, aHi uint64, aInt *big.Int) {
	var modulus big.Int
	var aIntMod big.Int
	modulus.SetInt64(1)
	modulus.Lsh(&modulus, 320)

	aIntMod.Mod(aInt, &modulus)

	bytes := aIntMod.Bytes()

	for i := 0; i < 33; i++ {
		var word uint64
		if i < 32 {
			word = a[i/8]
		} else {
			word = aHi
		}

		i2 := (i % 8) * 8
		byteA := byte(((255 << i2) & word) >> i2)
		var byteInt byte
		if i < len(bytes) {
			byteInt = bytes[len(bytes)-i-1]
		} else {
			byteInt = 0
		}

		if byteInt != byteA {
			panic("Bignum mismatch")
		}
	}
}

func (z *Element) ToVeryBigInt(i *big.Int, xHi uint64) {
	z.ToBigInt(i)
	var upperWord big.Int
	upperWord.SetUint64(xHi)
	upperWord.Lsh(&upperWord, 256)
	i.Add(&upperWord, i)
}

func (z *Element) log(base *Element, max uint) int {
	var bInv Element
	var current Element
	var currentInv Element

	current.SetOne()
	currentInv.SetOne()
	bInv.InverseOld(base)

	for i := 0; i < int(max); i++ {
		if current.Equal(z) {
			return i
		}
		if currentInv.Equal(z) {
			return -i
		}
		if i < int(max-1) {
			current.Mul(&current, base)
			currentInv.Mul(&currentInv, &bInv)
		}
	}
	return 1 //not found
}
