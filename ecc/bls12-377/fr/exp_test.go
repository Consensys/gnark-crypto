package fr

import (
	"crypto/rand"
	"math/big"
	"testing"
)

// expBinary is the old binary square-and-multiply implementation for reference comparison.
func expBinary(z *Element, x Element, k *big.Int) *Element {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		x.Inverse(&x)
		e = new(big.Int).Neg(k)
	}

	z.Set(&x)

	for i := e.BitLen() - 2; i >= 0; i-- {
		z.Square(z)
		if e.Bit(i) == 1 {
			z.Mul(z, &x)
		}
	}

	return z
}

// TestExpWindowedCorrectnessExhaustiveSmall tests all uint64 exponents 0..1024
func TestExpWindowedCorrectnessExhaustiveSmall(t *testing.T) {
	t.Parallel()
	var x Element
	x.MustSetRandom()

	for k := uint64(0); k <= 1024; k++ {
		kBig := new(big.Int).SetUint64(k)
		var got, want Element
		got.Exp(x, kBig)
		expBinary(&want, x, kBig)
		if !got.Equal(&want) {
			t.Fatalf("mismatch at k=%d", k)
		}
	}
}

// TestExpWindowedCorrectnessRandomUint64 tests random uint64 exponents
func TestExpWindowedCorrectnessRandomUint64(t *testing.T) {
	t.Parallel()
	for i := 0; i < 500; i++ {
		var x Element
		x.MustSetRandom()

		kBig, _ := rand.Int(rand.Reader, new(big.Int).SetUint64(^uint64(0)))

		var got, want Element
		got.Exp(x, kBig)
		expBinary(&want, x, kBig)
		if !got.Equal(&want) {
			t.Fatalf("mismatch at iteration %d, k=%s", i, kBig.String())
		}
	}
}

// TestExpWindowedCorrectnessRandomBig tests random big.Int exponents of various bit lengths
func TestExpWindowedCorrectnessRandomBig(t *testing.T) {
	t.Parallel()
	bitSizes := []int{1, 2, 3, 4, 5, 7, 8, 15, 16, 31, 32, 63, 64, 127, 128, 200, 253, 254, 255, 256, 384, 512, 1024}
	for _, bits := range bitSizes {
		for i := 0; i < 50; i++ {
			var x Element
			x.MustSetRandom()

			kBig, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), uint(bits)))

			var got, want Element
			got.Exp(x, kBig)
			expBinary(&want, x, kBig)
			if !got.Equal(&want) {
				t.Fatalf("mismatch at bits=%d iteration=%d, k=%s", bits, i, kBig.String())
			}
		}
	}
}

// TestExpWindowedEdgeCases tests specific edge cases
func TestExpWindowedEdgeCases(t *testing.T) {
	t.Parallel()
	var x, got, want Element
	x.MustSetRandom()

	// k = 0
	got.Exp(x, big.NewInt(0))
	if !got.IsOne() {
		t.Fatal("Exp(x, 0) should be 1")
	}

	// k = 1
	got.Exp(x, big.NewInt(1))
	if !got.Equal(&x) {
		t.Fatal("Exp(x, 1) should be x")
	}

	// k = 2
	got.Exp(x, big.NewInt(2))
	want.Square(&x)
	if !got.Equal(&want) {
		t.Fatal("Exp(x, 2) should be x²")
	}

	// k = 3
	got.Exp(x, big.NewInt(3))
	want.Mul(&want, &x)
	if !got.Equal(&want) {
		t.Fatal("Exp(x, 3) should be x³")
	}

	// powers of 2
	for p := 1; p <= 64; p++ {
		k := new(big.Int).Lsh(big.NewInt(1), uint(p))
		got.Exp(x, k)
		expBinary(&want, x, k)
		if !got.Equal(&want) {
			t.Fatalf("Exp(x, 2^%d) mismatch", p)
		}
	}

	// 2^k - 1 (all ones in binary)
	for p := 1; p <= 64; p++ {
		k := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(p)), big.NewInt(1))
		got.Exp(x, k)
		expBinary(&want, x, k)
		if !got.Equal(&want) {
			t.Fatalf("Exp(x, 2^%d - 1) mismatch", p)
		}
	}

	// x = 0
	var zero Element
	got.Exp(zero, big.NewInt(42))
	if !got.IsZero() {
		t.Fatal("Exp(0, 42) should be 0")
	}

	// x = 1
	var one Element
	one.SetOne()
	kBig, _ := rand.Int(rand.Reader, Modulus())
	got.Exp(one, kBig)
	if !got.IsOne() {
		t.Fatal("Exp(1, k) should be 1")
	}

	// negative exponent
	got.Exp(x, big.NewInt(-1))
	want.Inverse(&x)
	if !got.Equal(&want) {
		t.Fatal("Exp(x, -1) should be x⁻¹")
	}

	// Exp(x, q-1) should be 1 (Fermat's little theorem)
	qm1 := new(big.Int).Sub(Modulus(), big.NewInt(1))
	got.Exp(x, qm1)
	if !got.IsOne() {
		t.Fatal("Exp(x, q-1) should be 1")
	}

	// aliasing: z == &x
	got.Set(&x)
	got.Exp(got, big.NewInt(7))
	expBinary(&want, x, big.NewInt(7))
	if !got.Equal(&want) {
		t.Fatal("aliased Exp(x, 7) mismatch")
	}
}

// TestExpWindowedNegativeExponents tests negative exponents of various sizes
func TestExpWindowedNegativeExponents(t *testing.T) {
	t.Parallel()
	for i := 0; i < 200; i++ {
		var x Element
		x.MustSetRandom()

		kBig, _ := rand.Int(rand.Reader, Modulus())
		negK := new(big.Int).Neg(kBig)

		var got, want Element
		got.Exp(x, negK)
		expBinary(&want, x, negK)
		if !got.Equal(&want) {
			t.Fatalf("negative exp mismatch at iteration %d", i)
		}
	}
}

// TestExpWindowedConsistency verifies Exp with big.Int matches repeated multiplication
func TestExpWindowedConsistency(t *testing.T) {
	t.Parallel()
	var x Element
	x.MustSetRandom()

	// verify by repeated multiplication for small k
	for k := uint64(0); k <= 128; k++ {
		var got, want Element
		got.Exp(x, new(big.Int).SetUint64(k))

		want.SetOne()
		for j := uint64(0); j < k; j++ {
			want.Mul(&want, &x)
		}
		if !got.Equal(&want) {
			t.Fatalf("Exp(x, %d) != x*x*...*x (%d times)", k, k)
		}
	}
}

// Benchmarks with various exponent sizes
func BenchmarkExpBinary(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k, _ := rand.Int(rand.Reader, Modulus())
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		expBinary(&z, x, k)
	}
}

func BenchmarkExpWindowed(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k, _ := rand.Int(rand.Reader, Modulus())
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		z.Exp(x, k)
	}
}

func BenchmarkExpUint64Small(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k := new(big.Int).SetUint64(7)
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		z.Exp(x, k)
	}
}

func BenchmarkExpUint64Medium(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k := new(big.Int).SetUint64(1<<32 - 1)
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		z.Exp(x, k)
	}
}

func BenchmarkExpUint64Large(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k := new(big.Int).SetUint64(^uint64(0))
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		z.Exp(x, k)
	}
}

func BenchmarkExpBig128(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		z.Exp(x, k)
	}
}

func BenchmarkExpBig256(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		z.Exp(x, k)
	}
}

func BenchmarkExpBigModulus(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k, _ := rand.Int(rand.Reader, Modulus())
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		z.Exp(x, k)
	}
}

// Binary baseline benchmarks for comparison
func BenchmarkExpBinarySmall(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k := new(big.Int).SetUint64(7)
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		expBinary(&z, x, k)
	}
}

func BenchmarkExpBinaryUint64Large(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k := new(big.Int).SetUint64(^uint64(0))
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		expBinary(&z, x, k)
	}
}

func BenchmarkExpBinaryBig128(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		expBinary(&z, x, k)
	}
}

func BenchmarkExpBinaryBig256(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		expBinary(&z, x, k)
	}
}

func BenchmarkExpBinaryModulus(b *testing.B) {
	var x Element
	x.MustSetRandom()
	k, _ := rand.Int(rand.Reader, Modulus())
	b.ResetTimer()
	var z Element
	for i := 0; i < b.N; i++ {
		expBinary(&z, x, k)
	}
}
