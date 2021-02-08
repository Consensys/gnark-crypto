// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by goff (v0.3.12) DO NOT EDIT

package fp

import (
	"crypto/rand"
	"math/big"
	"math/bits"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// -------------------------------------------------------------------------------------------------
// benchmarks
// most benchmarks are rudimentary and should sample a large number of random inputs
// or be run multiple times to ensure it didn't measure the fastest path of the function

var benchResElement Element

func BenchmarkElementSetBytes(b *testing.B) {
	var x Element
	x.SetRandom()
	bb := x.Bytes()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchResElement.SetBytes(bb[:])
	}

}

func BenchmarkElementInverse(b *testing.B) {
	var x Element
	x.SetRandom()
	benchResElement.SetRandom()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchResElement.Inverse(&x)
	}

}

func BenchmarkElementExp(b *testing.B) {
	var x Element
	x.SetRandom()
	benchResElement.SetRandom()
	b1, _ := rand.Int(rand.Reader, Modulus())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Exp(x, b1)
	}
}

func BenchmarkElementDouble(b *testing.B) {
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Double(&benchResElement)
	}
}

func BenchmarkElementAdd(b *testing.B) {
	var x Element
	x.SetRandom()
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Add(&x, &benchResElement)
	}
}

func BenchmarkElementSub(b *testing.B) {
	var x Element
	x.SetRandom()
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Sub(&x, &benchResElement)
	}
}

func BenchmarkElementNeg(b *testing.B) {
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Neg(&benchResElement)
	}
}

func BenchmarkElementDiv(b *testing.B) {
	var x Element
	x.SetRandom()
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Div(&x, &benchResElement)
	}
}

func BenchmarkElementFromMont(b *testing.B) {
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.FromMont()
	}
}

func BenchmarkElementToMont(b *testing.B) {
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.ToMont()
	}
}
func BenchmarkElementSquare(b *testing.B) {
	benchResElement.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Square(&benchResElement)
	}
}

func BenchmarkElementSqrt(b *testing.B) {
	var a Element
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Sqrt(&a)
	}
}

func BenchmarkElementMul(b *testing.B) {
	x := Element{
		17644856173732828998,
		754043588434789617,
		10224657059481499349,
		7488229067341005760,
		11130996698012816685,
		1267921511277847466,
	}
	benchResElement.SetOne()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Mul(&benchResElement, &x)
	}
}

func BenchmarkElementCmp(b *testing.B) {
	x := Element{
		17644856173732828998,
		754043588434789617,
		10224657059481499349,
		7488229067341005760,
		11130996698012816685,
		1267921511277847466,
	}
	benchResElement = x
	benchResElement[0] = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchResElement.Cmp(&x)
	}
}

func TestElementCmp(t *testing.T) {
	var x, y Element

	if x.Cmp(&y) != 0 {
		t.Fatal("x == y")
	}

	one := One()
	y.Sub(&y, &one)

	if x.Cmp(&y) != -1 {
		t.Fatal("x < y")
	}
	if y.Cmp(&x) != 1 {
		t.Fatal("x < y")
	}

	x = y
	if x.Cmp(&y) != 0 {
		t.Fatal("x == y")
	}

	x.Sub(&x, &one)
	if x.Cmp(&y) != -1 {
		t.Fatal("x < y")
	}
	if y.Cmp(&x) != 1 {
		t.Fatal("x < y")
	}
}

func TestElementSetInterface(t *testing.T) {
	// TODO
	t.Skip("not implemented")
}

func TestElementIsRandom(t *testing.T) {
	for i := 0; i < 50; i++ {
		var x, y Element
		x.SetRandom()
		y.SetRandom()
		if x.Equal(&y) {
			t.Fatal("2 random numbers are unlikely to be equal")
		}
	}
}

// -------------------------------------------------------------------------------------------------
// Gopter tests
// most of them are generated with a template

const (
	nbFuzzShort = 200
	nbFuzz      = 1000
)

// special values to be used in tests
var staticTestValues []Element

func init() {
	staticTestValues = append(staticTestValues, Element{}) // zero
	staticTestValues = append(staticTestValues, One())     // one
	staticTestValues = append(staticTestValues, rSquare)   // r^2
	var e, one Element
	one.SetOne()
	e.Sub(&qElement, &one)
	staticTestValues = append(staticTestValues, e) // q - 1
	e.Double(&one)
	staticTestValues = append(staticTestValues, e) // 2

	{
		a := qElement
		a[5]--
		staticTestValues = append(staticTestValues, a)
	}
	{
		a := qElement
		a[0]--
		staticTestValues = append(staticTestValues, a)
	}

	{
		a := qElement
		a[5]--
		a[0]++
		staticTestValues = append(staticTestValues, a)
	}

}

func TestElementNegZero(t *testing.T) {
	var a, b Element
	b.SetZero()
	for a.IsZero() {
		a.SetRandom()
	}
	a.Neg(&b)
	if !a.IsZero() {
		t.Fatal("neg(0) != 0")
	}
}

func TestElementReduce(t *testing.T) {
	testValues := make([]Element, len(staticTestValues))
	copy(testValues, staticTestValues)

	for _, s := range testValues {
		expected := s
		reduce(&s)
		_reduceGeneric(&expected)
		if !s.Equal(&expected) {
			t.Fatal("reduce failed: asm and generic impl don't match")
		}
	}

	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := genFull()

	properties.Property("reduce should output a result smaller than modulus", prop.ForAll(
		func(a Element) bool {
			b := a
			reduce(&a)
			_reduceGeneric(&b)
			return !a.biggerOrEqualModulus() && a.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		t.Log("disabling ADX")
		supportAdx = false
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		supportAdx = true
	}

}

func TestElementBytes(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	properties.Property("SetBytes(Bytes()) should stayt constant", prop.ForAll(
		func(a testPairElement) bool {
			var b Element
			bytes := a.element.Bytes()
			b.SetBytes(bytes[:])
			return a.element.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestElementMulByConstants(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	implemented := []uint8{0, 1, 2, 3, 5}
	properties.Property("mulByConstant", prop.ForAll(
		func(a testPairElement) bool {
			for _, c := range implemented {
				var constant Element
				constant.SetUint64(uint64(c))

				b := a.element
				b.Mul(&b, &constant)

				aa := a.element
				mulByConstant(&aa, c)

				if !aa.Equal(&b) {
					return false
				}
			}

			return true
		},
		genA,
	))

	properties.Property("MulBy3(x) == Mul(x, 3)", prop.ForAll(
		func(a testPairElement) bool {
			var constant Element
			constant.SetUint64(3)

			b := a.element
			b.Mul(&b, &constant)

			MulBy3(&a.element)

			return a.element.Equal(&b)
		},
		genA,
	))

	properties.Property("MulBy5(x) == Mul(x, 5)", prop.ForAll(
		func(a testPairElement) bool {
			var constant Element
			constant.SetUint64(5)

			b := a.element
			b.Mul(&b, &constant)

			MulBy5(&a.element)

			return a.element.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		t.Log("disabling ADX")
		supportAdx = false
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		supportAdx = true
	}

}

func TestElementLegendre(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	properties.Property("legendre should output same result than big.Int.Jacobi", prop.ForAll(
		func(a testPairElement) bool {
			return a.element.Legendre() == big.Jacobi(&a.bigint, Modulus())
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		t.Log("disabling ADX")
		supportAdx = false
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		supportAdx = true
	}

}

func TestElementLexicographicallyLargest(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	properties.Property("element.Cmp should match LexicographicallyLargest output", prop.ForAll(
		func(a testPairElement) bool {
			var negA Element
			negA.Neg(&a.element)

			cmpResult := a.element.Cmp(&negA)
			lResult := a.element.LexicographicallyLargest()

			if lResult && cmpResult == 1 {
				return true
			}
			if !lResult && cmpResult != 1 {
				return true
			}
			return false
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		t.Log("disabling ADX")
		supportAdx = false
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		supportAdx = true
	}

}

func TestElementAdd(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()
	genB := gen()

	properties.Property("Add: having the receiver as operand should output the same result", prop.ForAll(
		func(a, b testPairElement) bool {
			var c, d Element
			d.Set(&a.element)

			c.Add(&a.element, &b.element)
			a.element.Add(&a.element, &b.element)
			b.element.Add(&d, &b.element)

			return a.element.Equal(&b.element) && a.element.Equal(&c) && b.element.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("Add: operation result must match big.Int result", prop.ForAll(
		func(a, b testPairElement) bool {
			{
				var c Element

				c.Add(&a.element, &b.element)

				var d, e big.Int
				d.Add(&a.bigint, &b.bigint).Mod(&d, Modulus())

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					return false
				}
			}

			// fixed elements
			// a is random
			// r takes special values
			testValues := make([]Element, len(staticTestValues))
			copy(testValues, staticTestValues)

			for _, r := range testValues {
				var d, e, rb big.Int
				r.ToBigIntRegular(&rb)

				var c Element
				c.Add(&a.element, &r)
				d.Add(&a.bigint, &rb).Mod(&d, Modulus())

				// checking generic impl against asm path
				var cGeneric Element
				_addGeneric(&cGeneric, &a.element, &r)
				if !cGeneric.Equal(&c) {
					// need to give context to failing error.
					return false
				}

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					return false
				}
			}
			return true
		},
		genA,
		genB,
	))

	properties.Property("Add: operation result must be smaller than modulus", prop.ForAll(
		func(a, b testPairElement) bool {
			var c Element

			c.Add(&a.element, &b.element)

			return !c.biggerOrEqualModulus()
		},
		genA,
		genB,
	))

	properties.Property("Add: assembly implementation must be consistent with generic one", prop.ForAll(
		func(a, b testPairElement) bool {
			var c, d Element
			c.Add(&a.element, &b.element)
			_addGeneric(&d, &a.element, &b.element)
			return c.Equal(&d)
		},
		genA,
		genB,
	))

	specialValueTest := func() {
		// test special values against special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			for _, b := range testValues {

				var bBig, d, e big.Int
				b.ToBigIntRegular(&bBig)

				var c Element
				c.Add(&a, &b)
				d.Add(&aBig, &bBig).Mod(&d, Modulus())

				// checking asm against generic impl
				var cGeneric Element
				_addGeneric(&cGeneric, &a, &b)
				if !cGeneric.Equal(&c) {
					t.Fatal("Add failed special test values: asm and generic impl don't match")
				}

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					t.Fatal("Add failed special test values")
				}
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		t.Log("disabling ADX")
		supportAdx = false
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}

func TestElementSub(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()
	genB := gen()

	properties.Property("Sub: having the receiver as operand should output the same result", prop.ForAll(
		func(a, b testPairElement) bool {
			var c, d Element
			d.Set(&a.element)

			c.Sub(&a.element, &b.element)
			a.element.Sub(&a.element, &b.element)
			b.element.Sub(&d, &b.element)

			return a.element.Equal(&b.element) && a.element.Equal(&c) && b.element.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("Sub: operation result must match big.Int result", prop.ForAll(
		func(a, b testPairElement) bool {
			{
				var c Element

				c.Sub(&a.element, &b.element)

				var d, e big.Int
				d.Sub(&a.bigint, &b.bigint).Mod(&d, Modulus())

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					return false
				}
			}

			// fixed elements
			// a is random
			// r takes special values
			testValues := make([]Element, len(staticTestValues))
			copy(testValues, staticTestValues)

			for _, r := range testValues {
				var d, e, rb big.Int
				r.ToBigIntRegular(&rb)

				var c Element
				c.Sub(&a.element, &r)
				d.Sub(&a.bigint, &rb).Mod(&d, Modulus())

				// checking generic impl against asm path
				var cGeneric Element
				_subGeneric(&cGeneric, &a.element, &r)
				if !cGeneric.Equal(&c) {
					// need to give context to failing error.
					return false
				}

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					return false
				}
			}
			return true
		},
		genA,
		genB,
	))

	properties.Property("Sub: operation result must be smaller than modulus", prop.ForAll(
		func(a, b testPairElement) bool {
			var c Element

			c.Sub(&a.element, &b.element)

			return !c.biggerOrEqualModulus()
		},
		genA,
		genB,
	))

	properties.Property("Sub: assembly implementation must be consistent with generic one", prop.ForAll(
		func(a, b testPairElement) bool {
			var c, d Element
			c.Sub(&a.element, &b.element)
			_subGeneric(&d, &a.element, &b.element)
			return c.Equal(&d)
		},
		genA,
		genB,
	))

	specialValueTest := func() {
		// test special values against special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			for _, b := range testValues {

				var bBig, d, e big.Int
				b.ToBigIntRegular(&bBig)

				var c Element
				c.Sub(&a, &b)
				d.Sub(&aBig, &bBig).Mod(&d, Modulus())

				// checking asm against generic impl
				var cGeneric Element
				_subGeneric(&cGeneric, &a, &b)
				if !cGeneric.Equal(&c) {
					t.Fatal("Sub failed special test values: asm and generic impl don't match")
				}

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					t.Fatal("Sub failed special test values")
				}
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		t.Log("disabling ADX")
		supportAdx = false
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}

func TestElementMul(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()
	genB := gen()

	properties.Property("Mul: having the receiver as operand should output the same result", prop.ForAll(
		func(a, b testPairElement) bool {
			var c, d Element
			d.Set(&a.element)

			c.Mul(&a.element, &b.element)
			a.element.Mul(&a.element, &b.element)
			b.element.Mul(&d, &b.element)

			return a.element.Equal(&b.element) && a.element.Equal(&c) && b.element.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("Mul: operation result must match big.Int result", prop.ForAll(
		func(a, b testPairElement) bool {
			{
				var c Element

				c.Mul(&a.element, &b.element)

				var d, e big.Int
				d.Mul(&a.bigint, &b.bigint).Mod(&d, Modulus())

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					return false
				}
			}

			// fixed elements
			// a is random
			// r takes special values
			testValues := make([]Element, len(staticTestValues))
			copy(testValues, staticTestValues)

			for _, r := range testValues {
				var d, e, rb big.Int
				r.ToBigIntRegular(&rb)

				var c Element
				c.Mul(&a.element, &r)
				d.Mul(&a.bigint, &rb).Mod(&d, Modulus())

				// checking generic impl against asm path
				var cGeneric Element
				_mulGeneric(&cGeneric, &a.element, &r)
				if !cGeneric.Equal(&c) {
					// need to give context to failing error.
					return false
				}

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					return false
				}
			}
			return true
		},
		genA,
		genB,
	))

	properties.Property("Mul: operation result must be smaller than modulus", prop.ForAll(
		func(a, b testPairElement) bool {
			var c Element

			c.Mul(&a.element, &b.element)

			return !c.biggerOrEqualModulus()
		},
		genA,
		genB,
	))

	properties.Property("Mul: assembly implementation must be consistent with generic one", prop.ForAll(
		func(a, b testPairElement) bool {
			var c, d Element
			c.Mul(&a.element, &b.element)
			_mulGeneric(&d, &a.element, &b.element)
			return c.Equal(&d)
		},
		genA,
		genB,
	))

	specialValueTest := func() {
		// test special values against special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			for _, b := range testValues {

				var bBig, d, e big.Int
				b.ToBigIntRegular(&bBig)

				var c Element
				c.Mul(&a, &b)
				d.Mul(&aBig, &bBig).Mod(&d, Modulus())

				// checking asm against generic impl
				var cGeneric Element
				_mulGeneric(&cGeneric, &a, &b)
				if !cGeneric.Equal(&c) {
					t.Fatal("Mul failed special test values: asm and generic impl don't match")
				}

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					t.Fatal("Mul failed special test values")
				}
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		t.Log("disabling ADX")
		supportAdx = false
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}

func TestElementDiv(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()
	genB := gen()

	properties.Property("Div: having the receiver as operand should output the same result", prop.ForAll(
		func(a, b testPairElement) bool {
			var c, d Element
			d.Set(&a.element)

			c.Div(&a.element, &b.element)
			a.element.Div(&a.element, &b.element)
			b.element.Div(&d, &b.element)

			return a.element.Equal(&b.element) && a.element.Equal(&c) && b.element.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("Div: operation result must match big.Int result", prop.ForAll(
		func(a, b testPairElement) bool {
			{
				var c Element

				c.Div(&a.element, &b.element)

				var d, e big.Int
				d.ModInverse(&b.bigint, Modulus())
				d.Mul(&d, &a.bigint).Mod(&d, Modulus())

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					return false
				}
			}

			// fixed elements
			// a is random
			// r takes special values
			testValues := make([]Element, len(staticTestValues))
			copy(testValues, staticTestValues)

			for _, r := range testValues {
				var d, e, rb big.Int
				r.ToBigIntRegular(&rb)

				var c Element
				c.Div(&a.element, &r)
				d.ModInverse(&rb, Modulus())
				d.Mul(&d, &a.bigint).Mod(&d, Modulus())

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					return false
				}
			}
			return true
		},
		genA,
		genB,
	))

	properties.Property("Div: operation result must be smaller than modulus", prop.ForAll(
		func(a, b testPairElement) bool {
			var c Element

			c.Div(&a.element, &b.element)

			return !c.biggerOrEqualModulus()
		},
		genA,
		genB,
	))

	specialValueTest := func() {
		// test special values against special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			for _, b := range testValues {

				var bBig, d, e big.Int
				b.ToBigIntRegular(&bBig)

				var c Element
				c.Div(&a, &b)
				d.ModInverse(&bBig, Modulus())
				d.Mul(&d, &aBig).Mod(&d, Modulus())

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					t.Fatal("Div failed special test values")
				}
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		t.Log("disabling ADX")
		supportAdx = false
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}

func TestElementExp(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()
	genB := gen()

	properties.Property("Exp: having the receiver as operand should output the same result", prop.ForAll(
		func(a, b testPairElement) bool {
			var c, d Element
			d.Set(&a.element)

			c.Exp(a.element, &b.bigint)
			a.element.Exp(a.element, &b.bigint)
			b.element.Exp(d, &b.bigint)

			return a.element.Equal(&b.element) && a.element.Equal(&c) && b.element.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("Exp: operation result must match big.Int result", prop.ForAll(
		func(a, b testPairElement) bool {
			{
				var c Element

				c.Exp(a.element, &b.bigint)

				var d, e big.Int
				d.Exp(&a.bigint, &b.bigint, Modulus())

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					return false
				}
			}

			// fixed elements
			// a is random
			// r takes special values
			testValues := make([]Element, len(staticTestValues))
			copy(testValues, staticTestValues)

			for _, r := range testValues {
				var d, e, rb big.Int
				r.ToBigIntRegular(&rb)

				var c Element
				c.Exp(a.element, &rb)
				d.Exp(&a.bigint, &rb, Modulus())

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					return false
				}
			}
			return true
		},
		genA,
		genB,
	))

	properties.Property("Exp: operation result must be smaller than modulus", prop.ForAll(
		func(a, b testPairElement) bool {
			var c Element

			c.Exp(a.element, &b.bigint)

			return !c.biggerOrEqualModulus()
		},
		genA,
		genB,
	))

	specialValueTest := func() {
		// test special values against special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			for _, b := range testValues {

				var bBig, d, e big.Int
				b.ToBigIntRegular(&bBig)

				var c Element
				c.Exp(a, &bBig)
				d.Exp(&aBig, &bBig, Modulus())

				if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
					t.Fatal("Exp failed special test values")
				}
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		t.Log("disabling ADX")
		supportAdx = false
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}

func TestElementSquare(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	properties.Property("Square: having the receiver as operand should output the same result", prop.ForAll(
		func(a testPairElement) bool {

			var b Element

			b.Square(&a.element)
			a.element.Square(&a.element)
			return a.element.Equal(&b)
		},
		genA,
	))

	properties.Property("Square: operation result must match big.Int result", prop.ForAll(
		func(a testPairElement) bool {
			var c Element
			c.Square(&a.element)

			var d, e big.Int
			d.Mul(&a.bigint, &a.bigint).Mod(&d, Modulus())

			return c.FromMont().ToBigInt(&e).Cmp(&d) == 0
		},
		genA,
	))

	properties.Property("Square: operation result must be smaller than modulus", prop.ForAll(
		func(a testPairElement) bool {
			var c Element
			c.Square(&a.element)
			return !c.biggerOrEqualModulus()
		},
		genA,
	))

	specialValueTest := func() {
		// test special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			var c Element
			c.Square(&a)

			var d, e big.Int
			d.Mul(&aBig, &aBig).Mod(&d, Modulus())

			if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
				t.Fatal("Square failed special test values")
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		supportAdx = false
		t.Log("disabling ADX")
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}

func TestElementInverse(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	properties.Property("Inverse: having the receiver as operand should output the same result", prop.ForAll(
		func(a testPairElement) bool {

			var b Element

			b.Inverse(&a.element)
			a.element.Inverse(&a.element)
			return a.element.Equal(&b)
		},
		genA,
	))

	properties.Property("Inverse: operation result must match big.Int result", prop.ForAll(
		func(a testPairElement) bool {
			var c Element
			c.Inverse(&a.element)

			var d, e big.Int
			d.ModInverse(&a.bigint, Modulus())

			return c.FromMont().ToBigInt(&e).Cmp(&d) == 0
		},
		genA,
	))

	properties.Property("Inverse: operation result must be smaller than modulus", prop.ForAll(
		func(a testPairElement) bool {
			var c Element
			c.Inverse(&a.element)
			return !c.biggerOrEqualModulus()
		},
		genA,
	))

	specialValueTest := func() {
		// test special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			var c Element
			c.Inverse(&a)

			var d, e big.Int
			d.ModInverse(&aBig, Modulus())

			if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
				t.Fatal("Inverse failed special test values")
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		supportAdx = false
		t.Log("disabling ADX")
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}

func TestElementSqrt(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	properties.Property("Sqrt: having the receiver as operand should output the same result", prop.ForAll(
		func(a testPairElement) bool {

			b := a.element

			b.Sqrt(&a.element)
			a.element.Sqrt(&a.element)
			return a.element.Equal(&b)
		},
		genA,
	))

	properties.Property("Sqrt: operation result must match big.Int result", prop.ForAll(
		func(a testPairElement) bool {
			var c Element
			c.Sqrt(&a.element)

			var d, e big.Int
			d.ModSqrt(&a.bigint, Modulus())

			return c.FromMont().ToBigInt(&e).Cmp(&d) == 0
		},
		genA,
	))

	properties.Property("Sqrt: operation result must be smaller than modulus", prop.ForAll(
		func(a testPairElement) bool {
			var c Element
			c.Sqrt(&a.element)
			return !c.biggerOrEqualModulus()
		},
		genA,
	))

	specialValueTest := func() {
		// test special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			var c Element
			c.Sqrt(&a)

			var d, e big.Int
			d.ModSqrt(&aBig, Modulus())

			if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
				t.Fatal("Sqrt failed special test values")
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		supportAdx = false
		t.Log("disabling ADX")
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}

func TestElementDouble(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	properties.Property("Double: having the receiver as operand should output the same result", prop.ForAll(
		func(a testPairElement) bool {

			var b Element

			b.Double(&a.element)
			a.element.Double(&a.element)
			return a.element.Equal(&b)
		},
		genA,
	))

	properties.Property("Double: operation result must match big.Int result", prop.ForAll(
		func(a testPairElement) bool {
			var c Element
			c.Double(&a.element)

			var d, e big.Int
			d.Lsh(&a.bigint, 1).Mod(&d, Modulus())

			return c.FromMont().ToBigInt(&e).Cmp(&d) == 0
		},
		genA,
	))

	properties.Property("Double: operation result must be smaller than modulus", prop.ForAll(
		func(a testPairElement) bool {
			var c Element
			c.Double(&a.element)
			return !c.biggerOrEqualModulus()
		},
		genA,
	))

	properties.Property("Double: assembly implementation must be consistent with generic one", prop.ForAll(
		func(a testPairElement) bool {
			var c, d Element
			c.Double(&a.element)
			_doubleGeneric(&d, &a.element)
			return c.Equal(&d)
		},
		genA,
	))

	specialValueTest := func() {
		// test special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			var c Element
			c.Double(&a)

			var d, e big.Int
			d.Lsh(&aBig, 1).Mod(&d, Modulus())

			// checking asm against generic impl
			var cGeneric Element
			_doubleGeneric(&cGeneric, &a)
			if !cGeneric.Equal(&c) {
				t.Fatal("Double failed special test values: asm and generic impl don't match")
			}

			if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
				t.Fatal("Double failed special test values")
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		supportAdx = false
		t.Log("disabling ADX")
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}

func TestElementNeg(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	properties.Property("Neg: having the receiver as operand should output the same result", prop.ForAll(
		func(a testPairElement) bool {

			var b Element

			b.Neg(&a.element)
			a.element.Neg(&a.element)
			return a.element.Equal(&b)
		},
		genA,
	))

	properties.Property("Neg: operation result must match big.Int result", prop.ForAll(
		func(a testPairElement) bool {
			var c Element
			c.Neg(&a.element)

			var d, e big.Int
			d.Neg(&a.bigint).Mod(&d, Modulus())

			return c.FromMont().ToBigInt(&e).Cmp(&d) == 0
		},
		genA,
	))

	properties.Property("Neg: operation result must be smaller than modulus", prop.ForAll(
		func(a testPairElement) bool {
			var c Element
			c.Neg(&a.element)
			return !c.biggerOrEqualModulus()
		},
		genA,
	))

	properties.Property("Neg: assembly implementation must be consistent with generic one", prop.ForAll(
		func(a testPairElement) bool {
			var c, d Element
			c.Neg(&a.element)
			_negGeneric(&d, &a.element)
			return c.Equal(&d)
		},
		genA,
	))

	specialValueTest := func() {
		// test special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			var c Element
			c.Neg(&a)

			var d, e big.Int
			d.Neg(&aBig).Mod(&d, Modulus())

			// checking asm against generic impl
			var cGeneric Element
			_negGeneric(&cGeneric, &a)
			if !cGeneric.Equal(&c) {
				t.Fatal("Neg failed special test values: asm and generic impl don't match")
			}

			if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
				t.Fatal("Neg failed special test values")
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		supportAdx = false
		t.Log("disabling ADX")
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}

func TestElementFromMont(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	properties.Property("Assembly implementation must be consistent with generic one", prop.ForAll(
		func(a testPairElement) bool {
			c := a.element
			d := a.element
			c.FromMont()
			_fromMontGeneric(&d)
			return c.Equal(&d)
		},
		genA,
	))

	properties.Property("x.FromMont().ToMont() == x", prop.ForAll(
		func(a testPairElement) bool {
			c := a.element
			c.FromMont().ToMont()
			return c.Equal(&a.element)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

type testPairElement struct {
	element Element
	bigint  big.Int
}

func (z *Element) biggerOrEqualModulus() bool {
	if z[5] > qElement[5] {
		return true
	}
	if z[5] < qElement[5] {
		return false
	}

	if z[4] > qElement[4] {
		return true
	}
	if z[4] < qElement[4] {
		return false
	}

	if z[3] > qElement[3] {
		return true
	}
	if z[3] < qElement[3] {
		return false
	}

	if z[2] > qElement[2] {
		return true
	}
	if z[2] < qElement[2] {
		return false
	}

	if z[1] > qElement[1] {
		return true
	}
	if z[1] < qElement[1] {
		return false
	}

	return z[0] >= qElement[0]
}

func gen() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var g testPairElement

		g.element = Element{
			genParams.NextUint64(),
			genParams.NextUint64(),
			genParams.NextUint64(),
			genParams.NextUint64(),
			genParams.NextUint64(),
			genParams.NextUint64(),
		}
		if qElement[5] != ^uint64(0) {
			g.element[5] %= (qElement[5] + 1)
		}

		for g.element.biggerOrEqualModulus() {
			g.element = Element{
				genParams.NextUint64(),
				genParams.NextUint64(),
				genParams.NextUint64(),
				genParams.NextUint64(),
				genParams.NextUint64(),
				genParams.NextUint64(),
			}
			if qElement[5] != ^uint64(0) {
				g.element[5] %= (qElement[5] + 1)
			}
		}

		g.element.ToBigIntRegular(&g.bigint)
		genResult := gopter.NewGenResult(g, gopter.NoShrinker)
		return genResult
	}
}

func genFull() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {

		genRandomFq := func() Element {
			var g Element

			g = Element{
				genParams.NextUint64(),
				genParams.NextUint64(),
				genParams.NextUint64(),
				genParams.NextUint64(),
				genParams.NextUint64(),
				genParams.NextUint64(),
			}

			if qElement[5] != ^uint64(0) {
				g[5] %= (qElement[5] + 1)
			}

			for g.biggerOrEqualModulus() {
				g = Element{
					genParams.NextUint64(),
					genParams.NextUint64(),
					genParams.NextUint64(),
					genParams.NextUint64(),
					genParams.NextUint64(),
					genParams.NextUint64(),
				}
				if qElement[5] != ^uint64(0) {
					g[5] %= (qElement[5] + 1)
				}
			}

			return g
		}
		a := genRandomFq()

		var carry uint64
		a[0], carry = bits.Add64(a[0], qElement[0], carry)
		a[1], carry = bits.Add64(a[1], qElement[1], carry)
		a[2], carry = bits.Add64(a[2], qElement[2], carry)
		a[3], carry = bits.Add64(a[3], qElement[3], carry)
		a[4], carry = bits.Add64(a[4], qElement[4], carry)
		a[5], _ = bits.Add64(a[5], qElement[5], carry)

		genResult := gopter.NewGenResult(a, gopter.NoShrinker)
		return genResult
	}
}
