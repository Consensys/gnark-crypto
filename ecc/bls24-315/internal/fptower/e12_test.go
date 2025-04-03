// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"math/big"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// tests

func TestE12ReceiverIsOperand(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE12()
	genB := GenE12()

	properties.Property("[BLS24-315] Having the receiver as operand (addition) should output the same result", prop.ForAll(
		func(a, b *E12) bool {
			var c, d E12
			d.Set(a)
			c.Add(a, b)
			a.Add(a, b)
			b.Add(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (sub) should output the same result", prop.ForAll(
		func(a, b *E12) bool {
			var c, d E12
			d.Set(a)
			c.Sub(a, b)
			a.Sub(a, b)
			b.Sub(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (mul) should output the same result", prop.ForAll(
		func(a, b *E12) bool {
			var c, d E12
			d.Set(a)
			c.Mul(a, b)
			a.Mul(a, b)
			b.Mul(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (square) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Square(a)
			a.Square(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (neg) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Neg(a)
			a.Neg(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (double) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Double(a)
			a.Double(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] Having the receiver as operand (Inverse) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Inverse(a)
			a.Inverse(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestE12Ops(t *testing.T) {
	t.Parallel()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE12()
	genB := GenE12()

	properties.Property("[BLS24-315] sub & add should leave an element invariant", prop.ForAll(
		func(a, b *E12) bool {
			var c E12
			c.Set(a)
			c.Add(&c, b).Sub(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] mul & inverse should leave an element invariant", prop.ForAll(
		func(a, b *E12) bool {
			var c, d E12
			d.Inverse(b)
			c.Set(a)
			c.Mul(&c, b).Mul(&c, &d)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[BLS24-315] inverse twice should leave an element invariant", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Inverse(a).Inverse(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] neg twice should leave an element invariant", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Neg(a).Neg(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS24-315] square and mul should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b, c E12
			b.Mul(a, a)
			c.Square(a)
			return b.Equal(&c)
		},
		genA,
	))

	properties.Property("[BLS24-315] Double and add twice should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Add(a, a)
			a.Double(a)
			return a.Equal(&b)
		},
		genA,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkE12Add(b *testing.B) {
	var a, c E12
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Add(&a, &c)
	}
}

func BenchmarkE12Sub(b *testing.B) {
	var a, c E12
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Sub(&a, &c)
	}
}

func BenchmarkE12Mul(b *testing.B) {
	var a, c E12
	a.MustSetRandom()
	c.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Mul(&a, &c)
	}
}

func BenchmarkE12Square(b *testing.B) {
	var a E12
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Square(&a)
	}
}

func BenchmarkE12Inverse(b *testing.B) {
	var a E12
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Inverse(&a)
	}
}

func BenchmarkE12ExpBySeed(b *testing.B) {
	var a E12
	var seed big.Int
	seed.SetString("3218079743", 10)
	a.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Exp(a, &seed).Conjugate(&a)
	}
}
