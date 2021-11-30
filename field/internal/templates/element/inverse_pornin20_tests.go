package element

const InversePornin20Tests = `

{{if eq .NoCarry true}}

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
	modulus.Lsh(&modulus, (Limbs + 1) * 64)
	aIntMod.Mod(aInt, &modulus)

	words := aIntMod.Bits()

	for i := 0; i < Limbs; i++ {
		assertMatch(words, z[i], i)
	}

	assertMatch(words, aHi, Limbs)
}

{{- end}}
`
