package fp

/*func testMonReduceNeg(x *Element, xHi uint64) {
	var asIs Element
	negRed := *x
	asIs.montReduceSigned(x, xHi)

	negHi := xHi
	neg := xHi&0x8000000000000000 != 0
	if neg {
		negHi = negRed.neg(x, xHi)
	}
	negRed.montReduceSigned(&negRed, negHi)

	if neg {
		negRed.Neg(&negRed)
	}

	var diff Element
	diff.sub(&negRed, &asIs)

	if !diff.IsZero() {
		panic(fmt.Sprint(xHi, x, ": expected", negRed, "got", asIs, "difference", diff))
	}
}*/

/*func testMontReduceRef(x *big.Int) {
	q := Modulus()
	var r big.Int
	r.SetInt64(1)
	r.Lsh(&r, Limbs * bits.UintSize)

	var u big.Int
	montReduce(&u, x)

	var ur big.Int
	ur.Mul(&u, &r)
	ur.Sub(&ur, x)
	ur.Mod(&ur, q)

	if ur.BitLen() != 0 {
		panic("Mismatch")
	}
}

func TestMontReduceRef(t *testing.T) {
	var max big.Int
	max.SetInt64(1)
	max.Lsh(&max, (Limbs+1)*bits.UintSize)
	for i := 0; i < 1000; i++ {
		y, _ := crand.Int(crand.Reader, &max)
		testMontReduceRef(y)
	}
}

func TestMontReduceRefSmall(t *testing.T) {
	var x big.Int

	x.SetString("1518345043075282886718915476446629923034923247403426348876984432860252403179691687682438634393061", 10)
	testMontReduceRef(&x)
}
*/

/*func TestMonReduceNegFixed(t *testing.T) {
	testMonReduceNeg(&Element{2625241524836463861, 14355433948910864505, 16319971849635632347, 821941842937000211}, 14800378828802555218)

}*/

/*
var rInv big.Int

func montReduce(res *big.Int, x *big.Int) {
	if rInv.BitLen() == 0 { //initialization
		rInv.SetUint64(1)
		rInv.Lsh(&rInv, Limbs*bits.UintSize)
		rInv.ModInverse(&rInv, Modulus())
	}
	res.Mul(x, &rInv)
	res.Mod(res, Modulus())
}

func testMontReduceSigned(x *Element, xHi uint64) {
	var res Element
	var xInt big.Int
	var resInt big.Int
	x.toVeryBigIntSigned(&xInt, xHi)
	res.montReduceSigned(x, xHi)
	montReduce(&resInt, &xInt)
	checkMatchBigInt(&res, 0, &resInt)
}

func TestMonReduceNeg(t *testing.T) {
	var x Element

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		testMontReduceSigned(&x, rand.Uint64()|signBitSelector)
	}
}

func TestMontReducePos(t *testing.T) {
	var x Element

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		testMontReduceSigned(&x, rand.Uint64() & ^signBitSelector)
	}
}

func TestMontNegMultipleOfR(t *testing.T) {
	zero := Element{0, 0, 0, 0}

	for i := 0; i < 1000; i++ {
		testMontReduceSigned(&zero, rand.Uint64()|signBitSelector)
	}
}
*/
