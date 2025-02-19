package vortex

import (
	"math/rand/v2"
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
		rng    = rand.New(rand.NewChaCha8([32]byte{}))
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
	modulus := uint32(koalabear.Modulus().Int64())
	return koalabear.Element{rng.Uint32N(modulus)}
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
