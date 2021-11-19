package fp

import (
	"fmt"
	"math/big"
	"math/bits"
	"math/rand"
	"strconv"
	"testing"
)

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
		panic(fmt.Sprint(xHi, x, ": expected", negRed, "difference", diff))
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

//montReduceSigned SOS algorithm; xHi must be at most 63 bits long.
func (z *Element) montReduceUnsigned(x *Element, xHi uint64) {

	const qInvNegLsb uint64 = 0x87d20782e4866389

	var t [7]uint64
	var C uint64
	{
		m := x[0] * qInvNegLsb

		C = madd0(m, qElement[0], x[0])
		C, t[1] = madd2(m, qElement[1], x[1], C)
		C, t[2] = madd2(m, qElement[2], x[2], C)
		C, t[3] = madd2(m, qElement[3], x[3], C)
		// the high word of m * qElement[3] is at most 62 bits
		// x[3] + C is at most 65 bits (high word at most 1 bit)
		// Thus the resulting C will be at most 63 bits
		t[4] = xHi + C
		// xHi and C are 63 bits, therefore no overflow

	}
	{
		const i = 1
		m := t[i] * qInvNegLsb

		C = madd0(m, qElement[0], t[i+0])
		C, t[i+1] = madd2(m, qElement[1], t[i+1], C)
		C, t[i+2] = madd2(m, qElement[2], t[i+2], C)
		C, t[i+3] = madd2(m, qElement[3], t[i+3], C)

		t[5] += C

	}
	{
		const i = 2
		m := t[i] * qInvNegLsb

		C = madd0(m, qElement[0], t[i+0])
		C, t[i+1] = madd2(m, qElement[1], t[i+1], C)
		C, t[i+2] = madd2(m, qElement[2], t[i+2], C)
		C, t[i+3] = madd2(m, qElement[3], t[i+3], C)

		t[6] += C
	}
	{
		const i = 3
		m := t[i] * qInvNegLsb

		C = madd0(m, qElement[0], t[i+0])
		C, z[0] = madd2(m, qElement[1], t[i+1], C)
		C, z[1] = madd2(m, qElement[2], t[i+2], C)
		z[3], z[2] = madd2(m, qElement[3], t[i+3], C)
	}

	// if z > q → z -= q
	// note: this is NOT constant time
	if !(z[3] < 3486998266802970665 || (z[3] == 3486998266802970665 && (z[2] < 13281191951274694749 || (z[2] == 13281191951274694749 && (z[1] < 10917124144477883021 || (z[1] == 10917124144477883021 && (z[0] < 4332616871279656263))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 4332616871279656263, 0)
		z[1], b = bits.Sub64(z[1], 10917124144477883021, b)
		z[2], b = bits.Sub64(z[2], 13281191951274694749, b)
		z[3], _ = bits.Sub64(z[3], 3486998266802970665, b)
	}
}

func (z *Element) linearCombSosUnsigned(x *Element, xC int64, y *Element, yC int64) {
	hi := z.linearCombNonModular(x, xC, y, yC)
	neg := (hi & 0x8000000000000000) != 0
	if neg {
		hi = z.neg(z, hi)
	}
	z.montReduceUnsigned(z, hi)
	if neg {
		z.Neg(z)
	}
}
