package fptower

func (z *E24) nSquareCompressed(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquareCompressed(z)
	}
}

// Expt set z to x^t in E24 and return z (t is the seed of the curve)
// -2**32+2**30+2**22-2**20+1
func (z *E24) Expt(x *E24) *E24 {

	var result, x20, x22, x30, x32 E24
	result.Set(x)

	result.nSquareCompressed(20)
	x20.Conjugate(&result)
	result.nSquareCompressed(2)
	x22.Set(&result)
	result.nSquareCompressed(8)
	x30.Set(&result)

	batch := BatchDecompress([]E24{x20, x22, x30})

	x32.CyclotomicSquare(&batch[2]).
		CyclotomicSquare(&x32).
		Conjugate(&x32)

	z.Mul(x, &batch[0]).
		Mul(z, &batch[1]).
		Mul(z, &batch[2]).
		Mul(z, &x32)

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
