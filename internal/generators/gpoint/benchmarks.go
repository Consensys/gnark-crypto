package gpoint

const Benchmarks = `

//--------------------//
//     benches		  //
//--------------------//

var benchRes{{.PName}} {{.PName}}Jac

func Benchmark{{.PName}}ScalarMul(b *testing.B) {

	curve := {{toUpper .Packag.PName}}()
	p := testPoints{{.PName}}()

	var scalar fr.Element
	scalar.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p[1].ScalarMul(curve, &p[1], scalar)
		b.StopTimer()
		scalar.SetRandom()
		b.StartTimer()
	}

}

func Benchmark{{.PName}}Add(b *testing.B) {

	curve := {{toUpper .Packag.PName}}()
	p := testPoints{{.PName}}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchRes{{.PName}}  = p[1]
		benchRes{{.PName}} .Add(curve, &p[2])
	}

}

func Benchmark{{.PName}}AddMixed(b *testing.B) {

	p := testPoints{{.PName}}()
	_p2 := {{.PName}}Affine{}
	p[2].ToAffineFromJac(&_p2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchRes{{.PName}} = p[1]
		benchRes{{.PName}} .AddMixed(&_p2)
	}


}

func Benchmark{{.PName}}Double(b *testing.B) {

	p := testPoints{{.PName}}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchRes{{.PName}} = p[1]
		benchRes{{.PName}}.Double()
	}

}

func Benchmark{{.PName}}WindowedMultiExp(b *testing.B) {
	curve := {{toUpper .Packag.PName}}()

	var G {{.PName}}Jac

	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	var nbSamples int
	nbSamples = 400000

	samplePoints := make([]{{.PName}}Jac, nbSamples)
	sampleScalars := make([]fr.Element, nbSamples)

	G.Set(&curve.{{toLower .PName}}Gen)

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
		samplePoints[i-1].Set(&curve.{{toLower .PName}}Gen)
	}

	var testPoint {{.PName}}Jac

	for i := 0; i < 8; i++ {
		b.Run(fmt.Sprintf("%d points", (i+1)*50000), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				testPoint.WindowedMultiExp(curve, samplePoints[:50000+i*50000], sampleScalars[:50000+i*50000])
			}
		})
	}
}

func BenchmarkMultiExp{{.PName}}(b *testing.B) {

	curve := {{toUpper .Packag.PName}}()

	var G {{.PName}}Jac

	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	var nbSamples int
	nbSamples = 800000

	samplePoints := make([]{{.PName}}Affine, nbSamples)
	sampleScalars := make([]fr.Element, nbSamples)

	G.Set(&curve.{{toLower .PName}}Gen)

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
		G.ToAffineFromJac(&samplePoints[i-1])
	}

	var testPoint {{.PName}}Jac

	for i := 0; i < 16; i++ {
		
		b.Run(fmt.Sprintf("%d points)", (i+1)*50000), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				<-testPoint.MultiExp(curve, samplePoints[:50000+i*50000], sampleScalars[:50000+i*50000])
			}
		})
	}
}
		
`
