package fptower

import (
	"github.com/consensys/gnark-crypto/ecc"
	"math/big"
)

// Expt set z to x^t in E12 and return z (t is the generator of the BN curve)
func (z *E12) Expt(x *E12) *E12 {

	var tAbsNAF [63]int8
	optimaAteLoop, _ := new(big.Int).SetString("4965661367192848881", 10)
	ecc.NafDecomposition(optimaAteLoop, tAbsNAF[:])

	var result, xInv E12
	result.Set(x)
	xInv.Conjugate(x)

	for i := 61; i >= 0; i-- {
		result.CyclotomicSquare(&result)
		if tAbsNAF[i] == 1 {
			result.Mul(&result, x)
		} else if tAbsNAF[i] == -1 {
			result.Mul(&result, &xInv)
		}
	}

	z.Set(&result)
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
