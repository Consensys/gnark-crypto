package extensions

import (
	fr "github.com/consensys/gnark-crypto/field/koalabear"
)

// q + r'.r = 1, i.e., qInvNeg = - q⁻¹ mod r
// used for Montgomery reduction
const qInvNeg uint64 = 2130706431

// Field modulus q (Fp)
const (
	q0 uint64 = 2130706433
	q  uint64 = q0
)

var qElement = fp.Element{
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
