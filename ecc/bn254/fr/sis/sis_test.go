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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

type TestCases struct {
	Inputs  [][]fr.Element `json:"inputs"`
	Entries []struct {
		Params struct {
			Seed                int64 `json:"seed"`
			LogTwoDegree        int   `json:"logTwoDegree"`
			LogTwoBound         int   `json:"logTwoBound"`
			MaxNbElementsToHash int   `json:"maxNbElementsToHash"`
		} `json:"params"`
		Expected [][]fr.Element `json:"expected"`
	} `json:"entries"`
}

func TestReference(t *testing.T) {
	if bits.UintSize == 32 {
		t.Skip("skipping this test in 32bit.")
	}
	assert := require.New(t)

	// read the test case file
	var testCases TestCases
	data, err := ioutil.ReadFile("test_cases.json")
	assert.NoError(err, "reading test cases failed")
	err = json.Unmarshal(data, &testCases)
	assert.NoError(err, "reading test cases failed")

	for _, testCase := range testCases.Entries {
		// create the SIS instance
		sis, err := NewRSis(testCase.Params.Seed, testCase.Params.LogTwoDegree, testCase.Params.LogTwoBound, testCase.Params.MaxNbElementsToHash)
		assert.NoError(err)

		// key generation same than in sage
		makeKeyDeterminitic(t, sis, testCase.Params.Seed)

		for i, in := range testCases.Inputs {
			sis.Reset()

			// hash test case entry input and compare with expected (computed by sage)
			got, err := sis.Hash(in)
			assert.NoError(err)
			if len(testCase.Expected[i]) == 0 {
				for _, e := range got {
					assert.True(e.IsZero(), "mismatch between reference test and computed value")
				}
			} else {
				assert.EqualValues(testCase.Expected[i], got, "mismatch between reference test and computed value")
			}

			// ensure max nb elements to hash has no incidence on result.
			if len(in) < testCase.Params.MaxNbElementsToHash {
				sis2, err := NewRSis(testCase.Params.Seed, testCase.Params.LogTwoDegree, testCase.Params.LogTwoBound, len(in))
				assert.NoError(err)
				makeKeyDeterminitic(t, sis2, testCase.Params.Seed)

				got2, err := sis2.Hash(in)
				assert.NoError(err)
				if len(testCase.Expected[i]) == 0 {
					for _, e := range got2 {
						assert.True(e.IsZero(), "mismatch between reference test and computed value")
					}
				} else {
					assert.EqualValues(got, got2, "max nb elements to hash change SIS result")
				}
			}

		}
	}

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

func makeKeyDeterminitic(t *testing.T, sis *RSis, _seed int64) {
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
	seed.SetInt64(_seed)
	for i := 0; i < len(sis.A); i++ {
		sis.A[i] = polyRand(seed, sis.Degree)
		copy(sis.Ag[i], sis.A[i])
		sis.Domain.FFT(sis.Ag[i], fft.DIF, fft.OnCoset())
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

	for _, param := range params128Bits {
		for n := 1 << 10; n <= nbInputs; n <<= 1 {
			in := inputs[:n]
			benchmarkSIS(b, in, false, param.logTwoBound, param.logTwoDegree, estimateSisTheory(param))
		}

	}
}

func benchmarkSIS(b *testing.B, input []fr.Element, sparse bool, logTwoBound, logTwoDegree int, theoritical float64) {
	b.Helper()

	n := len(input)

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

		startTime := time.Now()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err = instance.Hash(input)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()

		totalDuration := time.Since(startTime)
		nsPerField := totalDuration.Nanoseconds() / int64(b.N) / int64(n)

		b.ReportMetric(float64(nsPerField), "ns/field")

		b.ReportMetric(theoritical, "ns/field(theory)")

	})
}

// Hash interprets the input vector as a sequence of coefficients of size r.LogTwoBound bits long,
// and return the hash of the polynomial corresponding to the sum sum_i A[i]*m Mod X^{d}+1
//
// It is equivalent to calling r.Write(element.Marshal()); outBytes = r.Sum(nil);
// ! note @gbotrel: this is a place holder, may not make sense
func (r *RSis) Hash(v []fr.Element) ([]fr.Element, error) {
	if len(v) > r.maxNbElementsToHash {
		return nil, fmt.Errorf("can't hash more than %d elements with params provided in constructor", r.maxNbElementsToHash)
	}

	r.Reset()
	for _, e := range v {
		r.Write(e.Marshal())
	}
	sum := r.Sum(nil)
	var rlen [4]byte
	binary.BigEndian.PutUint32(rlen[:], uint32(len(sum)/fr.Bytes))
	reader := io.MultiReader(bytes.NewReader(rlen[:]), bytes.NewReader(sum))
	var result fr.Vector
	_, err := result.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Hash version without mem copy
// // Hash interprets the input vector as a sequence of coefficients of size r.LogTwoBound bits long,
// // and return the hash of the polynomial corresponding to the sum sum_i A[i]*m Mod X^{d}+1
// //
// // It is equivalent to calling r.Write(element.Marshal()); outBytes = r.Sum(nil);
// func (r *RSis) Hash(v []fr.Element) ([]fr.Element, error) {
// 	if len(v) > r.maxNbElementsToHash {
// 		return nil, fmt.Errorf("can't hash more than %d elements with params provided in constructor", r.maxNbElementsToHash)
// 	}

// 	// clear the buffers of the instance.
// 	defer func() {
// 		r.bufMValues.ClearAll()
// 		for i := 0; i < len(r.bufM); i++ {
// 			r.bufM[i].SetZero()
// 		}
// 	}()

// 	// bitwise decomposition of the buffer, in order to build m (the vector to hash)
// 	// as a list of polynomials, whose coefficients are less than r.B bits long.

// 	bitAt := func(v []fr.Element, i int) uint8 {
// 		// v --> slice of bits
// 		// return bit at position i
// 		const n = fr.Bytes * 8 // nb bits per element
// 		nbBits := len(v) * n

// 		if i >= nbBits {
// 			return 0
// 		}

// 		eIndex := i / n
// 		i %= n

// 		// we want bit i of v[eIndex]
// 		j := i / 64
// 		return uint8(v[eIndex][j] >> (i % 64) & 1)

// 	}

// 	// now we can construct m. The input to hash consists of the polynomials
// 	// m[k*r.Degree:(k+1)*r.Degree]
// 	m := r.bufM

// 	// mark blocks m[i*r.Degree : (i+1)*r.Degree] != [0...0]
// 	mValues := r.bufMValues

// 	// we process the input buffer by blocks of r.LogTwoBound bits
// 	// each of these block (<< 64bits) are interpreted as a coefficient
// 	mPos := 0
// 	nbBits := len(v) * fr.Bytes * 8
// 	for i := 0; i < nbBits; mPos++ {
// 		for j := 0; j < r.LogTwoBound; j++ {
// 			// r.LogTwoBound < 64; we just use the first word of our element here,
// 			// and set the bits from LSB to MSB.
// 			m[mPos][0] |= uint64(bitAt(v, i) << j)
// 			i++
// 		}
// 		if m[mPos][0] == 0 {
// 			continue
// 		}
// 		mValues.Set(uint(mPos / r.Degree))
// 	}

// 	// we can hash now.
// 	res := make(fr.Vector, r.Degree)

// 	// method 1: fft
// 	for i := 0; i < len(r.Ag); i++ {
// 		if !mValues.Test(uint(i)) {
// 			// means m[i*r.Degree : (i+1)*r.Degree] == [0...0]
// 			// we can skip this, FFT(0) = 0
// 			continue
// 		}
// 		k := m[i*r.Degree : (i+1)*r.Degree]
// 		r.Domain.FFT(k, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))
// 		mulModAcc(res, r.Ag[i], k)
// 	}
// 	r.Domain.FFTInverse(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1)) // -> reduces mod Xᵈ+1

// 	return res, nil
// }
