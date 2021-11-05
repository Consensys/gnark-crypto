package fp

import (
	"math/big"
	"math/bits"
)

func (z *Element) SetInt64(i int64) {
	z.MulWord(&rSquare, i)
}

//LtAsIs compares without changing back from Montgomery form
func (z *Element) LtAsIs(x *Element) bool {

	if z[3] == x[3] {
		if z[2] == x[2] {
			if z[1] == x[1] {
				return z[0] < x[0]
			}
			return z[1] < x[1]
		}
		return z[2] < x[2]
	}
	return z[3] < x[3]
}

func (z *Element) Inverse0(x *Element) *Element {

	if x.IsZero() {
		z.SetZero()
		return z
	}

	var a Element
	var b Element
	var u Element
	var v Element

	a = *x
	b = qElement
	u.SetOne()
	v.SetZero()

	//Loop bound according to P20 pg 3 footnote 2 (x is known to be invertible)
	//for iteration := 0; iteration < 2*Bits - 2; iteration++ {
	for !a.IsZero() {
		if a.Bit(0) == 0 {
			a.Halve()
			u.Halve()
		} else {
			if a.LtAsIs(&b) {
				a, b = b, a
				u, v = v, u
			}
			//TODO: Exploit the shrinking of the lengths of a,b? Nah
			a.Sub(&a, &b)
			a.Halve()
			u.Sub(&u, &v)
			u.Halve()
		}
	}

	z.Mul(&v, &rSquare) //TODO: Would it have been okay to store it in v itself?
	return z
}

var inversionCorrectionFactorOptimization1 = Element{
	5658184994089520847,
	11089936491052707196,
	18024563689369049901,
	817977532144045340,
}

func (z *Element) InverseOpt1(x *Element) *Element {

	if x.IsZero() {
		z.SetZero()
		return z
	}

	var a Element
	var b Element
	var u Element
	var v Element

	//Update factors: we get [u; v]:= [f0 g0; f1 g1] [u; v]
	//TODO: Better or worse to group two of them in the same var? Paper suggests it's used in "most optimized" implementation but algorithm 2 doesn't use it
	var f0 int64 = 1
	var g0 int64 = 0
	var f1 int64 = 0
	var g1 int64 = 1

	a = *x
	b = qElement
	u = Element{1, 0, 0, 0}
	v.SetZero()

	t := 1
	//Loop bound according to P20 pg 3 footnote 2 (x is known to be invertible)
	for iteration := 0; iteration < 2*Bits-2; iteration++ {
		//TODO: The two branches don't take the same amount of time
		if a.Bit(0) == 0 {
			a.Halve()
		} else {
			if a.LtAsIs(&b) {
				a, b = b, a
				f0, f1 = f1, f0
				g0, g1 = g1, g0
			}

			a.Sub(&a, &b)
			a.Halve()
			f0 -= f1
			g0 -= g1
		}

		f1 *= 2
		g1 *= 2

		if t == 62 || iteration == 2*Bits-3 { //TODO: 63 fails. Find out why. Anyway, same runtime
			t = 0

			var elementScratch Element
			var updateFactor Element
			//save u
			elementScratch.Set(&u)

			//update u
			u.MulWord(&u, f0)

			updateFactor.MulWord(&v, g0)
			u.Add(&u, &updateFactor)

			//update v
			elementScratch.MulWord(&elementScratch, f1)

			v.MulWord(&v, g1)
			v.Add(&v, &elementScratch)

			f0, g0, f1, g1 = 1, 0, 0, 1
		}
		t++
	}

	z.Mul(&v, &inversionCorrectionFactorOptimization1)
	return z
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func approximate(x *Element, n int) uint64 {

	if n <= 64 {
		return x[0]
	}

	mask := uint64(0x7FFFFFFF) //31 ones
	lo := mask & x[0]

	hiWordIndex := (n - 1) / 64

	hiWordBitsAvailable := n - hiWordIndex*64
	hiWordBitsUsed := min(hiWordBitsAvailable, 33)

	mask = ^((1 << (hiWordBitsAvailable - hiWordBitsUsed)) - 1)
	hi := (x[hiWordIndex] & mask) << (64 - hiWordBitsAvailable)

	mask = ^(1<<(31+hiWordBitsUsed) - 1)
	mid := (mask & x[hiWordIndex-1]) >> hiWordBitsUsed

	return lo | mid | hi
}

func approximatePair(x *Element, y *Element) (uint64, uint64) {
	n := max(x.BitLen(), y.BitLen())

	return approximate(x, n), approximate(y, n)
}

var inversionCorrectionFactorP20Full = Element{8862593707351107428, 14861862907286385237, 15773464367735268868, 1095622056137557639}

func (z *Element) Inverse(x *Element) *Element {
	if x.IsZero() {
		z.SetZero()
		return z
	}

	var a = *x
	var b = qElement
	var u = Element{1, 0, 0, 0}
	var v = Element{0, 0, 0, 0}

	// registers are 64bit, thus k = 32
	outerLoopIterations := 16 // ceil( (2* 254 - 1)/32 )

	for i := 0; i < outerLoopIterations; i++ {
		aApprox, bApprox := approximatePair(&a, &b)

		//Update factors: we get [u; v]:= [f0 g0; f1 g1] [u; v]
		var f0 int64 = 1
		var g0 int64 = 0
		var f1 int64 = 0
		var g1 int64 = 1

		for j := 0; j < 31; j++ {

			if aApprox&1 == 0 {
				aApprox /= 2
			} else {
				if aApprox < bApprox {
					aApprox, bApprox = bApprox, aApprox
					f0, f1 = f1, f0
					g0, g1 = g1, g0
				}

				aApprox = (aApprox - bApprox) / 2
				f0 -= f1
				g0 -= g1
			}

			f1 *= 2
			g1 *= 2
		}

		scratch := a
		aHi := a.bigNumLinearComb(&scratch, f0, &b, g0)
		bHi := b.bigNumLinearComb(&scratch, f1, &b, g1)

		if aHi&0x8000000000000000 != 0 {
			f0, g0 = -f0, -g0
			aHi = a.bigNumNeg(&a, aHi)
		}
		if bHi&0x8000000000000000 != 0 {
			f1, g1 = -f1, -g1
			bHi = b.bigNumNeg(&b, bHi)
		}

		a.bigNumRshBy31(&a, aHi)
		b.bigNumRshBy31(&b, bHi)

		var updateFactor Element
		//save u
		scratch.Set(&u)

		//update u
		u.MulWord(&u, f0)

		updateFactor.MulWord(&v, g0)
		u.Add(&u, &updateFactor)

		//update v
		scratch.MulWord(&scratch, f1)

		v.MulWord(&v, g1)
		v.Add(&v, &scratch)

		f0, g0, f1, g1 = 1, 0, 0, 1
	}

	z.Mul(&v, &inversionCorrectionFactorP20Full)
	return z
}

//TODO: Do this directly
//TODO: If not done directly, the absolute value of y is less than half of uint64's capacity, could mean more carries are certainly zero
func (z *Element) MulWord(x *Element, y int64) {
	var neg bool
	var abs uint64

	if y < 0 {
		neg = true
		abs = uint64(-y)
	} else {
		neg = false
		abs = uint64(y)
	}

	z.MulWordUnsigned(x, abs)

	if neg {
		z.Neg(z)
	}
}

func (z *Element) MulWordUnsigned(x *Element, y uint64) {

	var t [4]uint64
	var c [3]uint64
	{
		// round 0
		v := x[0]
		c[1], c[0] = bits.Mul64(v, y)
		m := c[0] * 9786893198990664585
		c[2] = madd0(m, 4332616871279656263, c[0])

		c[2], t[0] = madd2(m, 10917124144477883021, c[2], c[1])
		c[2], t[1] = madd2(m, 13281191951274694749, c[2], 0)
		t[3], t[2] = madd1(m, 3486998266802970665, c[2])

	}
	{
		// round 1
		v := x[1]
		c[1], c[0] = madd1(v, y, t[0])
		m := c[0] * 9786893198990664585
		c[2] = madd0(m, 4332616871279656263, c[0])

		c[0], c[1] = bits.Add64(c[1], t[1], 0)
		c[2], t[0] = madd2(m, 10917124144477883021, c[2], c[0])

		c[0], c[1] = bits.Add64(c[1], t[2], 0)
		c[2], t[1] = madd2(m, 13281191951274694749, c[2], c[0])

		c[0], c[1] = bits.Add64(c[1], t[3], 0)
		t[3], t[2] = madd3(m, 3486998266802970665, c[0], c[2], c[1])
	}
	{
		// round 2
		v := x[2]
		c[1], c[0] = madd1(v, y, t[0])
		m := c[0] * 9786893198990664585
		c[2] = madd0(m, 4332616871279656263, c[0])

		c[0], c[1] = bits.Add64(c[1], t[1], 0)
		c[2], t[0] = madd2(m, 10917124144477883021, c[2], c[0])

		c[0], c[1] = bits.Add64(c[1], t[2], 0)
		c[2], t[1] = madd2(m, 13281191951274694749, c[2], c[0])

		c[0], c[1] = bits.Add64(c[1], t[3], 0)
		t[3], t[2] = madd3(m, 3486998266802970665, c[0], c[2], c[1])
	}
	{
		// round 3
		v := x[3]
		c[1], c[0] = madd1(v, y, t[0])
		m := c[0] * 9786893198990664585
		c[2] = madd0(m, 4332616871279656263, c[0])

		c[0], c[1] = bits.Add64(c[1], t[1], 0)
		c[2], z[0] = madd2(m, 10917124144477883021, c[2], c[0])

		c[0], c[1] = bits.Add64(c[1], t[2], 0)
		c[2], z[1] = madd2(m, 13281191951274694749, c[2], c[0])

		c[0], c[1] = bits.Add64(c[1], t[3], 0)
		z[3], z[2] = madd3(m, 3486998266802970665, c[0], c[2], c[1])
	}

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[3] < 3486998266802970665 || (z[3] == 3486998266802970665 && (z[2] < 13281191951274694749 || (z[2] == 13281191951274694749 && (z[1] < 10917124144477883021 || (z[1] == 10917124144477883021 && (z[0] < 4332616871279656263))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 4332616871279656263, 0)
		z[1], b = bits.Sub64(z[1], 10917124144477883021, b)
		z[2], b = bits.Sub64(z[2], 13281191951274694749, b)
		z[3], _ = bits.Sub64(z[3], 3486998266802970665, b)
	}
}

func (z *Element) bigNumMultiply(x *Element, y int64) uint64 {
	var carry uint64

	var hi uint64
	var hi2 uint64 //these two variables alternate as holding the high word of the current multiplication and that of the previous one

	neg := y < 0
	var abs uint64

	if neg {
		abs = uint64(-y)
	} else {
		abs = uint64(y)
	}

	hi, z[0] = bits.Mul64(x[0], abs) //doesn't matter if z = x. We'll never again have to use the value x[0]

	hi2, z[1] = bits.Mul64(x[1], abs)
	z[1], carry = bits.Add64(z[1], hi, 0)

	hi, z[2] = bits.Mul64(x[2], abs)
	z[2], carry = bits.Add64(z[2], hi2, carry)

	hi2, z[3] = bits.Mul64(x[3], abs)
	z[3], carry = bits.Add64(z[3], hi, carry)

	carry = hi2 + carry //can we do this when not caring about carry?

	if neg {
		carry = z.bigNumNeg(z, carry)
	}

	return carry
}

func (z *Element) bigNumNeg(x *Element, xHi uint64) uint64 {

	z[0], z[1], z[2], z[3], xHi = ^x[0], ^x[1], ^x[2], ^x[3], ^xHi
	var carry uint64
	z[0], carry = bits.Add64(z[0], 1, 0)
	z[1], carry = bits.Add64(z[1], 0, carry)
	z[2], carry = bits.Add64(z[2], 0, carry)
	z[3], carry = bits.Add64(z[3], 0, carry)
	xHi, _ = bits.Add64(xHi, 0, carry)

	return xHi
}

func (z *Element) bigNumAdd(x *Element, xHi uint64, y *Element, yHi uint64) uint64 {
	var carry uint64
	z[0], carry = bits.Add64(x[0], y[0], 0)
	z[1], carry = bits.Add64(x[1], y[1], carry)
	z[2], carry = bits.Add64(x[2], y[2], carry)
	z[3], carry = bits.Add64(x[3], y[3], carry)
	carry, _ = bits.Add64(xHi, yHi, carry)

	return carry
}

func (z *Element) bigNumRshBy31(x *Element, xHi uint64) {
	mask := uint64(0x7FFFFFFF) //31 ones
	z[0] = (x[0] >> 31) | ((x[1] & mask) << 33)
	z[1] = (x[1] >> 31) | ((x[2] & mask) << 33)
	z[2] = (x[2] >> 31) | ((x[3] & mask) << 33)
	z[3] = (x[3] >> 31) | ((xHi & mask) << 33)
}

func (z *Element) bigNumLinearComb(x *Element, xCoeff int64, y *Element, yCoeff int64) uint64 {
	var xTimes Element
	var yTimes Element

	xHi := xTimes.bigNumMultiply(x, xCoeff)
	yHi := yTimes.bigNumMultiply(y, yCoeff)
	hi := z.bigNumAdd(&xTimes, xHi, &yTimes, yHi)

	return hi
}

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
