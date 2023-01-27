package gkr

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO: Move to test-utils
func createRandomized(nbOuter, nbInner int) [][]fr.Element {
	res := make([][]fr.Element, nbOuter)
	for i := range res {
		res[i] = make([]fr.Element, nbInner)
		for j := range res[i] {
			if _, err := res[i][j].SetRandom(); err != nil {
				panic(err)
			}
		}
	}
	return res
}

func testCollateExtrapolate(t *testing.T, nbPages, nbInstances, nbExtrapolations int) {
	diffs := createRandomized(nbPages, nbInstances)
	at0 := createRandomized(nbPages, nbInstances)

	atD := make([][][]fr.Element, nbExtrapolations+1) //indexes by extrapolation, then page, then instance
	atD[0] = at0

	for d := 1; d <= nbExtrapolations; d++ {
		atD[d] = make([][]fr.Element, nbPages)
		for i := range diffs {
			atD[d][i] = make([]fr.Element, nbInstances)
			for j := range diffs[i] {
				atD[d][i][j].Add(&atD[d-1][i][j], &diffs[i][j])
			}
		}
	}

	s := make([]polynomial.MultiLin, nbPages)
	for i := range s {
		s[i] = make(polynomial.MultiLin, nbInstances*2)
		copy(s[i][:nbInstances], atD[0][i])
		copy(s[i][nbInstances:], atD[1][i])
	}

	w := utils.NewWorkerPool()
	p := polynomial.NewPool(nbInstances * nbExtrapolations * nbPages)

	res := collateExtrapolate(s, nbExtrapolations, &w, &p)

	for d := 1; d < nbExtrapolations; d++ {
		for i := 0; i < nbInstances; i++ {
			for j := 0; j < nbPages; j++ {
				assert.Equal(t, atD[d][j][i], res[d-1][i*nbPages+j])
			}
		}
	}
}

func TestCollateExtrapolate(t *testing.T) {
	for i := 0; i < 10; i++ {
		testCollateExtrapolate(t, 1, 1, 1)
	}
}
