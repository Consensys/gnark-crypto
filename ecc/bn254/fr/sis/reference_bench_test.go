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
	{logTwoBound: 6, logTwoDegree: 5},
	{logTwoBound: 10, logTwoDegree: 6},
	{logTwoBound: 16, logTwoDegree: 7},
	{logTwoBound: 32, logTwoDegree: 8},
}

func BenchmarkSis(b *testing.B) {

	numFieldInput := 1 << 10

	// Assign the input with random bytes. In practice, theses bytes encodes
	// a string of field element. It would be more meaningful to take a slice
	// of field element directly because otherwise the conversion time is not
	// accounted for in the benchmark.
	inputs := make([]byte, numFieldInput*fr.Bytes)
	if _, err := rand.Read(inputs); err != nil {
		panic(err)
	}

	for _, param := range params128Bits {

		instance, err := sis.NewRSis(0, param.logTwoDegree, param.logTwoBound, numFieldInput*fr.Bits/param.logTwoBound)
		if err != nil {
			panic(err)
		}

		benchName := fmt.Sprintf("ring-sis/nb-input=%v-log-2-bound=%v-log-2-degree=%v", numFieldInput, param.logTwoBound, param.logTwoDegree)

		b.Run(benchName, func(b *testing.B) {

			// We introduce a custom metric which is the time per field element
			// Since the benchmark object allows to report extra metra but does
			// not allow accessing them. We measure the time ourself.

			startTime := time.Now()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = instance.Sum(inputs)
			}
			b.StopTimer()

			totalDuration := time.Since(startTime)
			nsPerField := totalDuration.Nanoseconds() / int64(b.N) / int64(numFieldInput)

			b.ReportMetric(float64(nsPerField), "ns/field")
		})
	}

}
