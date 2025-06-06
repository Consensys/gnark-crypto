package bls12381

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestIsInSubGroupBatch(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = 1
	} else {
		parameters.MinSuccessfulTests = 100
	}

	properties := gopter.NewProperties(parameters)

	// size of the multiExps
	const nbSamples = 100

	properties.Property("[BLS12-381] IsInSubGroupBatchNaive test should pass", prop.ForAll(
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

	properties.Property("[BLS12-381] IsInSubGroupBatchNaive test should not pass", prop.ForAll(
		func(mixer fr.Element, a fp.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}

			result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])
			h := fuzzCofactorOfG1(a)
			result[0].FromJacobian(&h)

			return !IsInSubGroupBatchNaive(result)
		},
		GenFr(),
		GenFp(),
	))

	properties.Property("[BLS12-381] IsInSubGroupBatch test should pass with probability 1-1/2^64", prop.ForAll(
		func(mixer fr.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}

			// random points in G1
			result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])

			bound := big.NewInt(10177)
			rounds := 5
			return IsInSubGroupBatch(result, bound, rounds)
		},
		GenFr(),
	))

	properties.Property("[BLS12-381] IsInSubGroupBatch test should not pass with probability 1-1/2^64", prop.ForAll(
		func(mixer fr.Element, a fp.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}

			// random points in G1
			result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])

			// random point in the h-torsion
			h := fuzzCofactorOfG1(a)
			result[0].FromJacobian(&h)

			bound := big.NewInt(10177)
			rounds := 5
			return !IsInSubGroupBatch(result, bound, rounds)
		},
		GenFr(),
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestIsInSubGroupBatchProbabilistic(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 1

	properties := gopter.NewProperties(parameters)

	// size of the multiExps
	const nbSamples = 100

	properties.Property("[BLS12-381] IsInSubGroupBatch should pass with probability 1/3^rounds although no point is in G1", prop.ForAll(
		func(mixer fr.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element
			result := make([]G1Affine, nbSamples)

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
				// all points are of order 3
				result[i-1].X.SetUint64(0)
				result[i-1].Y.SetUint64(2)
			}

			bound := big.NewInt(10177)
			rounds := 5
			return !IsInSubGroupBatch(result, bound, rounds)
		},
		GenFr(),
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// benches
func BenchmarkIsInSubGroupBatchNaive(b *testing.B) {
	const nbSamples = 100
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
	const nbSamples = 100
	// mixer ensures that all the words of a frElement are set
	var mixer fr.Element
	mixer.SetRandom()
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer)
	}

	result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])
	bound := big.NewInt(10177)
	round := 5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsInSubGroupBatch(result, bound, round)
	}

}
