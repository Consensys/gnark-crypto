package test_vector_utils

import (
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashNewElementSaved(t *testing.T) {
	var hash HashMap

	var fortyFour small_rational.SmallRational
	fortyFour.SetInt64(44)

	expected := hash.FindPair(&fortyFour, nil)
	for i := 0; i < 10; i++ {
		seen := hash.FindPair(&fortyFour, nil)
		if !expected.Equal(&seen) {
			t.Errorf("expected %s saw %s", expected.String(), seen.String())
		}
	}
}

func TestHashLoadConsistent(t *testing.T) {
	var expected small_rational.SmallRational
	var fortyFour small_rational.SmallRational

	fortyFour.SetInt64(44)

	for i := 0; i < 10; i++ {
		hash, err := GetHash("../../../gkr/test_vectors/resources/hash.json")
		assert.NoError(t, err)

		seen := hash.FindPair(&fortyFour, nil)
		if i == 0 {
			expected = seen
		} else {
			if !expected.Equal(&seen) {
				t.Errorf("expected %s saw %s", expected.String(), seen.String())
			}
		}
	}
}

func TestSaveHash(t *testing.T) {

	var one, two, three small_rational.SmallRational
	one.SetInt64(1)
	two.SetInt64(2)
	three.SetInt64(3)

	hash := HashMap{{
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
