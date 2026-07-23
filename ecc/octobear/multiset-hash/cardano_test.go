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

func TestCbrtE8Zero(t *testing.T) {
	var zero, got extensions.E8
	require.NotNil(t, cbrtE8(&got, &zero))
	require.True(t, got.IsZero())
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

func TestE16CbrtZero(t *testing.T) {
	var zero, got e16
	require.NotNil(t, got.Cbrt(&zero))
	require.True(t, got.A0.IsZero())
	require.True(t, got.A1.IsZero())
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

// cardanoDelta returns the dispatch discriminant delta = 108 - 27·c² used by
// cardanoRoots to choose a branch (zero → repeatedRoots, non-square →
// quadratic-extension branch, square → base-field branch).
func cardanoDelta(c *extensions.E8) extensions.E8 {
	var c2, delta extensions.E8
	c2.Square(c)
	delta.Mul(&c2, &e8TwentySeven)
	delta.Sub(&e8Neg4A3, &delta)
	return delta
}

func checkCubicRoots(t *testing.T, c extensions.E8, roots []extensions.E8) {
	t.Helper()
	require.NotEmpty(t, roots, "cardanoRoots returned no roots")
	for _, x := range roots {
		var lhs, x3 extensions.E8
		x3.Square(&x).Mul(&x3, &x)
		lhs.Set(&x3)
		lhs.Sub(&lhs, &x).Sub(&lhs, &x).Sub(&lhs, &x).Add(&lhs, &c)
		require.True(t, lhs.IsZero(), "root does not satisfy x^3 - 3x + c = 0")
	}
}

// TestCardanoRepeatedRootBranch exercises the delta = 0 branch with c = 2:
// x^3 - 3x + 2 = (x-1)^2 (x+2), so the dispatcher must hit repeatedRoots.
func TestCardanoRepeatedRootBranch(t *testing.T) {
	var c extensions.E8
	c.C0.B0.A0.SetUint64(2)

	delta := cardanoDelta(&c)
	require.True(t, delta.IsZero(), "c = 2 must drive delta to zero")

	roots := cardanoRoots(c)
	checkCubicRoots(t, c, roots)

	var one, negTwo extensions.E8
	one.SetOne()
	negTwo.SetOne().Double(&negTwo).Neg(&negTwo)
	var foundOne, foundNegTwo bool
	for _, x := range roots {
		if x.Equal(&one) {
			foundOne = true
		}
		if x.Equal(&negTwo) {
			foundNegTwo = true
		}
	}
	require.True(t, foundOne, "repeated-root branch must produce x = 1")
	require.True(t, foundNegTwo, "repeated-root branch must produce x = -2")
}

// TestCardanoBaseFieldBranch exercises the square-delta path. We build c from
// a known root x in the prime subfield: c = 3x - x^3 guarantees that x solves
// x^3 - 3x + c = 0, and any c whose components all lie in Fp produces a delta
// that is a square in E8 (since [E8 : Fp] = 8 is even, every element of Fp is
// a square in E8). Whether Cardano can recover roots through E8 depends on
// whether the cube root needed by the formula lies in E8, so we search across
// x values until the dispatcher returns a non-empty set including x.
func TestCardanoBaseFieldBranch(t *testing.T) {
	for n := uint64(3); n < 10_000; n++ {
		var x extensions.E8
		x.C0.B0.A0.SetUint64(n)

		var c, x3 extensions.E8
		x3.Square(&x).Mul(&x3, &x)
		c.Double(&x).Add(&c, &x).Sub(&c, &x3)

		delta := cardanoDelta(&c)
		if delta.IsZero() {
			continue
		}
		require.Equal(t, 1, delta.Legendre(), "c in Fp must give a square delta in E8")

		roots := cardanoRoots(c)
		if len(roots) == 0 {
			continue
		}
		checkCubicRoots(t, c, roots)
		var found bool
		for _, r := range roots {
			if r.Equal(&x) {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		return
	}
	t.Fatal("could not find a base-field-branch witness in [3, 10000)")
}

// TestCardanoQuadraticExtensionBranch exercises the non-square-delta path.
// To force delta to be a non-square in E8, c must have a non-trivial
// extension component (any element of the prime subfield is a square in E8).
// We construct x with both a base and an extension component, compute
// c = 3x - x^3, and search for one whose delta is a non-square.
func TestCardanoQuadraticExtensionBranch(t *testing.T) {
	for n := uint64(1); n < 4096; n++ {
		var x extensions.E8
		x.C0.B0.A0.SetUint64(n)
		x.C1.B0.A0.SetUint64(1)

		var c, x3 extensions.E8
		x3.Square(&x).Mul(&x3, &x)
		c.Double(&x).Add(&c, &x).Sub(&c, &x3)

		delta := cardanoDelta(&c)
		if delta.IsZero() || delta.Legendre() != -1 {
			continue
		}
		roots := cardanoRoots(c)
		checkCubicRoots(t, c, roots)
		var found bool
		for _, r := range roots {
			if r.Equal(&x) {
				found = true
				break
			}
		}
		require.True(t, found, "extension branch must recover x at n = %d", n)
		return
	}
	t.Fatal("could not find a quadratic-extension-branch witness in [1, 4096)")
}
