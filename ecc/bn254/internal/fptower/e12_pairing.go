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

// Expt set z to x^t in E12 and return z (t is the generator of the BN curve)
func (z *E12) Expt(x *E12) *E12 {

	var result, xInv E12
	xInv.Conjugate(x)

	result.Set(x)
	result.nSquare(4)
	result.Mul(&result, x)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(3)
	result.Mul(&result, &xInv)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(2)
	result.Mul(&result, &xInv)
	result.nSquare(3)
	result.Mul(&result, x)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(2)
	result.Mul(&result, &xInv)
	result.nSquare(2)
	result.Mul(&result, &xInv)
	result.nSquare(2)
	result.Mul(&result, &xInv)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(4)
	result.Mul(&result, x)
	result.nSquare(3)
	result.Mul(&result, x)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(2)
	result.Mul(&result, &xInv)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(3)
	result.Mul(&result, x)
	result.nSquare(5)
	result.Mul(&result, x)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(5)
	result.Mul(&result, &xInv)
	result.nSquare(4)
	z.Mul(&result, x)

	return z
}

// MulBy034 multiplication by sparse element (c0,0,0,c3,c4,0)
func (z *E12) MulBy034(c0, c3, c4 *E2) *E12 {

	var a, b, d E6

	a.MulByE2(&z.C0, c0)

	b.Set(&z.C1)
	b.MulBy01(c3, c4)

	c0.Add(c0, c3)
	d.Add(&z.C0, &z.C1)
	d.MulBy01(c0, c4)

	z.C1.Add(&a, &b).Neg(&z.C1).Add(&z.C1, &d)
	z.C0.MulByNonResidue(&b).Add(&z.C0, &a)

	return z
}

// Mul034By034 multiplication of sparse element (c0,0,0,c3,c4,0) by sparse element (d0,0,0,d3,d4,0)
func (z *E12) Mul034by034(d0, d3, d4, c0, c3, c4 *E2) *E12 {
	var tmp, x0, x3, x4, x04, x03, x34 E2
	x0.Mul(c0, d0)
	x3.Mul(c3, d3)
	x4.Mul(c4, d4)
	tmp.Add(c0, c4)
	x04.Add(d0, d4).
		Mul(&x04, &tmp).
		Sub(&x04, &x0).
		Sub(&x04, &x4)
	tmp.Add(c0, c3)
	x03.Add(d0, d3).
		Mul(&x03, &tmp).
		Sub(&x03, &x0).
		Sub(&x03, &x3)
	tmp.Add(c3, c4)
	x34.Add(d3, d4).
		Mul(&x34, &tmp).
		Sub(&x34, &x3).
		Sub(&x34, &x4)

	z.C0.B0.MulByNonResidue(&x4).
		Add(&z.C0.B0, &x0)
	z.C0.B1.Set(&x3)
	z.C0.B2.Set(&x34)
	z.C1.B0.Set(&x03)
	z.C1.B1.Set(&x04)
	z.C1.B2.SetZero()

	return z
}
