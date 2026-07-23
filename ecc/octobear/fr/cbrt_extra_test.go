package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestElementCbrtZero(t *testing.T) {
	var zero, got Element
	require.NotNil(t, got.Cbrt(&zero))
	require.True(t, got.IsZero())
}
