package extensions

import (
	fr "github.com/consensys/gnark-crypto/field/babybear"
)

// q + r'.r = 1, i.e., qInvNeg = - q⁻¹ mod r
// used for Montgomery reduction
const qInvNeg uint32 = 2013265919

// Field modulus q (Fr)
const (
	q0 uint32 = 2013265921
	q  uint32 = q0
)

var qElement = fr.Element{
	q0,
}

//go:noescape
func addE2(res, x, y *E2)

//go:noescape
func subE2(res, x, y *E2)

//go:noescape
func doubleE2(res, x *E2)

//go:noescape
func negE2(res, x *E2)

//go:noescape
func mulNonResE2(res, x *E2)

//go:noescape
func squareAdxE2(res, x *E2)

//go:noescape
func mulAdxE2(res, x, y *E2)

// MulByNonResidue multiplies a E2 by (9,1)
func (z *E2) MulByNonResidue(x *E2) *E2 {
	mulNonResE2(z, x)
	return z
}

// Mul sets z to the E2-product of x,y, returns z
func (z *E2) Mul(x, y *E2) *E2 {
	mulAdxE2(z, x, y)
	return z
}

// Square sets z to the E2-product of x,x, returns z
func (z *E2) Square(x *E2) *E2 {
	squareAdxE2(z, x)
	return z
}
