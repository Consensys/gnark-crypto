package fp

import (
	"fmt"
	"math/big"
	"math/bits"
	"math/rand"
	"testing"
)

func TestMulWRegularBf(t *testing.T) {
	w := rand.Int63()
	var x Element
	x.SetRandom()
	xBf := x
	xHi := x.mulWRegularBr(&x, w)
	xBfHi := xBf.mulWRegular(&xBf, w)

	if xHi != xBfHi || !x.Equal(&xBf) {
		panic("mismatch")
	}
}

// regular multiplication by one word regular (non montgomery)
func (z *Element) mulWRegularBr(x *Element, y int64) uint64 {

	w := abs(y)

	var c uint64
	c, z[0] = bits.Mul64(x[0], w)
	c, z[1] = madd1(x[1], w, c)
	c, z[2] = madd1(x[2], w, c)
	c, z[3] = madd1(x[3], w, c)

	if y < 0 {
		c = z.neg(z, c)
	}

	return c
}

func abs(y int64) uint64 {
	m := y >> 63
	return uint64((y ^ m) - m)
}

func (z *Element) add(x *Element, xHi uint64, y *Element, yHi uint64) uint64 {
	var carry uint64
	z[0], carry = bits.Add64(x[0], y[0], 0)
	z[1], carry = bits.Add64(x[1], y[1], carry)
	z[2], carry = bits.Add64(x[2], y[2], carry)
	z[3], carry = bits.Add64(x[3], y[3], carry)
	carry, _ = bits.Add64(xHi, yHi, carry)

	return carry
}

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

func (z *Element) sub(x *Element, y *Element) {
	var b uint64
	z[0], b = bits.Sub64(x[0], y[0], 0)
	z[1], b = bits.Sub64(x[1], y[1], b)
	z[2], b = bits.Sub64(x[2], y[2], b)
	z[3], _ = bits.Sub64(x[3], y[3], b)
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
	z.linearCombSosSigned(x, xC, y, yC)

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
	x.toVeryBigIntUnsigned(&xInt, xHi)
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
	x.toVeryBigIntUnsigned(&xInt, xHi)

	yHi := uint64(22264684)
	y := Element{
		4332616871279656263,
		10917124144477883021,
		13281191951274694749,
		3486998266802970665,
	}
	var yInt big.Int
	y.toVeryBigIntUnsigned(&yInt, yHi)

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

func (z *Element) toVeryBigIntUnsigned(i *big.Int, xHi uint64) {
	z.ToBigInt(i)
	var upperWord big.Int
	upperWord.SetUint64(xHi)
	upperWord.Lsh(&upperWord, Limbs*bits.UintSize)
	i.Add(&upperWord, i)
}

func (z *Element) toVeryBigIntSigned(i *big.Int, xHi uint64) {
	z.toVeryBigIntUnsigned(i, xHi)
	if signBitSelector&xHi != 0 {
		twosCompModulus := big.NewInt(1)
		twosCompModulus.Lsh(twosCompModulus, (Limbs+1)*bits.UintSize)
		i.Sub(i, twosCompModulus)
	}
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

func TestFindInversionCorrectionFactorFormula(t *testing.T) {
	fac := big.NewInt(1)

	var correctionFactor big.Int
	inversionCorrectionFactor.ToBigInt(&correctionFactor)
	correctionFactor.Mod(&correctionFactor, Modulus())

	for i := 1; i < 100000; i++ {

		fac.Lsh(fac, 1)
		fac.Mod(fac, Modulus())
		//fmt.Println(fac.Bits())

		if fac.Cmp(&correctionFactor) == 0 {
			fmt.Println("Match at", i)
			return
		}
	}
	panic("No match")
}

func TestVanillaInverseNeedsMulByRSq(t *testing.T) {
	var x Element
	if _, err := x.SetRandom(); err != nil {
		panic(err)
	}

	var xInt big.Int
	x.ToBigInt(&xInt)

	var vanillaInv big.Int
	vanillaInv.ModInverse(&xInt, Modulus())

	var montInv big.Int
	x.InverseOld(&x)
	x.ToBigInt(&montInv)

	vanillaInv.Lsh(&vanillaInv, 512)
	vanillaInv.Mod(&vanillaInv, Modulus())

	if montInv.Cmp(&vanillaInv) != 0 {
		panic("mismatch")
	}

}

func (z *Element) setBigIntNoMont(v *big.Int) {
	vBits := v.Bits()

	if bits.UintSize == 64 {
		for i := 0; i < len(vBits); i++ {
			z[i] = uint64(vBits[i])
		}
	} else {
		for i := 0; i < len(vBits); i++ {
			if i%2 == 0 {
				z[i/2] = uint64(vBits[i])
			} else {
				z[i/2] |= uint64(vBits[i]) << 32
			}
		}
	}

}

func TestVanillaInverseNeedsMulByRCbMont(t *testing.T) {
	var x Element
	if _, err := x.SetRandom(); err != nil {
		panic(err)
	}

	var xInt big.Int
	x.ToBigInt(&xInt)

	x.InverseOld(&x)

	var vanillaInvInt big.Int
	vanillaInvInt.ModInverse(&xInt, Modulus())
	var vanillaInv Element
	vanillaInv.setBigIntNoMont(&vanillaInvInt)

	correctionFactorInt := big.NewInt(1)
	correctionFactorInt.Lsh(correctionFactorInt, 3*256)
	correctionFactorInt.Mod(correctionFactorInt, Modulus())
	var correctionFactor Element
	correctionFactor.setBigIntNoMont(correctionFactorInt)

	vanillaInv.Mul(&vanillaInv, &correctionFactor)

	if !x.Equal(&vanillaInv) {
		panic("mismatch")
	}

}

func TestInversionCorrectionFactorFormula(t *testing.T) {
	const iterationN = 2 * ((2*Bits-2)/(2*k) + 1) // 2  ⌈ (2 * field size - 1) / 2k ⌉
	const kLimbs = k * Limbs
	const power = kLimbs*6 + iterationN*(kLimbs-k+1)
	factorInt := big.NewInt(1)
	factorInt.Lsh(factorInt, power)
	factorInt.Mod(factorInt, Modulus())

	var refFactorInt big.Int
	inversionCorrectionFactor.ToBigInt(&refFactorInt)

	if refFactorInt.Cmp(factorInt) != 0 {
		panic("mismatch")
	}
}
