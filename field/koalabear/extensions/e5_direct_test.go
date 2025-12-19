// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// genE5D generates an E5D elmt
func genE5D() gopter.Gen {
	return gopter.CombineGens(
		genFr(),
		genFr(),
		genFr(),
		genFr(),
		genFr(),
	).Map(func(values []interface{}) *E5D {
		return &E5D{A0: values[0].(fr.Element), A1: values[1].(fr.Element), A2: values[2].(fr.Element), A3: values[3].(fr.Element), A4: values[4].(fr.Element)}
	})
}

// ------------------------------------------------------------
// tests
func TestE5DReceiverIsOperand(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := genE5D()
	genB := genE5D()

	properties.Property("[(direct) koalabear] Having the receiver as operand (addition) should output the same result", prop.ForAll(
		func(a, b *E5D) bool {
			var c, d E5D
			d.Set(a)
			c.Add(a, b)
			a.Add(a, b)
			b.Add(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[(direct) koalabear] Having the receiver as operand (sub) should output the same result", prop.ForAll(
		func(a, b *E5D) bool {
			var c, d E5D
			d.Set(a)
			c.Sub(a, b)
			a.Sub(a, b)
			b.Sub(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[(direct) koalabear] Having the receiver as operand (mul) should output the same result", prop.ForAll(
		func(a, b *E5D) bool {
			var c, d E5D
			d.Set(a)
			c.Mul(a, b)
			a.Mul(a, b)
			b.Mul(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[(direct) koalabear] Having the receiver as operand (square) should output the same result", prop.ForAll(
		func(a *E5D) bool {
			var b E5D
			b.Square(a)
			a.Square(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[(direct) koalabear] Having the receiver as operand (double) should output the same result", prop.ForAll(
		func(a *E5D) bool {
			var b E5D
			b.Double(a)
			a.Double(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestE5DOps(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := genE5D()
	genB := genE5D()

	properties.Property("[(direct) koalabear] all mul algorithms match naive", prop.ForAll(
		func(a, b *E5D) bool {
			var c1, c2, c3 E5D
			c1.mulElMGuiIon5(a, b)
			c2.mulMontgomery5(a, b)
			c3.mulNaive(a, b)
			return c1.Equal(&c2) && c1.Equal(&c3)
		},
		genA,
		genB,
	))

	properties.Property("[(direct) koalabear] sub & add should leave an element invariant", prop.ForAll(
		func(a, b *E5D) bool {
			var c E5D
			c.Set(a)
			c.Add(&c, b).Sub(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[(direct) koalabear] mul & square should output the same result when squaring", prop.ForAll(
		func(a *E5D) bool {
			var b, c E5D
			b.Mul(a, a)
			c.Square(a)
			return b.Equal(&c)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// bench
func BenchmarkE5DMulElMGuiIon(b *testing.B) {
	var a, c E5D
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.mulElMGuiIon5(&a, &c)
	}
}

func BenchmarkE5DMulMontgomery(b *testing.B) {
	var a, c E5D
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.mulMontgomery5(&a, &c)
	}
}
