package element

const MontRedTests = `

//this is a hack so that there isn't an import error in case mrand is not used
//TODO: Do it properly
func useMRand() {
	_ = mrand.Uint64()
}

{{if eq .NoCarry true}}

func TestMontReducePos(t *testing.T) {
	var x Element

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		testMontReduceSigned(&x, mrand.Uint64() & ^signBitSelector)
	}
}

func TestMonReduceNeg(t *testing.T) {
	var x Element

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		testMontReduceSigned(&x, mrand.Uint64() | signBitSelector)
	}
}

func TestMontNegMultipleOfR(t *testing.T) {
	zero := Element{0, 0, 0, 0}

	for i := 0; i < 1000; i++ {
		testMontReduceSigned(&zero, mrand.Uint64() | signBitSelector)
	}
}

func testMontReduceSigned(x *Element, xHi uint64) {
	var res Element
	var xInt big.Int
	var resInt big.Int
	x.toVeryBigIntSigned(&xInt, xHi)
	res.montReduceSigned(x, xHi)
	montReduce(&resInt, &xInt)
	res.assertMatchBigInt(0, &resInt)
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

{{- end}}
`
