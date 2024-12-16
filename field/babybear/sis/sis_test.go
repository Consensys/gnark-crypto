// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package sis

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/bits"
	"os"
	"testing"
	"time"

	"github.com/bits-and-blooms/bitset"
	"github.com/consensys/gnark-crypto/field/babybear"
	"github.com/consensys/gnark-crypto/field/babybear/fft"
	"github.com/stretchr/testify/require"
)

func TestSage(t *testing.T) {

	sis, err := NewRSis(5, 2, 2, 10)
	if err != nil {
		t.Fatal(err)
	}
	makeKeyDeterministic(t, sis, 5)

	nbElmts := 10
	inputs := make([]babybear.Element, nbElmts)
	inputs[0].SetUint64(5)
	var five babybear.Element
	five.SetUint64(5)
	for i := 1; i < nbElmts; i++ {
		inputs[i].Mul(&inputs[i-1], &five)
	}

	res, err := sis.Hash(inputs)
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println("--key--")
	// for i := 0; i < len(sis.A); i++ {
	// 	for j := 0; j < len(sis.A[i]); j++ {
	// 		fmt.Printf("%s\n", sis.A[i][j].String())
	// 	}
	// 	fmt.Println("")
	// }
	// fmt.Println("-------")

	for i := 0; i < len(res); i++ {
		fmt.Println(res[i].String())
	}

}

type sisParams struct {
	logTwoBound, logTwoDegree int
}

var params128Bits []sisParams = []sisParams{

	// call to limbDecomposeBytes
	{logTwoDegree: 2, logTwoBound: 3},

	// call to limbDecomposeBytesSmallBound
	{logTwoDegree: 2, logTwoBound: 2},
	{logTwoDegree: 2, logTwoBound: 4},

	// call to limbDecomposeBytesSmallBound
	{logTwoDegree: 2, logTwoBound: 8},
	{logTwoDegree: 2, logTwoBound: 16},
	{logTwoDegree: 2, logTwoBound: 32},
}

type TestCases struct {
	Inputs  []babybear.Element `json:"inputs"`
	Entries []struct {
		Params struct {
			Seed                int64 `json:"seed"`
			LogTwoDegree        int   `json:"logTwoDegree"`
			LogTwoBound         int   `json:"logTwoBound"`
			MaxNbElementsToHash int   `json:"maxNbElementsToHash"`
		} `json:"params"`
		Expected []babybear.Element `json:"expected"`
	} `json:"entries"`
}

func TestReference(t *testing.T) {
	if bits.UintSize == 32 {
		t.Skip("skipping this test in 32bit.")
	}
	assert := require.New(t)

	// read the test case file
	var testCases TestCases
	data, err := os.ReadFile("test_cases.json")
	assert.NoError(err, "reading test cases failed")
	err = json.Unmarshal(data, &testCases)
	assert.NoError(err, "reading test cases failed")

	inputs := testCases.Inputs

	for testCaseID, testCase := range testCases.Entries {

		// create the SIS instance
		sis, err := NewRSis(testCase.Params.Seed, testCase.Params.LogTwoDegree, testCase.Params.LogTwoBound, testCase.Params.MaxNbElementsToHash)
		assert.NoError(err)

		// key generation same than in sage
		makeKeyDeterministic(t, sis, testCase.Params.Seed)

		sis.Reset()

		// hash test case entry input and compare with expected (computed by sage)
		goHash, err := sis.Hash(inputs)
		assert.NoError(err)

		assert.EqualValues(
			testCase.Expected, goHash,
			"mismatch between reference test and computed value (testcase %v)",
			testCaseID,
		)

	}

}

func TestLimbDecomposeBytesMiddleBound(t *testing.T) {

	var montConstant babybear.Element
	var bMontConstant big.Int
	bMontConstant.SetUint64(1)
	bMontConstant.Lsh(&bMontConstant, babybear.Bytes*8)
	montConstant.SetBigInt(&bMontConstant)

	nbElmts := 10
	a := make([]babybear.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		a[i].SetUint64(33)
	}
	var buf bytes.Buffer
	for i := 0; i < nbElmts; i++ {
		buf.Write(a[i].Marshal())
	}

	logTwoBound := 8

	for cc := 0; cc < 3; cc++ {
		m := make(babybear.Vector, nbElmts*babybear.Bytes*8/logTwoBound)
		limbDecomposeBytesMiddleBound(buf.Bytes(), m, logTwoBound, 4, nil)

		for i := 0; i < len(m); i++ {
			m[i].Mul(&m[i], &montConstant)
		}

		var x babybear.Element
		x.SetUint64(1 << logTwoBound)

		coeffsPerFieldsElmt := babybear.Bytes * 8 / logTwoBound
		for i := 0; i < nbElmts; i++ {
			r := eval(m[i*coeffsPerFieldsElmt:(i+1)*coeffsPerFieldsElmt], x)
			if !r.Equal(&a[i]) {
				t.Fatal("limbDecomposeBytes failed")
			}
		}
		logTwoBound *= 2
	}

}

func TestLimbDecomposeBytesSmallBound(t *testing.T) {

	var montConstant babybear.Element
	var bMontConstant big.Int
	bMontConstant.SetUint64(1)
	bMontConstant.Lsh(&bMontConstant, babybear.Bytes*8)
	montConstant.SetBigInt(&bMontConstant)

	nbElmts := 10
	a := make([]babybear.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		a[i].SetRandom()
	}
	var buf bytes.Buffer
	for i := 0; i < nbElmts; i++ {
		buf.Write(a[i].Marshal())
	}

	logTwoBound := 2

	for cc := 0; cc < 3; cc++ {

		m := make(babybear.Vector, nbElmts*babybear.Bytes*8/logTwoBound)
		m2 := make(babybear.Vector, nbElmts*babybear.Bytes*8/logTwoBound)

		// the limbs are set as is, they are NOT converted in Montgomery form
		limbDecomposeBytes(buf.Bytes(), m, logTwoBound, 4, nil)
		limbDecomposeBytesSmallBound(buf.Bytes(), m2, logTwoBound, 4, nil)

		for i := 0; i < len(m); i++ {
			m[i].Mul(&m[i], &montConstant)
			m2[i].Mul(&m2[i], &montConstant)
		}
		var x babybear.Element
		x.SetUint64(1 << logTwoBound)

		coeffsPerFieldsElmt := babybear.Bytes * 8 / logTwoBound
		for i := 0; i < nbElmts; i++ {
			r := eval(m[i*coeffsPerFieldsElmt:(i+1)*coeffsPerFieldsElmt], x)
			if !r.Equal(&a[i]) {
				t.Fatal("limbDecomposeBytes failed")
			}
			r = eval(m2[i*coeffsPerFieldsElmt:(i+1)*coeffsPerFieldsElmt], x)
			if !r.Equal(&a[i]) {
				t.Fatal("limbDecomposeBytesSmallBound failed")
			}
		}
		logTwoBound *= 2
	}

}

func eval(p []babybear.Element, x babybear.Element) babybear.Element {
	var res babybear.Element
	for i := len(p) - 1; i >= 0; i-- {
		res.Mul(&res, &x).Add(&res, &p[i])
	}
	return res
}

func TestMulMod(t *testing.T) {

	size := 4

	p := make([]babybear.Element, size)
	q := make([]babybear.Element, size)
	pCopy := make([]babybear.Element, size)
	qCopy := make([]babybear.Element, size)
	for i := 0; i < size; i++ {
		p[i].SetRandom()
		pCopy[i].Set(&p[i])
		q[i].SetRandom()
		qCopy[i].Set(&q[i])
	}

	// creation of the domain
	shift, err := babybear.Generator(uint64(2 * size))
	if err != nil {
		t.Fatal(err)
	}
	var g babybear.Element
	g.Square(&shift)
	domain := fft.NewDomain(uint64(size), fft.WithShift(shift))

	// mul mod
	domain.FFT(p, fft.DIF, fft.OnCoset())
	domain.FFT(q, fft.DIF, fft.OnCoset())
	r := mulMod(p, q)
	domain.FFTInverse(r, fft.DIT, fft.OnCoset())

	// manually check the product on the zeroes of X^4+1
	for i := 0; i < 4; i++ {
		u := eval(pCopy, shift)
		v := eval(qCopy, shift)
		w := eval(r, shift)
		u.Mul(&u, &v)
		if !w.Equal(&u) {
			t.Fatal("mul mol failed")
		}
		shift.Mul(&shift, &g)
	}

}

func makeKeyDeterministic(t *testing.T, sis *RSis, _seed int64) {
	t.Helper()
	// generate the key deterministically, the same way
	// we do in sage to generate the test vectors.

	polyRand := func(seed babybear.Element, deg int) []babybear.Element {
		res := make([]babybear.Element, deg)
		for i := 0; i < deg; i++ {
			res[i].Square(&seed)
			seed.Set(&res[i])
		}
		return res
	}

	var seed, one babybear.Element
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
	r := totalTimePoly * babybear.Bits / p.logTwoBound / (1 << p.logTwoDegree)
	return float64(r)
}

func BenchmarkDecomposition(b *testing.B) {

	nbElmts := 1000
	a := make([]babybear.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		a[i].SetRandom()
	}
	var buf bytes.Buffer
	for i := 0; i < nbElmts; i++ {
		buf.Write(a[i].Marshal())
	}
	logTwoBound := 4
	m := make(babybear.Vector, nbElmts*babybear.Bytes*8/logTwoBound)

	b.Run(fmt.Sprintf("limbDecomposeBytes logTwoBound=%d", logTwoBound), func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			limbDecomposeBytes(buf.Bytes(), m, logTwoBound, 4, nil)
		}
	})

	b.Run(fmt.Sprintf("limbDecomposeByteSmallBound logTwoBound=%d", logTwoBound), func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			limbDecomposeBytesSmallBound(buf.Bytes(), m, logTwoBound, 4, nil)
		}
	})

	logTwoBound = 16
	b.Run(fmt.Sprintf("limbDecomposeBytes logTwoBound=%d", logTwoBound), func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			limbDecomposeBytes(buf.Bytes(), m, logTwoBound, 4, nil)
		}
	})

	b.Run(fmt.Sprintf("limbDecomposeByteSmallBound logTwoBound=%d", logTwoBound), func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			limbDecomposeBytesMiddleBound(buf.Bytes(), m, logTwoBound, 4, nil)
		}
	})

}

func BenchmarkSIS(b *testing.B) {

	// max nb field elements to hash
	const nbInputs = 1 << 16

	// Assign the input with random bytes. In practice, theses bytes encodes
	// a string of field element. It would be more meaningful to take a slice
	// of field element directly because otherwise the conversion time is not
	// accounted for in the benchmark.
	inputs := make(babybear.Vector, nbInputs)
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

func benchmarkSIS(b *testing.B, input []babybear.Element, sparse bool, logTwoBound, logTwoDegree int, theoretical float64) {
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

		b.ReportMetric(theoretical, "ns/field(theory)")

	})
}

// Hash interprets the input vector as a sequence of coefficients of size r.LogTwoBound bits long,
// and return the hash of the polynomial corresponding to the sum sum_i A[i]*m Mod X^{d}+1
//
// It is equivalent to calling r.Write(element.Marshal()); outBytes = r.Sum(nil);
// ! note @gbotrel: this is a place holder, may not make sense
func (r *RSis) Hash(v []babybear.Element) ([]babybear.Element, error) {
	if len(v) > r.maxNbElementsToHash {
		return nil, fmt.Errorf("can't hash more than %d elements with params provided in constructor", r.maxNbElementsToHash)
	}

	r.Reset()
	for _, e := range v {
		r.Write(e.Marshal())
	}
	sum := r.Sum(nil)
	var rlen [4]byte
	binary.BigEndian.PutUint32(rlen[:], uint32(len(sum)/babybear.Bytes))
	reader := io.MultiReader(bytes.NewReader(rlen[:]), bytes.NewReader(sum))
	var result babybear.Vector
	_, err := result.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func TestLimbDecompositionFastPath(t *testing.T) {
	assert := require.New(t)

	for size := babybear.Bytes; size < 5*babybear.Bytes; size += babybear.Bytes {
		// Test the fast path of limbDecomposeBytes8_64
		buf := make([]byte, size)
		m := make([]babybear.Element, size)
		mValues := bitset.New(uint(size))
		n := make([]babybear.Element, size)
		nValues := bitset.New(uint(size))

		// Generate a random buffer
		_, err := rand.Read(buf)
		assert.NoError(err)

		limbDecomposeBytes8_64(buf, m, mValues)
		limbDecomposeBytes(buf, n, 8, 64, nValues)

		for i := 0; i < size; i++ {
			assert.Equal(mValues.Test(uint(i)), nValues.Test(uint(i)))
			assert.True(m[i].Equal(&n[i]))
		}
	}

}

func TestUnrolledFFT(t *testing.T) {

	var shift babybear.Element
	shift.SetRandom()

	const size = 64
	assert := require.New(t)
	domain := fft.NewDomain(size, fft.WithShift(shift))

	k1 := make([]babybear.Element, size)
	for i := 0; i < size; i++ {
		k1[i].SetRandom()
	}
	k2 := make([]babybear.Element, size)
	copy(k2, k1)

	// default FFT
	domain.FFT(k1, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))

	// unrolled FFT
	twiddlesCoset := PrecomputeTwiddlesCoset(domain.Generator, domain.FrMultiplicativeGen)
	FFT64(k2, twiddlesCoset)

	// compare results
	for i := 0; i < size; i++ {
		// fmt.Printf("i = %d, k1 = %v, k2 = %v\n", i, k1[i].String(), k2[i].String())
		assert.True(k1[i].Equal(&k2[i]), "i = %d", i)
	}
}
