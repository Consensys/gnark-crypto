package fptower

// MulBy034 multiplication by sparse element
func (z *E24) MulBy034(c0, c3, c4 *E4) *E24 {

    var z0, z1, z2, z3, z4, z5, tmp1, tmp2 E4
	var t [12]E4

    z0 = z.D0.C0
    z1 = z.D0.C1
    z2 = z.D1.C0
    z3 = z.D1.C1
    z4 = z.D2.C0
    z5 = z.D2.C1

	tmp1.MulByNonResidue(c3)
	tmp2.MulByNonResidue(c4)

	t[0].Mul(&tmp1, &z5)
	t[1].Mul(&tmp2, &z4)
	t[2].Mul(c3, &z3)
	t[3].Mul(&tmp2, &z5)
	t[4].Mul(c3, &z4)
	t[5].Mul(c4, &z3)
	t[6].Mul(c3, &z0)
	t[7].Mul(&tmp2, &z2)
	t[8].Mul(c3, &z1)
	t[9].Mul(c4, &z0)
	t[10].Mul(c3, &z2)
	t[11].Mul(c4, &z1)

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
