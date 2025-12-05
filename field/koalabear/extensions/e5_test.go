// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// tests
func TestE5ReceiverIsOperand(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := genE5()
	genB := genE5()

	properties.Property("[koalabear] Having the receiver as operand (addition) should output the same result", prop.ForAll(
		func(a, b *E5) bool {
			var c, d E5
			d.Set(a)
			c.Add(a, b)
			a.Add(a, b)
			b.Add(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[koalabear] Having the receiver as operand (sub) should output the same result", prop.ForAll(
		func(a, b *E5) bool {
			var c, d E5
			d.Set(a)
			c.Sub(a, b)
			a.Sub(a, b)
			b.Sub(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[koalabear] Having the receiver as operand (mul) should output the same result", prop.ForAll(
		func(a, b *E5) bool {
			var c, d E5
			d.Set(a)
			c.Mul(a, b)
			a.Mul(a, b)
			b.Mul(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[koalabear] Having the receiver as operand (square) should output the same result", prop.ForAll(
		func(a *E5) bool {
			var b E5
			b.Square(a)
			a.Square(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[koalabear] Having the receiver as operand (double) should output the same result", prop.ForAll(
		func(a *E5) bool {
			var b E5
			b.Double(a)
			a.Double(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[koalabear] Having the receiver as operand (Inverse) should output the same result", prop.ForAll(
		func(a *E5) bool {
			var b E5
			b.Inverse(a)
			a.Inverse(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestE5Ops(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := genE5()
	genB := genE5()

	properties.Property("[koalabear] sub & add should leave an element invariant", prop.ForAll(
		func(a, b *E5) bool {
			var c E5
			c.Set(a)
			c.Add(&c, b).Sub(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[koalabear] mul & inverse should leave an element invariant", prop.ForAll(
		func(a, b *E5) bool {
			var c, d E5
			d.Inverse(b)
			c.Set(a)
			c.Mul(&c, b).Mul(&c, &d)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[koalabear] inverse twice should leave an element invariant", prop.ForAll(
		func(a *E5) bool {
			var b E5
			b.Inverse(a).Inverse(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[koalabear] square and mul should output the same result", prop.ForAll(
		func(a *E5) bool {
			var b, c E5
			b.Mul(a, a)
			c.Square(a)
			return b.Equal(&c)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// bench
func BenchmarkE5Mul(b *testing.B) {
	var a, c E5
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Mul(&a, &c)
	}
}

// genE5 generates an E5 elmt
func genE5() gopter.Gen {
	return gopter.CombineGens(
		genFr(),
		genFr(),
		genFr(),
		genFr(),
		genFr(),
		genFr(),
	).Map(func(values []interface{}) *E5 {
		return &E5{A0: values[0].(fr.Element), A1: values[1].(fr.Element), A2: values[2].(fr.Element), A3: values[3].(fr.Element), A4: values[4].(fr.Element)}
	})
}
