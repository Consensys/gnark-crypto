package fptower

// MulByV2NRInv set z to x*(y*v^2*(1,1)^{-1}) and return z
func (z *E12) MulByV2NRInv(x *E12, y *E2) *E12 {

	var result E12
	var yNRInv E2
	yNRInv.MulByNonResidueInv(y)

	result.C0.B0.Mul(&x.C0.B1, y)
	result.C0.B1.Mul(&x.C0.B2, y)
	result.C0.B2.Mul(&x.C0.B0, &yNRInv)

	result.C1.B0.Mul(&x.C1.B1, y)
	result.C1.B1.Mul(&x.C1.B2, y)
	result.C1.B2.Mul(&x.C1.B0, &yNRInv)

	z.Set(&result)
	return z
}

// MulByVWNRInv set z to x*(y*v*w*(1,1)^{-1}) and return z
func (z *E12) MulByVWNRInv(x *E12, y *E2) *E12 {
	var result E12
	var yNRInv E2
	yNRInv.MulByNonResidueInv(y)

	result.C0.B0.Mul(&x.C1.B1, y)
	result.C0.B1.Mul(&x.C1.B2, y)
	result.C0.B2.Mul(&x.C1.B0, &yNRInv)

	result.C1.B0.Mul(&x.C0.B2, y)
	result.C1.B1.Mul(&x.C0.B0, &yNRInv)
	result.C1.B2.Mul(&x.C0.B1, &yNRInv)

	z.Set(&result)
	return z
}

// MulByWNRInv set z to x*(y*w*(1,1)^{-1}) and return z
func (z *E12) MulByWNRInv(x *E12, y *E2) *E12 {

	var result E12
	var yNRInv E2
	yNRInv.MulByNonResidueInv(y)

	result.C0.B0.Mul(&x.C1.B2, y)
	result.C0.B1.Mul(&x.C1.B0, &yNRInv)
	result.C0.B2.Mul(&x.C1.B1, &yNRInv)

	result.C1.B0.Mul(&x.C0.B0, &yNRInv)
	result.C1.B1.Mul(&x.C0.B1, &yNRInv)
	result.C1.B2.Mul(&x.C0.B2, &yNRInv)

	z.Set(&result)
	return z
}

// MulBy014 multiplication by sparse element
func (z *E12) MulBy014(c0, c1, c4 *E2) *E12 {

	var z0, z1, z2, z3, z4, z5, tmp1, tmp2 E2
	var t [12]E2

	z0 = z.C0.B0
	z1 = z.C0.B1
	z2 = z.C0.B2
	z3 = z.C1.B0
	z4 = z.C1.B1
	z5 = z.C1.B2

	tmp1.MulByNonResidue(c1)
	tmp2.MulByNonResidue(c4)

	t[0].Mul(&tmp1, &z2)
	t[1].Mul(&tmp2, &z4)
	t[2].Mul(c1, &z0)
	t[3].Mul(&tmp2, &z5)
	t[4].Mul(c1, &z1)
	t[5].Mul(c4, &z3)
	t[6].Mul(&tmp1, &z5)
	t[7].Mul(&tmp2, &z2)
	t[8].Mul(c1, &z3)
	t[9].Mul(c4, &z0)
	t[10].Mul(c1, &z4)
	t[11].Mul(c4, &z1)

	z.C0.B0.Mul(c0, &z0).
		Add(&z.C0.B0, &t[0]).
		Add(&z.C0.B0, &t[1])
	z.C0.B1.Mul(c0, &z1).
		Add(&z.C0.B1, &t[2]).
		Add(&z.C0.B1, &t[3])
	z.C0.B2.Mul(c0, &z2).
		Add(&z.C0.B2, &t[4]).
		Add(&z.C0.B2, &t[5])
	z.C1.B0.Mul(c0, &z3).
		Add(&z.C1.B0, &t[6]).
		Add(&z.C1.B0, &t[7])
	z.C1.B1.Mul(c0, &z4).
		Add(&z.C1.B1, &t[8]).
		Add(&z.C1.B1, &t[9])
	z.C1.B2.Mul(c0, &z5).
		Add(&z.C1.B2, &t[10]).
		Add(&z.C1.B2, &t[11])

	return z
}

func (z *E12) nSquare(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquare(z)
	}
}

// ExptHalf set z to x^(t/2) in E12 and return z
// const t/2 uint64 = 7566188111470821376 // negative
func (z *E12) ExptHalf(x *E12) *E12 {
	var result E12
	result.CyclotomicSquare(x)
	result.Mul(&result, x)
	result.nSquare(2)
	result.Mul(&result, x)
	result.nSquare(3)
	result.Mul(&result, x)
	result.nSquare(9)
	result.Mul(&result, x)
	result.nSquare(32)
	result.Mul(&result, x)
	result.nSquare(15)
	return z.Conjugate(&result) // because tAbsVal is negative
}

// Expt set z to x^t in E12 and return z
// const t uint64 = 15132376222941642752 // negative
func (z *E12) Expt(x *E12) *E12 {
	var result E12
	result.ExptHalf(x)
	return z.CyclotomicSquare(&result)
}
