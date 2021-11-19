package fp

import "math/bits"

func (z *Element) mulModR(x *Element, y *Element) {
	var res Element
	var carry [2]uint64 //can we stick them all in different bits in the same word?
	var a uint64
	var b uint64

	// word 0
	res[1], res[0] = bits.Mul64(x[0], y[0])

	//word 1
	res[2], b = bits.Mul64(x[0], y[1])
	res[1], carry[0] = bits.Add64(res[1], b, 0)
	a, b = bits.Mul64(x[1], y[0])
	res[1], carry[1] = bits.Add64(res[1], b, 0)

	//words 2,3
	res[2], carry[0] = bits.Add64(res[2], a, carry[0])

	res[3], a = bits.Mul64(x[2], y[0])
	res[2], carry[1] = bits.Add64(res[2], a, carry[1])
	a, b = bits.Mul64(x[1], y[1])

	res[3], _ = bits.Add64(res[3], a, carry[1])

	res[2], carry[1] = bits.Add64(res[2], b, 0)
	a, b = bits.Mul64(x[0], y[2])
	res[3], _ = bits.Add64(res[3], a, carry[1])
	res[2], carry[1] = bits.Add64(res[2], b, 0)

	//word 3
	_, a = bits.Mul64(x[3], y[0])
	res[3], _ = bits.Add64(res[3], a, carry[1])
	_, a = bits.Mul64(x[2], y[1])
	res[3], _ = bits.Add64(res[3], a, carry[0])

	res[3] += x[1]*y[2] + x[0]*y[3]

	*z = res
}

//9 working variables, 28 additions
func mulBig(hi *Element, x *Element, y *Element) {

	var a uint64
	var b uint64
	var c uint64
	var d uint64
	var e uint64

	var z Element

	z[1], _ = bits.Mul64(x[0], y[0]) //z[0] is available for scratch work
	z[2], z[0] = bits.Mul64(x[1], y[0])
	z[1], z[0] = bits.Add64(z[1], z[0], 0)
	a, b = bits.Mul64(x[0], y[1])
	_, z[1] = bits.Add64(z[1], b, 0) //final value for word 1

	z[2], z[1] = bits.Add64(z[2], a, z[1])
	z[3], b = bits.Mul64(x[2], y[0])
	z[2], z[0] = bits.Add64(z[2], b, z[0])

	a, b = bits.Mul64(x[1], y[1])
	z[2], b = bits.Add64(z[2], b, 0)
	z[3], c = bits.Add64(z[3], a, b)

	a, b = bits.Mul64(x[0], y[2])
	_, z[2] = bits.Add64(z[2], b, 0) //final value for word 2

	z[3], z[0] = bits.Add64(z[3], a, z[0])

	a, b = bits.Mul64(x[3], y[0])
	z[3], b = bits.Add64(z[3], b, z[1])

	z[1], d = bits.Mul64(x[2], y[1])
	z[3], d = bits.Add64(z[3], d, z[2])

	z[0], z[1] = bits.Add64(z[1], a, z[0])
	z[2], a = bits.Mul64(x[1], y[2])

	z[3], a = bits.Add64(z[3], a, 0)

	z[0], z[2] = bits.Add64(z[0], z[2], a)

	a, e = bits.Mul64(x[0], y[3])
	_, z[3] = bits.Add64(z[3], e, 0)

	z[0], a = bits.Add64(z[0], a, z[3])

	z[3], e = bits.Mul64(x[3], y[1])

	z[0], c = bits.Add64(z[0], e, c)
	z[1], z[3] = bits.Add64(z[3], z[1], z[2])

	z[2], e = bits.Mul64(x[2], y[2])
	z[0], e = bits.Add64(z[0], e, d)
	z[1], e = bits.Add64(z[1], z[2], e)

	z[2], d = bits.Mul64(x[1], y[3])
	hi[0], z[0] = bits.Add64(z[0], d, b)

	z[1], z[0] = bits.Add64(z[1], z[2], z[0])
	z[2], d = bits.Mul64(x[3], y[2])
	z[1], c = bits.Add64(z[1], d, c)

	b, d = bits.Mul64(x[2], y[3])
	hi[1], d = bits.Add64(z[1], d, a)

	z[2], z[3] = bits.Add64(z[2], b, z[3])
	z[1], b = bits.Mul64(x[3], y[3])
	z[2], c = bits.Add64(z[2], b, c)

	z[3], _ = bits.Add64(z[1], z[3], c)
	z[2], d = bits.Add64(z[2], z[0], d)
	hi[2], b = bits.Add64(z[2], 0, e)
	hi[3], _ = bits.Add64(z[3], b, d)
}

var qInvNeg = Element{9786893198990664585, 11447725176084130505, 15613922527736486528, 17688488658267049067}

//Vanilla Mont from Koc94 section 1
func (z *Element) classicMontReduceUnsigned(x *Element, xHi uint64) {

	//We know the low words of u and t will add up to 0. We just need to know if there's going to be a carry, i.e. if r doesn't divide x
	_, carry := bits.Add64(0xFFFFFFFFFFFFFFFF, x[0]|x[1]|x[2]|x[3], 0)

	z.mulModR(x, &qInvNeg)  //m, in the original algorithm
	mulBig(z, z, &qElement) //u, in the original algorithm

	z[0], carry = bits.Add64(z[0], xHi, carry)
	z[1], carry = bits.Add64(z[1], 0, carry)
	z[2], carry = bits.Add64(z[2], 0, carry)
	z[3], carry = bits.Add64(z[3], 0, carry)

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

//Vanilla Mont from Koc94 section 1
func (z *Element) classicMontReduceSigned(x *Element, xHi uint64) {

	//We know the low words of u and t will add up to 0. We just need to know if there's going to be a carry, i.e. if r doesn't divide x
	_, carry := bits.Add64(0xFFFFFFFFFFFFFFFF, x[0]|x[1]|x[2]|x[3], 0)

	z.mulModR(x, &qInvNeg)  //m, in the original algorithm
	mulBig(z, z, &qElement) //u, in the original algorithm

	z[0], carry = bits.Add64(z[0], xHi, carry)
	z[1], carry = bits.Add64(z[1], 0, carry)
	z[2], carry = bits.Add64(z[2], 0, carry)
	z[3], carry = bits.Add64(z[3], 0, carry)

	// if z > q → z -= q
	// note: this is NOT constant time
	if !(z[3] < 3486998266802970665 || (z[3] == 3486998266802970665 && (z[2] < 13281191951274694749 || (z[2] == 13281191951274694749 && (z[1] < 10917124144477883021 || (z[1] == 10917124144477883021 && (z[0] < 4332616871279656263))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 4332616871279656263, 0)
		z[1], b = bits.Sub64(z[1], 10917124144477883021, b)
		z[2], b = bits.Sub64(z[2], 13281191951274694749, b)
		z[3], _ = bits.Sub64(z[3], 3486998266802970665, b)
	}

	if xHi&0x8000000000000000 != 0 {
		var b uint64
		z[1], b = bits.Sub64(z[1], 1, 0)
		z[2], b = bits.Sub64(z[2], 0, b)
		z[3], b = bits.Sub64(z[3], 0, b)

		//very unlikely
		if b != 0 {
			// z[3] = -1
			//negative: add q
			z[0], b = bits.Add64(z[0], 4332616871279656263, 0)
			z[1], b = bits.Add64(z[1], 10917124144477883021, b)
			z[2], b = bits.Add64(z[2], 13281191951274694749, b)
			z[3], _ = bits.Add64(0xFFFFFFFFFFFFFFFF, 3486998266802970665, b)
		}
	}
}

func (z *Element) linearCombClassicSigned(x *Element, xC int64, y *Element, yC int64) {
	hi := z.linearCombNonModular(x, xC, y, yC)
	z.classicMontReduceSigned(z, hi)
}
