package pairing

// PairingTests ...
const PairingTests = `

import (
	"math/big"
	"fmt"
	"testing"

	"github.com/consensys/gurvy/{{ .Name }}/fr"    
    "github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// examples

func ExampleMillerLoop() {

	// samples a random scalar r
	var r big.Int
	var rFr fr.Element
	rFr.SetRandom()
	rFr.ToBigIntRegular(&r)

	// computes r*g1Gen, r*g2Gen
	var rg1 G1
	var rg2 G2
	rg1.ScalarMultiplication(&g1GenAff, &r)
	rg2.ScalarMultiplication(&g2GenAff, &r)

	// Computes e(g1GenAff, ag2) and e(ag1, g2GenAff)
	e1 := FinalExponentiation(MillerLoop(g1GenAff, rg2))
	E2 := FinalExponentiation(MillerLoop(rg1, g2GenAff))

	// checks that bilinearity property holds
	check := e1.Equal(&E2)

	fmt.Printf("%t\n", check)
	// Output: true

}

// ------------------------------------------------------------
// tests

func TestPairing(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE12()
	genR1 := GenFr()
	genR2 := GenFr()

	properties.Property("[{{ toUpper .Name}}] Having the receiver as operand (final expo) should output the same result", prop.ForAll(
		func(a *GT) bool {
			var b GT
			b.Set(a)
			b = FinalExponentiation(a)
			*a = FinalExponentiation(a)
			return a.Equal(&b)
		},
		genA,
	))

    properties.Property("[{{ toUpper .Name}}] Exponentiating FinalExpo(a) to r should output 1", prop.ForAll(
		func(a *GT) bool {
			var one GT
			e := fr.Modulus()
			one.SetOne()
			*a = FinalExponentiation(a)
			a.Exp(a, *e)
			return a.Equal(&one)
		},
		genA,
	))

	properties.Property("[{{ toUpper .Name}}] bilinearity", prop.ForAll(
		func(a, b fr.Element) bool {

			var res, resa, resb, resab, zero GT

			var ag1 G1
			var bg2 G2

			var abigint, bbigint, ab big.Int

			a.ToBigIntRegular(&abigint)
			b.ToBigIntRegular(&bbigint)
			ab.Mul(&abigint, &bbigint)

			ag1.ScalarMultiplication(&g1GenAff, &abigint)
			bg2.ScalarMultiplication(&g2GenAff, &bbigint)

			res = FinalExponentiation(MillerLoop(g1GenAff, g2GenAff))
			resa = FinalExponentiation(MillerLoop(ag1, g2GenAff))
			resb = FinalExponentiation(MillerLoop(g1GenAff, bg2))

			resab.Exp(&res, ab)
			resa.Exp(&resa, bbigint)
			resb.Exp(&resb, abigint)

			return resab.Equal(&resa) && resab.Equal(&resb) && !res.Equal(&zero)

		},
		genR1,
		genR2,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkPairing(b *testing.B) {

	var g1GenAff G1
	var g2GenAff G2

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FinalExponentiation(MillerLoop(g1GenAff, g2GenAff))
	}
}

func BenchmarkFinalExponentiation(b *testing.B) {

	var a GT
	a.SetRandom()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FinalExponentiation(&a)
	}

}

`
