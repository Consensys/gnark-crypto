package secp256r1

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/secp256r1/fp"
	"github.com/consensys/gnark-crypto/ecc/secp256r1/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestCardanoRoots(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[SECP256R1] CardanoRoots should return valid roots of x³ − 3x + c = 0", prop.ForAll(
		func(c fp.Element) bool {
			roots := CardanoRoots(c)
			var three fp.Element
			three.SetInt64(3)
			for _, r := range roots {
				// verify r³ − 3r + c = 0
				var r3, threex, lhs fp.Element
				r3.Square(&r).Mul(&r3, &r)
				threex.Mul(&three, &r)
				lhs.Sub(&r3, &threex).Add(&lhs, &c)
				if !lhs.IsZero() {
					return false
				}
			}
			return true
		},
		GenFp(),
	))

	properties.Property("[SECP256R1] CardanoRoots from curve points should find at least one root matching x", prop.ForAll(
		func(s fr.Element) bool {
			// generate a real curve point by scalar multiplication
			var sBig big.Int
			s.BigInt(&sBig)
			var p G1Jac
			p.ScalarMultiplication(&g1Gen, &sBig)
			var pAff G1Affine
			pAff.FromJacobian(&p)

			// c = b − y² so x³ − 3x + c = 0 must have pAff.X as a root
			var b fp.Element
			b.SetString("41058363725152142129326129780047268409114441015993725554835256314039467401291")
			var y2, c fp.Element
			y2.Square(&pAff.Y)
			c.Sub(&b, &y2)

			roots := CardanoRoots(c)
			if len(roots) == 0 {
				return false // must find at least one root
			}
			// verify at least one root matches the known x
			found := false
			for _, r := range roots {
				if r.Equal(&pAff.X) {
					found = true
					break
				}
			}
			return found
		},
		GenFr(),
	))

	properties.Property("[SECP256R1] CardanoRoots with c=0 should return roots of x³ − 3x = 0", prop.ForAll(
		func(_ fp.Element) bool {
			var c fp.Element // zero
			roots := CardanoRoots(c)
			if len(roots) == 0 {
				return false
			}
			var three fp.Element
			three.SetInt64(3)
			for _, r := range roots {
				var r3, threex, lhs fp.Element
				r3.Square(&r).Mul(&r3, &r)
				threex.Mul(&three, &r)
				lhs.Sub(&r3, &threex)
				if !lhs.IsZero() {
					return false
				}
			}
			return true
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
