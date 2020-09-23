package bw761

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// tests

func TestE6ReceiverIsOperand(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE6()
	genB := GenE6()

	properties.Property("[BW761] Having the receiver as operand (addition) should output the same result", prop.ForAll(
		func(a, b *e6) bool {
			var c, d e6
			d.Set(a)
			c.Add(a, b)
			a.Add(a, b)
			b.Add(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BW761] Having the receiver as operand (sub) should output the same result", prop.ForAll(
		func(a, b *e6) bool {
			var c, d e6
			d.Set(a)
			c.Sub(a, b)
			a.Sub(a, b)
			b.Sub(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BW761] Having the receiver as operand (mul) should output the same result", prop.ForAll(
		func(a, b *e6) bool {
			var c, d e6
			d.Set(a)
			c.Mul(a, b)
			a.Mul(a, b)
			b.Mul(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BW761] Having the receiver as operand (square) should output the same result", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.Square(a)
			a.Square(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (neg) should output the same result", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.Neg(a)
			a.Neg(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (double) should output the same result", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.Double(a)
			a.Double(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (Inverse) should output the same result", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.Inverse(a)
			a.Inverse(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (mul by e2) should output the same result", prop.ForAll(
		func(a *e6) bool {
			var b e6
			var c e2
			c.SetRandom()
			b.MulByE2(a, &c)
			a.MulByE2(a, &c)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (cyclotomic square) should output the same result", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.CyclotomicSquare(a)
			a.CyclotomicSquare(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (frobenius) should output the same result", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.Frobenius(a)
			a.Frobenius(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (frobenius square) should output the same result", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.FrobeniusSquare(a)
			a.FrobeniusSquare(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (frobenius cube) should output the same result", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.FrobeniusSquare(a)
			a.FrobeniusSquare(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestE6Ops(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 1

	properties := gopter.NewProperties(parameters)

	genA := GenE6()
	genB := GenE6()

	properties.Property("[BN256] sub & add should leave an element invariant", prop.ForAll(
		func(a, b *e6) bool {
			var c e6
			c.Set(a)
			c.Add(&c, b).Sub(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[BN256] mul & inverse should leave an element invariant", prop.ForAll(
		func(a, b *e6) bool {
			var c, d e6
			d.Inverse(b)
			c.Set(a)
			c.Mul(&c, b).Mul(&c, &d)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[BN256] inverse twice should leave an element invariant", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.Inverse(a).Inverse(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BN256] square and mul should output the same result", prop.ForAll(
		func(a *e6) bool {
			var b, c e6
			b.Mul(a, a)
			c.Square(a)
			return b.Equal(&c)
		},
		genA,
	))

	properties.Property("[BN256] pi**6=id", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.Frobenius(a).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b)
			return b.Equal(a)
		},
		genA,
	))

	properties.Property("[BN256] (pi**2)**3=id", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.FrobeniusSquare(a).
				FrobeniusSquare(&b).
				FrobeniusSquare(&b)
			return b.Equal(a)
		},
		genA,
	))

	properties.Property("[BN256] (pi**3)**2=id", prop.ForAll(
		func(a *e6) bool {
			var b e6
			b.FrobeniusCube(a).
				FrobeniusCube(&b)
			return b.Equal(a)
		},
		genA,
	))

	properties.Property("[BN256] cyclotomic square and square should be the same in the cyclotomic subgroup", prop.ForAll(
		func(a *e6) bool {
			var b, c, d e6
			b.Frobenius(a)
			c.FrobeniusSquare(a)
			a.Mul(a, &b).Mul(a, &c)
			b.Inverse(a)
			a.FrobeniusSquare(a).Mul(a, &b)
			c.CyclotomicSquare(a)
			d.Square(a)
			return c.Equal(&d)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

// ------------------------------------------------------------
// benches

func BenchmarkE6Add(b *testing.B) {
	var a, c e6
	a.SetRandom()
	c.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Add(&a, &c)
	}
}

func BenchmarkE6Sub(b *testing.B) {
	var a, c e6
	a.SetRandom()
	c.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Sub(&a, &c)
	}
}

func BenchmarkE6Mul(b *testing.B) {
	var a, c e6
	a.SetRandom()
	c.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Mul(&a, &c)
	}
}

func BenchmarkE6Square(b *testing.B) {
	var a e6
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Square(&a)
	}
}

func BenchmarkE6Inverse(b *testing.B) {
	var a e6
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Inverse(&a)
	}
}
