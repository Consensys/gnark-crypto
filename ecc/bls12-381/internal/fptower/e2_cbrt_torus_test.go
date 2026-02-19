package fptower

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

func TestCbrtTorusMatchesOriginal(t *testing.T) {
	for i := 0; i < 100; i++ {
		var t0, a E2
		t0.MustSetRandom()
		a.Square(&t0).Mul(&a, &t0) // a = t0³ (cubic residue)

		var got E2
		result := got.cbrtTorus(&a)
		if result == nil {
			t.Fatalf("cbrtTorus returned nil at iteration %d", i)
		}

		var check E2
		check.Square(&got).Mul(&check, &got)
		if !check.Equal(&a) {
			t.Fatalf("cbrtTorus failed: got³ != a at iteration %d", i)
		}
	}
}

func TestCbrtTorusEdgeCases(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		var x, z E2
		x.SetZero()
		z.cbrtTorus(&x)
		if !z.IsZero() {
			t.Fatal("cbrt(0) should be 0")
		}
	})

	t.Run("real", func(t *testing.T) {
		var x, z E2
		x.A0.SetUint64(8)
		x.A1.SetZero()
		result := z.cbrtTorus(&x)
		if result != nil {
			var check E2
			check.Square(&z).Mul(&check, &z)
			if !check.Equal(&x) {
				t.Fatal("cbrt(real)³ != real")
			}
		}
	})

	t.Run("imaginary", func(t *testing.T) {
		var t0 fp.Element
		t0.SetUint64(7)
		var x E2
		x.A0.SetZero()
		var neg fp.Element
		neg.Cube(&t0).Neg(&neg)
		x.A1.Set(&neg)

		var z E2
		result := z.cbrtTorus(&x)
		if result != nil {
			var check E2
			check.Square(&z).Mul(&check, &z)
			if !check.Equal(&x) {
				t.Fatal("cbrt(imaginary)³ != imaginary")
			}
		}
	})
}

func TestCbrtAndNormInverse(t *testing.T) {
	for i := 0; i < 100; i++ {
		var x E2
		x.MustSetRandom()

		// norm = x0² + x1²
		var norm, x0sq, x1sq fp.Element
		x0sq.Square(&x.A0)
		x1sq.Square(&x.A1)
		norm.Add(&x0sq, &x1sq)

		m, normInv, ok := cbrtAndNormInverse(&norm)
		if !ok {
			// norm might not be a cubic residue; skip
			continue
		}

		// Check m³ = norm
		var c fp.Element
		c.Cube(&m)
		if !c.Equal(&norm) {
			t.Fatalf("m³ != norm at iteration %d", i)
		}

		// Check normInv * norm = 1
		var one fp.Element
		one.Mul(&normInv, &norm)
		if !one.IsOne() {
			t.Fatalf("normInv * norm != 1 at iteration %d", i)
		}
	}
}

func BenchmarkE2CbrtTorus(b *testing.B) {
	var a, t0 E2
	t0.MustSetRandom()
	a.Square(&t0).Mul(&a, &t0)

	var c E2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.cbrtTorus(&a)
	}
	_ = c
}
