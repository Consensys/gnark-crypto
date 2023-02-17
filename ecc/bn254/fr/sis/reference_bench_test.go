package sis_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/sis"
)

type sisParams struct {
	logTwoBound, logTwoDegree int
}

var params128Bits []sisParams = []sisParams{
	{logTwoBound: 2, logTwoDegree: 3},
	{logTwoBound: 4, logTwoDegree: 4},
	// {logTwoBound: 6, logTwoDegree: 5},
	// {logTwoBound: 10, logTwoDegree: 6},
	// {logTwoBound: 16, logTwoDegree: 7},
	// {logTwoBound: 32, logTwoDegree: 8},
}

const (
	LATENCY_MUL_FIELD_NS int = 18
	LATENCY_ADD_FIELD_NS int = 4
)

// Estimate the theoretical performances that are achievable using ring-SIS
// operations. The time is obtained by counting the number of additions and
// multiplications occurring in the computation. This does not account for the
// possibilities to use SIMD instructions or for cache-locality issues. Thus, it
// does not represents a maximum even though it returns a good idea of what is
// achievable . This returns performances in term of ns/field. This also does not
// account for the time taken for "limb-splitting" the input.
func estimateSisTheory(p sisParams) int {

	// Since the FFT occurs over a coset, we need to multiply all the coefficients
	// of the input by some coset factors (for an entire polynomial)
	timeCosetShift := (1 << p.logTwoDegree) * LATENCY_MUL_FIELD_NS

	// The two additions are from the butterfly, and the multiplication represents
	// the one by the twiddle. (for an entire polynomial)
	timeFFT := (1 << p.logTwoDegree) * p.logTwoDegree * (2*LATENCY_ADD_FIELD_NS + LATENCY_MUL_FIELD_NS)

	// Time taken to multiply by the key and accumulate (for an entire polynomial)
	timeMulAddKey := (1 << p.logTwoDegree) * (LATENCY_MUL_FIELD_NS + LATENCY_ADD_FIELD_NS)

	// Total computation time for an entire polynomial
	totalTimePoly := timeCosetShift + timeFFT + timeMulAddKey

	// Convert this into a time per input field
	return totalTimePoly * fr.Bits / p.logTwoBound / (1 << p.logTwoDegree)
}

func BenchmarkSISRef(b *testing.B) {

	const numFieldInput = 1 << 14

	// Assign the input with random bytes. In practice, theses bytes encodes
	// a string of field element. It would be more meaningful to take a slice
	// of field element directly because otherwise the conversion time is not
	// accounted for in the benchmark.
	inputs := make([]byte, numFieldInput*fr.Bytes)
	if _, err := rand.Read(inputs); err != nil {
		b.Fatal(err)
	}

	for _, param := range params128Bits {

		benchName := fmt.Sprintf("ring-sis/nb-input=%v-log-2-bound=%v-log-2-degree=%v", numFieldInput, param.logTwoBound, param.logTwoDegree)

		b.Run(benchName, func(b *testing.B) {
			instance, err := sis.NewRSis(0, param.logTwoDegree, param.logTwoBound, numFieldInput*fr.Bits/param.logTwoBound)
			if err != nil {
				b.Fatal(err)
			}

			// We introduce a custom metric which is the time per field element
			// Since the benchmark object allows to report extra meta but does
			// not allow accessing them. We measure the time ourself.

			instance.Write(inputs)
			startTime := time.Now()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = instance.Sum(nil)
			}
			b.StopTimer()

			totalDuration := time.Since(startTime)
			nsPerField := totalDuration.Nanoseconds() / int64(b.N) / int64(numFieldInput)

			b.ReportMetric(float64(nsPerField), "ns/field")

			theoritical := estimateSisTheory(param)

			b.ReportMetric(float64(theoritical), "ns/field(theory)")

		})
	}
}

func BenchmarkSISSparseRef(b *testing.B) {

	const numFieldInput = 1 << 14
	const nNonZero = numFieldInput / 8

	// Assign the input with random bytes but als
	inputs := make([]byte, numFieldInput*fr.Bytes)

	if _, err := rand.Read(inputs[numFieldInput-nNonZero:]); err != nil {
		b.Fatal(err)
	}

	for _, param := range params128Bits {

		benchName := fmt.Sprintf("ring-sis/nb-input=%v-log-2-bound=%v-log-2-degree=%v", numFieldInput, param.logTwoBound, param.logTwoDegree)

		b.Run(benchName, func(b *testing.B) {
			instance, err := sis.NewRSis(0, param.logTwoDegree, param.logTwoBound, numFieldInput*fr.Bits/param.logTwoBound)
			if err != nil {
				b.Fatal(err)
			}
			// We introduce a custom metric which is the time per field element
			// Since the benchmark object allows to report extra meta but does
			// not allow accessing them. We measure the time ourself.

			instance.Write(inputs)
			startTime := time.Now()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = instance.Sum(nil)
			}
			b.StopTimer()

			totalDuration := time.Since(startTime)
			nsPerField := totalDuration.Nanoseconds() / int64(b.N) / int64(numFieldInput)

			b.ReportMetric(float64(nsPerField), "ns/field")

			theoritical := estimateSisTheory(param)

			b.ReportMetric(float64(theoritical), "ns/field(theory)")

		})
	}
}
