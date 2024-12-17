package extensions

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// q + r'.r = 1, i.e., qInvNeg = - q⁻¹ mod r
// used for Montgomery reduction
const qInvNeg uint64 = 725501752471715839

// Field modulus q (Fr)
const (
	q0 uint64 = 725501752471715841
	q1 uint64 = 6461107452199829505
	q2 uint64 = 6968279316240510977
	q3 uint64 = 1345280370688173398
)

var qElement = fr.Element{
	q0,
	q1,
	q2,
	q3,
}

//go:noescape
func addE2(res, x, y *E2)

//go:noescape
func subE2(res, x, y *E2)

//go:noescape
func doubleE2(res, x *E2)

//go:noescape
func negE2(res, x *E2)
