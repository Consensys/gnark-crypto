package element

const TestVector = `


import (
	"testing"
	"github.com/stretchr/testify/require"
	"sort"
	"reflect"
)



func TestVectorSort(t *testing.T) {
	assert := require.New(t)

	v := make(Vector, 3)
	v[0].SetUint64(2)
	v[1].SetUint64(3)
	v[2].SetUint64(1)

	sort.Sort(v)

	assert.Equal("[1,2,3]", v.String())
}

func TestVectorRoundTrip(t *testing.T) {
	assert := require.New(t)

	v1 := make(Vector, 3)
	v1[0].SetUint64(2)
	v1[1].SetUint64(3)
	v1[2].SetUint64(1)

	b, err := v1.MarshalBinary()
	assert.NoError(err)

	var v2 Vector

	err = v2.UnmarshalBinary(b)
	assert.NoError(err)

	assert.True(reflect.DeepEqual(v1,v2))
}

func TestVectorEmptyRoundTrip(t *testing.T) {
	assert := require.New(t)

	v1 := make(Vector, 0)

	b, err := v1.MarshalBinary()
	assert.NoError(err)

	var v2 Vector

	err = v2.UnmarshalBinary(b)
	assert.NoError(err)

	assert.True(reflect.DeepEqual(v1,v2))
}

`
