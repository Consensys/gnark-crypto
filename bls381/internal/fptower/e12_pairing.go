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
