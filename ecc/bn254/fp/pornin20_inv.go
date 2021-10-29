package fp

import (
	"math/big"
)

//Three possible ways to swap values:
//	1. Swap by value :D
//	2. Swap by reference
//	3. Run all assignments using conditionals on whether a or b is smaller
func swapBigInt(a *big.Int, b *big.Int, swapScratch *big.Int) {
	swapScratch.Set(a)
	a.Set(b)
	b.Set(swapScratch)
}

func (z *Element) swap(a *Element, b *Element) {
	z.Set(a)
	a.Set(b)
	b.Set(z)
}

//TODO: Very not constant time
func (z *Element) SetInt64(i int64) {
	var neg = i < 0
	var abs uint64
	if neg {
		abs = uint64(-i)
	} else {
		abs = uint64(i)
	}

	z.SetUint64(abs)
	z.Neg(z)
}

//z is known in the algorithm as u. Renamed for receiver name consistency
func (z *Element) Inverse(x *Element) *Element {

	if x.IsZero() {
		z.SetZero()
		return z
	}

	var a big.Int
	var b big.Int
	var u Element
	var v Element
	var intScratch big.Int //Copy values when
	var elementScratch Element

	x.ToBigInt(&a)
	b.Set(Modulus())
	u.SetOne() //TODO: Don't use u as working variable
	v.SetZero()

	//Loop bound according to P20 pg 3 footnote 2 (x is known to be invertible)
	//for iteration := 0; iteration < 2*Bits - 2; iteration++ {
	for a.BitLen() != 0 {
		//TODO: The two branches don't take the same amount of time
		if a.Bit(0) == 0 {
			a.Rsh(&a, 1) //TODO: okay to use a itself?
			u.Halve()
		} else {
			if a.Cmp(&b) < 0 {
				swapBigInt(&a, &b, &intScratch)
				elementScratch.swap(&u, &v)
			}
			//TODO: Exploit the shrinking of the lengths of a,b? Nah
			a.Sub(&a, &b)
			a.Rsh(&a, 1)
			u.Sub(&u, &v) //TODO: ok?
			u.Halve()
		}
	}

	z.Mul(&v, &rSquare) //TODO: Would it have been okay to store it in v itself?
	return z
}

func (z *Element) InverseOptimization1(x *Element) *Element {
	if x.IsZero() {
		z.SetZero()
		return z
	}

	var a big.Int
	var b big.Int
	var v Element
	var bigIntScratch big.Int //Copy values when
	var elementScratch Element
	var intScratch int64

	//Update factors: we get [u; v]:= [f0 g0; f1 g1] [u; v]
	//TODO: Better or worse to group two of them in the same var? Paper suggests it's used in "most optimized" implementation but algorithm 2 doesn't use it
	var f0 int64
	var g0 int64
	var f1 int64
	var g1 int64

	x.ToBigInt(&a)
	b.Set(Modulus())
	z.SetOne()
	v.SetZero()

	//Loop bound according to P20 pg 3 footnote 2 (x is known to be invertible)
	for iteration := 0; iteration < 2*Bits-2; iteration++ {
		var t = iteration % 63
		//TODO: The two branches don't take the same amount of time
		if a.Bit(0) == 0 {
			a.Rsh(&a, 1) //TODO: okay to use a itself?
			f1 *= 2
			g1 *= 2
		} else {
			if a.Cmp(&b) < 0 {
				swapBigInt(&a, &b, &bigIntScratch)
				intScratch = f1
				f1 = f0
				f0 = intScratch
				intScratch = g1
				g1 = g0
				g0 = intScratch
			}
			//TODO: Exploit the shrinking of the lengths of a,b?
			a.Sub(&a, &b) //TODO: This okay?
			a.Rsh(&a, 1)  //TODO: ok?
			f0 -= f1
			g0 -= g1
			f1 *= 2
			g1 *= 2
		}

		if t == 63 {
			t = 0
			var updateFactor Element
			//save u
			elementScratch.Set(z)

			//update u
			updateFactor.SetInt64(f0)
			//TODO: Exploit the fact that the update factor is small?
			z.Mul(z, &updateFactor)
			updateFactor.SetInt64(g0)
			updateFactor.Mul(&v, &updateFactor)
			z.Add(z, &updateFactor)

			//update v
			updateFactor.SetInt64(f1)
			elementScratch.Mul(&elementScratch, &updateFactor)
			updateFactor.SetInt64(g1)
			v.Mul(&v, &updateFactor)
			v.Add(&v, &elementScratch)
		}
	}

	//TODO: Multiply by rSquare x 2^(-2*Bits+2) instead
	z.Mul(&v, &rSquare) //TODO: Would it have been okay to store it in v itself?
	return z

}

/*func (z *Element) P20Inverse(x *Element) *Element {
	if x.IsZero() {
		z.SetZero()
		return z
	}


}*/
