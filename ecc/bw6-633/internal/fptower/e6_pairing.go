package fptower

import "github.com/consensys/gnark-crypto/ecc/bw6-633/fp"

func (z *E6) nSquare(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquare(z)
	}
}

// Expt set z to x^t in E6 and return z (t is the seed of the curve)
func (z *E6) Expt(x *E6) *E6 {

	var result, xInv E6
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

// MulBy014 multiplication by sparse element (c0,c1,0,0,c4,0)
func (z *E6) MulBy014(c0, c1, c4 *fp.Element) *E6 {

	var a, b E3
	var d fp.Element

	a.Set(&z.B0)
	a.MulBy01(c0, c1)

	b.Set(&z.B1)
	b.MulBy1(c4)
	d.Add(c1, c4)

	z.B1.Add(&z.B1, &z.B0)
	z.B1.MulBy01(c0, &d)
	z.B1.Sub(&z.B1, &a)
	z.B1.Sub(&z.B1, &b)
	z.B0.MulByNonResidue(&b)
	z.B0.Add(&z.B0, &a)

	return z
}
