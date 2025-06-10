package bls12377

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// For highly 2-adic curves the bound is always 2.
// For a failure probability of 2⁻ᵝ we need to set rounds=β.
// For example β=64 gives rounds=64 and β=128 gives rounds=128.
var rounds = 64

func TestIsInSubGroupBatch(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = 1
	} else {
		parameters.MinSuccessfulTests = 100
	}

	properties := gopter.NewProperties(parameters)

	// number of points to test
	const nbSamples = 100

	properties.Property("[BLS12-377] IsInSubGroupBatchNaive test should pass", prop.ForAll(
		func(mixer fr.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}
			result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])

			return IsInSubGroupBatchNaive(result)
		},
		GenFr(),
	))

	properties.Property("[BLS12-377] IsInSubGroupBatchNaive test should not pass", prop.ForAll(
		func(mixer fr.Element, a fp.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}
			result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])

			// random point in the h-torsion
			h := fuzzCofactorOfG1(a)
			result[0].FromJacobian(&h)

			return !IsInSubGroupBatchNaive(result)
		},
		GenFr(),
		GenFp(),
	))

	properties.Property("[BLS12-377] IsInSubGroupBatch test should pass with probability 1-1/2^64", prop.ForAll(
		func(mixer fr.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}
			result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])

			return IsInSubGroupBatch(result, rounds)
		},
		GenFr(),
	))

	properties.Property("[BLS12-377] IsInSubGroupBatch test should not pass", prop.ForAll(
		func(mixer fr.Element, a fp.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}
			result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])

			// random points in the h-torsion
			h := fuzzCofactorOfG1(a)
			result[0].FromJacobian(&h)
			h = fuzzCofactorOfG1(a)
			result[nbSamples-1].FromJacobian(&h)

			return !IsInSubGroupBatch(result, rounds)
		},
		GenFr(),
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// benches
func BenchmarkIsInSubGroupBatchNaive(b *testing.B) {
	const nbSamples = 1000
	// mixer ensures that all the words of a frElement are set
	var mixer fr.Element
	mixer.SetRandom()
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer)
	}
	result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsInSubGroupBatchNaive(result)
	}

}

func BenchmarkIsInSubGroupBatch(b *testing.B) {
	const nbSamples = 1000
	// mixer ensures that all the words of a frElement are set
	var mixer fr.Element
	mixer.SetRandom()
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer)
	}
	result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsInSubGroupBatch(result, rounds)
	}

}

// utils
func fuzzCofactorOfG1(f fp.Element) G1Jac {
	var res, jac G1Jac
	aff := MapToCurve1(&f)
	jac.FromAffine(&aff)
	// p+x²ϕ(p) = [r]p
	res.phi(&jac).
		mulBySeed(&res).
		mulBySeed(&res)
	res.AddAssign(&jac)
	return res
}
