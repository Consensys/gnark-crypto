package koalabear

import (
	"fmt"
	"os"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestVector16Ops(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = 2
	} else {
		parameters.MinSuccessfulTests = 10
	}
	properties := gopter.NewProperties(parameters)

	sumVectorNaive := func(a Vector) bool {
		var sum Element
		computed := a.Sum16_Naive()
		for i := 0; i < len(a); i++ {
			sum.Add(&sum, &a[i])
		}

		return sum.Equal(&computed)
	}

	sumVectorAvx := func(a Vector) bool {
		var sum Element
		computed := a.Sum16_avx512()
		for i := 0; i < len(a); i++ {
			sum.Add(&sum, &a[i])
		}

		return sum.Equal(&computed)
	}

	sizes := []int{16}
	type genPair struct {
		g1, g2 gopter.Gen
		label  string
	}

	for _, size := range sizes {
		generators := []genPair{
			{genZeroVector(size), genZeroVector(size), "zero vectors"},
			{genMaxVector(size), genMaxVector(size), "max vectors"},
			{genVector(size), genVector(size), "random vectors"},
			{genVector(size), genZeroVector(size), "random and zero vectors"},
		}
		for _, gp := range generators {
			properties.Property(fmt.Sprintf("vector sum naive %d - %s", size, gp.label), prop.ForAll(
				sumVectorNaive,
				gp.g1,
			))
			properties.Property(fmt.Sprintf("vector sum avx %d - %s", size, gp.label), prop.ForAll(
				sumVectorAvx,
				gp.g1,
			))
		}
	}

	properties.TestingRun(t, gopter.NewFormatedReporter(false, 260, os.Stdout))
}

func BenchmarkVector16Ops(b *testing.B) {
	// note; to benchmark against "no asm" version, use the following
	// build tag: -tags purego
	const N = 16
	a1 := make(Vector, N)
	b1 := make(Vector, N)
	var mixer Element
	mixer.SetRandom()
	for i := 1; i < N; i++ {
		a1[i-1].SetUint64(uint64(i)).
			Mul(&a1[i-1], &mixer)
		b1[i-1].SetUint64(^uint64(i)).
			Mul(&b1[i-1], &mixer)
	}

	b.Run(fmt.Sprintf("sum naive %d", N), func(b *testing.B) {
		_a := a1[:N]
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = _a.Sum16_Naive()
		}
	})

	b.Run(fmt.Sprintf("sum avx %d", N), func(b *testing.B) {
		_a := a1[:N]
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = _a.Sum16_avx512()
		}
	})

	// sum generic
	b.Run(fmt.Sprintf("sum generic %d", N), func(b *testing.B) {
		_a := a1[:N]
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = _a.Sum()
		}
	})

}
