package unsafe_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/utils/unsafe"
	"github.com/stretchr/testify/require"
)

func TestPointDump(t *testing.T) {
	assert := require.New(t)
	samplePoints := make([]bn254.G2Affine, 10)
	fillBenchBasesG2(samplePoints)

	var buf bytes.Buffer

	err := unsafe.WriteSlice(&buf, samplePoints)
	assert.NoError(err)

	readPoints, _, err := unsafe.ReadSlice[[]bn254.G2Affine](&buf)
	assert.NoError(err)

	assert.Equal(samplePoints, readPoints)
}

func TestMarker(t *testing.T) {
	assert := require.New(t)
	var buf bytes.Buffer

	err := unsafe.WriteMarker(&buf)
	assert.NoError(err)

	err = unsafe.ReadMarker(&buf)
	assert.NoError(err)
}

func fillBenchBasesG2(samplePoints []bn254.G2Affine) {
	var r big.Int
	r.SetString("340444420969191673093399857471996460938405", 10)
	samplePoints[0].ScalarMultiplication(&samplePoints[0], &r)

	one := samplePoints[0].X
	one.SetOne()

	for i := 1; i < len(samplePoints); i++ {
		samplePoints[i].X.Add(&samplePoints[i-1].X, &one)
		samplePoints[i].Y.Sub(&samplePoints[i-1].Y, &one)
	}
}
