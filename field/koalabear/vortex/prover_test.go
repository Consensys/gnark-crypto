package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
)

func TestZeroMatrix(t *testing.T) {

	var (
		numCol             = 16
		numRow             = 8
		reedSolomonInvRate = 2
		numSelectedColumns = 4
		sisParams, _       = sis.NewRSis(0, 9, 16, numRow)
		params             = NewParams(numCol, numRow, sisParams, reedSolomonInvRate, numSelectedColumns)
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

	proverState, err := Commit(params, m)
	if err != nil {
		t.Fatal(err)
	}

	proverState.OpenLinComb(*alpha)

	proof, err := proverState.OpenColumns(selectedColumns)
	if err != nil {
		t.Fatal(err)
	}

	err = params.Verify(VerifierInput{
		Proof:           proof,
		MerkleRoot:      proverState.GetCommitment(),
		ClaimedValues:   y,
		EvaluationPoint: x,
		Alpha:           *alpha,
		SelectedColumns: selectedColumns,
	})

	if err != nil {
		t.Fatal(err)
	}
}
