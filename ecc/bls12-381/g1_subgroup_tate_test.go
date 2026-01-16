package bls12381

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"testing"
)

func TestG1SubGroupMembershipTate(t *testing.T) {
	t.Parallel()
	tab := precomputeTableDefault()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS12-381] IsInSubGroupTate should output true for points on G1", prop.ForAll(
		func(a fp.Element) bool {
			p := MapToG1(a)
			return p.IsInSubGroupTate(tab)
		},
		GenFp(),
	))

	properties.Property("[BLS12-381] IsInSubGroupTate should output false for points not on G1", prop.ForAll(
		func(a fp.Element) bool {
			p := fuzzCofactorOfG1(a)
			var paff G1Affine
			paff.FromJacobian(&p)
			return !paff.IsInSubGroupTate(tab)
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// bench
func BenchmarkG1IsInSubGroupTate(b *testing.B) {
	var p G1Affine
	p.Set(&g1GenAff)
	tab := precomputeTableDefault()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.IsInSubGroupTate(tab)
	}
}
