package fp

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
)

func testMonReduceNeg(x *Element, xHi uint64) {
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
}

func TestMonReduceNegFixed(t *testing.T) {
	testMonReduceNeg(&Element{2625241524836463861, 14355433948910864505, 16319971849635632347, 821941842937000211}, 14800378828802555218)

}

func TestMonReduceNeg(t *testing.T) {
	var x Element

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		xHi := rand.Uint64()
		xHi |= 0x8000000000000000
		testMonReduceNeg(&x, xHi)
	}
}

func TestMontReduceUnsignedRand(t *testing.T) {
	for i := 0; i < 1000; i++ {
		xHi := rand.Uint64()
		//xHi |= 0x8000000000000000	//make sure it "overflows"
		xHi &= 0x7FFFFFFFFFFFFFFF //make sure it doesn't "overflow"
		var x Element
		var res Element
		x.SetRandom()
		res.montReduceSigned(&x, xHi)

		var xInt big.Int
		var resInt big.Int
		x.ToVeryBigInt(&xInt, xHi)
		res.ToBigInt(&resInt)

		resInt.Lsh(&resInt, 256)
		resInt.Sub(&resInt, &xInt)
		resInt.Mod(&resInt, Modulus())

		if resInt.BitLen() != 0 {
			panic("incorrect")
		}
	}
}
