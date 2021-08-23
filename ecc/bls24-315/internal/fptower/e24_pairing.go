package fptower

func (z *E24) nSquare(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquare(z)
	}
}

// Expt set z to x^t in E24 and return z (t is the seed of the curve)
func (z *E24) Expt(x *E24) *E24 {

	var result, xInv E24
	result.Set(x)
	xInv.Conjugate(x)

	result.nSquare(2)
	result.Mul(&result, &xInv)
	result.nSquare(8)
	result.Mul(&result, &xInv)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(20)
	result.Mul(&result, &xInv)

	z.Conjugate(&result)

	return z
}

// MulBy012 multiplication by sparse element
// https://eprint.iacr.org/2019/077.pdf
func (z *E24) MulBy012(c0, c1, c2 *E4) *E24 {

	var d0, v0, v1, tmp E8

	d0.C0.Set(c0)
	d0.C1.Set(c1)

	v0.Mul(&z.D0, &d0)
	v1.C0.Mul(&z.D1.C0, c2)
	v1.C1.Mul(&z.D1.C1, c2)

	z.D1.Add(&z.D1, &z.D0)
	tmp.Set(&d0)
	tmp.C0.Add(&tmp.C0, c2)
	z.D1.Mul(&z.D1, &tmp)
	z.D1.Sub(&z.D1, &v0)
	z.D1.Sub(&z.D1, &v1)

	z.D0.C0.Mul(&z.D2.C0, c2)
	z.D0.C1.Mul(&z.D2.C1, c2)
	z.D0.MulByNonResidue(&z.D0)
	z.D0.Add(&z.D0, &v0)

	z.D2.Mul(&z.D2, &d0)
	z.D2.Add(&z.D2, &v1)

	return z
}
