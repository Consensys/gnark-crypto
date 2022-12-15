package test_vector_utils

import (
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashNewElementSaved(t *testing.T) {
	var hash ElementMap

	var fortyFour small_rational.SmallRational
	fortyFour.SetInt64(44)

	expected, err := hash.FindPair(&fortyFour, nil)
	assert.NoError(t, err)
	for i := 0; i < 10; i++ {
		seen, err := hash.FindPair(&fortyFour, nil)
		assert.NoError(t, err)
		if !expected.Equal(&seen) {
			t.Errorf("expected %s saw %s", expected.String(), seen.String())
		}
	}
}

func TestHashConsistency(t *testing.T) {
	var one small_rational.SmallRational
	var mp ElementMap
	one.SetOne()
	bytes := one.Bytes()

	t1 := fiatshamir.NewTranscript(&MapHash{Map: &mp}, "0")
	assert.NoError(t, t1.Bind("0", bytes[:]))
	c1, err := t1.ComputeChallenge("0")
	assert.NoError(t, err)

	t2 := fiatshamir.NewTranscript(&MapHash{Map: &mp}, "0")
	assert.NoError(t, t2.Bind("0", bytes[:]))
	c2, err := t2.ComputeChallenge("0")
	assert.NoError(t, err)

	assert.Equal(t, c1, c2)
}

func TestSaveHash(t *testing.T) {

	var one, two, three small_rational.SmallRational
	one.SetInt64(1)
	two.SetInt64(2)
	three.SetInt64(3)

	hash := ElementMap{{
		key1:        one,
		key2:        small_rational.SmallRational{},
		key2Present: false,
		value:       two,
		used:        true,
	}, {
		key1:        one,
		key2:        one,
		key2Present: true,
		value:       three,
		used:        true,
	}, {
		key1:        two,
		key2:        one,
		key2Present: true,
		value:       two,
		used:        false,
	}}

	serialized, err := hash.serializedUsedEntries()
	assert.NoError(t, err)
	assert.Equal(t, "{\n\t\"1\":2,\n\t\"1,1\":3\n}", serialized)
}
