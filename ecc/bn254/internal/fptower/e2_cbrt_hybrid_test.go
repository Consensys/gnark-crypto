package fptower

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
)

func TestCbrtHybridMatchesOriginal(t *testing.T) {
	for i := 0; i < 100; i++ {
		var t0, a E2
		t0.MustSetRandom()
		a.Square(&t0).Mul(&a, &t0) // a = t0³ (cubic residue)

		var got E2
		result := got.cbrtHybrid(&a)
		if result == nil {
			t.Fatalf("cbrtHybrid returned nil at iteration %d", i)
		}

		var check E2
		check.Square(&got).Mul(&check, &got)
		if !check.Equal(&a) {
			t.Fatalf("cbrtHybrid failed: got³ != a at iteration %d", i)
		}
	}
}

func TestCbrtHybridEdgeCases(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		var x, z E2
		x.SetZero()
		z.cbrtHybrid(&x)
		if !z.IsZero() {
			t.Fatal("cbrt(0) should be 0")
		}
	})

	t.Run("real", func(t *testing.T) {
		var x, z E2
		x.A0.SetUint64(8)
		x.A1.SetZero()
		result := z.cbrtHybrid(&x)
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
		result := z.cbrtHybrid(&x)
		if result != nil {
			var check E2
			check.Square(&z).Mul(&check, &z)
			if !check.Equal(&x) {
				t.Fatal("cbrt(imaginary)³ != imaginary")
			}
		}
	})
}

func BenchmarkE2CbrtHybrid(b *testing.B) {
	var a, t0 E2
	t0.MustSetRandom()
	a.Square(&t0).Mul(&a, &t0)

	var c E2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.cbrtHybrid(&a)
	}
	_ = c
}
