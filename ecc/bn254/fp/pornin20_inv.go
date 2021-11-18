package fp

import (
	"math/bits"
)

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

const k = 32 // word size / 2

func approximate(x *Element, n int) uint64 {

	if n <= 64 {
		return x[0]
	}

	const mask = (uint64(1) << (k - 1)) - 1 //k-1 ones
	lo := mask & x[0]

	hiWordIndex := (n - 1) / 64

	hiWordBitsAvailable := n - hiWordIndex*64
	hiWordBitsUsed := min(hiWordBitsAvailable, 33)

	mask_ := uint64(^((1 << (hiWordBitsAvailable - hiWordBitsUsed)) - 1))
	hi := (x[hiWordIndex] & mask_) << (64 - hiWordBitsAvailable)

	mask_ = ^(1<<(31+hiWordBitsUsed) - 1)
	mid := (mask_ & x[hiWordIndex-1]) >> hiWordBitsUsed

	return lo | mid | hi
}

var inversionCorrectionFactor = Element{5743661648749932980, 12551916556084744593, 23273105902916091, 802172129993363311}

func (z *Element) Inverse(x *Element) *Element {
	if x.IsZero() {
		z.SetZero()
		return z
	}

	var a = *x
	var b = qElement
	var u = Element{1}

	//Update factors: we get [u; v]:= [f0 g0; f1 g1] [u; v]
	var f0, g0, f1, g1 int64

	//Saved update factors to reduce the number of field multiplications
	var pf0, pg0, pf1, pg1 int64

	var i uint

	var v, s Element

	const iterationN = 16 // 2 \ceil{ (2 * field size - 1) / 2k }
	const approxLowBitsN = k - 1
	const approxHighBitsN = k + 1

	//Since u,v are updated every other iteration, we must make sure we terminate after evenly many iterations
	//This also lets us get away with 8 update factors instead of 16
	//To make this constant-time-ish, replace the condition with i < iterationN
	for i = 0; i&1 == 1 || !a.IsZero(); i++ {
		n := max(a.BitLen(), b.BitLen())
		aApprox, bApprox := approximate(&a, n), approximate(&b, n)

		f0, g0, f1, g1 = 1, 0, 0, 1

		for j := 0; j < approxLowBitsN; j++ {

			if aApprox&1 == 0 {
				aApprox /= 2
			} else {
				s, borrow := bits.Sub64(aApprox, bApprox, 0)
				if borrow == 1 {
					s = bApprox - aApprox
					bApprox = aApprox
					f0, f1 = f1, f0
					g0, g1 = g1, g0
				}

				aApprox = s / 2
				f0 -= f1
				g0 -= g1

			}

			f1 *= 2
			g1 *= 2

		}

		s = a
		aHi := a.linearCombNonModular(&s, f0, &b, g0)
		if aHi&(0b1<<63) != 0 {
			// if aHi < 0
			f0, g0 = -f0, -g0
			aHi = a.neg(&a, aHi)
		}
		//right-shift a by k-1 bits
		a[0] = (a[0] >> approxLowBitsN) | ((a[1]) << approxHighBitsN)
		a[1] = (a[1] >> approxLowBitsN) | ((a[2]) << approxHighBitsN)
		a[2] = (a[2] >> approxLowBitsN) | ((a[3]) << approxHighBitsN)
		a[3] = (a[3] >> approxLowBitsN) | ((aHi) << approxHighBitsN)

		bHi := b.linearCombNonModular(&s, f1, &b, g1)
		if bHi&(0b1<<63) != 0 {
			// if bHi < 0
			f1, g1 = -f1, -g1
			bHi = b.neg(&b, bHi)
		}
		//right-shift b by k-1 bits
		b[0] = (b[0] >> approxLowBitsN) | ((b[1]) << approxHighBitsN)
		b[1] = (b[1] >> approxLowBitsN) | ((b[2]) << approxHighBitsN)
		b[2] = (b[2] >> approxLowBitsN) | ((b[3]) << approxHighBitsN)
		b[3] = (b[3] >> approxLowBitsN) | ((bHi) << approxHighBitsN)

		if i&1 == 1 {
			//Combine current update factors with previously stored ones
			// [f₀, g₀; f₁, g₁] ← [f₀, g₀; f₁, g₀] [pf0, pg0; pf1, pg1]
			f0, g0, f1, g1 = f0*pf0+g0*pf1,
				f0*pg0+g0*pg1,
				f1*pf0+g1*pf1,
				f1*pg0+g1*pg1

			s = u
			u.linearComb(&u, f0, &v, g0)
			v.linearComb(&s, f1, &v, g1)

		} else {
			//Save update factors
			pf0, pg0, pf1, pg1 = f0, g0, f1, g1
		}

	}

	//For every iteration that we miss, v is not being multiplied by 2²ᵏ⁻²
	const pSq int64 = 1 << (2 * (k - 1))
	//If the function is constant-time ish, this loop will not run (probably no need to take it out explicitly)
	for ; i < iterationN; i += 2 {
		v.mulWSigned(&v, pSq)
	}

	z.Mul(&v, &inversionCorrectionFactor)
	return z
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

// On ARM, using the branch free version gives 21% speedup. On x86 it slows things down.
// mulWRegular branch-free regular multiplication by one word (non montgomery)
func (z *Element) mulWRegular(x *Element, y int64) uint64 {

	w := uint64(y)
	allNeg := uint64(y >> 63)

	//z1, z2 so results are not stored immediately in z.
	//x[i] will be needed in the i+1 th iteration. We don't want to overwrite it in case x = z
	var h1, h2, b, c, z1, z2 uint64

	h1, z1 = bits.Mul64(x[0], w)

	h2, z2 = bits.Mul64(x[1], w)
	z2, c = bits.Add64(z2, h1, 0)
	z2, b = bits.Sub64(z2, allNeg&x[0], 0) //x[0] no longer useful, safe to write to z[0]
	z[0] = z1

	h1, z1 = bits.Mul64(x[2], w)
	z1, c = bits.Add64(z1, h2, c)
	z1, b = bits.Sub64(z1, allNeg&x[1], b) //x[1] no longer useful, safe to write to z[1]
	z[1] = z2

	h2, z2 = bits.Mul64(x[3], w)
	z2, c = bits.Add64(z2, h1, c)
	z2, b = bits.Sub64(z2, allNeg&x[2], b)
	z[2] = z1

	z1, _ = bits.Sub64(h2, allNeg&x[3], b)
	z[3] = z2
	return z1 + c
}

// mulWSigned mul word signed (w/ montgomery reduction)
func (z *Element) mulWSigned(x *Element, y int64) {
	_mulWGeneric(z, x, abs(y))
	if y < 0 {
		z.Neg(z)
	}
}

func (z *Element) neg(x *Element, xHi uint64) uint64 {
	var b uint64
	z[0], b = bits.Sub64(0, x[0], 0)
	z[1], b = bits.Sub64(0, x[1], b)
	z[2], b = bits.Sub64(0, x[2], b)
	z[3], b = bits.Sub64(0, x[3], b)
	xHi, _ = bits.Sub64(0, xHi, b)
	return xHi
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

//WARNING: Might need an extra high word (last carry) if BitLen(x) or BitLen(y) are 256. Not a problem here since len(p) = 254
func (z *Element) linearCombNonModular(x *Element, xC int64, y *Element, yC int64) uint64 {
	var yTimes Element

	yHi := yTimes.mulWRegular(y, yC)
	xHi := z.mulWRegular(x, xC)

	var carry uint64
	z[0], carry = bits.Add64(z[0], yTimes[0], 0)
	z[1], carry = bits.Add64(z[1], yTimes[1], carry)
	z[2], carry = bits.Add64(z[2], yTimes[2], carry)
	z[3], carry = bits.Add64(z[3], yTimes[3], carry)
	yHi, _ = bits.Add64(xHi, yHi, carry)

	return yHi
}

func (z *Element) linearComb(x *Element, xC int64, y *Element, yC int64) {
	hi := z.linearCombNonModular(x, xC, y, yC)
	z.montReduceSigned(z, hi)
}

var montNegativeCorrectionBias = Element{13555988908134432071, 10917124144477883020, 13281191951274694749, 3486998266802970665}

//montReduceSigned SOS algorithm; xHi must be at most 63 bits long. Last bit of xHi may be used as a sign bit
func (z *Element) montReduceSigned(x *Element, xHi uint64) {

	const qInvNegLsb uint64 = 0x87d20782e4866389

	neg := (int64(xHi) >> 63) != 0
	xHi &= 0x7FFFFFFFFFFFFFFF

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

	if neg {
		z.Add(z, &montNegativeCorrectionBias)
	}
}
