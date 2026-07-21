package fr

import (
	"testing"

	"github.com/consensys/gnark-crypto/utils/cpu"
	"github.com/stretchr/testify/require"
)

func TestVectorMulIFMACarryRegression(t *testing.T) {
	if !cpu.SupportAVX512IFMA {
		t.Skip("requires AVX-512 IFMA")
	}

	assert := require.New(t)
	parse := func(s string) Element {
		var z Element
		_, err := z.SetString(s)
		assert.NoError(err)
		return z
	}

	a := make(Vector, 8)
	b := make(Vector, 8)
	a[0] = parse("7545707920004054512758986885880204442125303275406040989987309426702336485362")
	b[0] = parse("8048480940844806781267238101067768890083750858835167793058558758481368829585")

	want := make(Vector, len(a))
	mulVecGeneric(want, a, b)

	got := make(Vector, len(a))
	got.Mul(a, b)
	assert.True(got.Equal(want), "Mul should handle the IFMA carry-chain regression with a distinct destination")

	inPlace := make(Vector, len(a))
	copy(inPlace, a)
	inPlace.Mul(inPlace, b)
	assert.True(inPlace.Equal(want), "Mul should handle the same carry-chain regression in place")
}
