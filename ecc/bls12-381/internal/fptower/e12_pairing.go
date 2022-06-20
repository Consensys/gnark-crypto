package fptower

func (z *E12) nSquare(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquare(z)
	}
}

func (z *E12) nSquareCompressed(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquareCompressed(z)
	}
}

// ExptHalf set z to x^(t/2) in E12 and return z
// const t/2 uint64 = 7566188111470821376 // negative
func (z *E12) ExptHalf(x *E12) *E12 {
	var result E12
	var t [2]E12
	result.Set(x)
	result.nSquareCompressed(15)
	t[0].Set(&result)
	result.nSquareCompressed(32)
	t[1].Set(&result)
	batch := BatchDecompressKarabina([]E12{t[0], t[1]})
	result.Mul(&batch[0], &batch[1])
	batch[1].nSquare(9)
	result.Mul(&result, &batch[1])
	batch[1].nSquare(3)
	result.Mul(&result, &batch[1])
	batch[1].nSquare(2)
	result.Mul(&result, &batch[1])
	batch[1].CyclotomicSquare(&batch[1])
	result.Mul(&result, &batch[1])
	return z.Conjugate(&result) // because tAbsVal is negative
}

// Expt set z to xᵗ in E12 and return z
// const t uint64 = 15132376222941642752 // negative
func (z *E12) Expt(x *E12) *E12 {
	var result E12
	result.ExptHalf(x)
	return z.CyclotomicSquare(&result)
}

// MulBy014 multiplication by sparse element (c0, c1, 0, 0, c4)
func (z *E12) MulBy014(c0, c1, c4 *E2) *E12 {

	var a, b E6
	var d E2

	a.Set(&z.C0)
	a.MulBy01(c0, c1)

	b.Set(&z.C1)
	b.MulBy1(c4)
	d.Add(c1, c4)

	z.C1.Add(&z.C1, &z.C0)
	z.C1.MulBy01(c0, &d)
	z.C1.Sub(&z.C1, &a)
	z.C1.Sub(&z.C1, &b)
	z.C0.MulByNonResidue(&b)
	z.C0.Add(&z.C0, &a)

	return z
}

// Mul014By014 multiplication of sparse element (c0,c1,0,0,c4,0) by sparse element (d0,d1,0,0,d4,0)
func (z *E12) Mul014By014(d0, d1, d4, c0, c1, c4 *E2) *E12 {
	var tmp, x0, x1, x4, x04, x01, x14 E2
	x0.Mul(c0, d0)
	x1.Mul(c1, d1)
	x4.Mul(c4, d4)
	tmp.Add(c0, c4)
	x04.Add(d0, d4).
		Mul(&x04, &tmp).
		Sub(&x04, &x0).
		Sub(&x04, &x4)
	tmp.Add(c0, c1)
	x01.Add(d0, d1).
		Mul(&x01, &tmp).
		Sub(&x01, &x0).
		Sub(&x01, &x1)
	tmp.Add(c1, c4)
	x14.Add(d1, d4).
		Mul(&x14, &tmp).
		Sub(&x14, &x1).
		Sub(&x14, &x4)

	z.C0.B0.MulByNonResidue(&x4).
		Add(&z.C0.B0, &x0)
	z.C0.B1.Set(&x01)
	z.C0.B2.Set(&x1)
	z.C1.B0.SetZero()
	z.C1.B1.Set(&x04)
	z.C1.B2.Set(&x14)

	return z
}
