package element

const InversePornin20Tests = `

//this is a hack so that there isn't an import error in case mrand is not used
//TODO: Do it properly
func useMRand() {
	_ = mrand.Uint64()
}

{{if eq .NoCarry true}}

func Benchmark{{.ElementName}}InverseNew(b *testing.B) {
	var x {{.ElementName}}
	x.SetRandom()

	b.Run("inverseNew", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchRes{{.ElementName}}.Inverse(&x)
		}
	})
}

func TestP20InversionApproximation(t *testing.T) {
	var x {{.ElementName}}
	for i := 0; i < 1000; i++ {
		x.SetRandom()

		//Normally small elements are unlikely. Here we give them a higher chance
		xZeros := mrand.Int() % Limbs
		for j := 1; j < xZeros; j++ {
			x[Limbs - j] = 0
		}

		a := approximate(&x, x.BitLen())
		aRef := approximateRef(&x)

		if a != aRef {
			t.Fatal("Approximation mismatch")
		}
	}
}

func TestP20InversionCorrectionFactorFormula(t *testing.T) {
	const kLimbs = k * Limbs
	const power = kLimbs*6 + invIterationsN*(kLimbs-k+1)
	factorInt := big.NewInt(1)
	factorInt.Lsh(factorInt, power)
	factorInt.Mod(factorInt, Modulus())

	var refFactorInt big.Int
	inversionCorrectionFactor.ToBigInt(&refFactorInt)

	if refFactorInt.Cmp(factorInt) != 0 {
		t.Fatal("mismatch")
	}
}

func TestLinearComb(t *testing.T) {
	var x {{.ElementName}}
	var y {{.ElementName}}

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		y.SetRandom()
		testLinearComb(t, &x, mrand.Int63(), &y, mrand.Int63())
	}
}

//Probably unnecessary post-dev. In case the output of inv is wrong, this checks whether it's only off by a constant factor.
func TestP20InversionCorrectionFactor(t *testing.T) {

	//(1/x)/inv(x) = (1/1)/inv(1) ⇔ inv(1) = x inv(x)

	var one {{.ElementName}}
	var oneInv {{.ElementName}}
	one.SetOne()
	oneInv.Inverse(&one)

	for i := 0; i < 100; i++ {
		var x {{.ElementName}}
		var xInv {{.ElementName}}
		x.SetRandom()
		xInv.Inverse(&x)

		x.Mul(&x, &xInv)
		if !x.Equal(&oneInv) {
			t.Fatal("Correction factor is inconsistent")
		}
	}

	if !oneInv.Equal(&one) {
		var i big.Int
		oneInv.ToBigIntRegular(&i)	//no montgomery
		i.ModInverse(&i, Modulus())
		var fac {{.ElementName}}
		fac.setBigInt(&i)	//back to montgomery

		var facTimesFac {{.ElementName}}
		facTimesFac.Mul(&inversionCorrectionFactor, &fac)

		t.Fatal("Correction factor is consistently off by", fac, "Should be", facTimesFac)
	}
}

func TestBigNumNeg(t *testing.T) {
	var a {{.ElementName}}
	aHi := a.neg(&a, 0)
	if !a.IsZero() || aHi != 0 {
		t.Fatal("-0 != 0")
	}
}

func TestBigNumWMul(t *testing.T) {
	var x {{.ElementName}}

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		w := mrand.Int63()
		testBigNumWMul(t, &x, w)
	}
}

func TestBigNumWMulBr(t *testing.T) {
	var x {{.ElementName}}

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		w := mrand.Int63()
		testBigNumWMulBr(t, &x, w)
	}
}

func TestVeryBigIntConversion(t *testing.T) {
	xHi := mrand.Uint64()
	var x {{.ElementName}}
	x.SetRandom()
	var xInt big.Int
	x.toVeryBigIntSigned(&xInt, xHi)
	x.assertMatchVeryBigInt(t, xHi, &xInt)
}

func TestMontReducePos(t *testing.T) {
	var x {{.ElementName}}

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		testMontReduceSigned(t, &x, mrand.Uint64() & ^signBitSelector)
	}
}

func TestMonReduceNeg(t *testing.T) {
	var x {{.ElementName}}

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		testMontReduceSigned(t, &x, mrand.Uint64() | signBitSelector)
	}
}

func TestMontNegMultipleOfR(t *testing.T) {
	var zero {{.ElementName}}

	for i := 0; i < 1000; i++ {
		testMontReduceSigned(t, &zero, mrand.Uint64() | signBitSelector)
	}
}

func testLinearComb(t *testing.T, x *{{.ElementName}}, xC int64, y *{{.ElementName}}, yC int64) {

	var p1 big.Int
	x.ToBigInt(&p1)
	p1.Mul(&p1, big.NewInt(xC))

	var p2 big.Int
	y.ToBigInt(&p2)
	p2.Mul(&p2, big.NewInt(yC))

	p1.Add(&p1, &p2)
	p1.Mod(&p1, Modulus())
	montReduce(&p1, &p1)

	var z {{.ElementName}}
	z.linearCombSosSigned(x, xC, y, yC)
	z.assertMatchVeryBigInt(t, 0, &p1)

}

func testBigNumWMulBr(t *testing.T, a *{{.ElementName}}, c int64) {
	var aHi uint64
	var aTimes {{.ElementName}}
	aHi = aTimes.mulWRegularBr(a, c)

	assertMulProduct(t, a, c, &aTimes, aHi)
}

func testBigNumWMul(t *testing.T, a *{{.ElementName}}, c int64) {
	var aHi uint64
	var aTimes {{.ElementName}}
	aHi = aTimes.mulWRegular(a, c)

	assertMulProduct(t, a, c, &aTimes, aHi)
}

func testMontReduceSigned(t *testing.T, x *{{.ElementName}}, xHi uint64) {
	var res {{.ElementName}}
	var xInt big.Int
	var resInt big.Int
	x.toVeryBigIntSigned(&xInt, xHi)
	res.montReduceSigned(x, xHi)
	montReduce(&resInt, &xInt)
	res.assertMatchVeryBigInt(t, 0, &resInt)
}

var rInv big.Int
func montReduce(res *big.Int, x *big.Int) {
	if rInv.BitLen() == 0 {	//initialization
		rInv.SetUint64(1)
		rInv.Lsh(&rInv, Limbs * bits.UintSize)
		rInv.ModInverse(&rInv, Modulus())
	}
	res.Mul(x, &rInv)
	res.Mod(res, Modulus())
}

func (z *{{.ElementName}}) toVeryBigIntUnsigned(i *big.Int, xHi uint64) {
	z.ToBigInt(i)
	var upperWord big.Int
	upperWord.SetUint64(xHi)
	upperWord.Lsh(&upperWord, Limbs*bits.UintSize)
	i.Add(&upperWord, i)
}

func (z *{{.ElementName}}) toVeryBigIntSigned(i *big.Int, xHi uint64) {
	z.toVeryBigIntUnsigned(i, xHi)
	if signBitSelector&xHi != 0 {
		twosCompModulus := big.NewInt(1)
		twosCompModulus.Lsh(twosCompModulus, (Limbs+1)*bits.UintSize)
		i.Sub(i, twosCompModulus)
	}
}

func assertMulProduct(t *testing.T, x *{{.ElementName}}, c int64, result *{{.ElementName}}, resultHi uint64) big.Int {
	var xInt big.Int
	x.ToBigInt(&xInt)

	xInt.Mul(&xInt, big.NewInt(c))

	result.assertMatchVeryBigInt(t, resultHi, &xInt)
	return xInt
}

func assertMatch(t *testing.T, w []big.Word, a uint64, index int) {
	var wI big.Word

	if index < len(w) {
		wI = w[index]
	}

	if uint64(wI) != a {
		t.Fatal("Bignum mismatch: disagreement on word", index)
	}
}

func (z *{{.ElementName}}) assertMatchVeryBigInt(t *testing.T, aHi uint64, aInt *big.Int) {

	if bits.UintSize != 64 {
		panic("Word size 64 expected")
	}

	var modulus big.Int
	var aIntMod big.Int
	modulus.SetInt64(1)
	modulus.Lsh(&modulus, (Limbs + 1) * 64)
	aIntMod.Mod(aInt, &modulus)

	words := aIntMod.Bits()

	for i := 0; i < Limbs; i++ {
		assertMatch(t, words, z[i], i)
	}

	assertMatch(t, words, aHi, Limbs)
}

func approximateRef(x *{{.ElementName}}) uint64 {

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
{{- end}}
`
