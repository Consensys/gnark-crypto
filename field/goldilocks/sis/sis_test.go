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

	"github.com/consensys/gnark-crypto/field/goldilocks"
	"github.com/consensys/gnark-crypto/field/goldilocks/fft"
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
	Inputs  []goldilocks.Element `json:"inputs"`
	Entries []struct {
		Params struct {
			Seed                int64 `json:"seed"`
			LogTwoDegree        int   `json:"logTwoDegree"`
			LogTwoBound         int   `json:"logTwoBound"`
			MaxNbElementsToHash int   `json:"maxNbElementsToHash"`
		} `json:"params"`
		Expected []goldilocks.Element `json:"expected"`
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
		if testCase.Params.LogTwoBound > goldilocks.Bits {
			t.Logf("skipping test case %d, logTwoBound %d is greater than field bit size (%d)", testCaseID, testCase.Params.LogTwoBound, goldilocks.Bits)
			continue
		}
		t.Logf("logTwoBound = %d, logTwoDegree = %d", testCase.Params.LogTwoBound, testCase.Params.LogTwoDegree)

		// create the SIS instance
		sis, err := NewRSis(testCase.Params.Seed, testCase.Params.LogTwoDegree, testCase.Params.LogTwoBound, testCase.Params.MaxNbElementsToHash)
		assert.NoError(err)

		// key generation same than in sage
		makeKeyDeterministic(t, sis, testCase.Params.Seed)

		// hash test case entry input and compare with expected (computed by sage)
		goHash := make([]goldilocks.Element, 1<<testCase.Params.LogTwoDegree)
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

	var montConstant goldilocks.Element
	var bMontConstant big.Int
	bMontConstant.SetUint64(1)
	bMontConstant.Lsh(&bMontConstant, goldilocks.Bytes*8)
	montConstant.SetBigInt(&bMontConstant)

	nbElmts := 10
	a := make([]goldilocks.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		a[i].MustSetRandom()
	}

	logTwoBound := 8

	for cc := 0; cc < 3; cc++ {
		vr := NewLimbIterator(&VectorIterator{v: a}, logTwoBound/8)
		m := make(goldilocks.Vector, nbElmts*goldilocks.Bytes*8/logTwoBound)
		var ok bool
		for i := 0; i < len(m); i++ {
			m[i][0], ok = vr.NextLimb()
			assert.True(ok)
		}

		for i := 0; i < len(m); i++ {
			m[i].Mul(&m[i], &montConstant)
		}

		var x goldilocks.Element
		x.SetUint64(1 << logTwoBound)

		coeffsPerFieldsElmt := goldilocks.Bytes * 8 / logTwoBound
		for i := 0; i < nbElmts; i++ {
			r := eval(m[i*coeffsPerFieldsElmt:(i+1)*coeffsPerFieldsElmt], x)
			assert.True(r.Equal(&a[i]), "limbDecomposeBytes failed")
		}
		logTwoBound *= 2
	}

}

func eval(p []goldilocks.Element, x goldilocks.Element) goldilocks.Element {
	var res goldilocks.Element
	for i := len(p) - 1; i >= 0; i-- {
		res.Mul(&res, &x).Add(&res, &p[i])
	}
	return res
}

func makeKeyDeterministic(t *testing.T, sis *RSis, _seed int64) {
	t.Helper()
	// generate the key deterministically, the same way
	// we do in sage to generate the test vectors.

	polyRand := func(seed goldilocks.Element, deg int) []goldilocks.Element {
		res := make([]goldilocks.Element, deg)
		for i := 0; i < deg; i++ {
			res[i].Square(&seed)
			seed.Set(&res[i])
		}
		return res
	}

	var seed, one goldilocks.Element
	one.SetOne()
	seed.SetInt64(_seed)
	for i := 0; i < len(sis.A); i++ {
		sis.A[i] = polyRand(seed, sis.Degree)
		copy(sis.Ag[i], sis.A[i])
		sis.Domain.FFT(sis.Ag[i], fft.DIF, fft.OnCoset())
		seed.Add(&seed, &one)
	}
}

func BenchmarkSIS(b *testing.B) {

	// max nb field elements to hash
	const nbInputs = 1 << 16

	// Assign the input with random bytes. In practice, theses bytes encodes
	// a string of field element. It would be more meaningful to take a slice
	// of field element directly because otherwise the conversion time is not
	// accounted for in the benchmark.
	inputs := make(goldilocks.Vector, nbInputs)
	for i := 0; i < len(inputs); i++ {
		inputs[i].MustSetRandom()
	}

	for _, param := range params128Bits {
		for n := 1 << 10; n <= nbInputs; n <<= 1 {
			in := inputs[:n]
			benchmarkSIS(b, in, false, param.logTwoBound, param.logTwoDegree)
		}

	}
}

func benchmarkSIS(b *testing.B, input []goldilocks.Element, sparse bool, logTwoBound, logTwoDegree int) {
	b.Helper()

	n := len(input)

	benchName := "ring-sis/"
	if sparse {
		benchName += "sparse/"
	}
	benchName += fmt.Sprintf("inputs=%v/log2-bound=%v/log2-degree=%v", n, logTwoBound, logTwoDegree)

	b.Run(benchName, func(b *testing.B) {
		// report the throughput in MB/s
		b.SetBytes(int64(len(input)) * goldilocks.Bytes)

		instance, err := NewRSis(0, logTwoDegree, logTwoBound, n)
		if err != nil {
			b.Fatal(err)
		}

		res := make([]goldilocks.Element, 1<<logTwoDegree)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = instance.Hash(input, res)
		}
	})
}
