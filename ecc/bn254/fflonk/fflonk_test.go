package fflonk

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	"github.com/stretchr/testify/require"
)

// Test SRS re-used across tests of the KZG scheme
var testSrs *kzg.SRS
var bAlpha *big.Int

func init() {
	const srsSize = 230
	bAlpha = new(big.Int).SetInt64(42) // randomise ?
	testSrs, _ = kzg.NewSRS(ecc.NextPowerOfTwo(srsSize), bAlpha)
}

func TestCommit(t *testing.T) {

	assert := require.New(t)

	// sample polynomials
	nbPolys := 2
	p := make([][]fr.Element, nbPolys)
	for i := 0; i < nbPolys; i++ {
		p[i] = make([]fr.Element, i+10)
		for j := 0; j < i+10; j++ {
			p[i][j].SetRandom()
		}
	}

	// fflonk commit to them
	var x fr.Element
	x.SetRandom()
	proof, err := kzg.Open(Fold(p), x, testSrs.Pk)
	assert.NoError(err)

	// check that Open(C, x) = ∑_{i<t}Pᵢ(xᵗ)xⁱ
	var xt fr.Element
	var expo big.Int
	expo.SetUint64(uint64(nbPolys))
	xt.Exp(x, &expo)
	px := make([]fr.Element, nbPolys)
	for i := 0; i < nbPolys; i++ {
		px[i] = eval(p[i], xt)
	}
	y := eval(px, x)
	assert.True(y.Equal(&proof.ClaimedValue))
}

func TestGetIthRootOne(t *testing.T) {

	assert := require.New(t)

	order := 9
	omega, err := getIthRootOne(order)
	assert.NoError(err)
	var orderBigInt big.Int
	orderBigInt.SetUint64(uint64(order))
	omega.Exp(omega, &orderBigInt)
	assert.True(omega.IsOne())

	order = 7
	_, err = getIthRootOne(order)
	assert.Error(err)
}
