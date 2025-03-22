package vortex

import (
	"encoding/binary"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
)

type testcaseVortex struct {
	M               [][]koalabear.Element
	X               fext.E4
	Ys              []fext.E4
	Alpha           fext.E4
	SelectedColumns []int
}

func TestZeroMatrix(t *testing.T) {

	var (
		numCol = 16
		numRow = 8
	)

	var (
		m               = make([][]koalabear.Element, numRow)
		x               = fext.E4{}
		y               = make([]fext.E4, numRow)
		alpha, _        = new(fext.E4).SetRandom()
		selectedColumns = []int{0, 1, 2, 3}
	)

	for i := range m {
		m[i] = make([]koalabear.Element, numCol)
	}

	runTest(t, &testcaseVortex{
		M:               m,
		X:               x,
		Ys:              y,
		Alpha:           *alpha,
		SelectedColumns: selectedColumns,
	})

}

func TestFullRandom(t *testing.T) {

	var (
		numCol = 16
		numRow = 8
		// #nosec G404 -- test case generation does not require a cryptographic PRNG
		rng = rand.New(rand.NewChaCha8([32]byte{}))
	)

	var (
		m               = make([][]koalabear.Element, numRow)
		x               = randFext(rng)
		ys              = make([]fext.E4, numRow)
		alpha           = randFext(rng)
		selectedColumns = []int{0, 1, 2, 3}
		err             error
	)

	for i := range m {
		m[i] = make([]koalabear.Element, numCol)
		for j := range m[i] {
			m[i][j] = randElement(rng)
		}

		ys[i], err = EvalBasePolyLagrange(m[i], x)
		if err != nil {
			t.Fatal(err)
		}
	}

	runTest(t, &testcaseVortex{
		M:               m,
		X:               x,
		Ys:              ys,
		Alpha:           alpha,
		SelectedColumns: selectedColumns,
	})
}

func randElement(rng *rand.Rand) koalabear.Element {
	return koalabear.Element{rng.Uint32N(2130706433)}
}

func randFext(rng *rand.Rand) fext.E4 {
	return fext.E4{
		B0: fext.E2{
			A0: randElement(rng),
			A1: randElement(rng),
		},
		B1: fext.E2{
			A0: randElement(rng),
			A1: randElement(rng),
		},
	}
}

func runTest(t *testing.T, tc *testcaseVortex) {

	var (
		numCol             = len(tc.M[0])
		numRow             = len(tc.M)
		reedSolomonInvRate = 2
		numSelectedColumns = len(tc.SelectedColumns)
		sisParams, _       = sis.NewRSis(0, 9, 16, numRow)
		params             = NewParams(numCol, numRow, sisParams, reedSolomonInvRate, numSelectedColumns)
	)

	proverState, err := Commit(params, tc.M)
	if err != nil {
		t.Fatal(err)
	}

	proverState.OpenLinComb(tc.Alpha)

	proof, err := proverState.OpenColumns(tc.SelectedColumns)
	if err != nil {
		t.Fatal(err)
	}

	err = params.Verify(VerifierInput{
		Proof:           proof,
		MerkleRoot:      proverState.GetCommitment(),
		ClaimedValues:   tc.Ys,
		EvaluationPoint: tc.X,
		Alpha:           tc.Alpha,
		SelectedColumns: tc.SelectedColumns,
	})

	if err != nil {
		t.Fatal(err)
	}
}

// BenchmarkVortexReal benchmarks Vortex in (estimated) production conditions for the
// zkEVM. We aim to have it commit to 4GiB of data. So about 1<<30 koalabear elements.
func BenchmarkVortexReal(b *testing.B) {

	var (
		numCol             = 1 << 19
		numRow             = 1 << 11
		invRate            = 2
		numSelectedColumns = 256
		wg                 sync.WaitGroup
		sisParams, _       = sis.NewRSis(0, 9, 16, numRow)
		params             = NewParams(numCol, numRow, sisParams, invRate, numSelectedColumns)
		// #nosec G404 -- test case generation does not require a cryptographic PRNG
		topRng          = rand.New(rand.NewChaCha8([32]byte{}))
		alpha           = randFext(topRng)
		selectedColumns = make([]int, 256)
	)

	for i := range selectedColumns {
		selectedColumns[i] = topRng.IntN(numCol * 2)
	}

	// Generating the matrix and filling it with PRNG elements on a single-thread would
	// be very time-consuming so we parallelize it, giving it different seeds for each
	// row.
	m := make([][]koalabear.Element, numRow)
	for row := range m {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			m[row] = make([]koalabear.Element, numCol)
			seed := [32]byte{}
			binary.PutVarint(seed[:], int64(row))

			// #nosec G404 -- test case generation does not require a cryptographic PRNG
			rng := rand.New(rand.NewChaCha8(seed))
			for j := range m[row] {
				m[row][j] = randElement(rng)
			}
		}(row)
	}

	wg.Wait()

	var (
		proverState *ProverState
		err         error
	)

	b.Run("committing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			proverState, err = Commit(params, m)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	_ = proverState
	_ = alpha

	b.Run("opening-alpha", func(b *testing.B) {
		proverState, err = Commit(params, m)
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			proverState.OpenLinComb(alpha)
		}
	})

	// b.Run("opening-columns", func(b *testing.B) {
	// 	b.ResetTimer()
	// 	for i := 0; i < b.N; i++ {
	// 		_, err := proverState.OpenColumns(selectedColumns)
	// 		if err != nil {
	// 			b.Fatal(err)
	// 		}
	// 	}
	// })

}
