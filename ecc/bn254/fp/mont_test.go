package fp

import (
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"testing"
)

func TestClassicMontReduceUnsignedRand(t *testing.T) {
	for i := 0; i < 1000; i++ {
		xHi := rand.Uint64()
		xHi |= 0x8000000000000000 //make sure it "overflows"
		//xHi &= 0x7FFFFFFFFFFFFFFF  //make sure it doesn't "overflow"
		var x Element
		var res Element
		x.SetRandom()
		res.classicMontReduceUnsigned(&x, xHi)

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

func testClassicMonReduceNeg(x *Element, xHi uint64) {
	var asIs Element
	negRed := *x
	asIs.classicMontReduceSigned(x, xHi)

	negHi := xHi
	neg := xHi&0x8000000000000000 != 0
	if neg {
		negHi = negRed.neg(x, xHi)
	}
	negRed.classicMontReduceSigned(&negRed, negHi)

	if neg {
		negRed.Neg(&negRed)
	}

	var diff Element
	diff.sub(&negRed, &asIs)

	if !diff.IsZero() {
		panic(fmt.Sprint(xHi, x, ":", diff))
	}
}

func TestClassicMonReduceNeg(t *testing.T) {
	var x Element

	for i := 0; i < 1000; i++ {
		x.SetRandom()
		xHi := rand.Uint64()
		xHi |= 0x8000000000000000
		testClassicMonReduceNeg(&x, xHi)
	}
}

func TestCompute2Pow192Neg(t *testing.T) {
	var twoPow192Neg big.Int

	x := big.NewInt(1)
	x.Lsh(x, 192)
	twoPow192Neg.Neg(x)
	twoPow192Neg.Mod(&twoPow192Neg, Modulus())
	fmt.Println("2^192", twoPow192Neg)

	x.Lsh(x, 1)
	x.Neg(x)
	x.Mod(x, Modulus())
	fmt.Println("2^193", *x)

	fmt.Println("Highest word of 2^192 in Hex", strconv.FormatUint(3486998266802970663, 16))

	computedBySub := Element{0, 0, 0, 1}
	computedBySub.Neg(&computedBySub)
	fmt.Println("2^192", computedBySub)

	fmt.Println("Modulus", qElement)
}
