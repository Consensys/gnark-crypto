package point

const PointTests = `

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gurvy/{{ toLower .CurveName}}/fp"
	"github.com/consensys/gurvy/{{ toLower .CurveName}}/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// utils

func fuzzJacobian(p *{{ toUpper .PointName }}Jac, f *{{.CoordType}}) {{ toUpper .PointName }}Jac {
	var res {{ toUpper .PointName }}Jac
	res.X.Mul(&p.X, f).Mul(&res.X, f)
	res.Y.Mul(&p.Y, f).Mul(&res.Y, f).Mul(&res.Y, f)
	res.Z.Mul(&p.Z, f)
	return res
}

func fuzzProjective(p *{{ toUpper .PointName }}Proj, f *{{.CoordType}}) {{ toUpper .PointName }}Proj {
	var res {{ toUpper .PointName }}Proj
	res.X.Mul(&p.X, f)
	res.Y.Mul(&p.Y, f)
	res.Z.Mul(&p.Z, f)
	return res
}

func fuzzExtendedJacobian(p *{{ toLower .PointName }}JacExtended, f *{{.CoordType}}) {{ toLower .PointName }}JacExtended {
	var res {{ toLower .PointName }}JacExtended
	var ff, fff {{.CoordType}}
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

func Test{{ toUpper .PointName }}Conversions(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genFuzz1 := GenFp()
	genFuzz2 := GenFp()

	properties.Property("Affine representation should be independent of the Jacobian representative", prop.ForAll(
		func(u {{.CoordType}}) bool {
			g := fuzzJacobian(&{{ toLower .PointName }}Gen, &u)
			var {{ toLower .PointName }} {{ toUpper .PointName }}Affine
			{{ toLower .PointName }}.FromJacobian(&g)
			return {{ toLower .PointName }}.X.Equal(&{{ toLower .PointName }}Gen.X) && {{ toLower .PointName }}.Y.Equal(&{{ toLower .PointName }}Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("Affine representation should be independent of a Extended Jacobian representative", prop.ForAll(
		func(u {{.CoordType}}) bool {
			var g {{ toLower .PointName }}JacExtended
			g.X.Set(&{{ toLower .PointName }}Gen.X)
			g.Y.Set(&{{ toLower .PointName }}Gen.Y)
			g.ZZ.Set(&{{ toLower .PointName }}Gen.Z)
			g.ZZZ.Set(&{{ toLower .PointName }}Gen.Z)
			gfuzz := fuzzExtendedJacobian(&g, &u)

			var {{ toLower .PointName }} {{ toUpper .PointName }}Affine
			gfuzz.ToAffine(&{{ toLower .PointName }})
			return {{ toLower .PointName }}.X.Equal(&{{ toLower .PointName }}Gen.X) && {{ toLower .PointName }}.Y.Equal(&{{ toLower .PointName }}Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("Projective representation should be independent of a Jacobian representative", prop.ForAll(
		func(u {{.CoordType}}) bool {

			g := fuzzJacobian(&{{ toLower .PointName }}Gen, &u)

			var {{ toLower .PointName }} {{ toUpper .PointName }}Proj
			{{ toLower .PointName }}.FromJacobian(&g)
			var a, c {{.CoordType}}
			a.Mul(&g.X, &g.Z)
			c.Square(&g.Z).Mul(&c, &g.Z)

			return {{ toLower .PointName }}.X.Equal(&a) && {{ toLower .PointName }}.Y.Equal(&g.Y) && {{ toLower .PointName }}.Z.Equal(&c)
		},
		genFuzz1,
	))

	properties.Property("Jacobian representation should be the same as the affine representative", prop.ForAll(
		func(u {{.CoordType}}) bool {
			var g {{ toUpper .PointName }}Jac
			var {{ toLower .PointName }} {{ toUpper .PointName }}Affine
			{{ toLower .PointName }}.X.Set(&{{ toLower .PointName }}Gen.X)
			{{ toLower .PointName }}.Y.Set(&{{ toLower .PointName }}Gen.Y)

			var one {{.CoordType}}
			one.SetOne()

			g.FromAffine(&{{ toLower .PointName }})

			return g.X.Equal(&{{ toLower .PointName }}Gen.X) && g.Y.Equal(&{{ toLower .PointName }}Gen.Y) && g.Z.Equal(&one)
		},
		genFuzz1,
	))

	properties.Property("Converting affine symbol for infinity to Jacobian should output correct infinity in Jacobian", prop.ForAll(
		func() bool {
			var g {{ toUpper .PointName }}Affine
			g.X.SetZero()
			g.Y.SetZero()
			var {{ toLower .PointName }} {{ toUpper .PointName }}Jac
			{{ toLower .PointName }}.FromAffine(&g)
			var one, zero {{.CoordType}}
			one.SetOne()
			return {{ toLower .PointName }}.X.Equal(&one) && {{ toLower .PointName }}.Y.Equal(&one) && {{ toLower .PointName }}.Z.Equal(&zero)
		},
	))

	properties.Property("Converting infinity in extended Jacobian to affine should output infinity symbol in Affine", prop.ForAll(
		func() bool {
			var g {{ toUpper .PointName }}Affine
			var {{ toLower .PointName }} {{ toLower .PointName }}JacExtended
			var zero {{.CoordType}}
			{{ toLower .PointName }}.X.Set(&{{ toLower .PointName }}Gen.X)
			{{ toLower .PointName }}.Y.Set(&{{ toLower .PointName }}Gen.Y)
			{{ toLower .PointName }}.ToAffine(&g)
			return g.X.Equal(&zero) && g.Y.Equal(&zero)
		},
	))

	properties.Property("Converting infinity in extended Jacobian to Jacobian should output infinity in Jacobian", prop.ForAll(
		func() bool {
			var g {{ toUpper .PointName }}Jac
			var {{ toLower .PointName }} {{ toLower .PointName }}JacExtended
			var zero, one {{.CoordType}}
			one.SetOne()
			{{ toLower .PointName }}.X.Set(&{{ toLower .PointName }}Gen.X)
			{{ toLower .PointName }}.Y.Set(&{{ toLower .PointName }}Gen.Y)
			{{ toLower .PointName }}.ToJac(&g)
			return g.X.Equal(&one) && g.Y.Equal(&one) && g.Z.Equal(&zero)
		},
	))

	properties.Property("[Jacobian] Two representatives of the same class should be equal", prop.ForAll(
		func(a, b {{.CoordType}}) bool {
			{{ toLower .PointName }} := fuzzJacobian(&{{ toLower .PointName }}Gen, &a)
			g2 := fuzzJacobian(&{{ toLower .PointName }}Gen, &b)
			return {{ toLower .PointName }}.Equal(&g2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func Test{{ toUpper .PointName }}Ops(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genFuzz1 := GenFp()
	genFuzz2 := GenFp()

	genScalar := GenFr()

	properties.Property("[Jacobian] Add should call double when having adding the same point", prop.ForAll(
		func(a, b {{.CoordType}}) bool {
			f{{ toLower .PointName }} := fuzzJacobian(&{{ toLower .PointName }}Gen, &a)
			fg2 := fuzzJacobian(&{{ toLower .PointName }}Gen, &b)
			var {{ toLower .PointName }}, g2 {{ toUpper .PointName }}Jac
			{{ toLower .PointName }}.Set(&f{{ toLower .PointName }}).AddAssign(&fg2)
			g2.Double(&fg2)
			return {{ toLower .PointName }}.Equal(&g2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[Jacobian] Adding the opposite of a point to itself should output inf", prop.ForAll(
		func(a, b {{.CoordType}}) bool {
			{{ toLower .PointName }} := fuzzJacobian(&{{ toLower .PointName }}Gen, &a)
			g2 := fuzzJacobian(&{{ toLower .PointName }}Gen, &b)
			g2.Neg(&g2)
			{{ toLower .PointName }}.AddAssign(&g2)
			return {{ toLower .PointName }}.Equal(&{{ toLower .PointName }}Infinity)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[Jacobian] Adding the inf to a point should not modify the point", prop.ForAll(
		func(a {{.CoordType}}) bool {
			{{ toLower .PointName }} := fuzzJacobian(&{{ toLower .PointName }}Gen, &a)
			{{ toLower .PointName }}.AddAssign(&{{ toLower .PointName }}Infinity)
			var g2 {{ toUpper .PointName }}Jac
			g2.Set(&{{ toLower .PointName }}Infinity)
			g2.AddAssign(&{{ toLower .PointName }}Gen)
			return {{ toLower .PointName }}.Equal(&{{ toLower .PointName }}Gen) && g2.Equal(&{{ toLower .PointName }}Gen)
		},
		genFuzz1,
	))

	properties.Property("[Jacobian] Addmix the negation to itself should output 0", prop.ForAll(
		func(a {{.CoordType}}) bool {
			{{ toLower .PointName }} := fuzzJacobian(&{{ toLower .PointName }}Gen, &a)
			{{ toLower .PointName }}.Neg(&{{ toLower .PointName }})
			var g2 {{ toUpper .PointName }}Affine
			g2.FromJacobian(&{{ toLower .PointName }}Gen)
			{{ toLower .PointName }}.AddMixed(&g2)
			return {{ toLower .PointName }}.Equal(&{{ toLower .PointName }}Infinity)
		},
		genFuzz1,
	))

	properties.Property("scalar multiplication (double and add) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.ElementModulus()
			var g {{ toUpper .PointName }}Jac
			var gaff {{ toUpper .PointName }}Affine
			gaff.FromJacobian(&{{ toLower .PointName }}Gen)
			g.ScalarMultiplication(&gaff, r)

			var scalar, blindedScalard, rminusone big.Int
			var {{ toLower .PointName }}, g2, g3, gneg {{ toUpper .PointName }}Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			g3.ScalarMultiplication(&gaff, &rminusone)
			gneg.Neg(&{{ toLower .PointName }}Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			{{ toLower .PointName }}.ScalarMultiplication(&gaff, &scalar)
			g2.ScalarMultiplication(&gaff, &blindedScalard)

			return {{ toLower .PointName }}.Equal(&g2) && g.Equal(&{{ toLower .PointName }}Infinity) && !{{ toLower .PointName }}.Equal(&{{ toLower .PointName }}Infinity) && gneg.Equal(&g3)

		},
		genScalar,
	))

	properties.Property("scalar multiplication (GLV) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.ElementModulus()
			var g {{ toUpper .PointName }}Jac
			var gaff {{ toUpper .PointName }}Affine
			gaff.FromJacobian(&{{ toLower .PointName }}Gen)
			g.ScalarMulGLV(&gaff, r)

			var scalar, blindedScalard, rminusone big.Int
			var {{ toLower .PointName }}, g2, g3, gneg {{ toUpper .PointName }}Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			g3.ScalarMulGLV(&gaff, &rminusone)
			gneg.Neg(&{{ toLower .PointName }}Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			{{ toLower .PointName }}.ScalarMulGLV(&gaff, &scalar)
			g2.ScalarMulGLV(&gaff, &blindedScalard)

			return {{ toLower .PointName }}.Equal(&g2) && g.Equal(&{{ toLower .PointName }}Infinity) && !{{ toLower .PointName }}.Equal(&{{ toLower .PointName }}Infinity) && gneg.Equal(&g3)

		},
		genScalar,
	))

	properties.Property("GLV and Double and Add should output the same result", prop.ForAll(
		func(s fr.Element) bool {

			var r big.Int
			var {{ toLower .PointName }}, g2 {{ toUpper .PointName }}Jac
			var gaff {{ toUpper .PointName }}Affine
			s.ToBigIntRegular(&r)
			gaff.FromJacobian(&{{ toLower .PointName }}Gen)
			{{ toLower .PointName }}.ScalarMultiplication(&gaff, &r)
			g2.ScalarMulGLV(&gaff, &r)
			return {{ toLower .PointName }}.Equal(&g2) && !{{ toLower .PointName }}.Equal(&{{ toLower .PointName }}Infinity)

		},
		genScalar,
	))

	properties.Property("Multi exponentation (>50points) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var g {{ toUpper .PointName }}Jac
			g.Set(&{{ toLower .PointName }}Gen)

			// mixer ensures that all the words of a fpElement are set
			samplePoints := make([]{{ toUpper .PointName }}Affine, 3000)
			sampleScalars := make([]fr.Element, 3000)

			for i := 1; i <= 3000; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
				samplePoints[i-1].FromJacobian(&g)
				g.AddAssign(&{{ toLower .PointName }}Gen)
			}

			var {{ toLower .PointName }}MultiExp {{ toUpper .PointName }}Jac
			<-{{ toLower .PointName }}MultiExp.MultiExp(samplePoints, sampleScalars)

			var finalBigScalar fr.Element
			var finalBigScalarBi big.Int
			var {{ toLower .PointName }}ScalarMul {{ toUpper .PointName }}Jac
			var {{ toLower .PointName }}Aff {{ toUpper .PointName }}Affine
			{{ toLower .PointName }}Aff.FromJacobian(&{{ toLower .PointName }}Gen)
			finalBigScalar.SetString("9004500500").MulAssign(&mixer)
			finalBigScalar.ToBigIntRegular(&finalBigScalarBi)
			{{ toLower .PointName }}ScalarMul.ScalarMultiplication(&{{ toLower .PointName }}Aff, &finalBigScalarBi)

			return {{ toLower .PointName }}ScalarMul.Equal(&{{ toLower .PointName }}MultiExp)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (<50points) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var g {{ toUpper .PointName }}Jac
			g.Set(&{{ toLower .PointName }}Gen)

			// mixer ensures that all the words of a fpElement are set
			samplePoints := make([]{{ toUpper .PointName }}Affine, 30)
			sampleScalars := make([]fr.Element, 30)

			for i := 1; i <= 30; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
				samplePoints[i-1].FromJacobian(&g)
				g.AddAssign(&{{ toLower .PointName }}Gen)
			}

			var {{ toLower .PointName }}MultiExp {{ toUpper .PointName }}Jac
			<-{{ toLower .PointName }}MultiExp.MultiExp(samplePoints, sampleScalars)

			var finalBigScalar fr.Element
			var finalBigScalarBi big.Int
			var {{ toLower .PointName }}ScalarMul {{ toUpper .PointName }}Jac
			var {{ toLower .PointName }}Aff {{ toUpper .PointName }}Affine
			{{ toLower .PointName }}Aff.FromJacobian(&{{ toLower .PointName }}Gen)
			finalBigScalar.SetString("9455").MulAssign(&mixer)
			finalBigScalar.ToBigIntRegular(&finalBigScalarBi)
			{{ toLower .PointName }}ScalarMul.ScalarMultiplication(&{{ toLower .PointName }}Aff, &finalBigScalarBi)

			return {{ toLower .PointName }}ScalarMul.Equal(&{{ toLower .PointName }}MultiExp)
		},
		genScalar,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkGLV(b *testing.B) {
	var g {{ toUpper .PointName }}Affine
	g.FromJacobian(&{{ toLower .PointName }}Gen)
	var {{ toLower .PointName }} {{ toUpper .PointName }}Jac
	var s big.Int
	s.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		{{ toLower .PointName }}.ScalarMulGLV(&g, &s)
	}

}

func BenchmarkDoubleAndAdd(b *testing.B) {

	var g {{ toUpper .PointName }}Affine
	g.FromJacobian(&{{ toLower .PointName }}Gen)

	var {{ toLower .PointName }} {{ toUpper .PointName }}Jac
	var s big.Int
	s.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		{{ toLower .PointName }}.ScalarMultiplication(&g, &s)
	}

}

func Benchmark{{ toUpper .PointName }}Add(b *testing.B) {
	var a {{ toUpper .PointName }}Jac
	a.Double(&{{ toLower .PointName }}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddAssign(&{{ toLower .PointName }}Gen)
	}
}

func Benchmark{{ toUpper .PointName }}AddMixed(b *testing.B) {
	var a {{ toUpper .PointName }}Jac
	a.Double(&{{ toLower .PointName }}Gen)

	var c {{ toUpper .PointName }}Affine
	c.FromJacobian(&{{ toLower .PointName }}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddMixed(&c)
	}

}

func Benchmark{{ toUpper .PointName }}Double(b *testing.B) {
	var a {{ toUpper .PointName }}Jac
	a.Set(&{{ toLower .PointName }}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.DoubleAssign()
	}

}

func BenchmarkMultiExp{{ toUpper .PointName }}(b *testing.B) {

	var G {{ toUpper .PointName }}Jac

	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	var nbSamples int
	nbSamples = 800000

	samplePoints := make([]{{ toUpper .PointName }}Affine, nbSamples)
	sampleScalars := make([]fr.Element, nbSamples)

	G.Set(&{{ toLower .PointName }}Gen)

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
		samplePoints[i-1].FromJacobian(&G)
	}

	var testPoint {{ toUpper .PointName }}Jac

	for i := 0; i < 16; i++ {

		b.Run(fmt.Sprintf("%d points)", (i+1)*50000), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				<-testPoint.MultiExp(samplePoints[:50000+i*50000], sampleScalars[:50000+i*50000])
			}
		})
	}
}
`
