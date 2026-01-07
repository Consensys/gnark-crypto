// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package secp256r1

import (
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/secp256r1/fp"

	"github.com/consensys/gnark-crypto/ecc/secp256r1/fr"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestIsOnG1(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[SECP256R1] g1Gen (affine) should be on the curve", prop.ForAll(
		func(a fp.Element) bool {
			var op1, op2 G1Affine
			op1.FromJacobian(&g1Gen)
			op2.Set(&op1)
			op2.Y.Mul(&op2.Y, &a)
			return op1.IsOnCurve() && !op2.IsOnCurve()
		},
		GenFp(),
	))

	properties.Property("[SECP256R1] g1Gen (Jacobian) should be on the curve", prop.ForAll(
		func(a fp.Element) bool {
			var op1, op2, op3 G1Jac
			op1.Set(&g1Gen)
			op3.Set(&g1Gen)

			op2 = fuzzG1Jac(&g1Gen, a)
			op3.Y.Mul(&op3.Y, &a)
			return op1.IsOnCurve() && op2.IsOnCurve() && !op3.IsOnCurve()
		},
		GenFp(),
	))

	properties.Property("[SECP256R1] IsInSubGroup and MulBy subgroup order should be the same", prop.ForAll(
		func(a fp.Element) bool {
			var op1, op2 G1Jac
			op1 = fuzzG1Jac(&g1Gen, a)
			_r := fr.Modulus()
			op2.mulWindowed(&op1, _r)
			return op1.IsInSubGroup() && op2.Z.IsZero()
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG1Conversions(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[SECP256R1] Affine representation should be independent of the Jacobian representative", prop.ForAll(
		func(a fp.Element) bool {
			g := fuzzG1Jac(&g1Gen, a)
			var op1 G1Affine
			op1.FromJacobian(&g)
			return op1.X.Equal(&g1Gen.X) && op1.Y.Equal(&g1Gen.Y)
		},
		GenFp(),
	))

	properties.Property("[SECP256R1] Affine representation should be independent of a Extended Jacobian representative", prop.ForAll(
		func(a fp.Element) bool {
			var g g1JacExtended
			g.X.Set(&g1Gen.X)
			g.Y.Set(&g1Gen.Y)
			g.ZZ.Set(&g1Gen.Z)
			g.ZZZ.Set(&g1Gen.Z)
			gfuzz := fuzzg1JacExtended(&g, a)

			var op1 G1Affine
			op1.fromJacExtended(&gfuzz)
			return op1.X.Equal(&g1Gen.X) && op1.Y.Equal(&g1Gen.Y)
		},
		GenFp(),
	))

	properties.Property("[SECP256R1] Jacobian representation should be the same as the affine representative", prop.ForAll(
		func(a fp.Element) bool {
			var g G1Jac
			var op1 G1Affine
			op1.X.Set(&g1Gen.X)
			op1.Y.Set(&g1Gen.Y)

			var one fp.Element
			one.SetOne()

			g.FromAffine(&op1)

			return g.X.Equal(&g1Gen.X) && g.Y.Equal(&g1Gen.Y) && g.Z.Equal(&one)
		},
		GenFp(),
	))

	properties.Property("[SECP256R1] Converting affine symbol for infinity to Jacobian should output correct infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G1Affine
			g.X.SetZero()
			g.Y.SetZero()
			var op1 G1Jac
			op1.FromAffine(&g)
			var one, zero fp.Element
			one.SetOne()
			return op1.X.Equal(&one) && op1.Y.Equal(&one) && op1.Z.Equal(&zero)
		},
	))

	properties.Property("[SECP256R1] Converting infinity in extended Jacobian to affine should output infinity symbol in Affine", prop.ForAll(
		func() bool {
			var g G1Affine
			var op1 g1JacExtended
			var zero fp.Element
			op1.X.Set(&g1Gen.X)
			op1.Y.Set(&g1Gen.Y)
			g.fromJacExtended(&op1)
			return g.X.Equal(&zero) && g.Y.Equal(&zero)
		},
	))

	properties.Property("[SECP256R1] Converting infinity in extended Jacobian to Jacobian should output infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G1Jac
			var op1 g1JacExtended
			var zero, one fp.Element
			one.SetOne()
			op1.X.Set(&g1Gen.X)
			op1.Y.Set(&g1Gen.Y)
			g.fromJacExtended(&op1)
			return g.X.Equal(&one) && g.Y.Equal(&one) && g.Z.Equal(&zero)
		},
	))

	properties.Property("[SECP256R1] [Jacobian] Two representatives of the same class should be equal", prop.ForAll(
		func(a, b fp.Element) bool {
			op1 := fuzzG1Jac(&g1Gen, a)
			op2 := fuzzG1Jac(&g1Gen, b)
			return op1.Equal(&op2)
		},
		GenFp(),
		GenFp(),
	))
	properties.Property("[SECP256R1] BatchJacobianToAffineG1 and FromJacobian should output the same result", prop.ForAll(
		func(a, b fp.Element) bool {
			g1 := fuzzG1Jac(&g1Gen, a)
			g2 := fuzzG1Jac(&g1Gen, b)
			var op1, op2 G1Affine
			op1.FromJacobian(&g1)
			op2.FromJacobian(&g2)
			baseTableAff := BatchJacobianToAffineG1([]G1Jac{g1, g2})
			return op1.Equal(&baseTableAff[0]) && op2.Equal(&baseTableAff[1])
		},
		GenFp(),
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG1AffineOps(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	genScalar := GenFr()

	properties.Property("[SECP256R1] Add(P,-P) should return the point at infinity", prop.ForAll(
		func(s fr.Element) bool {
			var op1, op2 G1Affine
			var sInt big.Int
			g := g1GenAff
			s.BigInt(&sInt)
			op1.ScalarMultiplication(&g, &sInt)
			op2.Neg(&op1)

			op1.Add(&op1, &op2)
			return op1.IsInfinity()

		},
		GenFr(),
	))

	properties.Property("[SECP256R1] Add(P,0) and Add(0,P) should return P", prop.ForAll(
		func(s fr.Element) bool {
			var op1, op2 G1Affine
			var sInt big.Int
			g := g1GenAff
			s.BigInt(&sInt)
			op1.ScalarMultiplication(&g, &sInt)
			op2.SetInfinity()

			op1.Add(&op1, &op2)
			op2.Add(&op2, &op1)
			return op1.Equal(&op2)

		},
		GenFr(),
	))

	properties.Property("[SECP256R1] Add should call double when adding the same point", prop.ForAll(
		func(s fr.Element) bool {
			var op1, op2 G1Affine
			var sInt big.Int
			g := g1GenAff
			s.BigInt(&sInt)
			op1.ScalarMultiplication(&g, &sInt)

			op2.Double(&op1)
			op1.Add(&op1, &op1)
			return op1.Equal(&op2)

		},
		GenFr(),
	))

	properties.Property("[SECP256R1] [2]G = double(G) + G - G", prop.ForAll(
		func(s fr.Element) bool {
			var sInt big.Int
			g := g1GenAff
			s.BigInt(&sInt)
			g.ScalarMultiplication(&g, &sInt)
			var op1, op2 G1Affine
			op1.ScalarMultiplication(&g, big.NewInt(2))
			op2.Double(&g)
			op2.Add(&op2, &g)
			op2.Sub(&op2, &g)
			return op1.Equal(&op2)
		},
		GenFr(),
	))

	properties.Property("[SECP256R1] [-s]G = -[s]G", prop.ForAll(
		func(s fr.Element) bool {
			var nbs, bs big.Int
			s.BigInt(&bs)
			nbs.Neg(&bs)

			var res = true

			var op1, op2 G1Jac
			op1.mulWindowed(&g1Gen, &bs).Neg(&op1)
			op2.mulWindowed(&g1Gen, &nbs)
			res = res && op1.Equal(&op2)

			return res
		},
		GenFr(),
	))

	properties.Property("[SECP256R1] [Jacobian] Add should call double when adding the same point", prop.ForAll(
		func(a, b fp.Element) bool {
			fop1 := fuzzG1Jac(&g1Gen, a)
			fop2 := fuzzG1Jac(&g1Gen, b)
			var op1, op2 G1Jac
			op1.Set(&fop1).AddAssign(&fop2)
			op2.Double(&fop2)
			return op1.Equal(&op2)
		},
		GenFp(),
		GenFp(),
	))

	properties.Property("[SECP256R1] [Jacobian] Adding the opposite of a point to itself should output inf", prop.ForAll(
		func(a, b fp.Element) bool {
			fop1 := fuzzG1Jac(&g1Gen, a)
			fop2 := fuzzG1Jac(&g1Gen, b)
			fop2.Neg(&fop2)
			fop1.AddAssign(&fop2)
			return fop1.Equal(&g1Infinity)
		},
		GenFp(),
		GenFp(),
	))

	properties.Property("[SECP256R1] [Jacobian] Adding the inf to a point should not modify the point", prop.ForAll(
		func(a fp.Element) bool {
			fop1 := fuzzG1Jac(&g1Gen, a)
			fop1.AddAssign(&g1Infinity)
			var op2 G1Jac
			op2.Set(&g1Infinity)
			op2.AddAssign(&g1Gen)
			return fop1.Equal(&g1Gen) && op2.Equal(&g1Gen)
		},
		GenFp(),
	))

	properties.Property("[SECP256R1] [Jacobian Extended] addMixed (-G) should equal subMixed(G)", prop.ForAll(
		func(a fp.Element) bool {
			fop1 := fuzzG1Jac(&g1Gen, a)
			var p1, p1Neg G1Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 g1JacExtended
			o1.addMixed(&p1Neg)
			o2.subMixed(&p1)

			return o1.X.Equal(&o2.X) &&
				o1.Y.Equal(&o2.Y) &&
				o1.ZZ.Equal(&o2.ZZ) &&
				o1.ZZZ.Equal(&o2.ZZZ)
		},
		GenFp(),
	))

	properties.Property("[SECP256R1] [Jacobian Extended] doubleMixed (-G) should equal doubleNegMixed(G)", prop.ForAll(
		func(a fp.Element) bool {
			fop1 := fuzzG1Jac(&g1Gen, a)
			var p1, p1Neg G1Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 g1JacExtended
			o1.doubleMixed(&p1Neg)
			o2.doubleNegMixed(&p1)

			return o1.X.Equal(&o2.X) &&
				o1.Y.Equal(&o2.Y) &&
				o1.ZZ.Equal(&o2.ZZ) &&
				o1.ZZZ.Equal(&o2.ZZZ)
		},
		GenFp(),
	))

	properties.Property("[SECP256R1] [Jacobian] Addmix the negation to itself should output 0", prop.ForAll(
		func(a fp.Element) bool {
			fop1 := fuzzG1Jac(&g1Gen, a)
			fop1.Neg(&fop1)
			var op2 G1Affine
			op2.FromJacobian(&g1Gen)
			fop1.AddMixed(&op2)
			return fop1.Equal(&g1Infinity)
		},
		GenFp(),
	))

	properties.Property("[SECP256R1] scalar multiplication (double and add) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g G1Jac
			g.ScalarMultiplication(&g1Gen, r)

			var scalar, blindedScalar, rminusone big.Int
			var op1, op2, op3, gneg G1Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.mulWindowed(&g1Gen, &rminusone)
			gneg.Neg(&g1Gen)
			s.BigInt(&scalar)
			blindedScalar.Mul(&scalar, r).Add(&blindedScalar, &scalar)
			op1.mulWindowed(&g1Gen, &scalar)
			op2.mulWindowed(&g1Gen, &blindedScalar)

			return op1.Equal(&op2) && g.Equal(&g1Infinity) && !op1.Equal(&g1Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))
	properties.Property("[SECP256R1] JointScalarMultiplicationBase and ScalarMultiplication should output the same results", prop.ForAll(
		func(s1, s2 fr.Element) bool {

			var op1, op2, temp G1Jac

			op1.JointScalarMultiplicationBase(&g1GenAff, s1.BigInt(new(big.Int)), s2.BigInt(new(big.Int)))
			temp.ScalarMultiplication(&g1Gen, s2.BigInt(new(big.Int)))
			op2.ScalarMultiplication(&g1Gen, s1.BigInt(new(big.Int))).
				AddAssign(&temp)

			return op1.Equal(&op2)

		},
		genScalar,
		genScalar,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG1CrossImplementations(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[SECP256R1] ScalarMultiplication should output the same result as crypto/elliptic", prop.ForAll(
		func(s fr.Element) bool {
			curve := elliptic.P256()
			Gx, Gy := curve.Params().Gx, curve.Params().Gy
			var S big.Int
			s.BigInt(&S)
			Qx, Qy := curve.ScalarMult(Gx, Gy, S.Bytes())
			var Q G1Affine
			Q.ScalarMultiplication(&g1GenAff, &S)
			var qx, qy fp.Element
			qx.SetBigInt(Qx)
			qy.SetBigInt(Qy)
			return Q.X.Equal(&qx) && Q.Y.Equal(&qy)
		},
		GenFr(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkG1JacIsInSubGroup(b *testing.B) {
	var a G1Jac
	a.Set(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.IsInSubGroup()
	}

}

func BenchmarkG1JacEqual(b *testing.B) {
	var scalar fp.Element
	scalar.MustSetRandom()

	var a G1Jac
	a.ScalarMultiplication(&g1Gen, big.NewInt(42))

	b.Run("equal", func(b *testing.B) {
		var scalarSquared fp.Element
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
		var aPlus1 G1Jac
		aPlus1.AddAssign(&g1Gen)

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

func BenchmarkG1JacScalarMultiplication(b *testing.B) {
	for i := 0; i <= fr.Modulus().BitLen(); i += 8 {
		bound := new(big.Int).Lsh(big.NewInt(1), uint(i))
		scalar, err := crand.Int(crand.Reader, bound)
		if err != nil {
			b.Fatalf("failed to generate random scalar: %v", err)
		}

		var doubleAndAdd G1Jac
		b.Run(fmt.Sprintf("method=window/scalarwidth=%d", i), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				doubleAndAdd.mulWindowed(&g1Gen, scalar)
			}
		})
	}
}

func BenchmarkG1JacScalarMultiplicationMethod(b *testing.B) {
	for i := 0; i <= fr.Modulus().BitLen(); i += 8 {
		bound := new(big.Int).Lsh(big.NewInt(1), uint(i))
		scalar, err := crand.Int(crand.Reader, bound)
		if err != nil {
			b.Fatalf("failed to generate random scalar: %v", err)
		}

		var res G1Jac
		b.Run(fmt.Sprintf("scalarwidth=%d", i), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				res.ScalarMultiplication(&g1Gen, scalar)
			}
		})
	}
}

func BenchmarkG1JacAdd(b *testing.B) {
	var a G1Jac
	a.Double(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddAssign(&g1Gen)
	}
}

func BenchmarkG1JacAddMixed(b *testing.B) {
	var a G1Jac
	a.Double(&g1Gen)

	var c G1Affine
	c.FromJacobian(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddMixed(&c)
	}

}

func BenchmarkG1JacDouble(b *testing.B) {
	var a G1Jac
	a.Set(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.DoubleAssign()
	}

}

func BenchmarkG1JacExtAddMixed(b *testing.B) {
	var a g1JacExtended
	a.doubleMixed(&g1GenAff)

	var c G1Affine
	c.FromJacobian(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.addMixed(&c)
	}
}

func BenchmarkG1JacExtSubMixed(b *testing.B) {
	var a g1JacExtended
	a.doubleMixed(&g1GenAff)

	var c G1Affine
	c.FromJacobian(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.subMixed(&c)
	}
}

func BenchmarkG1JacExtDoubleMixed(b *testing.B) {
	var a g1JacExtended
	a.doubleMixed(&g1GenAff)

	var c G1Affine
	c.FromJacobian(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.doubleMixed(&c)
	}
}

func BenchmarkG1JacExtDoubleNegMixed(b *testing.B) {
	var a g1JacExtended
	a.doubleMixed(&g1GenAff)

	var c G1Affine
	c.FromJacobian(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.doubleNegMixed(&c)
	}
}

func BenchmarkG1JacExtAdd(b *testing.B) {
	var a, c g1JacExtended
	a.doubleMixed(&g1GenAff)
	c.double(&a)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.add(&c)
	}
}

func BenchmarkG1JacExtDouble(b *testing.B) {
	var a g1JacExtended
	a.doubleMixed(&g1GenAff)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.double(&a)
	}
}

func BenchmarkG1AffineAdd(b *testing.B) {
	var a G1Affine
	a.Double(&g1GenAff)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Add(&a, &g1GenAff)
	}
}

func BenchmarkG1AffineDouble(b *testing.B) {
	var a G1Affine
	a.Double(&g1GenAff)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Double(&a)
	}
}

func fuzzG1Jac(p *G1Jac, f fp.Element) G1Jac {
	var res G1Jac
	res.X.Mul(&p.X, &f).Mul(&res.X, &f)
	res.Y.Mul(&p.Y, &f).Mul(&res.Y, &f).Mul(&res.Y, &f)
	res.Z.Mul(&p.Z, &f)
	return res
}

func fuzzg1JacExtended(p *g1JacExtended, f fp.Element) g1JacExtended {
	var res g1JacExtended
	var ff, fff fp.Element
	ff.Square(&f)
	fff.Mul(&ff, &f)
	res.X.Mul(&p.X, &ff)
	res.Y.Mul(&p.Y, &fff)
	res.ZZ.Mul(&p.ZZ, &ff)
	res.ZZZ.Mul(&p.ZZZ, &fff)
	return res
}

const (
	nbFuzzShort = 10
	nbFuzz      = 100
)

// define Gopters generators

// GenFr generates an Fr element
func GenFr() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var elmt fr.Element
		elmt.MustSetRandom()

		return gopter.NewGenResult(elmt, gopter.NoShrinker)
	}
}

// GenFp generates an Fp element
func GenFp() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var elmt fp.Element
		elmt.MustSetRandom()

		return gopter.NewGenResult(elmt, gopter.NoShrinker)
	}
}

// GenBigInt generates a big.Int
func GenBigInt() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var s big.Int
		var b [fp.Bytes]byte
		_, err := crand.Read(b[:]) //#nosec G404 weak rng is fine here
		if err != nil {
			panic(err)
		}
		s.SetBytes(b[:])
		genResult := gopter.NewGenResult(s, gopter.NoShrinker)
		return genResult
	}
}
