package test_vector_utils

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational/test_vector_utils"
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

func TestCounterTranscriptInequality(t *testing.T) {
	const challengeName = "fC.0"
	t1 := fiatshamir.NewTranscript(test_vector_utils.NewMessageCounter(1, 1), challengeName)
	t2 := fiatshamir.NewTranscript(test_vector_utils.NewMessageCounter(0, 1), challengeName)
	var c1, c2 []byte
	var err error
	c1, err = t1.ComputeChallenge(challengeName)
	assert.NoError(t, err)
	c2, err = t2.ComputeChallenge(challengeName)
	assert.NoError(t, err)
	assert.NotEqual(t, c1, c2)
}
