package bls12381

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestG1SubGroupMembershipTateChain(t *testing.T) {
	t.Parallel()
	tab := precomputeChainTableDefault()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS12-381] IsInSubGroupTateChain should output true for points on G1", prop.ForAll(
		func(a fp.Element) bool {
			p := MapToG1(a)
			return p.IsInSubGroupTateChain(tab)
		},
		GenFp(),
	))

	properties.Property("[BLS12-381] IsInSubGroupTateChain should output false for points not on G1", prop.ForAll(
		func(a fp.Element) bool {
			p := fuzzCofactorOfG1(a)
			var paff G1Affine
			paff.FromJacobian(&p)
			return !paff.IsInSubGroupTateChain(tab)
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG1SubGroupTateChainConsistency(t *testing.T) {
	t.Parallel()
	tabOriginal := precomputeTableDefault()
	tabChain := precomputeChainTableDefault()

	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS12-381] IsInSubGroupTateChain should match IsInSubGroupTate for G1 points", prop.ForAll(
		func(a fp.Element) bool {
			p := MapToG1(a)
			return p.IsInSubGroupTate(tabOriginal) == p.IsInSubGroupTateChain(tabChain)
		},
		GenFp(),
	))

	properties.Property("[BLS12-381] IsInSubGroupTateChain should match IsInSubGroupTate for non-G1 points", prop.ForAll(
		func(a fp.Element) bool {
			p := fuzzCofactorOfG1(a)
			var paff G1Affine
			paff.FromJacobian(&p)
			return paff.IsInSubGroupTate(tabOriginal) == paff.IsInSubGroupTateChain(tabChain)
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Benchmarks
func BenchmarkG1IsInSubGroupTateChain(b *testing.B) {
	var p G1Affine
	p.Set(&g1GenAff)
	tab := precomputeChainTableDefault()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.IsInSubGroupTateChain(tab)
	}
}

func BenchmarkG1IsInSubGroupTateChainVsOriginal(b *testing.B) {
	var p G1Affine
	p.Set(&g1GenAff)

	b.Run("Original", func(b *testing.B) {
		tab := precomputeTableDefault()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.IsInSubGroupTate(tab)
		}
	})

	b.Run("Chain", func(b *testing.B) {
		tab := precomputeChainTableDefault()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.IsInSubGroupTateChain(tab)
		}
	})

	b.Run("Fast", func(b *testing.B) {
		tab := precomputeChainTableDefault()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.IsInSubGroupTateFast(tab)
		}
	})

	b.Run("Combined", func(b *testing.B) {
		tab := precomputeChainTableDefault()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.IsInSubGroupTateCombined(tab)
		}
	})

	b.Run("Probabilistic", func(b *testing.B) {
		tab := precomputeChainTableDefault()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.IsInSubGroupTateProbabilistic(tab)
		}
	})

	b.Run("Scott", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.IsInSubGroup()
		}
	})
}

func TestG1SubGroupTateFastConsistency(t *testing.T) {
	t.Parallel()
	tab := precomputeChainTableDefault()

	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS12-381] IsInSubGroupTateFast should match IsInSubGroupTateChain for G1 points", prop.ForAll(
		func(a fp.Element) bool {
			p := MapToG1(a)
			return p.IsInSubGroupTateChain(tab) == p.IsInSubGroupTateFast(tab)
		},
		GenFp(),
	))

	properties.Property("[BLS12-381] IsInSubGroupTateFast should match IsInSubGroupTateChain for non-G1 points", prop.ForAll(
		func(a fp.Element) bool {
			p := fuzzCofactorOfG1(a)
			var paff G1Affine
			paff.FromJacobian(&p)
			return paff.IsInSubGroupTateChain(tab) == paff.IsInSubGroupTateFast(tab)
		},
		GenFp(),
	))

	properties.Property("[BLS12-381] IsInSubGroupTateCombined should match IsInSubGroupTateChain for G1 points", prop.ForAll(
		func(a fp.Element) bool {
			p := MapToG1(a)
			return p.IsInSubGroupTateChain(tab) == p.IsInSubGroupTateCombined(tab)
		},
		GenFp(),
	))

	properties.Property("[BLS12-381] IsInSubGroupTateCombined should match IsInSubGroupTateChain for non-G1 points", prop.ForAll(
		func(a fp.Element) bool {
			p := fuzzCofactorOfG1(a)
			var paff G1Affine
			paff.FromJacobian(&p)
			return paff.IsInSubGroupTateChain(tab) == paff.IsInSubGroupTateCombined(tab)
		},
		GenFp(),
	))

	properties.Property("[BLS12-381] IsInSubGroupTateProbabilistic should output true for G1 points", prop.ForAll(
		func(a fp.Element) bool {
			p := MapToG1(a)
			return p.IsInSubGroupTateProbabilistic(tab)
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
