package pairing

// PairingTests ...
const PairingTests = `

import (
	"math/big"
	"testing"

	"github.com/consensys/gurvy/{{ .CurveName }}/fr"    
    "github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// tests

func TestPairing(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE12()
	genR1 := GenFr()
	genR2 := GenFr()

	properties.Property("[{{ toUpper .CurveName}}] Having the receiver as operand (final expo) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Set(a)
			b.FinalExponentiation(a)
			a.FinalExponentiation(a)
			return a.Equal(&b)
		},
		genA,
	))

    properties.Property("[{{ toUpper .CurveName}}] Exponentiating FinalExpo(a) to r should output 1", prop.ForAll(
		func(a *E12) bool {
			var one E12
			var e big.Int
			e.SetString("{{ .RTorsion }}", 10)
			one.SetOne()
			a.FinalExponentiation(a).Exp(a, e)
			return a.Equal(&one)
		},
		genA,
	))

	properties.Property("[{{ toUpper .CurveName}}] bilinearity", prop.ForAll(
		func(a, b fr.Element) bool {

			var res, resa, resb, resab, zero PairingResult

			var aG1 G1Jac
			var bG2 G2Jac

			var g1affine, ag1 G1Affine
			var g2affine, bg2 G2Affine

			var abigint, bbigint, ab big.Int

			a.ToBigIntRegular(&abigint)
			b.ToBigIntRegular(&bbigint)
			ab.Mul(&abigint, &bbigint)

			g1affine.FromJacobian(&g1Gen)
			g2affine.FromJacobian(&g2Gen)

			aG1.ScalarMultiplication(&g1Gen, &abigint)
			bG2.ScalarMultiplication(&g2Gen, &bbigint)
			ag1.FromJacobian(&aG1)
			bg2.FromJacobian(&bG2)

			res = FinalExponentiation(MillerLoop(g1affine, g2affine))
			resa = FinalExponentiation(MillerLoop(ag1, g2affine))
			resb = FinalExponentiation(MillerLoop(g1affine, bg2))
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

	var g1GenAff G1Affine
	var g2GenAff G2Affine

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FinalExponentiation(MillerLoop(g1GenAff, g2GenAff))
	}
}

func BenchmarkFinalExponentiation(b *testing.B) {

	var a E12
	a.SetRandom()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FinalExponentiation(&a)
	}

}

`
