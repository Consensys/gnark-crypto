package test_vector_utils

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestTranscript(t *testing.T) {

	mp, err := CreateElementMap(map[string]interface{}{
		strconv.Itoa('0'): 2,
		"3,2":             5,
	})
	assert.NoError(t, err)

	hsh := MapHash{Map: &mp}
	transcript := fiatshamir.NewTranscript(&hsh, "0", "1")
	bytes := ToElement(3).Bytes()
	err = transcript.Bind("0", bytes[:])
	assert.NoError(t, err)
	var cBytes []byte
	cBytes, err = transcript.ComputeChallenge("0")
	assert.NoError(t, err)
	var res fr.Element
	res.SetBytes(cBytes)
	assert.True(t, ToElement(5).Equal(&res))
}
