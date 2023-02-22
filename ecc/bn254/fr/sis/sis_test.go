// Copyright 2023 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sis

import (
	"fmt"
	"math/bits"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"github.com/stretchr/testify/require"
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

func TestReference(t *testing.T) {
	if bits.UintSize == 32 {
		t.Skip("skipping this test in 32bit.")
	}
	assert := require.New(t)

	const (
		logTwoBound = 4
		degree      = 4
	)

	sis, err := NewRSis(5, 2, logTwoBound, 1)
	assert.NoError(err)
	ssis := sis.(*RSis)
	makeKeyDeterminitic(t, ssis)

	// message to hash
	var m fr.Element
	m.SetString("21888242871839275222246405745257275088548364400416034343698204186575808495614")
	_, err = ssis.Write(m.Marshal())
	assert.NoError(err)

	got := ssis.Sum(nil)

	// compare expected against computed
	expected := []byte{0x17, 0xcd, 0xe4, 0x27, 0xaa, 0x1, 0x3e, 0xd1, 0xc5, 0x4d, 0x1, 0xef, 0xa4, 0x6b, 0x6, 0xfc, 0xc4, 0xbe, 0x86, 0x91, 0xfc, 0xd7, 0x4a, 0xcf, 0x33, 0x8d, 0xc0, 0x80, 0xa1, 0x86, 0x7, 0x3b, 0xd, 0x50, 0x3d, 0x4, 0xa9, 0x88, 0xd5, 0xd3, 0x1c, 0x85, 0xe9, 0xea, 0x22, 0x6f, 0xc0, 0xac, 0x8c, 0xa4, 0xc4, 0x5f, 0x3b, 0x65, 0xac, 0xfc, 0xd8, 0x53, 0xf1, 0xf8, 0xf5, 0xe2, 0x6f, 0x9d, 0x23, 0xb9, 0x8b, 0x41, 0xb3, 0xab, 0xbd, 0x38, 0x28, 0xd8, 0xe6, 0x54, 0xee, 0x5f, 0x17, 0x43, 0xf9, 0x9b, 0x51, 0x2d, 0xfb, 0xeb, 0xc8, 0x60, 0x6c, 0x9a, 0x2d, 0xaa, 0x1c, 0xc0, 0x49, 0xa8, 0x12, 0xad, 0xc0, 0x9, 0x27, 0x9a, 0x90, 0xea, 0x95, 0x68, 0x57, 0x3f, 0x3a, 0x3d, 0xc1, 0x19, 0x63, 0xcb, 0xcc, 0x35, 0xd3, 0x18, 0xa5, 0x7c, 0x18, 0x71, 0xf7, 0xec, 0xd1, 0x2, 0xab, 0xa5}
	assert.EqualValues(expected, got, "hash does not match expected result")

	// [ Sage comparison ]
	// m = Fr(21888242871839275222246405745257275088548364400416034343698204186575808495614)
	// mb = toBytes(m)
	// mb = toBytes(m, 32)
	// sis = Sis(5, 16, 4,4)
	// h = sis.sum(mc)
	// res =[]
	// for i in range(4):
	// 		res += toBytes(lift(h.coefficients()[i]), 32)

}

func TestMulMod(t *testing.T) {

	size := 4

	p := make([]fr.Element, size)
	p[0].SetString("2389")
	p[1].SetString("987192")
	p[2].SetString("623")
	p[3].SetString("91")

	q := make([]fr.Element, size)
	q[0].SetString("76755")
	q[1].SetString("232893720")
	q[2].SetString("989273")
	q[3].SetString("675273")

	// creation of the domain
	var shift fr.Element
	shift.SetString("19540430494807482326159819597004422086093766032135589407132600596362845576832")
	domain := fft.NewDomain(uint64(size), shift)

	// mul mod
	domain.FFT(p, fft.DIF, fft.OnCoset())
	domain.FFT(q, fft.DIF, fft.OnCoset())
	r := mulMod(p, q)
	domain.FFTInverse(r, fft.DIT, fft.OnCoset())

	// expected result
	expectedr := make([]fr.Element, 4)
	expectedr[0].SetString("21888242871839275222246405745257275088548364400416034343698204185887558114297")
	expectedr[1].SetString("631644300118")
	expectedr[2].SetString("229913166975959")
	expectedr[3].SetString("1123315390878")

	for i := 0; i < 4; i++ {
		if !expectedr[i].Equal(&r[i]) {
			t.Fatal("product failed")
		}
	}

}

func makeKeyDeterminitic(t *testing.T, ssis *RSis) {
	t.Helper()
	// generate the key deterministically, the same way
	// we do in sage to generate the test vectors.

	polyRand := func(seed fr.Element, deg int) []fr.Element {
		res := make([]fr.Element, deg)
		for i := 0; i < deg; i++ {
			res[i].Square(&seed)
			seed.Set(&res[i])
		}
		return res
	}

	var seed, one fr.Element
	one.SetOne()
	seed.SetUint64(5)
	for i := 0; i < len(ssis.A); i++ {
		ssis.A[i] = polyRand(seed, ssis.Degree)
		copy(ssis.Ag[i], ssis.A[i])
		ssis.Domain.FFT(ssis.Ag[i], fft.DIF, fft.OnCoset())
		seed.Add(&seed, &one)
	}
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
func estimateSisTheory(p sisParams) float64 {

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
	r := totalTimePoly * fr.Bits / p.logTwoBound / (1 << p.logTwoDegree)
	return float64(r)
}

func BenchmarkSIS(b *testing.B) {

	// max nb field elements to hash
	const nbInputs = 1 << 16

	// Assign the input with random bytes. In practice, theses bytes encodes
	// a string of field element. It would be more meaningful to take a slice
	// of field element directly because otherwise the conversion time is not
	// accounted for in the benchmark.
	inputs := make(fr.Vector, nbInputs)
	for i := 0; i < len(inputs); i++ {
		inputs[i].SetRandom()
	}
	bInputs, err := inputs.MarshalBinary()
	if err != nil {
		b.Fatal(err)
	}
	bInputs = bInputs[4:] // ignore first 4 bytes that encode len.

	for _, param := range params128Bits {
		for n := 1 << 10; n <= nbInputs; n <<= 1 {
			in := bInputs[:n*fr.Bytes]
			benchmarkSIS(b, in, false, param.logTwoBound, param.logTwoDegree, estimateSisTheory(param))
		}

	}
}

func benchmarkSIS(b *testing.B, input []byte, sparse bool, logTwoBound, logTwoDegree int, theoritical float64) {
	b.Helper()

	n := len(input) / fr.Bytes

	benchName := "ring-sis/"
	if sparse {
		benchName += "sparse/"
	}
	benchName += fmt.Sprintf("inputs=%v/log2-bound=%v/log2-degree=%v", n, logTwoBound, logTwoDegree)

	b.Run(benchName, func(b *testing.B) {
		instance, err := NewRSis(0, logTwoDegree, logTwoBound, n)
		if err != nil {
			b.Fatal(err)
		}

		// We introduce a custom metric which is the time per field element
		// Since the benchmark object allows to report extra meta but does
		// not allow accessing them. We measure the time ourself.

		instance.Write(input)

		startTime := time.Now()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = instance.Sum(nil)
		}
		b.StopTimer()

		totalDuration := time.Since(startTime)
		nsPerField := totalDuration.Nanoseconds() / int64(b.N) / int64(n)

		b.ReportMetric(float64(nsPerField), "ns/field")

		b.ReportMetric(theoritical, "ns/field(theory)")

	})
}
