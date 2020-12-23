// Copyright 2020 ConsenSys AG
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

package bw761

import (
	"math/big"
	"testing"

	"github.com/consensys/gurvy/bw761/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// tests

func TestPairing(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE6()
	genR1 := GenFr()
	genR2 := GenFr()

	properties.Property("Having the receiver as operand (final expo) should output the same result", prop.ForAll(
		func(a *GT) bool {
			var b GT
			b.Set(a)
			b = FinalExponentiation(a)
			*a = FinalExponentiation(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("Exponentiating FinalExpo(a) to r should output 1", prop.ForAll(
		func(a *GT) bool {
			var one GT
			var e big.Int
			e.SetString("258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177", 10)
			one.SetOne()
			*a = FinalExponentiation(a)
			a.Exp(a, e)
			return a.Equal(&one)
		},
		genA,
	))

	properties.Property("bilinearity", prop.ForAll(
		func(a, b fr.Element) bool {

			var res, resa, resb, resab, zero GT

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

			tmp, err := MillerLoop([]G1Affine{g1affine}, []G2Affine{g2affine})
			if err != nil {
				t.Fatal(err)
			}
			res = FinalExponentiation(&tmp)
			tmp, err = MillerLoop([]G1Affine{ag1}, []G2Affine{g2affine})
			if err != nil {
				t.Fatal(err)
			}
			resa = FinalExponentiation(&tmp)
			tmp, err = MillerLoop([]G1Affine{g1affine}, []G2Affine{bg2})
			if err != nil {
				t.Fatal(err)
			}
			resb = FinalExponentiation(&tmp)
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
		tmp, _ := MillerLoop([]G1Affine{g1GenAff}, []G2Affine{g2GenAff})
		FinalExponentiation(&tmp)
	}
}

func BenchmarkMillerLoop(b *testing.B) {

	var g1GenAff G1Affine
	var g2GenAff G2Affine

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MillerLoop([]G1Affine{g1GenAff}, []G2Affine{g2GenAff})
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
