package point

// PointTests ...
const PointTests = `

import (
	"fmt"
	"math/big"
	"math/bits"
	"runtime"
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
{{- else if eq .CoordType "e2" }}
	func fuzzJacobian{{ toUpper .PointName}}(p *{{ toUpper .PointName}}Jac, f *e2) {{ toUpper .PointName}}Jac {
		var res {{ toUpper .PointName}}Jac
		res.X.Mul(&p.X, f).Mul(&res.X, f)
		res.Y.Mul(&p.Y, f).Mul(&res.Y, f).Mul(&res.Y, f)
		res.Z.Mul(&p.Z, f)
		return res
	}

	func fuzzExtendedJacobian{{ toUpper .PointName}}(p *{{ toLower .PointName }}JacExtended, f *e2) {{ toLower .PointName }}JacExtended {
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

func Test{{ toUpper .PointName}}IsOnCurve(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)
	{{- if eq .CoordType "fp.Element" }}
		genFuzz1 := GenFp()
	{{- else if eq .CoordType "e2" }}
		genFuzz1 := GenE2()
	{{- end}}
	properties.Property("[{{ toUpper .CurveName }}] {{ toLower .PointName}}Gen (affine) should be on the curve", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a *e2) bool {
		{{- end}}
			var op1, op2 {{ toUpper .PointName}}Affine
			op1.FromJacobian(&{{ toLower .PointName}}Gen)
			op2.FromJacobian(&{{ toLower .PointName}}Gen)
			{{- if eq .CoordType "fp.Element" }}
				op2.Y.Mul(&op2.Y, &a)
			{{- else if eq .CoordType "e2" }}
			op2.Y.Mul(&op2.Y, a)
			{{- end}}
			return op1.IsOnCurve() && !op2.IsOnCurve()
		},
		genFuzz1,
	))

	properties.Property("[{{ toUpper .CurveName }}] {{ toLower .PointName}}Gen (Jacobian) should be on the curve", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a *e2) bool {
		{{- end}}
			var op1, op2, op3 {{ toUpper .PointName}}Jac
			op1.Set(&{{ toLower .PointName}}Gen)
			op3.Set(&{{ toLower .PointName}}Gen)

			op2 = fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName}}Gen, a)
			{{- if eq .CoordType "fp.Element" }}
				op3.Y.Mul(&op3.Y, &a)
			{{- else if eq .CoordType "e2" }}
				op3.Y.Mul(&op3.Y, a)
			{{- end}}
			return op1.IsOnCurve() && op2.IsOnCurve() && !op3.IsOnCurve()
		},
		genFuzz1,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func Test{{ toUpper .PointName}}Conversions(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)
	{{- if eq .CoordType "fp.Element" }}
		genFuzz1 := GenFp()
		genFuzz2 := GenFp()
	{{- else if eq .CoordType "e2" }}
		genFuzz1 := GenE2()
		genFuzz2 := GenE2()
	{{- end}}

	properties.Property("[{{ toUpper .CurveName }}] Affine representation should be independent of the Jacobian representative", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a *e2) bool {
		{{- end}}
			g := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)
			var op1 {{ toUpper .PointName}}Affine
			op1.FromJacobian(&g)
			return op1.X.Equal(&{{ toLower .PointName }}Gen.X) && op1.Y.Equal(&{{ toLower .PointName }}Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("[{{ toUpper .CurveName }}] Affine representation should be independent of a Extended Jacobian representative", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a *e2) bool {
		{{- end}}
			var g {{ toLower .PointName }}JacExtended
			g.X.Set(&{{ toLower .PointName }}Gen.X)
			g.Y.Set(&{{ toLower .PointName }}Gen.Y)
			g.ZZ.Set(&{{ toLower .PointName }}Gen.Z)
			g.ZZZ.Set(&{{ toLower .PointName }}Gen.Z)
			gfuzz := fuzzExtendedJacobian{{ toUpper .PointName}}(&g, a)

			var op1 {{ toUpper .PointName}}Affine
			op1.fromJacExtended(&gfuzz)
			return op1.X.Equal(&{{ toLower .PointName }}Gen.X) && op1.Y.Equal(&{{ toLower .PointName }}Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("[{{ toUpper .CurveName }}] Jacobian representation should be the same as the affine representative", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a *e2) bool {
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

	properties.Property("[{{ toUpper .CurveName }}] Converting affine symbol for infinity to Jacobian should output correct infinity in Jacobian", prop.ForAll(
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

	properties.Property("[{{ toUpper .CurveName }}] Converting infinity in extended Jacobian to affine should output infinity symbol in Affine", prop.ForAll(
		func() bool {
			var g {{ toUpper .PointName}}Affine
			var op1 {{ toLower .PointName }}JacExtended
			var zero {{ .CoordType}}
			op1.X.Set(&{{ toLower .PointName }}Gen.X)
			op1.Y.Set(&{{ toLower .PointName }}Gen.Y)
			g.fromJacExtended(&op1)
			return g.X.Equal(&zero) && g.Y.Equal(&zero)
		},
	))

	properties.Property("[{{ toUpper .CurveName }}] Converting infinity in extended Jacobian to Jacobian should output infinity in Jacobian", prop.ForAll(
		func() bool {
			var g {{ toUpper .PointName}}Jac
			var op1 {{ toLower .PointName }}JacExtended
			var zero, one {{ .CoordType}}
			one.SetOne()
			op1.X.Set(&{{ toLower .PointName }}Gen.X)
			op1.Y.Set(&{{ toLower .PointName }}Gen.Y)
			g.fromJacExtended(&op1)
			return g.X.Equal(&one) && g.Y.Equal(&one) && g.Z.Equal(&zero)
		},
	))

	properties.Property("[{{ toUpper .CurveName }}] [Jacobian] Two representatives of the same class should be equal", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a, b {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a, b *e2) bool {
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
	{{- else if eq .CoordType "e2" }}
		genFuzz1 := GenE2()
		genFuzz2 := GenE2()
	{{- end}}

	genScalar := GenFr()

	properties.Property("[{{ toUpper .CurveName }}] [Jacobian] Add should call double when having adding the same point", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a, b {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a, b *e2) bool {
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

	properties.Property("[{{ toUpper .CurveName }}] [Jacobian] Adding the opposite of a point to itself should output inf", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a, b {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a, b *e2) bool {
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

	properties.Property("[{{ toUpper .CurveName }}] [Jacobian] Adding the inf to a point should not modify the point", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a *e2) bool {
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

	properties.Property("[{{ toUpper .CurveName }}] [Jacobian Extended] add (-G) should equal sub(G)", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a *e2) bool {
		{{- end}}
			fop1 := fuzzJacobian{{ toUpper .PointName}}(&{{ toLower .PointName }}Gen, a)
			var p1,p1Neg {{ toUpper .PointName}}Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 {{ toLower .PointName}}JacExtended 
			o1.add(&p1Neg)
			o2.sub(&p1)

			return 	o1.X.Equal(&o2.X) && 
					o1.Y.Equal(&o2.Y) && 
					o1.ZZ.Equal(&o2.ZZ) && 
					o1.ZZZ.Equal(&o2.ZZZ) 
		},
		genFuzz1,
	))

	properties.Property("[{{ toUpper .CurveName }}] [Jacobian Extended] double (-G) should equal doubleNeg(G)", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a *e2) bool {
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

	properties.Property("[{{ toUpper .CurveName }}] [Jacobian] Addmix the negation to itself should output 0", prop.ForAll(
		{{- if eq .CoordType "fp.Element" }}
			func(a {{ .CoordType}}) bool {
		{{- else if eq .CoordType "e2" }}
			func(a *e2) bool {
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

	properties.Property("[{{ toUpper .CurveName }}] scalar multiplication (double and add) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g {{ toUpper .PointName}}Jac
			g.ScalarMultiplication(&{{ toLower .PointName}}Gen, r)

			var scalar, blindedScalard, rminusone big.Int
			var op1, op2, op3, gneg {{ toUpper .PointName}}Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.ScalarMultiplication(&{{ toLower .PointName}}Gen, &rminusone)
			gneg.Neg(&{{ toLower .PointName}}Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			op1.ScalarMultiplication(&{{ toLower .PointName}}Gen, &scalar)
			op2.ScalarMultiplication(&{{ toLower .PointName}}Gen, &blindedScalard)

			return op1.Equal(&op2) && g.Equal(&{{ toLower .PointName}}Infinity) && !op1.Equal(&{{ toLower .PointName}}Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	{{ if eq .CoordType "e2" }}
		properties.Property("[{{ toUpper .CurveName }}] psi should map points from E' to itself", prop.ForAll(
			func() bool {
				var a {{ toUpper .PointName }}Jac
				a.psi(&{{ toLower .PointName }}Gen)
				return a.IsOnCurve() && !a.Equal(&g2Gen)
			},
		))
	{{ end }}

    {{if .GLV}}
        properties.Property("[{{ toUpper .CurveName }}] scalar multiplication (GLV) should depend only on the scalar mod r", prop.ForAll(
            func(s fr.Element) bool {

                r := fr.Modulus()
                var g {{ toUpper .PointName}}Jac
                g.mulGLV(&{{ toLower .PointName}}Gen, r)

                var scalar, blindedScalard, rminusone big.Int
                var op1, op2, op3, gneg {{ toUpper .PointName}}Jac
                rminusone.SetUint64(1).Sub(r, &rminusone)
                op3.mulGLV(&{{ toLower .PointName}}Gen, &rminusone)
                gneg.Neg(&{{ toLower .PointName}}Gen)
                s.ToBigIntRegular(&scalar)
                blindedScalard.Add(&scalar, r)
                op1.mulGLV(&{{ toLower .PointName}}Gen, &scalar)
                op2.mulGLV(&{{ toLower .PointName}}Gen, &blindedScalard)

                return op1.Equal(&op2) && g.Equal(&{{ toLower .PointName}}Infinity) && !op1.Equal(&{{ toLower .PointName}}Infinity) && gneg.Equal(&op3)

            },
            genScalar,
        ))

        properties.Property("[{{ toUpper .CurveName }}] GLV and Double and Add should output the same result", prop.ForAll(
            func(s fr.Element) bool {

                var r big.Int
                var op1, op2 {{ toUpper .PointName}}Jac
                s.ToBigIntRegular(&r)
                op1.mulWindowed(&{{ toLower .PointName}}Gen, &r)
                op2.mulGLV(&{{ toLower .PointName}}Gen, &r)
                return op1.Equal(&op2) && !op1.Equal(&{{ toLower .PointName}}Infinity)

            },
            genScalar,
        ))
    {{end}}

	// note : this test is here as we expect to have a different multiExp than the above bucket method
	// for small number of points
	properties.Property("[{{ toUpper .CurveName }}] Multi exponentation (<50points) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var g {{ toUpper .PointName}}Jac
			g.Set(&{{ toLower .PointName}}Gen)

			// mixer ensures that all the words of a fpElement are set
			samplePoints := make([]{{ toUpper .PointName}}Affine, 30)
			sampleScalars := make([]fr.Element, 30)

			for i := 1; i <= 30; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
				samplePoints[i-1].FromJacobian(&g)
				g.AddAssign(&{{ toLower .PointName}}Gen)
			}

			var op1MultiExp {{ toUpper .PointName}}Jac
			op1MultiExp.MultiExp(samplePoints, sampleScalars)

			var finalBigScalar fr.Element
			var finalBigScalarBi big.Int
			var op1ScalarMul {{ toUpper .PointName}}Jac
			finalBigScalar.SetString("9455").MulAssign(&mixer)
			finalBigScalar.ToBigIntRegular(&finalBigScalarBi)
			op1ScalarMul.ScalarMultiplication(&{{ toLower .PointName}}Gen, &finalBigScalarBi)

			return op1ScalarMul.Equal(&op1MultiExp)
		},
		genScalar,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func Test{{ toUpper .PointName}}MultiExp(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	genScalar := GenFr()
	
	// size of the multiExps 
	const nbSamples = 500

	// multi exp points
	var samplePoints [nbSamples]{{ toUpper $.PointName}}Affine
	var g {{ toUpper $.PointName}}Jac
	g.Set(&{{ toLower $.PointName }}Gen)
	for i := 1; i <= nbSamples; i++ {
		samplePoints[i-1].FromJacobian(&g)
		g.AddAssign(&{{ toLower $.PointName }}Gen)
	}

	// final scalar to use in double and add method (without mixer factor)
	// n(n+1)(2n+1)/6  (sum of the squares from 1 to n)
	var scalar big.Int
	scalar.SetInt64(nbSamples)
	scalar.Mul(&scalar, new(big.Int).SetInt64(nbSamples+1))
	scalar.Mul(&scalar, new(big.Int).SetInt64(2*nbSamples+1))
	scalar.Div(&scalar, new(big.Int).SetInt64(6))

	{{range $c :=  .CRange}}
	
	{{if gt $c 15}}
	if !testing.Short() {
	{{end}}
	properties.Property("[{{ toUpper $.CurveName }}] Multi exponentation (c={{$c}}) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {
	
			var result, expected {{ toUpper $.PointName}}Jac
	
	
			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element
	
			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewCPUSemaphore(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], {{$c}})
			result.msmC{{$c}}(samplePoints[:], scalars, opt)
	
	
			// compute expected result with double and add
			var finalScalar,mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&{{ toLower $.PointName }}Gen, &finalScalar)
	
			return result.Equal(&expected)
		},
		genScalar,
	))

	{{if gt $c 15}}
	}
	{{end}}

	{{end}}
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

{{if .CofactorCleaning }}
func Test{{ toUpper .PointName}}CofactorCleaning(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	properties.Property("[{{ toUpper .CurveName }}] Clearing the cofactor of a random point should set it in the r-torsion", prop.ForAll(
		func() bool {
			var a, x, b {{ .CoordType }}
			a.SetRandom()
			{{if eq .CoordType "fp.Element" }}
				{{if eq .PointName "g2" }}
					x.Square(&a).Mul(&x, &a).Add(&x, &bTwistCurveCoeff)
				{{else}}
					x.Square(&a).Mul(&x, &a).Add(&x, &bCurveCoeff)
				{{end}}	
				for x.Legendre() != 1 {
					a.SetRandom()
					{{if eq .PointName "g2" }}
						x.Square(&a).Mul(&x, &a).Add(&x, &bTwistCurveCoeff)
					{{else}}
						x.Square(&a).Mul(&x, &a).Add(&x, &bCurveCoeff)
					{{end}}
				}
			{{else if eq .CoordType "e2" }}
				x.Square(&a).Mul(&x, &a).Add(&x, &bTwistCurveCoeff)
				for x.Legendre() != 1 {
					a.SetRandom()
					x.Square(&a).Mul(&x, &a).Add(&x, &bTwistCurveCoeff)
				}
			{{end}}
			b.Sqrt(&x)
			var point, pointCleared, infinity {{ toUpper .PointName}}Jac
			point.X.Set(&a)
			point.Y.Set(&b)
			point.Z.SetOne()
			pointCleared.ClearCofactor(&point)
			infinity.Set(&{{ toLower .PointName}}Infinity)
			return point.IsOnCurve() && pointCleared.IsInSubGroup() && !pointCleared.Equal(&infinity)
		},
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))

}
{{end}}

func Test{{ toUpper .PointName}}BatchScalarMultiplication(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	genScalar := GenFr()

	// size of the multiExps
	const nbSamples = 500

	properties.Property("[{{ toUpper .CurveName }}] BatchScalarMultiplication should be consistant with individual scalar multiplications", prop.ForAll(
		func(mixer fr.Element) bool {
			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			result := BatchScalarMultiplication{{ toUpper .PointName}}(&{{ toLower .PointName}}GenAff, sampleScalars[:])

			if len(result) != len(sampleScalars) {
				return false
			}

			for i := 0; i < len(result); i++ {
				var expectedJac {{ toUpper .PointName}}Jac
				var expected {{ toUpper .PointName}}Affine
				var b big.Int
				expectedJac.mulGLV(&{{ toLower .PointName}}Gen, sampleScalars[i].ToBigInt(&b))
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

func Benchmark{{ toUpper .PointName}}BatchScalarMul(b *testing.B) {
	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	const pow = 15
	const nbSamples = 1 << pow

	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
	}

	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_ = BatchScalarMultiplication{{ toUpper .PointName}}(&{{ toLower .PointName}}GenAff, sampleScalars[:using])
			}
		})
	}
}

func Benchmark{{ toUpper .PointName}}ScalarMul(b *testing.B) {

	var scalar big.Int
	r := fr.Modulus()
	scalar.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)
	scalar.Add(&scalar, r)

	var doubleAndAdd {{ toUpper .PointName}}Jac

	b.Run("double and add", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			doubleAndAdd.mulWindowed(&{{ toLower .PointName}}Gen, &scalar)
		}
	})

    {{if .GLV}}
	var glv {{ toUpper .PointName}}Jac
	b.Run("GLV", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			glv.mulGLV(&{{ toLower .PointName}}Gen, &scalar)
		}
	})
    {{end}}

}


{{if .CofactorCleaning}}
func Benchmark{{ toUpper .PointName }}CofactorClearing(b *testing.B) {
	var a {{ toUpper .PointName }}Jac
	a.Set(&{{ toLower .PointName }}Gen)
	for i := 0; i < b.N; i++ {
		a.ClearCofactor(&a)
	}
}
{{end}}

func Benchmark{{ toUpper .PointName}}Add(b *testing.B) {
	var a {{ toUpper .PointName}}Jac
	a.Double(&{{ toLower .PointName}}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddAssign(&{{ toLower .PointName}}Gen)
	}
}

func Benchmark{{ toUpper .PointName}}JacExtendedAdd(b *testing.B) {
	var a {{ toLower .PointName}}JacExtended
	a.double(&{{ toLower .PointName}}GenAff)

	var c {{ toUpper .PointName}}Affine
	c.FromJacobian(&{{ toLower .PointName}}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.add(&c)
	}
}

func Benchmark{{ toUpper .PointName}}JacExtendedSub(b *testing.B) {
	var a {{ toLower .PointName}}JacExtended
	a.double(&{{ toLower .PointName}}GenAff)

	var c {{ toUpper .PointName}}Affine
	c.FromJacobian(&{{ toLower .PointName}}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.sub(&c)
	}
}

func Benchmark{{ toUpper .PointName}}JacExtendedDouble(b *testing.B) {
	var a {{ toLower .PointName}}JacExtended
	a.double(&{{ toLower .PointName}}GenAff)

	var c {{ toUpper .PointName}}Affine
	c.FromJacobian(&{{ toLower .PointName}}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.double(&c)
	}
}


func Benchmark{{ toUpper .PointName}}JacExtendedDoubleNeg(b *testing.B) {
	var a {{ toLower .PointName}}JacExtended
	a.double(&{{ toLower .PointName}}GenAff)

	var c {{ toUpper .PointName}}Affine
	c.FromJacobian(&{{ toLower .PointName}}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.doubleNeg(&c)
	}
}

func Benchmark{{ toUpper .PointName}}AddMixed(b *testing.B) {
	var a {{ toUpper .PointName}}Jac
	a.Double(&{{ toLower .PointName}}Gen)

	var c {{ toUpper .PointName}}Affine
	c.FromJacobian(&{{ toLower .PointName}}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddMixed(&c)
	}

}

func Benchmark{{ toUpper .PointName}}Double(b *testing.B) {
	var a {{ toUpper .PointName}}Jac
	a.Set(&{{ toLower .PointName}}Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.DoubleAssign()
	}

}

func Benchmark{{ toUpper .PointName}}MultiExp{{ toUpper .PointName}}(b *testing.B) {
	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	const pow = (bits.UintSize / 2 ) - (bits.UintSize / 8) // 24 on 64 bits arch, 12 on 32 bits 
	const nbSamples = 1 << pow

	var samplePoints [nbSamples]{{ toUpper .PointName}}Affine
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
		samplePoints[i-1] = {{ toLower .PointName}}GenAff
	}

	var testPoint {{ toUpper .PointName}}Jac

	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				testPoint.MultiExp(samplePoints[:using], sampleScalars[:using])
			}
		})
	}
}

`
