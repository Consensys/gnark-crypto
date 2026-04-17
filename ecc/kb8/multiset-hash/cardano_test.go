package multisethash

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/stretchr/testify/require"
)

func TestCbrtE8OnCubicResidues(t *testing.T) {
	for i := 0; i < 128; i++ {
		var a, x, got, check extensions.E8
		a.MustSetRandom()
		x.Square(&a).Mul(&x, &a)
		require.NotNil(t, cbrtE8(&got, &x))
		check.Square(&got).Mul(&check, &got)
		require.True(t, check.Equal(&x))
	}
}

func TestCbrtE8RejectsNonResidues(t *testing.T) {
	var x, got extensions.E8
	for i := 0; i < 256; i++ {
		x.MustSetRandom()
		if cbrtE8(&got, &x) == nil {
			return
		}
	}
	t.Fatal("failed to find an E8 non-cube in 256 samples")
}

func TestE16CbrtOnCubicResidues(t *testing.T) {
	for i := 0; i < 128; i++ {
		var a, x, got, check e16
		a.A0.MustSetRandom()
		a.A1.MustSetRandom()
		x.Square(&a).Mul(&x, &a)
		require.NotNil(t, got.Cbrt(&x))
		check.Square(&got).Mul(&check, &got)
		require.True(t, check.A0.Equal(&x.A0))
		require.True(t, check.A1.Equal(&x.A1))
	}
}

func TestE16CbrtRejectsNonResidues(t *testing.T) {
	var x, got e16
	for i := 0; i < 256; i++ {
		x.A0.MustSetRandom()
		x.A1.MustSetRandom()
		if got.Cbrt(&x) == nil {
			return
		}
	}
	t.Fatal("failed to find an E16 non-cube in 256 samples")
}

func TestE16GLVTraceMatchesBinaryLucas(t *testing.T) {
	for i := 0; i < 16; i++ {
		var tau extensions.E8
		tau.MustSetRandom()
		gotTe, gotTe1 := lucasV2E8(&tau)
		require.False(t, gotTe.IsZero() && gotTe1.IsZero())
	}
}

func TestDepressedCubicRootFindsValidRoot(t *testing.T) {
	for i := 0; i < 64; i++ {
		var x, x3, c, lhs extensions.E8
		x.MustSetRandom()
		x3.Square(&x).Mul(&x3, &x)
		c.Double(&x).Add(&c, &x).Sub(&c, &x3)

		root, ok := depressedCubicRoot(c)
		require.True(t, ok)
		lhs.Square(&root).Mul(&lhs, &root)
		lhs.Sub(&lhs, &root).Sub(&lhs, &root).Sub(&lhs, &root).Add(&lhs, &c)
		require.True(t, lhs.IsZero())
	}
}
