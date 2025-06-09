package bls12377strong

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377-strong/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-377-strong/fr"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// Let h be the cofactor of (E/𝔽p).
// h = 3 * (2 * 1553806976791259819)²
// For a failure probability of 2⁻ᵝ we need to set rounds=⌈β⌉.
// For example β=64 gives rounds=5 and β=128 gives rounds=2.
var bound = big.NewInt(1553806976791259819)
var rounds = 1

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

	properties.Property("[BLS12-377-STRONG] IsInSubGroupBatchNaive test should pass", prop.ForAll(
		func(mixer fr.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}

			_, _, g, _ := Generators()
			result := BatchScalarMultiplicationG1(&g, sampleScalars[:])

			return IsInSubGroupBatchNaive(result)
		},
		GenFr(),
	))

	properties.Property("[BLS12-377-STRONG] IsInSubGroupBatchNaive test should not pass", prop.ForAll(
		func(mixer fr.Element, a fp.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}

			// random points in G1
			_, _, g, _ := Generators()
			result := BatchScalarMultiplicationG1(&g, sampleScalars[:])
			// random points in the h-torsion
			h := fuzzCofactorOfG1(a)
			result[0].FromJacobian(&h)
			h = fuzzCofactorOfG1(a)
			result[nbSamples-1].FromJacobian(&h)

			return !IsInSubGroupBatchNaive(result)
		},
		GenFr(),
		GenFp(),
	))

	properties.Property("[BLS12-377-STRONG] IsInSubGroupBatch test should pass with probability 1-1/2^64", prop.ForAll(
		func(mixer fr.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}

			// random points in G1
			_, _, g, _ := Generators()
			result := BatchScalarMultiplicationG1(&g, sampleScalars[:])

			return IsInSubGroupBatch(result, bound, rounds)
		},
		GenFr(),
	))

	properties.Property("[BLS12-377-STRONG] IsInSubGroupBatch test should not pass", prop.ForAll(
		func(mixer fr.Element, a fp.Element) bool {
			// mixer ensures that all the words of a frElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}

			// random points in G1
			_, _, g, _ := Generators()
			result := BatchScalarMultiplicationG1(&g, sampleScalars[:])

			// random points in the h-torsion
			h := fuzzCofactorOfG1(a)
			result[0].FromJacobian(&h)
			h = fuzzCofactorOfG1(a)
			result[nbSamples-1].FromJacobian(&h)

			return !IsInSubGroupBatch(result, bound, rounds)
		},
		GenFr(),
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestTatePairings(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 1

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS12-377-STRONG] Tate(P3,Q) should be 1", prop.ForAll(
		func(a fr.Element) bool {
			var s big.Int
			a.BigInt(&s)
			_, _, g, _ := Generators()
			g.ScalarMultiplication(&g, &s)
			return isFirstTateOne(g)
		},
		GenFr(),
	))

	properties.Property("[BLS12-377-STRONG] Tate(P11,Q) should be 1", prop.ForAll(
		func(a fr.Element) bool {
			var s big.Int
			a.BigInt(&s)
			_, _, g, _ := Generators()
			g.ScalarMultiplication(&g, &s)
			return isSecondTateOne(g)
		},
		GenFr(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// benches
func BenchmarkIsInSubGroupBatchNaive(b *testing.B) {
	const nbSamples = 1000000

	// mixer ensures that all the words of a frElement are set
	var mixer fr.Element
	mixer.SetRandom()
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer)
	}

	_, _, g, _ := Generators()
	result := BatchScalarMultiplicationG1(&g, sampleScalars[:])

	for _, i := range []uint64{10, 100, 1000, 10000, 100000, 1000000} {
		b.Run(fmt.Sprintf("%d points", i), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				IsInSubGroupBatchNaive(result[:i])
			}
		})
	}
}

func BenchmarkIsInSubGroupBatch(b *testing.B) {
	const nbSamples = 1000000

	// mixer ensures that all the words of a frElement are set
	var mixer fr.Element
	mixer.SetRandom()
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer)
	}

	_, _, g, _ := Generators()
	result := BatchScalarMultiplicationG1(&g, sampleScalars[:])

	for _, i := range []uint64{10, 100, 1000, 10000, 100000, 1000000} {
		b.Run(fmt.Sprintf("%d points", i), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				IsInSubGroupBatch(result[:i], bound, rounds)
			}
		})
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
