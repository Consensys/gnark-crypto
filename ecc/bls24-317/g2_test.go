// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package bls24317

import (
	"fmt"
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls24-317/internal/fptower"

	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestG2Endomorphism(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS24-317] check that phi(P) = lambdaGLV * P", prop.ForAll(
		func(a fptower.E4) bool {
			var p, res1, res2 G2Jac
			g := MapToG2(a)
			p.FromAffine(&g)
			res1.phi(&p)
			res2.mulWindowed(&p, &lambdaGLV)

			return p.IsInSubGroup() && res1.Equal(&res2)
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] check that phi^2(P) + phi(P) + P = 0", prop.ForAll(
		func(a fptower.E4) bool {
			var p, res, tmp G2Jac
			g := MapToG2(a)
			p.FromAffine(&g)
			tmp.phi(&p)
			res.phi(&tmp).
				AddAssign(&tmp).
				AddAssign(&p)

			return res.Z.IsZero()
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] check that psi^2(P) = -phi(P)", prop.ForAll(
		func(a fptower.E4) bool {
			var p, res1, res2 G2Jac
			g := MapToG2(a)
			p.FromAffine(&g)
			res1.psi(&p).psi(&res1).psi(&res1).psi(&res1).Neg(&res1)
			res2.Set(&p)
			res2.X.MulByElement(&res2.X, &thirdRootOneG1)

			return p.IsInSubGroup() && res1.Equal(&res2)
		},
		GenE4(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestIsOnG2(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS24-317] g2Gen (affine) should be on the curve", prop.ForAll(
		func(a fptower.E4) bool {
			var op1, op2 G2Affine
			op1.FromJacobian(&g2Gen)
			op2.Set(&op1)
			op2.Y.Mul(&op2.Y, &a)
			return op1.IsOnCurve() && !op2.IsOnCurve()
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] g2Gen (Jacobian) should be on the curve", prop.ForAll(
		func(a fptower.E4) bool {
			var op1, op2, op3 G2Jac
			op1.Set(&g2Gen)
			op3.Set(&g2Gen)

			op2 = fuzzG2Jac(&g2Gen, a)
			op3.Y.Mul(&op3.Y, &a)
			return op1.IsOnCurve() && op2.IsOnCurve() && !op3.IsOnCurve()
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] IsInSubGroup and MulBy subgroup order should be the same", prop.ForAll(
		func(a fptower.E4) bool {
			var op1, op2 G2Jac
			op1 = fuzzG2Jac(&g2Gen, a)
			_r := fr.Modulus()
			op2.mulWindowed(&op1, _r)
			return op1.IsInSubGroup() && op2.Z.IsZero()
		},
		GenE4(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG2Conversions(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS24-317] Affine representation should be independent of the Jacobian representative", prop.ForAll(
		func(a fptower.E4) bool {
			g := fuzzG2Jac(&g2Gen, a)
			var op1 G2Affine
			op1.FromJacobian(&g)
			return op1.X.Equal(&g2Gen.X) && op1.Y.Equal(&g2Gen.Y)
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] Affine representation should be independent of a Extended Jacobian representative", prop.ForAll(
		func(a fptower.E4) bool {
			var g g2JacExtended
			g.X.Set(&g2Gen.X)
			g.Y.Set(&g2Gen.Y)
			g.ZZ.Set(&g2Gen.Z)
			g.ZZZ.Set(&g2Gen.Z)
			gfuzz := fuzzg2JacExtended(&g, a)

			var op1 G2Affine
			op1.fromJacExtended(&gfuzz)
			return op1.X.Equal(&g2Gen.X) && op1.Y.Equal(&g2Gen.Y)
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] Jacobian representation should be the same as the affine representative", prop.ForAll(
		func(a fptower.E4) bool {
			var g G2Jac
			var op1 G2Affine
			op1.X.Set(&g2Gen.X)
			op1.Y.Set(&g2Gen.Y)

			var one fptower.E4
			one.SetOne()

			g.FromAffine(&op1)

			return g.X.Equal(&g2Gen.X) && g.Y.Equal(&g2Gen.Y) && g.Z.Equal(&one)
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] Converting affine symbol for infinity to Jacobian should output correct infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G2Affine
			g.X.SetZero()
			g.Y.SetZero()
			var op1 G2Jac
			op1.FromAffine(&g)
			var one, zero fptower.E4
			one.SetOne()
			return op1.X.Equal(&one) && op1.Y.Equal(&one) && op1.Z.Equal(&zero)
		},
	))

	properties.Property("[BLS24-317] Converting infinity in extended Jacobian to affine should output infinity symbol in Affine", prop.ForAll(
		func() bool {
			var g G2Affine
			var op1 g2JacExtended
			var zero fptower.E4
			op1.X.Set(&g2Gen.X)
			op1.Y.Set(&g2Gen.Y)
			g.fromJacExtended(&op1)
			return g.X.Equal(&zero) && g.Y.Equal(&zero)
		},
	))

	properties.Property("[BLS24-317] Converting infinity in extended Jacobian to Jacobian should output infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G2Jac
			var op1 g2JacExtended
			var zero, one fptower.E4
			one.SetOne()
			op1.X.Set(&g2Gen.X)
			op1.Y.Set(&g2Gen.Y)
			g.fromJacExtended(&op1)
			return g.X.Equal(&one) && g.Y.Equal(&one) && g.Z.Equal(&zero)
		},
	))

	properties.Property("[BLS24-317] [Jacobian] Two representatives of the same class should be equal", prop.ForAll(
		func(a, b fptower.E4) bool {
			op1 := fuzzG2Jac(&g2Gen, a)
			op2 := fuzzG2Jac(&g2Gen, b)
			return op1.Equal(&op2)
		},
		GenE4(),
		GenE4(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG2AffineOps(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	genScalar := GenFr()

	properties.Property("[BLS24-317] Add(P,-P) should return the point at infinity", prop.ForAll(
		func(s fr.Element) bool {
			var op1, op2 G2Affine
			var sInt big.Int
			g := g2GenAff
			s.BigInt(&sInt)
			op1.ScalarMultiplication(&g, &sInt)
			op2.Neg(&op1)

			op1.Add(&op1, &op2)
			return op1.IsInfinity()

		},
		GenFr(),
	))

	properties.Property("[BLS24-317] Add(P,0) and Add(0,P) should return P", prop.ForAll(
		func(s fr.Element) bool {
			var op1, op2 G2Affine
			var sInt big.Int
			g := g2GenAff
			s.BigInt(&sInt)
			op1.ScalarMultiplication(&g, &sInt)
			op2.SetInfinity()

			op1.Add(&op1, &op2)
			op2.Add(&op2, &op1)
			return op1.Equal(&op2)

		},
		GenFr(),
	))

	properties.Property("[BLS24-317] Add should call double when adding the same point", prop.ForAll(
		func(s fr.Element) bool {
			var op1, op2 G2Affine
			var sInt big.Int
			g := g2GenAff
			s.BigInt(&sInt)
			op1.ScalarMultiplication(&g, &sInt)

			op2.Double(&op1)
			op1.Add(&op1, &op1)
			return op1.Equal(&op2)

		},
		GenFr(),
	))

	properties.Property("[BLS24-317] [2]G = double(G) + G - G", prop.ForAll(
		func(s fr.Element) bool {
			var sInt big.Int
			g := g2GenAff
			s.BigInt(&sInt)
			g.ScalarMultiplication(&g, &sInt)
			var op1, op2 G2Affine
			op1.ScalarMultiplication(&g, big.NewInt(2))
			op2.Double(&g)
			op2.Add(&op2, &g)
			op2.Sub(&op2, &g)
			return op1.Equal(&op2)
		},
		GenFr(),
	))

	properties.Property("[BLS24-317] [-s]G = -[s]G", prop.ForAll(
		func(s fr.Element) bool {
			g := g2GenAff
			var gj G2Jac
			var nbs, bs big.Int
			s.BigInt(&bs)
			nbs.Neg(&bs)

			var res = true

			// mulGLV
			{
				var op1, op2 G2Affine
				op1.ScalarMultiplication(&g, &bs).Neg(&op1)
				op2.ScalarMultiplication(&g, &nbs)
				res = res && op1.Equal(&op2)
			}

			// mulWindowed
			{
				var op1, op2 G2Jac
				op1.mulWindowed(&gj, &bs).Neg(&op1)
				op2.mulWindowed(&gj, &nbs)
				res = res && op1.Equal(&op2)
			}

			return res
		},
		GenFr(),
	))

	properties.Property("[BLS24-317] [Jacobian] Add should call double when adding the same point", prop.ForAll(
		func(a, b fptower.E4) bool {
			fop1 := fuzzG2Jac(&g2Gen, a)
			fop2 := fuzzG2Jac(&g2Gen, b)
			var op1, op2 G2Jac
			op1.Set(&fop1).AddAssign(&fop2)
			op2.Double(&fop2)
			return op1.Equal(&op2)
		},
		GenE4(),
		GenE4(),
	))

	properties.Property("[BLS24-317] [Jacobian] Adding the opposite of a point to itself should output inf", prop.ForAll(
		func(a, b fptower.E4) bool {
			fop1 := fuzzG2Jac(&g2Gen, a)
			fop2 := fuzzG2Jac(&g2Gen, b)
			fop2.Neg(&fop2)
			fop1.AddAssign(&fop2)
			return fop1.Equal(&g2Infinity)
		},
		GenE4(),
		GenE4(),
	))

	properties.Property("[BLS24-317] [Jacobian] Adding the inf to a point should not modify the point", prop.ForAll(
		func(a fptower.E4) bool {
			fop1 := fuzzG2Jac(&g2Gen, a)
			fop1.AddAssign(&g2Infinity)
			var op2 G2Jac
			op2.Set(&g2Infinity)
			op2.AddAssign(&g2Gen)
			return fop1.Equal(&g2Gen) && op2.Equal(&g2Gen)
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] [Jacobian Extended] addMixed (-G) should equal subMixed(G)", prop.ForAll(
		func(a fptower.E4) bool {
			fop1 := fuzzG2Jac(&g2Gen, a)
			var p1, p1Neg G2Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 g2JacExtended
			o1.addMixed(&p1Neg)
			o2.subMixed(&p1)

			return o1.X.Equal(&o2.X) &&
				o1.Y.Equal(&o2.Y) &&
				o1.ZZ.Equal(&o2.ZZ) &&
				o1.ZZZ.Equal(&o2.ZZZ)
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] [Jacobian Extended] doubleMixed (-G) should equal doubleNegMixed(G)", prop.ForAll(
		func(a fptower.E4) bool {
			fop1 := fuzzG2Jac(&g2Gen, a)
			var p1, p1Neg G2Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 g2JacExtended
			o1.doubleMixed(&p1Neg)
			o2.doubleNegMixed(&p1)

			return o1.X.Equal(&o2.X) &&
				o1.Y.Equal(&o2.Y) &&
				o1.ZZ.Equal(&o2.ZZ) &&
				o1.ZZZ.Equal(&o2.ZZZ)
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] [Jacobian] Addmix the negation to itself should output 0", prop.ForAll(
		func(a fptower.E4) bool {
			fop1 := fuzzG2Jac(&g2Gen, a)
			fop1.Neg(&fop1)
			var op2 G2Affine
			op2.FromJacobian(&g2Gen)
			fop1.AddMixed(&op2)
			return fop1.Equal(&g2Infinity)
		},
		GenE4(),
	))

	properties.Property("[BLS24-317] scalar multiplication (double and add) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g G2Jac
			g.ScalarMultiplication(&g2Gen, r)

			var scalar, blindedScalar, rminusone big.Int
			var op1, op2, op3, gneg G2Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.mulWindowed(&g2Gen, &rminusone)
			gneg.Neg(&g2Gen)
			s.BigInt(&scalar)
			blindedScalar.Mul(&scalar, r).Add(&blindedScalar, &scalar)
			op1.mulWindowed(&g2Gen, &scalar)
			op2.mulWindowed(&g2Gen, &blindedScalar)

			return op1.Equal(&op2) && g.Equal(&g2Infinity) && !op1.Equal(&g2Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	properties.Property("[BLS24-317] psi should map points from E' to itself", prop.ForAll(
		func() bool {
			var a G2Jac
			a.psi(&g2Gen)
			return a.IsOnCurve() && !a.Equal(&g2Gen)
		},
	))

	properties.Property("[BLS24-317] scalar multiplication (GLV) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g G2Jac
			g.mulGLV(&g2Gen, r)

			var scalar, blindedScalar, rminusone big.Int
			var op1, op2, op3, gneg G2Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.ScalarMultiplication(&g2Gen, &rminusone)
			gneg.Neg(&g2Gen)
			s.BigInt(&scalar)
			blindedScalar.Mul(&scalar, r).Add(&blindedScalar, &scalar)
			op1.ScalarMultiplication(&g2Gen, &scalar)
			op2.ScalarMultiplication(&g2Gen, &blindedScalar)

			return op1.Equal(&op2) && g.Equal(&g2Infinity) && !op1.Equal(&g2Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	properties.Property("[BLS24-317] GLV and Double and Add should output the same result", prop.ForAll(
		func(s fr.Element) bool {

			var r big.Int
			var op1, op2 G2Jac
			s.BigInt(&r)
			op1.mulWindowed(&g2Gen, &r)
			op2.mulGLV(&g2Gen, &r)
			return op1.Equal(&op2) && !op1.Equal(&g2Infinity)

		},
		genScalar,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG2CofactorClearing(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS24-317] Clearing the cofactor of a random point should set it in the r-torsion", prop.ForAll(
		func() bool {
			var a, x, b fptower.E4
			a.MustSetRandom()

			x.Square(&a).Mul(&x, &a).Add(&x, &bTwistCurveCoeff)
			for x.Legendre() != 1 {
				a.MustSetRandom()
				x.Square(&a).Mul(&x, &a).Add(&x, &bTwistCurveCoeff)
			}

			b.Sqrt(&x)
			var point, pointCleared, infinity G2Jac
			point.X.Set(&a)
			point.Y.Set(&b)
			point.Z.SetOne()
			pointCleared.ClearCofactor(&point)
			infinity.Set(&g2Infinity)
			return point.IsOnCurve() && pointCleared.IsInSubGroup() && !pointCleared.Equal(&infinity)
		},
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

func TestG2BatchScalarMultiplication(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzzShort
	}

	properties := gopter.NewProperties(parameters)

	genScalar := GenFr()

	// size of the multiExps
	const nbSamples = 10

	properties.Property("[BLS24-317] BatchScalarMultiplication should be consistent with individual scalar multiplications", prop.ForAll(
		func(mixer fr.Element) bool {
			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					Mul(&sampleScalars[i-1], &mixer)
			}

			result := BatchScalarMultiplicationG2(&g2GenAff, sampleScalars[:])

			if len(result) != len(sampleScalars) {
				return false
			}

			for i := 0; i < len(result); i++ {
				var expectedJac G2Jac
				var expected G2Affine
				var b big.Int
				expectedJac.ScalarMultiplication(&g2Gen, sampleScalars[i].BigInt(&b))
				expected.FromJacobian(&expectedJac)
				if !result[i].Equal(&expected) {
					return false
				}
			}
			return true
		},
		genScalar,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkG2JacIsInSubGroup(b *testing.B) {
	var a G2Jac
	a.Set(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.IsInSubGroup()
	}

}

func BenchmarkG2JacEqual(b *testing.B) {
	var scalar fptower.E4
	scalar.MustSetRandom()

	var a G2Jac
	a.ScalarMultiplication(&g2Gen, big.NewInt(42))

	b.Run("equal", func(b *testing.B) {
		var scalarSquared fptower.E4
		scalarSquared.Square(&scalar)

		aZScaled := a
		aZScaled.X.Mul(&aZScaled.X, &scalarSquared)
		aZScaled.Y.Mul(&aZScaled.Y, &scalarSquared).Mul(&aZScaled.Y, &scalar)
		aZScaled.Z.Mul(&aZScaled.Z, &scalar)

		// Check the setup.
		if !a.Equal(&aZScaled) {
			b.Fatalf("invalid test setup")
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			a.Equal(&aZScaled)
		}
	})

	b.Run("not equal", func(b *testing.B) {
		var aPlus1 G2Jac
		aPlus1.AddAssign(&g2Gen)

		// Check the setup.
		if a.Equal(&aPlus1) {
			b.Fatalf("invalid test setup")
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			a.Equal(&aPlus1)
		}
	})
}

func BenchmarkBatchAddG2Affine(b *testing.B) {

	var P, R pG2AffineC16
	var RR ppG2AffineC16
	ridx := make([]int, len(P))

	// TODO P == R may produce skewed benches
	fillBenchBasesG2(P[:])
	fillBenchBasesG2(R[:])

	for i := 0; i < len(ridx); i++ {
		ridx[i] = i
	}

	// random permute
	rand.Shuffle(len(ridx), func(i, j int) { ridx[i], ridx[j] = ridx[j], ridx[i] })

	for i, ri := range ridx {
		RR[i] = &R[ri]
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batchAddG2Affine[pG2AffineC16, ppG2AffineC16, cG2AffineC16](&RR, &P, len(P))
	}
}

func BenchmarkG2AffineBatchScalarMultiplication(b *testing.B) {
	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	const pow = 15
	const nbSamples = 1 << pow

	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer)
	}

	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_ = BatchScalarMultiplicationG2(&g2GenAff, sampleScalars[:using])
			}
		})
	}
}

func BenchmarkG2JacScalarMultiplication(b *testing.B) {

	var scalar big.Int
	r := fr.Modulus()
	scalar.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)
	scalar.Add(&scalar, r)

	var doubleAndAdd G2Jac

	b.Run("double and add", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			doubleAndAdd.mulWindowed(&g2Gen, &scalar)
		}
	})

	var glv G2Jac
	b.Run("GLV", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			glv.mulGLV(&g2Gen, &scalar)
		}
	})

}

func BenchmarkG2AffineCofactorClearing(b *testing.B) {
	var a G2Jac
	a.Set(&g2Gen)
	for i := 0; i < b.N; i++ {
		a.ClearCofactor(&a)
	}
}

func BenchmarkG2JacAdd(b *testing.B) {
	var a G2Jac
	a.Double(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddAssign(&g2Gen)
	}
}

func BenchmarkG2JacAddMixed(b *testing.B) {
	var a G2Jac
	a.Double(&g2Gen)

	var c G2Affine
	c.FromJacobian(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddMixed(&c)
	}

}

func BenchmarkG2JacDouble(b *testing.B) {
	var a G2Jac
	a.Set(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.DoubleAssign()
	}

}

func BenchmarkG2JacExtAddMixed(b *testing.B) {
	var a g2JacExtended
	a.doubleMixed(&g2GenAff)

	var c G2Affine
	c.FromJacobian(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.addMixed(&c)
	}
}

func BenchmarkG2JacExtSubMixed(b *testing.B) {
	var a g2JacExtended
	a.doubleMixed(&g2GenAff)

	var c G2Affine
	c.FromJacobian(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.subMixed(&c)
	}
}

func BenchmarkG2JacExtDoubleMixed(b *testing.B) {
	var a g2JacExtended
	a.doubleMixed(&g2GenAff)

	var c G2Affine
	c.FromJacobian(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.doubleMixed(&c)
	}
}

func BenchmarkG2JacExtDoubleNegMixed(b *testing.B) {
	var a g2JacExtended
	a.doubleMixed(&g2GenAff)

	var c G2Affine
	c.FromJacobian(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.doubleNegMixed(&c)
	}
}

func BenchmarkG2JacExtAdd(b *testing.B) {
	var a, c g2JacExtended
	a.doubleMixed(&g2GenAff)
	c.double(&a)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.add(&c)
	}
}

func BenchmarkG2JacExtDouble(b *testing.B) {
	var a g2JacExtended
	a.doubleMixed(&g2GenAff)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.double(&a)
	}
}

func BenchmarkG2AffineAdd(b *testing.B) {
	var a G2Affine
	a.Double(&g2GenAff)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Add(&a, &g2GenAff)
	}
}

func BenchmarkG2AffineDouble(b *testing.B) {
	var a G2Affine
	a.Double(&g2GenAff)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Double(&a)
	}
}

func fuzzG2Jac(p *G2Jac, f fptower.E4) G2Jac {
	var res G2Jac
	res.X.Mul(&p.X, &f).Mul(&res.X, &f)
	res.Y.Mul(&p.Y, &f).Mul(&res.Y, &f).Mul(&res.Y, &f)
	res.Z.Mul(&p.Z, &f)
	return res
}

func fuzzg2JacExtended(p *g2JacExtended, f fptower.E4) g2JacExtended {
	var res g2JacExtended
	var ff, fff fptower.E4
	ff.Square(&f)
	fff.Mul(&ff, &f)
	res.X.Mul(&p.X, &ff)
	res.Y.Mul(&p.Y, &fff)
	res.ZZ.Mul(&p.ZZ, &ff)
	res.ZZZ.Mul(&p.ZZZ, &fff)
	return res
}
