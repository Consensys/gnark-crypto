package fp

import (
	"math/bits"
)

//This is not being inlined and I don't understand why
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
	//Compute n jointly and inline?. Code for that follows. Currently, BitLen and max together are taking 1.2% of exec time. Worth it?

	/*n := 192
	var msw uint64 //Most significant word
	msw = x[3] | y[3]
	if msw == 0 {
		n = 128
		msw = x[2] | y[2]

		if msw == 0 {
			n = 64
			msw = x[1] | y[1]

			if msw == 0 {
				n = 0
				msw = x[0] | y[0]
			}
		}
	}
	n |= bits.Len64(msw)*/

	return approximate(x, n), approximate(y, n)
}

//Which correction factor to use depends on how many iterations the outer loop takes
var inversionCorrectionFactors = [8]Element{
	{9294402098508299643, 16236581287374362326, 1806700940207652208, 128304151138745798},
	{3785369258512301398, 3447191806671807780, 17892925251185020671, 628989039686645193},
	{3640683342331600137, 9590128738288309796, 14138712235514295312, 1231420490468424357},
	{4521516680493641497, 8084381843320164072, 9724766311162352044, 2024159453010255379},
	{15621838106149573218, 3484330101846812783, 657711689591423763, 1264074572563695769},
	{1576046162781523005, 3026941236205245694, 13031833993062009898, 554036701478437490},
	{5738979239160164595, 3911769744532092421, 6476601505093438411, 2879139492355964105},
	{5743661648749932980, 12551916556084744593, 23273105902916091, 802172129993363311},
}

func (z *Element) Inverse(x *Element) *Element {
	if x.IsZero() {
		z.SetZero()
		return z
	}

	var a = *x
	var b = qElement
	var u = Element{1, 0, 0, 0}
	var v = Element{0, 0, 0, 0}

	//Update factors: we get [u; v]:= [f0 g0; f1 g1] [u; v]
	var f0, g0, f1, g1 int64

	//Saved update factors to reduce the number of field multiplications
	var pf0, pg0, pf1, pg1 int64

	//Used for updating u,v
	var scratch, updateFactor Element

	var i uint

	//Since u,v are updated every other iteration, we must make sure we terminate after evenly many iterations
	//This also lets us get away with 8 update factors instead of 16
	for i = 0; i%2 == 1 || !a.IsZero(); i++ {
		aApprox, bApprox := approximatePair(&a, &b)
		f0, g0, f1, g1 = 1, 0, 0, 1

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

		scratch = a
		aHi := a.bigNumLinearComb(&scratch, f0, &b, g0)
		bHi := b.bigNumLinearComb(&scratch, f1, &b, g1)

		//The condition means "negative"
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

		if i%2 == 1 {
			//Combine current update factors with previously stored ones
			f0, g0, f1, g1 = f0*pf0+g0*pf1,
				f0*pg0+g0*pg1,
				f1*pf0+g1*pf1,
				f1*pg0+g1*pg1

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
		} else {
			//Save update factors
			pf0, pg0, pf1, pg1 = f0, g0, f1, g1
		}

	}

	//Alternative to storing many correction factors. Not much slower
	/*const pSq int64 = 0x4000000000000000
	for ; i < 16; i+=2 {
		v.MulWord(&v, pSq)
	}*/

	//Multiply by the appropriate correction factor
	z.Mul(&v, &inversionCorrectionFactors[i/2-1])

	return z
}

//TODO: Do this directly
//TODO: If not done directly, the absolute value of y is less than half of uint64's capacity, could mean more carries are certainly zero
func (z *Element) MulWord(x *Element, y int64) {
	var neg bool
	var abs uint64

	//This logic is somehow taking 2.38% of execution time?!
	if y < 0 {
		neg = true
		abs = uint64(-y)
	} else {
		neg = false
		abs = uint64(y)
	}
	//</this logic>

	z.mulWordUnsigned(x, abs)

	if neg {
		z.Neg(z)
	}
}

func (z *Element) mulWordUnsigned(x *Element, y uint64) {

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

	carry += hi2

	if neg {
		carry = z.bigNumNeg(z, carry)
	}

	return carry
}

func (z *Element) bigNumNeg(x *Element, xHi uint64) uint64 {

	//Bad to use z as working variable? It's on stack anyway
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
	const mask = uint64(0x7FFFFFFF) //31 ones
	z[0] = (x[0] >> 31) | ((x[1] & mask) << 33)
	z[1] = (x[1] >> 31) | ((x[2] & mask) << 33)
	z[2] = (x[2] >> 31) | ((x[3] & mask) << 33)
	z[3] = (x[3] >> 31) | ((xHi & mask) << 33)
}

func (z *Element) bigNumLinearComb(x *Element, xCoeff int64, y *Element, yCoeff int64) uint64 {
	var xTimes Element

	//Removed working variable yTimes and used z instead. Seems to have hurt performance instead of improving.
	//Generally discouraged practice?

	xHi := xTimes.bigNumMultiply(x, xCoeff)
	yHi := z.bigNumMultiply(y, yCoeff)
	hi := z.bigNumAdd(&xTimes, xHi, z, yHi)

	return hi
}
