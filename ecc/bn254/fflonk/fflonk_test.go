package fflonk

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/stretchr/testify/require"
)

func TestExtractRoots(t *testing.T) {

	assert := require.New(t)

	m := 9
	var x fr.Element
	x.SetRandom()
	roots, err := extractRoots(x, m)
	assert.NoError(err)

	// check that (yᵐ-x)=Πᵢ(y-ωⁱᵗ√(x)) for a random y
	var y fr.Element
	y.SetRandom()
	expo := big.NewInt(int64(m))
	y.Exp(x, expo).Sub(&y, &x)
	var rhs, tmp fr.Element
	rhs.SetOne()
	for i := 0; i < m; i++ {
		tmp.Sub(&y, &roots[i])
		rhs.Mul(&rhs, &tmp)
	}
	if !rhs.Equal(&y) {
		assert.Fail("(yᵐ-x) != Πᵢ(y-ωⁱᵗ√(x)))")
	}

}
