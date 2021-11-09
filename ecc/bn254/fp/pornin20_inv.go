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

	const mask = uint64(0x7FFFFFFF) //31 ones
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
	var u = Element{1}

	//Update factors: we get [u; v]:= [f0 g0; f1 g1] [u; v]
	var f0, g0, f1, g1 int64

	//Saved update factors to reduce the number of field multiplications
	var pf0, pg0, pf1, pg1 int64

	var i uint
	var carry uint64

	var v, s, r Element

	//Since u,v are updated every other iteration, we must make sure we terminate after evenly many iterations
	//This also lets us get away with 8 update factors instead of 16
	for i = 0; i&1 == 1 || !a.IsZero(); i++ {
		n := max(a.BitLen(), b.BitLen())
		aApprox, bApprox := approximate(&a, n), approximate(&b, n)

		f0, g0, f1, g1 = 1, 0, 0, 1

		for j := 0; j < 31; j++ {

			if aApprox&1 == 0 {
				aApprox /= 2
			} else {
				s, borrow := bits.Sub64(aApprox, bApprox, 0)
				if borrow == 1 {
					s = (bApprox - aApprox)
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
		aHi := a.linearComb(&s, f0, &b, g0)
		if aHi&(0b1<<63) != 0 {
			// if aHi < 0
			f0, g0 = -f0, -g0
			aHi = a.neg(&a, aHi)
		}
		a.rsh31(&a, aHi)

		bHi := b.linearComb(&s, f1, &b, g1)
		if bHi&(0b1<<63) != 0 {
			// if bHi < 0
			f1, g1 = -f1, -g1
			bHi = b.neg(&b, bHi)
		}
		b.rsh31(&b, bHi)

		if i&1 == 1 {
			//Combine current update factors with previously stored ones
			f0, g0, f1, g1 = f0*pf0+g0*pf1,
				f0*pg0+g0*pg1,
				f1*pf0+g1*pf1,
				f1*pg0+g1*pg1

				// save u in s
			s = u

			//update u
			u.mulWSigned(&u, f0)

			r.mulWSigned(&v, g0)

			u[0], carry = bits.Add64(u[0], r[0], 0)
			u[1], carry = bits.Add64(u[1], r[1], carry)
			u[2], carry = bits.Add64(u[2], r[2], carry)
			u[3], _ = bits.Add64(u[3], r[3], carry)

			//update v
			s.mulWSigned(&s, f1)
			v.mulWSigned(&v, g1)

			v[0], carry = bits.Add64(v[0], s[0], 0)
			v[1], carry = bits.Add64(v[1], s[1], carry)
			v[2], carry = bits.Add64(v[2], s[2], carry)
			v[3], _ = bits.Add64(v[3], s[3], carry)
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

// mulWSigned mul word signed (w/ montgomery reduction)
func (z *Element) mulWSigned(x *Element, y int64) {
	_mulWGeneric(z, x, abs(y))
	if y < 0 {
		z.Neg(z)
	}
}

// regular multiplication by one word regular (non montgomery)
func (z *Element) mulWRegular(x *Element, y int64) uint64 {

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

func (z *Element) neg(x *Element, xHi uint64) uint64 {
	var b uint64
	z[0], b = bits.Sub64(0, x[0], 0)
	z[1], b = bits.Sub64(0, x[1], b)
	z[2], b = bits.Sub64(0, x[2], b)
	z[3], b = bits.Sub64(0, x[3], b)
	xHi, _ = bits.Sub64(0, xHi, b)
	return xHi
}

func (z *Element) add(xTimes *Element, xHi uint64, yTimes *Element, yHi uint64) uint64 {
	var carry uint64
	z[0], carry = bits.Add64(xTimes[0], yTimes[0], 0)
	z[1], carry = bits.Add64(xTimes[1], yTimes[1], carry)
	z[2], carry = bits.Add64(xTimes[2], yTimes[2], carry)
	z[3], carry = bits.Add64(xTimes[3], yTimes[3], carry)
	carry, _ = bits.Add64(xHi, yHi, carry)

	return carry
}

func (z *Element) rsh31(x *Element, xHi uint64) {
	z[0] = (x[0] >> 31) | ((x[1]) << 33)
	z[1] = (x[1] >> 31) | ((x[2]) << 33)
	z[2] = (x[2] >> 31) | ((x[3]) << 33)
	z[3] = (x[3] >> 31) | ((xHi) << 33)
}

func (z *Element) linearComb(x *Element, xCoeff int64, y *Element, yCoeff int64) uint64 {
	var yTimes Element

	yHi := yTimes.mulWRegular(y, yCoeff)
	xHi := z.mulWRegular(x, xCoeff)

	var carry uint64
	z[0], carry = bits.Add64(z[0], yTimes[0], 0)
	z[1], carry = bits.Add64(z[1], yTimes[1], carry)
	z[2], carry = bits.Add64(z[2], yTimes[2], carry)
	z[3], carry = bits.Add64(z[3], yTimes[3], carry)
	yHi, _ = bits.Add64(xHi, yHi, carry)

	return yHi
}
