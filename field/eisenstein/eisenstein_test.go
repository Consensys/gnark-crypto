package eisenstein

import (
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

const (
	nbFuzzShort = 10
	nbFuzz      = 50
	boundSize   = 128
)

func TestEisensteinReceiverIsOperand(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genE := GenComplexNumber(boundSize)

	properties.Property("Having the receiver as operand (addition) should output the same result", prop.ForAll(
		func(a, b *ComplexNumber) bool {
			var c, d ComplexNumber
			d.Set(a)
			c.Add(a, b)
			a.Add(a, b)
			b.Add(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genE,
		genE,
	))

	properties.Property("Having the receiver as operand (sub) should output the same result", prop.ForAll(
		func(a, b *ComplexNumber) bool {
			var c, d ComplexNumber
			d.Set(a)
			c.Sub(a, b)
			a.Sub(a, b)
			b.Sub(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genE,
		genE,
	))

	properties.Property("Having the receiver as operand (mul) should output the same result", prop.ForAll(
		func(a, b *ComplexNumber) bool {
			var c, d ComplexNumber
			d.Set(a)
			c.Mul(a, b)
			a.Mul(a, b)
			b.Mul(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genE,
		genE,
	))

	properties.Property("Having the receiver as operand (neg) should output the same result", prop.ForAll(
		func(a *ComplexNumber) bool {
			var b ComplexNumber
			b.Neg(a)
			a.Neg(a)
			return a.Equal(&b)
		},
		genE,
	))

	properties.Property("Having the receiver as operand (conjugate) should output the same result", prop.ForAll(
		func(a *ComplexNumber) bool {
			var b ComplexNumber
			b.Conjugate(a)
			a.Conjugate(a)
			return a.Equal(&b)
		},
		genE,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestEisensteinArithmetic(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genE := GenComplexNumber(boundSize)

	properties.Property("sub & add should leave an element invariant", prop.ForAll(
		func(a, b *ComplexNumber) bool {
			var c ComplexNumber
			c.Set(a)
			c.Add(&c, b).Sub(&c, b)
			return c.Equal(a)
		},
		genE,
		genE,
	))

	properties.Property("neg twice should leave an element invariant", prop.ForAll(
		func(a *ComplexNumber) bool {
			var b ComplexNumber
			b.Neg(a).Neg(&b)
			return a.Equal(&b)
		},
		genE,
	))

	properties.Property("conj twice should leave an element invariant", prop.ForAll(
		func(a *ComplexNumber) bool {
			var b ComplexNumber
			b.Conjugate(a).Conjugate(&b)
			return a.Equal(&b)
		},
		genE,
	))

	properties.Property("add zero should leave element invariant", prop.ForAll(
		func(a *ComplexNumber) bool {
			var b, zero ComplexNumber
			zero.SetZero()
			b.Add(a, &zero)
			return a.Equal(&b)
		},
		genE,
	))

	properties.Property("mul by one should leave element invariant", prop.ForAll(
		func(a *ComplexNumber) bool {
			var b, one ComplexNumber
			one.SetOne()
			b.Mul(a, &one)
			return a.Equal(&b)
		},
		genE,
	))

	properties.Property("add should be commutative", prop.ForAll(
		func(a, b *ComplexNumber) bool {
			var c, d ComplexNumber
			c.Add(a, b)
			d.Add(b, a)
			return c.Equal(&d)
		},
		genE,
		genE,
	))

	properties.Property("add should be assiocative", prop.ForAll(
		func(a, b, c *ComplexNumber) bool {
			var d, e ComplexNumber
			d.Add(a, b).Add(&d, c)
			e.Add(c, b).Add(&e, a)
			return e.Equal(&d)
		},
		genE,
		genE,
		genE,
	))

	properties.Property("mul should be commutative", prop.ForAll(
		func(a, b *ComplexNumber) bool {
			var c, d ComplexNumber
			c.Mul(a, b)
			d.Mul(b, a)
			return c.Equal(&d)
		},
		genE,
		genE,
	))

	properties.Property("mul should be assiocative", prop.ForAll(
		func(a, b, c *ComplexNumber) bool {
			var d, e ComplexNumber
			d.Mul(a, b).Mul(&d, c)
			e.Mul(c, b).Mul(&e, a)
			return e.Equal(&d)
		},
		genE,
		genE,
		genE,
	))

	properties.Property("norm should always be positive", prop.ForAll(
		func(a *ComplexNumber) bool {
			return a.Norm().Sign() >= 0
		},
		genE,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestEisensteinHalfGCD(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genE := GenComplexNumber(boundSize)

	properties.Property("half-GCD", prop.ForAll(
		func(a, b *ComplexNumber) bool {
			res := HalfGCD(a, b)
			var c, d ComplexNumber
			c.Mul(b, res[1])
			d.Mul(a, res[2])
			d.Add(&c, &d)
			return d.Equal(res[0])
		},
		genE,
		genE,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestEisensteinQuoRem(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)
	genE := GenComplexNumber(boundSize)

	properties.Property("QuoRem should be correct", prop.ForAll(
		func(a, b *ComplexNumber) bool {
			var z, rem ComplexNumber
			z.QuoRem(a, b, &rem)
			var res ComplexNumber
			res.Mul(b, &z)
			res.Add(&res, &rem)
			return res.Equal(a)
		},
		genE,
		genE,
	))

	properties.Property("QuoRem remainder should be smaller than divisor", prop.ForAll(
		func(a, b *ComplexNumber) bool {
			var z, rem ComplexNumber
			z.QuoRem(a, b, &rem)
			return rem.Norm().Cmp(b.Norm()) == -1
		},
		genE,
		genE,
	))
}

func TestRegressionHalfGCD1483(t *testing.T) {
	// This test is a regression test for issue #1483 in gnark
	a0, _ := new(big.Int).SetString("64502973549206556628585045361533709077", 10)
	a1, _ := new(big.Int).SetString("-303414439467246543595250775667605759171", 10)
	c0, _ := new(big.Int).SetString("-432420386565659656852420866390673177323", 10)
	c1, _ := new(big.Int).SetString("238911465918039986966665730306072050094", 10)
	a := ComplexNumber{A0: a0, A1: a1}
	c := ComplexNumber{A0: c0, A1: c1}

	ticker := time.NewTimer(time.Second * 3)
	doneCh := make(chan struct{})
	go func() {
		HalfGCD(&a, &c)
		close(doneCh)
	}()

	select {
	case <-ticker.C:
		t.Error("HalfGCD took too long to compute")
	case <-doneCh:
		// Test passed
	}
}

// GenNumber generates a random integer
func GenNumber(boundSize int64) gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var bound big.Int
		bound.Exp(big.NewInt(2), big.NewInt(boundSize), nil)
		elmt, _ := rand.Int(genParams.Rng, &bound)
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}

// GenComplexNumber generates a random integer
func GenComplexNumber(boundSize int64) gopter.Gen {
	return gopter.CombineGens(
		GenNumber(boundSize),
		GenNumber(boundSize),
	).Map(func(values []interface{}) *ComplexNumber {
		return &ComplexNumber{A0: values[0].(*big.Int), A1: values[1].(*big.Int)}
	})
}

// bench
var benchRes [3]*ComplexNumber

func BenchmarkHalfGCD(b *testing.B) {
	var n, _ = new(big.Int).SetString("100000000000000000000000000000000", 16) // 2^128
	a0, _ := rand.Int(rand.Reader, n)
	a1, _ := rand.Int(rand.Reader, n)
	c0, _ := rand.Int(rand.Reader, n)
	c1, _ := rand.Int(rand.Reader, n)
	a := ComplexNumber{A0: a0, A1: a1}
	c := ComplexNumber{A0: c0, A1: c1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchRes = HalfGCD(&a, &c)
	}
}
