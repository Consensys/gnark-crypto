package bls381

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gurvy/bls381/fp"
	"github.com/consensys/gurvy/bls381/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// utils

func fuzzJacobian(p *G1Jac, f *fp.Element) G1Jac {
	var res G1Jac
	res.X.Mul(&p.X, f).Mul(&res.X, f)
	res.Y.Mul(&p.Y, f).Mul(&res.Y, f).Mul(&res.Y, f)
	res.Z.Mul(&p.Z, f)
	return res
}

func fuzzProjective(p *G1Proj, f *fp.Element) G1Proj {
	var res G1Proj
	res.X.Mul(&p.X, f)
	res.Y.Mul(&p.Y, f)
	res.Z.Mul(&p.Z, f)
	return res
}

func fuzzExtendedJacobian(p *g1JacExtended, f *fp.Element) g1JacExtended {
	var res g1JacExtended
	var ff, fff fp.Element
	ff.Square(f)
	fff.Mul(&ff, f)
	res.X.Mul(&p.X, &ff)
	res.Y.Mul(&p.Y, &fff)
	res.ZZ.Mul(&p.ZZ, &ff)
	res.ZZZ.Mul(&p.ZZZ, &fff)
	return res
}

// ------------------------------------------------------------
// tests

func TestG1Conversions(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genFuzz1 := GenFp()
	genFuzz2 := GenFp()

	properties.Property("Affine representation should be independent of the Jacobian representative", prop.ForAll(
		func(u fp.Element) bool {
			g := fuzzJacobian(&g1Gen, &u)
			var g1 G1Affine
			g1.FromJacobian(&g)
			return g1.X.Equal(&g1Gen.X) && g1.Y.Equal(&g1Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("Affine representation should be independent of a Extended Jacobian representative", prop.ForAll(
		func(u fp.Element) bool {
			var g g1JacExtended
			g.X.Set(&g1Gen.X)
			g.Y.Set(&g1Gen.Y)
			g.ZZ.Set(&g1Gen.Z)
			g.ZZZ.Set(&g1Gen.Z)
			gfuzz := fuzzExtendedJacobian(&g, &u)

			var g1 G1Affine
			gfuzz.ToAffine(&g1)
			return g1.X.Equal(&g1Gen.X) && g1.Y.Equal(&g1Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("Projective representation should be independent of a Jacobian representative", prop.ForAll(
		func(u fp.Element) bool {

			g := fuzzJacobian(&g1Gen, &u)

			var g1 G1Proj
			g1.FromJacobian(&g)
			var a, c fp.Element
			a.Mul(&g.X, &g.Z)
			c.Square(&g.Z).Mul(&c, &g.Z)

			return g1.X.Equal(&a) && g1.Y.Equal(&g.Y) && g1.Z.Equal(&c)
		},
		genFuzz1,
	))

	properties.Property("Jacobian representation should be the same as the affine representative", prop.ForAll(
		func(u fp.Element) bool {
			var g G1Jac
			var g1 G1Affine
			g1.X.Set(&g1Gen.X)
			g1.Y.Set(&g1Gen.Y)

			var one fp.Element
			one.SetOne()

			g.FromAffine(&g1)

			return g.X.Equal(&g1Gen.X) && g.Y.Equal(&g1Gen.Y) && g.Z.Equal(&one)
		},
		genFuzz1,
	))

	properties.Property("Converting affine symbol for infinity to Jacobian should output correct infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G1Affine
			g.X.SetZero()
			g.Y.SetZero()
			var g1 G1Jac
			g1.FromAffine(&g)
			var one, zero fp.Element
			one.SetOne()
			return g1.X.Equal(&one) && g1.Y.Equal(&one) && g1.Z.Equal(&zero)
		},
	))

	properties.Property("Converting infinity in extended Jacobian to affine should output infinity symbol in Affine", prop.ForAll(
		func() bool {
			var g G1Affine
			var g1 g1JacExtended
			var zero fp.Element
			g1.X.Set(&g1Gen.X)
			g1.Y.Set(&g1Gen.Y)
			g1.ToAffine(&g)
			return g.X.Equal(&zero) && g.Y.Equal(&zero)
		},
	))

	properties.Property("Converting infinity in extended Jacobian to Jacobian should output infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G1Jac
			var g1 g1JacExtended
			var zero, one fp.Element
			one.SetOne()
			g1.X.Set(&g1Gen.X)
			g1.Y.Set(&g1Gen.Y)
			g1.ToJac(&g)
			return g.X.Equal(&one) && g.Y.Equal(&one) && g.Z.Equal(&zero)
		},
	))

	properties.Property("[Jacobian] Two representatives of the same class should be equal", prop.ForAll(
		func(a, b fp.Element) bool {
			g1 := fuzzJacobian(&g1Gen, &a)
			g2 := fuzzJacobian(&g1Gen, &b)
			return g1.Equal(&g2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG1Ops(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genFuzz1 := GenFp()
	genFuzz2 := GenFp()

	genScalar := GenFr()

	properties.Property("[Jacobian] Add should call double when having adding the same point", prop.ForAll(
		func(a, b fp.Element) bool {
			fg1 := fuzzJacobian(&g1Gen, &a)
			fg2 := fuzzJacobian(&g1Gen, &b)
			var g1, g2 G1Jac
			g1.Set(&fg1).AddAssign(&fg2)
			g2.Double(&fg2)
			return g1.Equal(&g2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[Jacobian] Adding the opposite of a point to itself should output inf", prop.ForAll(
		func(a, b fp.Element) bool {
			g1 := fuzzJacobian(&g1Gen, &a)
			g2 := fuzzJacobian(&g1Gen, &b)
			g2.Neg(&g2)
			g1.AddAssign(&g2)
			return g1.Equal(&g1Infinity)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[Jacobian] Adding the inf to a point should not modify the point", prop.ForAll(
		func(a fp.Element) bool {
			g1 := fuzzJacobian(&g1Gen, &a)
			g1.AddAssign(&g1Infinity)
			var g2 G1Jac
			g2.Set(&g1Infinity)
			g2.AddAssign(&g1Gen)
			return g1.Equal(&g1Gen) && g2.Equal(&g1Gen)
		},
		genFuzz1,
	))

	properties.Property("[Jacobian] Addmix the negation to itself should output 0", prop.ForAll(
		func(a fp.Element) bool {
			g1 := fuzzJacobian(&g1Gen, &a)
			g1.Neg(&g1)
			var g2 G1Affine
			g2.FromJacobian(&g1Gen)
			g1.AddMixed(&g2)
			return g1.Equal(&g1Infinity)
		},
		genFuzz1,
	))

	properties.Property("scalar multiplication (double and add) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.ElementModulus()
			var g G1Jac
			var gaff G1Affine
			gaff.FromJacobian(&g1Gen)
			g.ScalarMultiplication(&gaff, r)

			var scalar, blindedScalard, rminusone big.Int
			var g1, g2, g3, gneg G1Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			g3.ScalarMultiplication(&gaff, &rminusone)
			gneg.Neg(&g1Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			g1.ScalarMultiplication(&gaff, &scalar)
			g2.ScalarMultiplication(&gaff, &blindedScalard)

			return g1.Equal(&g2) && g.Equal(&g1Infinity) && !g1.Equal(&g1Infinity) && gneg.Equal(&g3)

		},
		genScalar,
	))

	properties.Property("scalar multiplication (GLV) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.ElementModulus()
			var g G1Jac
			var gaff G1Affine
			gaff.FromJacobian(&g1Gen)
			g.ScalarMulGLV(&gaff, r)

			var scalar, blindedScalard, rminusone big.Int
			var g1, g2, g3, gneg G1Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			g3.ScalarMulGLV(&gaff, &rminusone)
			gneg.Neg(&g1Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			g1.ScalarMulGLV(&gaff, &scalar)
			g2.ScalarMulGLV(&gaff, &blindedScalard)

			return g1.Equal(&g2) && g.Equal(&g1Infinity) && !g1.Equal(&g1Infinity) && gneg.Equal(&g3)

		},
		genScalar,
	))

	properties.Property("GLV and Double and Add should output the same result", prop.ForAll(
		func(s fr.Element) bool {

			var r big.Int
			var g1, g2 G1Jac
			var gaff G1Affine
			s.ToBigIntRegular(&r)
			gaff.FromJacobian(&g1Gen)
			g1.ScalarMultiplication(&gaff, &r)
			g2.ScalarMulGLV(&gaff, &r)
			return g1.Equal(&g2) && !g1.Equal(&g1Infinity)

		},
		genScalar,
	))

	properties.Property("Multi exponentation (>50points) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var g G1Jac
			g.Set(&g1Gen)

			// mixer ensures that all the words of a fpElement are set
			samplePoints := make([]G1Affine, 3000)
			sampleScalars := make([]fr.Element, 3000)

			for i := 1; i <= 3000; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
				samplePoints[i-1].FromJacobian(&g)
				g.AddAssign(&g1Gen)
			}

			var g1MultiExp G1Jac
			<-g1MultiExp.MultiExp(samplePoints, sampleScalars)

			var finalBigScalar fr.Element
			var finalBigScalarBi big.Int
			var g1ScalarMul G1Jac
			var g1Aff G1Affine
			g1Aff.FromJacobian(&g1Gen)
			finalBigScalar.SetString("9004500500").MulAssign(&mixer)
			finalBigScalar.ToBigIntRegular(&finalBigScalarBi)
			g1ScalarMul.ScalarMultiplication(&g1Aff, &finalBigScalarBi)

			return g1ScalarMul.Equal(&g1MultiExp)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (<50points) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var g G1Jac
			g.Set(&g1Gen)

			// mixer ensures that all the words of a fpElement are set
			samplePoints := make([]G1Affine, 30)
			sampleScalars := make([]fr.Element, 30)

			for i := 1; i <= 30; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
				samplePoints[i-1].FromJacobian(&g)
				g.AddAssign(&g1Gen)
			}

			var g1MultiExp G1Jac
			<-g1MultiExp.MultiExp(samplePoints, sampleScalars)

			var finalBigScalar fr.Element
			var finalBigScalarBi big.Int
			var g1ScalarMul G1Jac
			var g1Aff G1Affine
			g1Aff.FromJacobian(&g1Gen)
			finalBigScalar.SetString("9455").MulAssign(&mixer)
			finalBigScalar.ToBigIntRegular(&finalBigScalarBi)
			g1ScalarMul.ScalarMultiplication(&g1Aff, &finalBigScalarBi)

			return g1ScalarMul.Equal(&g1MultiExp)
		},
		genScalar,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkGLV(b *testing.B) {
	var g G1Affine
	g.FromJacobian(&g1Gen)
	var g1 G1Jac
	var s big.Int
	s.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g1.ScalarMulGLV(&g, &s)
	}

}

func BenchmarkDoubleAndAdd(b *testing.B) {

	var g G1Affine
	g.FromJacobian(&g1Gen)

	var g1 G1Jac
	var s big.Int
	s.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g1.ScalarMultiplication(&g, &s)
	}

}

func BenchmarkG1Add(b *testing.B) {
	var a G1Jac
	a.Double(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddAssign(&g1Gen)
	}
}

func BenchmarkG1AddMixed(b *testing.B) {
	var a G1Jac
	a.Double(&g1Gen)

	var c G1Affine
	c.FromJacobian(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddMixed(&c)
	}

}

func BenchmarkG1Double(b *testing.B) {
	var a G1Jac
	a.Set(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.DoubleAssign()
	}

}

func BenchmarkMultiExpG1(b *testing.B) {

	var G G1Jac

	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	var nbSamples int
	nbSamples = 800000

	samplePoints := make([]G1Affine, nbSamples)
	sampleScalars := make([]fr.Element, nbSamples)

	G.Set(&g1Gen)

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
		samplePoints[i-1].FromJacobian(&G)
	}

	var testPoint G1Jac

	for i := 0; i < 16; i++ {

		b.Run(fmt.Sprintf("%d points)", (i+1)*50000), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				<-testPoint.MultiExp(samplePoints[:50000+i*50000], sampleScalars[:50000+i*50000])
			}
		})
	}
}
