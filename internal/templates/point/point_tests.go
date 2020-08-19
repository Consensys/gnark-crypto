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

{{- if eq .CoordType "fp.Element" }}
	func fuzzJacobian{{ toUpper .PointName}}(p *{{ toUpper .PointName}}Jac, f {{ .CoordType}}) {{ toUpper .PointName}}Jac {
		var res {{ toUpper .PointName}}Jac
		res.X.Mul(&p.X, &f).Mul(&res.X, &f)
		res.Y.Mul(&p.Y, &f).Mul(&res.Y, &f).Mul(&res.Y, &f)
		res.Z.Mul(&p.Z, &f)
		return res
	}

	func fuzzProjective{{ toUpper .PointName}}(p *{{ toUpper .PointName}}Proj, f {{ .CoordType}}) {{ toUpper .PointName}}Proj {
		var res {{ toUpper .PointName}}Proj
		res.X.Mul(&p.X, &f)
		res.Y.Mul(&p.Y, &f)
		res.Z.Mul(&p.Z, &f)
		return res
	}

	func fuzzExtendedJacobian{{ toUpper .PointName}}(p *{{ toLower .PointName }}JacExtended, f {{ .CoordType}}) {{ toLower .PointName }}JacExtended {
		var res {{ toLower .PointName }}JacExtended
		var ff, fff {{ .CoordType}}
		ff.Square(&f)
		fff.Mul(&ff, &f)
		res.X.Mul(&p.X, &ff)
		res.Y.Mul(&p.Y, &fff)
		res.ZZ.Mul(&p.ZZ, &ff)
		res.ZZZ.Mul(&p.ZZZ, &fff)
		return res
	}
{{- else if eq .CoordType "E2" }}
	func fuzzJacobian{{ toUpper .PointName}}(p *{{ toUpper .PointName}}Jac, f *E2) {{ toUpper .PointName}}Jac {
		var res {{ toUpper .PointName}}Jac
		res.X.Mul(&p.X, f).Mul(&res.X, f)
		res.Y.Mul(&p.Y, f).Mul(&res.Y, f).Mul(&res.Y, f)
		res.Z.Mul(&p.Z, f)
		return res
	}

	func fuzzProjective{{ toUpper .PointName}}(p *{{ toUpper .PointName}}Proj, f *E2) {{ toUpper .PointName}}Proj {
		var res {{ toUpper .PointName}}Proj
		res.X.Mul(&p.X, f)
		res.Y.Mul(&p.Y, f)
		res.Z.Mul(&p.Z, f)
		return res
	}

	func fuzzExtendedJacobian{{ toUpper .PointName}}(p *{{ toLower .PointName }}JacExtended, f *E2) {{ toLower .PointName }}JacExtended {
		var res {{ toLower .PointName }}JacExtended
		var ff, fff {{ .CoordType}}
		ff.Square(f)
		fff.Mul(&ff, f)
		res.X.Mul(&p.X, &ff)
		res.Y.Mul(&p.Y, &fff)
		res.ZZ.Mul(&p.ZZ, &ff)
		res.ZZZ.Mul(&p.ZZZ, &fff)
		return res
	}
{{- end}}

// ------------------------------------------------------------
// tests

func Test{{ toUpper .PointName}}Conversions(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)
	{{- if eq .CoordType "fp.Element" }}
		genFuzz1 := GenFp()
		genFuzz2 := GenFp()
	{{- else if eq .CoordType "E2" }}
		genFuzz1 := GenE2()
		genFuzz2 := GenE2()
	{{- end}}

	properties.Property("Affine representation should be independent of the Jacobian representative", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a *E2) bool {
		{{- end}}
			g := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)
			var op1 {{ toUpper .PointName}}Affine
			op1.FromJacobian(&g)
			return op1.X.Equal(&{{ toLower .PointName }}Gen.X) && op1.Y.Equal(&{{ toLower .PointName }}Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("Affine representation should be independent of a Extended Jacobian representative", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a *E2) bool {
		{{- end}}
			var g {{ toLower .PointName }}JacExtended
			g.X.Set(&{{ toLower .PointName }}Gen.X)
			g.Y.Set(&{{ toLower .PointName }}Gen.Y)
			g.ZZ.Set(&{{ toLower .PointName }}Gen.Z)
			g.ZZZ.Set(&{{ toLower .PointName }}Gen.Z)
			gfuzz := fuzzExtendedJacobian{{ toUpper .PointName}}(&g, a)

			var op1 {{ toUpper .PointName}}Affine
			gfuzz.ToAffine(&op1)
			return op1.X.Equal(&{{ toLower .PointName }}Gen.X) && op1.Y.Equal(&{{ toLower .PointName }}Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("Projective representation should be independent of a Jacobian representative", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a *E2) bool {
		{{- end}}

			g := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)

			var op1 {{ toUpper .PointName}}Proj
			op1.FromJacobian(&g)
			var u, v {{ .CoordType}}
			u.Mul(&g.X, &g.Z)
			v.Square(&g.Z).Mul(&v, &g.Z)

			return op1.X.Equal(&u) && op1.Y.Equal(&g.Y) && op1.Z.Equal(&v)
		},
		genFuzz1,
	))

	properties.Property("Jacobian representation should be the same as the affine representative", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a *E2) bool {
		{{- end}}
			var g {{ toUpper .PointName}}Jac
			var op1 {{ toUpper .PointName}}Affine
			op1.X.Set(&{{ toLower .PointName }}Gen.X)
			op1.Y.Set(&{{ toLower .PointName }}Gen.Y)

			var one {{ .CoordType}}
			one.SetOne()

			g.FromAffine(&op1)

			return g.X.Equal(&{{ toLower .PointName }}Gen.X) && g.Y.Equal(&{{ toLower .PointName }}Gen.Y) && g.Z.Equal(&one)
		},
		genFuzz1,
	))

	properties.Property("Converting affine symbol for infinity to Jacobian should output correct infinity in Jacobian", prop.ForAll(
		func() bool {
			var g {{ toUpper .PointName}}Affine
			g.X.SetZero()
			g.Y.SetZero()
			var op1 {{ toUpper .PointName}}Jac
			op1.FromAffine(&g)
			var one, zero {{ .CoordType}}
			one.SetOne()
			return op1.X.Equal(&one) && op1.Y.Equal(&one) && op1.Z.Equal(&zero)
		},
	))

	properties.Property("Converting infinity in extended Jacobian to affine should output infinity symbol in Affine", prop.ForAll(
		func() bool {
			var g {{ toUpper .PointName}}Affine
			var op1 {{ toLower .PointName }}JacExtended
			var zero {{ .CoordType}}
			op1.X.Set(&{{ toLower .PointName }}Gen.X)
			op1.Y.Set(&{{ toLower .PointName }}Gen.Y)
			op1.ToAffine(&g)
			return g.X.Equal(&zero) && g.Y.Equal(&zero)
		},
	))

	properties.Property("Converting infinity in extended Jacobian to Jacobian should output infinity in Jacobian", prop.ForAll(
		func() bool {
			var g {{ toUpper .PointName}}Jac
			var op1 {{ toLower .PointName }}JacExtended
			var zero, one {{ .CoordType}}
			one.SetOne()
			op1.X.Set(&{{ toLower .PointName }}Gen.X)
			op1.Y.Set(&{{ toLower .PointName }}Gen.Y)
			op1.ToJac(&g)
			return g.X.Equal(&one) && g.Y.Equal(&one) && g.Z.Equal(&zero)
		},
	))

	properties.Property("[Jacobian] Two representatives of the same class should be equal", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a, b {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a, b *E2) bool {
		{{- end}}
			op1 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)
			op2 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, b)
			return op1.Equal(&op2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func Test{{ toUpper .PointName}}Ops(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)
	{{- if eq .CoordType "fp.Element" }}
		genFuzz1 := GenFp()
		genFuzz2 := GenFp()
	{{- else if eq .CoordType "E2" }}
		genFuzz1 := GenE2()
		genFuzz2 := GenE2()
	{{- end}}

	genScalar := GenFr()

	properties.Property("[Jacobian] Add should call double when having adding the same point", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a, b {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a, b *E2) bool {
		{{- end}}
			fop1 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)
			fop2 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, b)
			var op1, op2 {{ toUpper .PointName}}Jac
			op1.Set(&fop1).AddAssign(&fop2)
			op2.Double(&fop2)
			return op1.Equal(&op2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[Jacobian] Adding the opposite of a point to itself should output inf", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a, b {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a, b *E2) bool {
		{{- end}}
			fop1 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)
			fop2 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, b)
			fop2.Neg(&fop2)
			fop1.AddAssign(&fop2)
			return fop1.Equal(&{{ toLower .PointName }}Infinity)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[Jacobian] Adding the inf to a point should not modify the point", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a *E2) bool {
		{{- end}}
			fop1 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)
			fop1.AddAssign(&{{ toLower .PointName }}Infinity)
			var op2 {{ toUpper .PointName}}Jac
			op2.Set(&{{ toLower .PointName }}Infinity)
			op2.AddAssign(&{{ toLower .PointName }}Gen)
			return fop1.Equal(&{{ toLower .PointName }}Gen) && op2.Equal(&{{ toLower .PointName }}Gen)
		},
		genFuzz1,
	))

	properties.Property("[Jacobian Extended] mAdd (-G) should equal mSub(G)", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a *E2) bool {
		{{- end}}
			fop1 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)
			var p1,p1Neg {{ toUpper .PointName}}Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 {{ toLower .PointName}}JacExtended 
			o1.mAdd(&p1Neg)
			o2.mSub(&p1)

			return 	o1.X.Equal(&o2.X) && 
					o1.Y.Equal(&o2.Y) && 
					o1.ZZ.Equal(&o2.ZZ) && 
					o1.ZZZ.Equal(&o2.ZZZ) 
		},
		genFuzz1,
	))

	properties.Property("[Jacobian Extended] double (-G) should equal doubleNeg(G)", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a *E2) bool {
		{{- end}}
			fop1 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)
			var p1,p1Neg {{ toUpper .PointName}}Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 {{ toLower .PointName}}JacExtended 
			o1.double(&p1Neg)
			o2.doubleNeg(&p1)

			return 	o1.X.Equal(&o2.X) && 
					o1.Y.Equal(&o2.Y) && 
					o1.ZZ.Equal(&o2.ZZ) && 
					o1.ZZZ.Equal(&o2.ZZZ) 
		},
		genFuzz1,
	))

	properties.Property("[Jacobian] Addmix the negation to itself should output 0", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "E2" }}
			func(a *E2) bool {
		{{- end}}
			fop1 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)
			fop1.Neg(&fop1)
			var op2 {{ toUpper .PointName}}Affine
			op2.FromJacobian(&{{ toLower .PointName }}Gen)
			fop1.AddMixed(&op2)
			return fop1.Equal(&{{ toLower .PointName }}Infinity)
		},
		genFuzz1,
	))

	properties.Property("scalar multiplication (double and add) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g {{ toUpper .PointName}}Jac
			var gaff {{ toUpper .PointName}}Affine
			gaff.FromJacobian(&{{ toLower .PointName }}Gen)
			g.ScalarMultiplication(&gaff, r)

			var scalar, blindedScalard, rminusone big.Int
			var op1, op2, op3, gneg {{ toUpper .PointName}}Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.ScalarMultiplication(&gaff, &rminusone)
			gneg.Neg(&{{ toLower .PointName }}Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			op1.ScalarMultiplication(&gaff, &scalar)
			op2.ScalarMultiplication(&gaff, &blindedScalard)

			return op1.Equal(&op2) && g.Equal(&{{ toLower .PointName }}Infinity) && !op1.Equal(&{{ toLower .PointName }}Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	properties.Property("scalar multiplication (GLV) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g {{ toUpper .PointName}}Jac
			var gaff {{ toUpper .PointName}}Affine
			gaff.FromJacobian(&{{ toLower .PointName }}Gen)
			g.ScalarMulGLV(&gaff, r)

			var scalar, blindedScalard, rminusone big.Int
			var op1, op2, op3, gneg {{ toUpper .PointName}}Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.ScalarMulGLV(&gaff, &rminusone)
			gneg.Neg(&{{ toLower .PointName }}Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			op1.ScalarMulGLV(&gaff, &scalar)
			op2.ScalarMulGLV(&gaff, &blindedScalard)

			return op1.Equal(&op2) && g.Equal(&{{ toLower .PointName }}Infinity) && !op1.Equal(&{{ toLower .PointName }}Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	properties.Property("GLV and Double and Add should output the same result", prop.ForAll(
		func(s fr.Element) bool {

			var r big.Int
			var op1, op2 {{ toUpper .PointName}}Jac
			var gaff {{ toUpper .PointName}}Affine
			s.ToBigIntRegular(&r)
			gaff.FromJacobian(&{{ toLower .PointName }}Gen)
			op1.ScalarMultiplication(&gaff, &r)
			op2.ScalarMulGLV(&gaff, &r)
			return op1.Equal(&op2) && !op1.Equal(&{{ toLower .PointName }}Infinity)

		},
		genScalar,
	))

	

	{{ template "test_multiexp" dict "all" . "C" "4"}}
	{{ template "test_multiexp" dict "all" . "C" "8"}}
	{{ template "test_multiexp" dict "all" . "C" "11"}}
	{{ template "test_multiexp" dict "all" . "C" "14"}}
	{{ template "test_multiexp" dict "all" . "C" "16"}}

	{{define "test_multiexp"}}
		properties.Property("Multi exponentation (c={{.C}}) should be consistant with sum of square", prop.ForAll(
			func(mixer fr.Element) bool {

				const nbSamples = 3000
				var g {{ toUpper .all.PointName}}Jac
				g.Set(&{{ toLower .all.PointName }}Gen)

				// mixer ensures that all the words of a fpElement are set
				var samplePoints [nbSamples]{{ toUpper .all.PointName}}Affine
				var sampleScalars [nbSamples]fr.Element

				for i := 1; i <= nbSamples; i++ {
					sampleScalars[i-1].SetUint64(uint64(i)).
						MulAssign(&mixer).
						FromMont()
					samplePoints[i-1].FromJacobian(&g)
					g.AddAssign(&{{ toLower .all.PointName }}Gen)
				}

				// compare multiExp with double and add
				var result, expected {{ toUpper .all.PointName}}Jac
				<-result.multiExpc{{.C}}(samplePoints[:], sampleScalars[:])
				var finalBigScalar fr.Element
				var finalBigScalarBi big.Int

				// TODO make this a function of nbSamples so that we can reduce test time..
				finalBigScalar.SetString("9004500500").MulAssign(&mixer)
				finalBigScalar.ToBigIntRegular(&finalBigScalarBi)
				expected.ScalarMultiplication(&{{ toLower .all.PointName }}GenAff, &finalBigScalarBi)

				return result.Equal(&expected)
			},
			genScalar,
		))
	{{end}}



	properties.Property("Multi exponentation (<50points) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var g {{ toUpper .PointName}}Jac
			g.Set(&{{ toLower .PointName }}Gen)

			// mixer ensures that all the words of a fpElement are set
			samplePoints := make([]{{ toUpper .PointName}}Affine, 30)
			sampleScalars := make([]fr.Element, 30)

			for i := 1; i <= 30; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
				samplePoints[i-1].FromJacobian(&g)
				g.AddAssign(&{{ toLower .PointName }}Gen)
			}

			var op1MultiExp {{ toUpper .PointName}}Jac
			<-op1MultiExp.MultiExp(samplePoints, sampleScalars)

			var finalBigScalar fr.Element
			var finalBigScalarBi big.Int
			var op1ScalarMul {{ toUpper .PointName}}Jac
			var op1Aff {{ toUpper .PointName}}Affine
			op1Aff.FromJacobian(&{{ toLower .PointName }}Gen)
			finalBigScalar.SetString("9455").MulAssign(&mixer)
			finalBigScalar.ToBigIntRegular(&finalBigScalarBi)
			op1ScalarMul.ScalarMultiplication(&op1Aff, &finalBigScalarBi)

			return op1ScalarMul.Equal(&op1MultiExp)
		},
		genScalar,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func Benchmark{{ toUpper .PointName}}GLV(b *testing.B) {
	var g {{ toUpper .PointName}}Affine
	g.FromJacobian(&{{ toLower .PointName }}Gen)
	var op1 {{ toUpper .PointName}}Jac
	var s big.Int
	s.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op1.ScalarMulGLV(&g, &s)
	}

}



func Benchmark{{ toUpper .PointName}}DoubleAndAdd(b *testing.B) {

	var g {{ toUpper .PointName}}Affine
	g.FromJacobian(&{{ toLower .PointName }}Gen)

	var op1 {{ toUpper .PointName}}Jac
	var s big.Int
	s.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op1.ScalarMultiplication(&g, &s)
	}

}

func Benchmark{{ toUpper .PointName}}Add(b *testing.B) {
	var a {{ toUpper .PointName}}Jac
	a.Double(&{{ toLower .PointName }}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddAssign(&{{ toLower .PointName }}Gen)
	}
}

func Benchmark{{ toUpper .PointName}}mAdd(b *testing.B) {
	var a {{ toLower .PointName}}JacExtended
	a.double(&{{ toLower .PointName }}GenAff)

	var c {{ toUpper .PointName}}Affine
	c.FromJacobian(&{{ toLower .PointName }}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.mAdd(&c)
	}

}

func Benchmark{{ toUpper .PointName}}AddMixed(b *testing.B) {
	var a {{ toUpper .PointName}}Jac
	a.Double(&{{ toLower .PointName }}Gen)

	var c {{ toUpper .PointName}}Affine
	c.FromJacobian(&{{ toLower .PointName }}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddMixed(&c)
	}

}

func Benchmark{{ toUpper .PointName}}Double(b *testing.B) {
	var a {{ toUpper .PointName}}Jac
	a.Set(&{{ toLower .PointName }}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.DoubleAssign()
	}

}


func Benchmark{{ toUpper .PointName}}MultiExp{{ toUpper .PointName}}(b *testing.B) {
	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	const pow = 24
	const nbSamples = 1 << pow

	var samplePoints [nbSamples]{{ toUpper .PointName}}Affine
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
		samplePoints[i-1]= {{ toLower .PointName }}GenAff
	}

	var testPoint {{ toUpper .PointName}}Jac


	
	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points",using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				<-testPoint.MultiExp(samplePoints[:using], sampleScalars[:using])
			}
		})
	}
}



`
