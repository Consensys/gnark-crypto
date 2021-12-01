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

// MulBy034 multiplication by sparse element (c0,0,0,c3,c4,0)
func (z *E24) MulBy034(c0, c3, c4 *E4) *E24 {

	var a, b, d E12

	a.MulByE2(&z.D0, c0)

	b.Set(&z.D1)
	b.MulBy01(c3, c4)

	c0.Add(c0, c3)
	d.Add(&z.D0, &z.D1)
	d.MulBy01(c0, c4)

	z.D1.Add(&a, &b).Neg(&z.D1).Add(&z.D1, &d)
	z.D0.MulByNonResidue(&b).Add(&z.D0, &a)

	return z
}
