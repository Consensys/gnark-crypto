// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls12377

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// tests

func TestPairingCubical(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genR1 := GenFr()
	genR2 := GenFr()

	properties.Property("[BLS12-377] bilinearity", prop.ForAll(
		func(a, b fr.Element) bool {

			var res, resa, resb, resab, zero GT

			var ag1 G1Affine
			var bg2 G2Affine

			var abigint, bbigint, ab big.Int

			a.BigInt(&abigint)
			b.BigInt(&bbigint)
			ab.Mul(&abigint, &bbigint)

			ag1.ScalarMultiplication(&g1GenAff, &abigint)
			bg2.ScalarMultiplication(&g2GenAff, &bbigint)

			res, _ = PairCubical([]G1Affine{g1GenAff}, []G2Affine{g2GenAff})
			resa, _ = PairCubical([]G1Affine{ag1}, []G2Affine{g2GenAff})
			resb, _ = PairCubical([]G1Affine{g1GenAff}, []G2Affine{bg2})

			resab.Exp(res, &ab)
			resa.Exp(resa, &bbigint)
			resb.Exp(resb, &abigint)

			return resab.Equal(&resa) && resab.Equal(&resb) && !res.Equal(&zero)

		},
		genR1,
		genR2,
	))

	properties.Property("[BLS12-377] PairingCubicalCheck", prop.ForAll(
		func(a, b fr.Element) bool {

			var g1GenAffNeg G1Affine
			g1GenAffNeg.Neg(&g1GenAff)
			tabP := []G1Affine{g1GenAff, g1GenAffNeg}
			tabQ := []G2Affine{g2GenAff, g2GenAff}

			res, _ := PairingCubicalCheck(tabP, tabQ)

			return res
		},
		genR1,
		genR2,
	))

	// Cross-check: cubical and Miller loop pairings should agree on CHECK results.
	// This is the key property for cryptographic applications (BLS signatures, zkSNARKs).
	// Both pairings compute e(P, Q) * e(-P, Q) == 1 correctly.
	properties.Property("[BLS12-377] cubical and Miller agree on pairing check", prop.ForAll(
		func(a, b fr.Element) bool {

			var ag1 G1Affine
			var bg2 G2Affine

			var abigint, bbigint big.Int

			a.BigInt(&abigint)
			b.BigInt(&bbigint)

			ag1.ScalarMultiplication(&g1GenAff, &abigint)
			bg2.ScalarMultiplication(&g2GenAff, &bbigint)

			// Create -ag1
			var ag1Neg G1Affine
			ag1Neg.Neg(&ag1)

			// Both should return true for e(P, Q) * e(-P, Q)
			millerCheck, err1 := PairingCheck(
				[]G1Affine{ag1, ag1Neg},
				[]G2Affine{bg2, bg2},
			)
			cubicalCheck, err2 := PairingCubicalCheck(
				[]G1Affine{ag1, ag1Neg},
				[]G2Affine{bg2, bg2},
			)

			if err1 != nil || err2 != nil {
				return false
			}

			return millerCheck == cubicalCheck && millerCheck == true
		},
		genR1,
		genR2,
	))

	// Cross-check: incorrect pairing equations should return false for both
	properties.Property("[BLS12-377] cubical and Miller agree on failed pairing check", prop.ForAll(
		func(a, b fr.Element) bool {

			var ag1 G1Affine
			var bg2 G2Affine

			var abigint, bbigint big.Int

			a.BigInt(&abigint)
			b.BigInt(&bbigint)

			// Ensure non-trivial scalars
			if abigint.Sign() == 0 {
				abigint.SetInt64(1)
			}
			if bbigint.Sign() == 0 {
				bbigint.SetInt64(1)
			}

			ag1.ScalarMultiplication(&g1GenAff, &abigint)
			bg2.ScalarMultiplication(&g2GenAff, &bbigint)

			// e(P, Q) * e(P, Q) should NOT equal 1 (unless P or Q is infinity)
			millerCheck, err1 := PairingCheck(
				[]G1Affine{ag1, ag1},
				[]G2Affine{bg2, bg2},
			)
			cubicalCheck, err2 := PairingCubicalCheck(
				[]G1Affine{ag1, ag1},
				[]G2Affine{bg2, bg2},
			)

			if err1 != nil || err2 != nil {
				return false
			}

			// Both should return false
			return millerCheck == cubicalCheck && millerCheck == false
		},
		genR1,
		genR2,
	))

	// Cross-check: both pairings should be in the correct subgroup
	properties.Property("[BLS12-377] cubical pairing is in GT subgroup", prop.ForAll(
		func(a, b fr.Element) bool {

			var ag1 G1Affine
			var bg2 G2Affine

			var abigint, bbigint big.Int

			a.BigInt(&abigint)
			b.BigInt(&bbigint)

			ag1.ScalarMultiplication(&g1GenAff, &abigint)
			bg2.ScalarMultiplication(&g2GenAff, &bbigint)

			cubical, err := PairCubical([]G1Affine{ag1}, []G2Affine{bg2})
			if err != nil {
				return false
			}

			return cubical.IsInSubGroup()
		},
		genR1,
		genR2,
	))

	// Verify that e_cub = e_mil^k for some fixed k ∈ Z_r*.
	// We can't compute k directly (would require solving discrete log in 253-bit group),
	// but we can verify the relationship holds by checking that the ratio e_cub/e_mil
	// scales consistently with bilinearity: ratio(aP, bQ) = ratio(P, Q)^{ab}.
	properties.Property("[BLS12-377] cubical and Miller pairings differ by fixed exponent", prop.ForAll(
		func(a, b fr.Element) bool {

			var ag1 G1Affine
			var bg2 G2Affine

			var abigint, bbigint, ab big.Int

			a.BigInt(&abigint)
			b.BigInt(&bbigint)
			ab.Mul(&abigint, &bbigint)

			ag1.ScalarMultiplication(&g1GenAff, &abigint)
			bg2.ScalarMultiplication(&g2GenAff, &bbigint)

			// Compute ratio for generators: ratio = e_cub(G1, G2) / e_mil(G1, G2)
			eMilGen, _ := Pair([]G1Affine{g1GenAff}, []G2Affine{g2GenAff})
			eCubGen, _ := PairCubical([]G1Affine{g1GenAff}, []G2Affine{g2GenAff})

			var ratioGen GT
			ratioGen.Div(&eCubGen, &eMilGen)

			// Compute ratio for (aP, bQ): ratio_ab = e_cub(aP, bQ) / e_mil(aP, bQ)
			eMilAB, _ := Pair([]G1Affine{ag1}, []G2Affine{bg2})
			eCubAB, _ := PairCubical([]G1Affine{ag1}, []G2Affine{bg2})

			var ratioAB GT
			ratioAB.Div(&eCubAB, &eMilAB)

			// If e_cub = e_mil^k, then:
			//   ratio_ab = e_cub(aP, bQ) / e_mil(aP, bQ)
			//            = e_mil(aP, bQ)^{k-1}
			//            = (e_mil(P, Q)^{ab})^{k-1}
			//            = (e_mil(P, Q)^{k-1})^{ab}
			//            = ratio^{ab}
			var expectedRatio GT
			expectedRatio.Exp(ratioGen, &ab)

			return ratioAB.Equal(&expectedRatio)
		},
		genR1,
		genR2,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkPairingCubical(b *testing.B) {

	var g1GenAff G1Affine
	var g2GenAff G2Affine

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PairCubical([]G1Affine{g1GenAff}, []G2Affine{g2GenAff})
	}
}
