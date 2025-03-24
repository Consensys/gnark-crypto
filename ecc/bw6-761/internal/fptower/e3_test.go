package fptower

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// tests

func TestE3ReceiverIsOperand(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE3()
	genB := GenE3()
	genfp := GenFp()

	properties.Property("[BW761] Having the receiver as operand (addition) should output the same result", prop.ForAll(
		func(a, b *E3) bool {
			var c, d E3
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
		func(a, b *E3) bool {
			var c, d E3
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
		func(a, b *E3) bool {
			var c, d E3
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
		func(a *E3) bool {
			var b E3
			b.Square(a)
			a.Square(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (neg) should output the same result", prop.ForAll(
		func(a *E3) bool {
			var b E3
			b.Neg(a)
			a.Neg(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (double) should output the same result", prop.ForAll(
		func(a *E3) bool {
			var b E3
			b.Double(a)
			a.Double(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (mul by non residue) should output the same result", prop.ForAll(
		func(a *E3) bool {
			var b E3
			b.MulByNonResidue(a)
			a.MulByNonResidue(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (Inverse) should output the same result", prop.ForAll(
		func(a *E3) bool {
			var b E3
			b.Inverse(a)
			a.Inverse(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Having the receiver as operand (mul by element) should output the same result", prop.ForAll(
		func(a *E3, b fp.Element) bool {
			var c E3
			c.MulByElement(a, &b)
			a.MulByElement(a, &b)
			return a.Equal(&c)
		},
		genA,
		genfp,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestE3Ops(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE3()
	genB := GenE3()
	genfp := GenFp()

	properties.Property("[BW761] sub & add should leave an element invariant", prop.ForAll(
		func(a, b *E3) bool {
			var c E3
			c.Set(a)
			c.Add(&c, b).Sub(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[BW761] mul & inverse should leave an element invariant", prop.ForAll(
		func(a, b *E3) bool {
			var c, d E3
			d.Inverse(b)
			c.Set(a)
			c.Mul(&c, b).Mul(&c, &d)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[BW761] inverse twice should leave an element invariant", prop.ForAll(
		func(a *E3) bool {
			var b E3
			b.Inverse(a).Inverse(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] BatchInvertE3 should output the same result as Inverse", prop.ForAll(
		func(a, b, c *E3) bool {

			batch := BatchInvertE3([]E3{*a, *b, *c})
			a.Inverse(a)
			b.Inverse(b)
			c.Inverse(c)
			return a.Equal(&batch[0]) && b.Equal(&batch[1]) && c.Equal(&batch[2])
		},
		genA,
		genA,
		genA,
	))

	properties.Property("[BW761] neg twice should leave an element invariant", prop.ForAll(
		func(a *E3) bool {
			var b E3
			b.Neg(a).Neg(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] square and mul should output the same result", prop.ForAll(
		func(a *E3) bool {
			var b, c E3
			b.Mul(a, a)
			c.Square(a)
			return b.Equal(&c)
		},
		genA,
	))

	properties.Property("[BW761] MulByElement MulByElement inverse should leave an element invariant", prop.ForAll(
		func(a *E3, b fp.Element) bool {
			var c E3
			var d fp.Element
			d.Inverse(&b)
			c.MulByElement(a, &b).MulByElement(&c, &d)
			return c.Equal(a)
		},
		genA,
		genfp,
	))

	properties.Property("[BW761] Double and mul by 2 should output the same result", prop.ForAll(
		func(a *E3) bool {
			var b E3
			var c fp.Element
			c.SetUint64(2)
			b.Double(a)
			a.MulByElement(a, &c)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BW761] Mulbynonres should be the same as multiplying by (0,1)", prop.ForAll(
		func(a *E3) bool {
			var b, c, d E3
			b.A1.SetOne()
			c.MulByNonResidue(a)
			d.Mul(a, &b)
			return c.Equal(&d)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkE3Add(b *testing.B) {
	var a, c E3
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Add(&a, &c)
	}
}

func BenchmarkE3Sub(b *testing.B) {
	var a, c E3
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Sub(&a, &c)
	}
}

func BenchmarkE3Mul(b *testing.B) {
	var a, c E3
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Mul(&a, &c)
	}
}

func BenchmarkE3MulByElement(b *testing.B) {
	var a E3
	var c fp.Element
	c.MustSetRandom()
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.MulByElement(&a, &c)
	}
}

func BenchmarkE3Square(b *testing.B) {
	var a E3
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Square(&a)
	}
}

func BenchmarkE3Inverse(b *testing.B) {
	var a E3
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Inverse(&a)
	}
}

func BenchmarkE3MulNonRes(b *testing.B) {
	var a E3
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.MulByNonResidue(&a)
	}
}
