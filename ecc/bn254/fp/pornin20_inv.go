package fp

import (
	"math/bits"
)

//TODO: Very not constant time
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

var inversionCorrectionFactor = Element{
	5658184994089520847,
	11089936491052707196,
	18024563689369049901,
	817977532144045340,
}

func (z *Element) Inverse(x *Element) *Element {

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
			f1 *= 2
			g1 *= 2
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
			f1 *= 2
			g1 *= 2
		}

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

	z.Mul(&v, &inversionCorrectionFactor)
	return z
}

/*func (z *Element) P20Inverse(x *Element) *Element {
	if x.IsZero() {
		z.SetZero()
		return z
	}
}*/

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

		//c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[0], c[1] = bits.Add64(c[1], t[1], 0)
		c[2], t[0] = madd2(m, 10917124144477883021, c[2], c[0])

		//c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[0], c[1] = bits.Add64(c[1], t[2], 0)
		c[2], t[1] = madd2(m, 13281191951274694749, c[2], c[0])

		//c[1], c[0] = madd2(v, y[3], c[1], t[3])
		c[0], c[1] = bits.Add64(c[1], t[3], 0)
		t[3], t[2] = madd3(m, 3486998266802970665, c[0], c[2], c[1])
	}
	{
		// round 2
		v := x[2]
		c[1], c[0] = madd1(v, y, t[0])
		m := c[0] * 9786893198990664585
		c[2] = madd0(m, 4332616871279656263, c[0])

		//c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[0], c[1] = bits.Add64(c[1], t[1], 0)
		c[2], t[0] = madd2(m, 10917124144477883021, c[2], c[0])

		//c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[0], c[1] = bits.Add64(c[1], t[2], 0)
		c[2], t[1] = madd2(m, 13281191951274694749, c[2], c[0])

		//c[1], c[0] = madd2(v, y[3], c[1], t[3])
		c[0], c[1] = bits.Add64(c[1], t[3], 0)
		t[3], t[2] = madd3(m, 3486998266802970665, c[0], c[2], c[1])
	}
	{
		// round 3
		v := x[3]
		c[1], c[0] = madd1(v, y, t[0])
		m := c[0] * 9786893198990664585
		c[2] = madd0(m, 4332616871279656263, c[0])

		//c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[0], c[1] = bits.Add64(c[1], t[1], 0)
		c[2], z[0] = madd2(m, 10917124144477883021, c[2], c[0])

		//c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[0], c[1] = bits.Add64(c[1], t[2], 0)
		c[2], z[1] = madd2(m, 13281191951274694749, c[2], c[0])

		//c[1], c[0] = madd2(v, y[3], c[1], t[3])
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
