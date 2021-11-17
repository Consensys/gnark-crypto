package fp

import (
	_ "crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

func TestEuclideanAlgo(t *testing.T) {
	//q:= Modulus()
	var qInvNeg big.Int
	//qInvNegLsb.SetString("-4759646384140481320982610724935209484903937857060724391493050186936685796471", 10)
	qInvNeg.SetString("111032442853175714102588374283752698368366046808579839647964533820976443843465", 10)
	//var rInv big.Int
	//rInv.SetString("-899718596722274150243595920809187510076580371697509328435252918265935168272", 10)

	r := big.NewInt(1)
	r.Lsh(r, 256)

	q := Modulus()

	var u big.Int
	u.Mul(q, &qInvNeg)
	fmt.Println(u.String())
	u.Add(&u, big.NewInt(1))
	fmt.Println(u, u.String())

	qInvNeg.Mod(&qInvNeg, r)
	fmt.Println("Reduced qInv", qInvNeg, qInvNeg.String())
	u.Mul(q, &qInvNeg)
	fmt.Println(u.String())
	u.Add(&u, big.NewInt(1))
	fmt.Println(u, u.String())

}

func TestMontReduce(t *testing.T) {
	var xInt big.Int
	xInt.SetString("1518345043075282886718915476446629923034923247403426348876984432860252403179691687682438634393061", 10)
	var xRedInt big.Int
	xRedInt.SetString("635522307507233077051145701706764522087344188384121418491036574122751191340", 10)

	var x = Element{5632619021051350501, 15614519151907775586, 16162948170853994679, 14978651708485828588}
	var xHi uint64 = 13112683716790278092

	x.montReduce(&x, xHi)
	checkMatchBigInt(&x, 0, &xRedInt)
}

func TestMontReduceRef(t *testing.T) {
	q := Modulus()
	var r big.Int
	r.SetInt64(1)
	r.Lsh(&r, 256)

	var x big.Int
	/*{
		y, _ := rand.Int(rand.Reader, &r)
		x = *y
		fmt.Println(x)
	}*/

	x.SetString("1518345043075282886718915476446629923034923247403426348876984432860252403179691687682438634393061", 10)

	var u big.Int
	montReduceRef(&u, &x)

	fmt.Println(u, u.String())

	var ur big.Int
	ur.Mul(&u, &r)
	ur.Sub(&ur, &x)
	ur.Mod(&ur, q)

	if ur.BitLen() != 0 {
		panic("Mismatch")
	}
}

func montReduceRef(u *big.Int, t *big.Int) {
	q := Modulus()
	var qInvNeg big.Int
	/*_qInvNeg := Element{9786893198990664585, 11447725176084130505, 15613922527736486528, 17688488658267049067}
	_qInvNeg.ToBigInt(&qInvNegLsb)*/
	qInvNeg.SetString("111032442853175714102588374283752698368366046808579839647964533820976443843465", 10)
	r := big.NewInt(1)
	r.Lsh(r, 256)

	var m big.Int
	m.Mul(t, &qInvNeg)
	m.Mod(&m, r)

	u.Mul(&m, q)
	u.Add(u, t)
	u.Div(u, r)

	if u.Cmp(q) >= 0 {
		u.Sub(u, q)
	}
}

func BenchmarkElementInverseNew(b *testing.B) {
	var x Element
	x.SetString("9537083524586879850302283710748940119696335591071039437516223462175307793360")

	// b.Run("inverseOld", func(b *testing.B) {
	// 	b.ResetTimer()
	// 	for i := 0; i < b.N; i++ {
	// 		benchResElement.InverseOld(&x)
	// 	}
	// })

	b.Run("inverseNew", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchResElement.Inverse(&x)
		}
	})

}

//Copied from field.go
// https://en.wikipedia.org/wiki/Extended_Euclidean_algorithm
// r > q, modifies rinv and qinv such that rinv.r - qinv.q = 1
func extendedEuclideanAlgo(r, q, rInv, qInv *big.Int) {
	var s1, s2, t1, t2, qi, tmpMuls, riPlusOne, tmpMult, a, b big.Int
	t1.SetUint64(1)
	rInv.Set(big.NewInt(1))
	qInv.Set(big.NewInt(0))
	a.Set(r)
	b.Set(q)

	// r_i+1 = r_i-1 - q_i.r_i
	// s_i+1 = s_i-1 - q_i.s_i
	// t_i+1 = t_i-1 - q_i.s_i
	for b.Sign() > 0 {
		qi.Div(&a, &b)
		riPlusOne.Mod(&a, &b)

		tmpMuls.Mul(&s1, &qi)
		tmpMult.Mul(&t1, &qi)

		s2.Set(&s1)
		t2.Set(&t1)

		s1.Sub(rInv, &tmpMuls)
		t1.Sub(qInv, &tmpMult)
		rInv.Set(&s2)
		qInv.Set(&t2)

		a.Set(&b)
		b.Set(&riPlusOne)
	}
	qInv.Neg(qInv)
}

func toUint64Slice(b *big.Int, nbWords ...int) (s []uint64) {
	if len(nbWords) > 0 && nbWords[0] > len(b.Bits()) {
		s = make([]uint64, nbWords[0])
	} else {
		s = make([]uint64, len(b.Bits()))
	}

	for i, v := range b.Bits() {
		s[i] = (uint64)(v)
	}
	return
}

func TestComputeMontConstants(t *testing.T) {
	// taken from field.go
	_r := big.NewInt(1)
	_r.Lsh(_r, 256)
	_rInv := big.NewInt(1)
	_qInv := big.NewInt(0)
	extendedEuclideanAlgo(_r, Modulus(), _rInv, _qInv)
	fmt.Println("qInv", _qInv)
	fmt.Println("rInv", _rInv)

	_qInv.Mod(_qInv, _r)
	qInv := toUint64Slice(_qInv, 256)

	/*_r.Mul(_rInv, _r)
	_qInv.Mul(_qInv, Modulus())
	_r.Sub(_r, _qInv)

	if _r.Cmp(big.NewInt(1)) != 0 {
		panic("Not inverses?")
	}*/
	rInv := _rInv.Bits()

	fmt.Println("qInv", qInv)
	fmt.Println("rInv", _rInv.Sign(), rInv)
}

func testLinearComb(x *Element, xC int64, y *Element, yC int64) {
	var z Element
	z.linearComb(x, xC, y, yC)

	var p1, p2 Element
	p1.mulWSigned(x, xC)
	p2.mulWSigned(y, yC)
	p1.Add(&p1, &p2)

	if !p1.Equal(&z) {
		panic("mismatch")
	}
}

func TestLinearComb(t *testing.T) {
	testLinearComb(&Element{1, 0, 0, 0}, -871749566700742666, &Element{0, 0, 0, 0}, 252938184923674574)
}

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
		factor.mulWSigned(&factor, b)
		fmt.Println(factor)
	}
}

func testBigNumMult(a *Element, c int64) {
	var aHi uint64
	var aTimes Element
	aHi = aTimes.mulWRegular(a, c)

	checkMult(a, c, &aTimes, aHi)
}

func TestBigNumNeg(t *testing.T) {
	var a = Element{0, 0, 0, 0}
	aHi := a.neg(&a, 0)
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
	sumHi := sum.add(&x, xHi, &y, yHi)

	checkMatchBigInt(&sum, sumHi, &sumInt)
}

func computeCorrectiveFactor(c *Element) {
	c.SetOne()
	c.Inverse(c)
	c.InverseOld(c)
}

func TestLinearCombNonModular(t *testing.T) {
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
	bHi := b.linearCombNonModular(&a, f1, &b, g1)
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
	prodFast.mulWSigned(&x, yWord)
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
	a.rsh31(&a, aHi)

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

	a.rsh31(&a, aHi)

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
	quickRes.mulWSigned(&u, coeff)

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
