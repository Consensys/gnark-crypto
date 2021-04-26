package fptower

import (
	"math/bits"
)

// Expt set z to x^t in E24 and return z (t is the seed of the curve)
func (z *E24) Expt(x *E24) *E24 {

	const tAbsVal uint64 = 3218079743

	var result E24
	result.Set(x)

	l := bits.Len64(tAbsVal) - 2
	for i := l; i >= 0; i-- {
		result.Square(&result)
		if tAbsVal&(1<<uint(i)) != 0 {
			result.Mul(&result, x)
		}
	}

	z.Conjugate(&result)

	return z
}

// MulBy012 multiplication by sparse element
// https://eprint.iacr.org/2019/077.pdf
func (z *E24) MulBy012(c0, c1, c2 *E4) *E24 {

	var z0, z1, z2, z3, z4, z5, tmp1, tmp2 E4
	var t [12]E4

	z0 = z.D0.C0
	z1 = z.D0.C1
	z2 = z.D1.C0
	z3 = z.D1.C1
	z4 = z.D2.C0
	z5 = z.D2.C1

	tmp1.MulByNonResidue(c1)
	tmp2.MulByNonResidue(c2)

	t[0].Mul(&tmp1, &z1)
	t[1].Mul(&tmp2, &z5)
	t[2].Mul(c1, &z0)
	t[3].Mul(c2, &z4)
	t[4].Mul(&tmp1, &z3)
	t[5].Mul(c2, &z0)
	t[6].Mul(c1, &z2)
	t[7].Mul(c2, &z1)
	t[8].Mul(&tmp1, &z5)
	t[9].Mul(c2, &z2)
	t[10].Mul(c1, &z4)
	t[11].Mul(c2, &z3)

	z.D0.C0.Mul(c0, &z0).
		Add(&z.D0.C0, &t[0]).
		Add(&z.D0.C0, &t[1])
	z.D0.C1.Mul(c0, &z1).
		Add(&z.D0.C1, &t[2]).
		Add(&z.D0.C1, &t[3])
	z.D1.C0.Mul(c0, &z2).
		Add(&z.D1.C0, &t[4]).
		Add(&z.D1.C0, &t[5])
	z.D1.C1.Mul(c0, &z3).
		Add(&z.D1.C1, &t[6]).
		Add(&z.D1.C1, &t[7])
	z.D2.C0.Mul(c0, &z4).
		Add(&z.D2.C0, &t[8]).
		Add(&z.D2.C0, &t[9])
	z.D2.C1.Mul(c0, &z5).
		Add(&z.D2.C1, &t[10]).
		Add(&z.D2.C1, &t[11])

	return z
}
