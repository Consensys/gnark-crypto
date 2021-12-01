package fp

func abs(y int64) uint64 {
	m := y >> 63
	return uint64((y ^ m) - m)
}

/*
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



func assertMatch(w []big.Word, a uint64, index int) {
	var wI big.Word

	if index < len(w) {
		wI = w[index]
	}

	if uint64(wI) != a {
		fmt.Printf("Disagreement on word %d\n", index)
		panic("Bignum mismatch")
	}
}

func (z *Element) assertMatchBigInt(aHi uint64, aInt *big.Int) {

	if bits.UintSize != 64 {
		panic("Word size 64 expected")
	}

	var modulus big.Int
	var aIntMod big.Int
	modulus.SetInt64(1)
	modulus.Lsh(&modulus, (Limbs+1)*64)
	aIntMod.Mod(aInt, &modulus)

	words := aIntMod.Bits()

	for i := 0; i < Limbs; i++ {
		assertMatch(words, z[i], i)
	}

	assertMatch(words, aHi, Limbs)
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
*/
