// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package sis

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/bits"
	"os"
	"testing"

	"encoding/binary"
	"github.com/consensys/gnark-crypto/field/babybear"
	"github.com/consensys/gnark-crypto/field/babybear/fft"
	"github.com/stretchr/testify/require"
)

type sisParams struct {
	logTwoBound, logTwoDegree int
}

var params128Bits []sisParams = []sisParams{
	{logTwoBound: 8, logTwoDegree: 5},
	{logTwoBound: 8, logTwoDegree: 6},
	{logTwoBound: 16, logTwoDegree: 6},
	{logTwoBound: 16, logTwoDegree: 9},
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
		if testCase.Params.LogTwoBound%8 != 0 {
			t.Logf("skipping test case %d, logTwoBound is not a multiple of 8", testCaseID)
			continue
		}
		if testCase.Params.LogTwoBound > babybear.Bits {
			t.Logf("skipping test case %d, logTwoBound %d is greater than field bit size (%d)", testCaseID, testCase.Params.LogTwoBound, babybear.Bits)
			continue
		}
		t.Logf("logTwoBound = %d, logTwoDegree = %d", testCase.Params.LogTwoBound, testCase.Params.LogTwoDegree)

		// create the SIS instance
		sis, err := NewRSis(testCase.Params.Seed, testCase.Params.LogTwoDegree, testCase.Params.LogTwoBound, testCase.Params.MaxNbElementsToHash)
		assert.NoError(err)

		// key generation same than in sage
		makeKeyDeterministic(t, sis, testCase.Params.Seed)

		// hash test case entry input and compare with expected (computed by sage)
		goHash := make([]babybear.Element, 1<<testCase.Params.LogTwoDegree)
		err = sis.Hash(inputs, goHash)
		assert.NoError(err)

		assert.EqualValues(
			testCase.Expected, goHash,
			"mismatch between reference test and computed value (testcase %v)",
			testCaseID,
		)

	}

}

func TestLimbDecomposeBytes(t *testing.T) {
	assert := require.New(t)

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

	logTwoBound := 8

	for cc := 0; cc < 1; cc++ {
		vr := NewLimbIterator(&VectorIterator{v: a}, logTwoBound/8)
		m := make(babybear.Vector, nbElmts*babybear.Bytes*8/logTwoBound)
		var ok bool
		for i := 0; i < len(m); i++ {
			m[i][0], ok = vr.NextLimb()
			assert.True(ok)
		}

		for i := 0; i < len(m); i++ {
			m[i].Mul(&m[i], &montConstant)
		}

		var x babybear.Element
		x.SetUint64(1 << logTwoBound)

		coeffsPerFieldsElmt := babybear.Bytes * 8 / logTwoBound
		for i := 0; i < nbElmts; i++ {
			r := eval(m[i*coeffsPerFieldsElmt:(i+1)*coeffsPerFieldsElmt], x)
			assert.True(r.Equal(&a[i]), "limbDecomposeBytes failed")
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
	if sis.hasFast512_16 {
		sis.agShuffled = make([][]babybear.Element, len(sis.Ag))
		for i := range sis.agShuffled {
			sis.agShuffled[i] = make([]babybear.Element, sis.Degree)
			copy(sis.agShuffled[i], sis.Ag[i])
			sisShuffle_avx512(sis.agShuffled[i])
		}
	}
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
			benchmarkSIS(b, in, false, param.logTwoBound, param.logTwoDegree)
		}

	}
}

func benchmarkSIS(b *testing.B, input []babybear.Element, sparse bool, logTwoBound, logTwoDegree int) {
	b.Helper()

	n := len(input)

	benchName := "ring-sis/"
	if sparse {
		benchName += "sparse/"
	}
	benchName += fmt.Sprintf("inputs=%v/log2-bound=%v/log2-degree=%v", n, logTwoBound, logTwoDegree)

	b.Run(benchName, func(b *testing.B) {
		// report the throughput in MB/s
		b.SetBytes(int64(len(input)) * babybear.Bytes)

		instance, err := NewRSis(0, logTwoDegree, logTwoBound, n)
		if err != nil {
			b.Fatal(err)
		}

		res := make([]babybear.Element, 1<<logTwoDegree)

		// We introduce a custom metric which is the time per field element
		// Since the benchmark object allows to report extra meta but does
		// not allow accessing them. We measure the time ourself.

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = instance.Hash(input, res)
		}
	})
}
func FuzzSISAvx512(f *testing.F) {
	if !supportAVX512 {
		f.Skip("AVX512 not supported")
	}

	const logTwoBound = 16
	const logTwoDegree = 9

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < babybear.Bytes+8 {
			t.Skip("not enough data")
		}

		// Extract the seed from the data
		seed := int64(binary.LittleEndian.Uint64(data[:8]))
		data = data[8:]

		// Create a new RSIS instance
		instance, err := NewRSis(seed, logTwoDegree, logTwoBound, len(data)/babybear.Bytes)
		if err != nil {
			t.Fatal(err)
		}

		a0 := make([]babybear.Element, len(data)/babybear.Bytes)
		a1 := make([]babybear.Element, len(data)/babybear.Bytes)

		for i := range a0 {
			a0[i][0] = binary.LittleEndian.Uint32(data[i*babybear.Bytes:])
			a0[i][0] %= 2013265921
		}

		copy(a1[:], a0[:])

		// Call the AVX512
		var res0, res1 [512]babybear.Element
		err = instance.Hash(a0, res0[:])
		if err != nil {
			t.Fatal(err)
		}

		instance.hasFast512_16 = false
		// call the generic --> note that this still may call FFT avx512 code
		err = instance.Hash(a1, res1[:])
		if err != nil {
			t.Fatal(err)
		}

		// compare the results
		for i := range res0 {
			if res0[i][0] != res1[i][0] {
				t.Fatal("results differ")
			}
		}

	})
}
