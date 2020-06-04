package pairing

const Frobenius = `
{{- if (eq $.EmbeddingDegree 12) }}

	// Frobenius set z to Frobenius(x), return z
	func (z *PairingResult) Frobenius(x *PairingResult) *PairingResult {
		// Algorithm 28 from https://eprint.iacr.org/2010/354.pdf (beware typos!)
		var t [6]{{.Fp2Name}}

		// Frobenius acts on fp2 by conjugation
		t[0].Conjugate(&x.C0.B0)
		t[1].Conjugate(&x.C0.B1)
		t[2].Conjugate(&x.C0.B2)
		t[3].Conjugate(&x.C1.B0)
		t[4].Conjugate(&x.C1.B1)
		t[5].Conjugate(&x.C1.B2)

		t[1].MulByNonResidue1Power2(&t[1])
		t[2].MulByNonResidue1Power4(&t[2])
		t[3].MulByNonResidue1Power1(&t[3])
		t[4].MulByNonResidue1Power3(&t[4])
		t[5].MulByNonResidue1Power5(&t[5])

		z.C0.B0 = t[0]
		z.C0.B1 = t[1]
		z.C0.B2 = t[2]
		z.C1.B0 = t[3]
		z.C1.B1 = t[4]
		z.C1.B2 = t[5]

		return z
	}

	// FrobeniusSquare set z to Frobenius^2(x), and return z
	func (z *PairingResult) FrobeniusSquare(x *PairingResult) *PairingResult {
		// Algorithm 29 from https://eprint.iacr.org/2010/354.pdf (beware typos!)
		var t [6]{{.Fp2Name}}

		t[1].MulByNonResidue2Power2(&x.C0.B1)
		t[2].MulByNonResidue2Power4(&x.C0.B2)
		t[3].MulByNonResidue2Power1(&x.C1.B0)
		t[4].MulByNonResidue2Power3(&x.C1.B1)
		t[5].MulByNonResidue2Power5(&x.C1.B2)

		z.C0.B0 = x.C0.B0
		z.C0.B1 = t[1]
		z.C0.B2 = t[2]
		z.C1.B0 = t[3]
		z.C1.B1 = t[4]
		z.C1.B2 = t[5]

		return z
	}

	// FrobeniusCube set z to Frobenius^3(x), return z
	func (z *PairingResult) FrobeniusCube(x *PairingResult) *PairingResult {
		// Algorithm 30 from https://eprint.iacr.org/2010/354.pdf (beware typos!)
		var t [6]{{.Fp2Name}}

		// Frobenius^3 acts on fp2 by conjugation
		t[0].Conjugate(&x.C0.B0)
		t[1].Conjugate(&x.C0.B1)
		t[2].Conjugate(&x.C0.B2)
		t[3].Conjugate(&x.C1.B0)
		t[4].Conjugate(&x.C1.B1)
		t[5].Conjugate(&x.C1.B2)

		t[1].MulByNonResidue3Power2(&t[1])
		t[2].MulByNonResidue3Power4(&t[2])
		t[3].MulByNonResidue3Power1(&t[3])
		t[4].MulByNonResidue3Power3(&t[4])
		t[5].MulByNonResidue3Power5(&t[5])

		z.C0.B0 = t[0]
		z.C0.B1 = t[1]
		z.C0.B2 = t[2]
		z.C1.B0 = t[3]
		z.C1.B1 = t[4]
		z.C1.B2 = t[5]

		return z
	}

{{- else if (eq $.EmbeddingDegree 6) }}
	
	// Frobenius set z to Frobenius(x), return z
	func (z *PairingResult) Frobenius(x *PairingResult) *PairingResult {
		// Adapted from https://eprint.iacr.org/2010/354.pdf (Section 3.2)

		z.B0.Conjugate(&x.B0)
		z.B1.Conjugate(&x.B1)
		z.B2.Conjugate(&x.B2)
	
		z.B1.MulByNonResidue1Power1(&z.B1)
		z.B2.MulByNonResidue1Power2(&z.B2)
	
		return z
	}

	// FrobeniusSquare set z to Frobenius^2(x), and return z
	func (z *PairingResult) FrobeniusSquare(x *PairingResult) *PairingResult {
		// Adapted from https://eprint.iacr.org/2010/354.pdf (Section 3.2)
	
		z.Set(x)
	
		z.B1.MulByNonResidue2Power1(&z.B1)
		z.B2.MulByNonResidue2Power2(&z.B2)
	
		return z
	}
	
	// FrobeniusCube set z to Frobenius^3(x), return z
	func (z *PairingResult) FrobeniusCube(x *PairingResult) *PairingResult {
		// Adapted from https://eprint.iacr.org/2010/354.pdf (Section 3.2)
	
		z.B0.Conjugate(&x.B0)
		z.B1.Conjugate(&x.B1)
		z.B2.Conjugate(&x.B2)
	
		z.B1.MulByNonResidue3Power1(&z.B1)
		z.B2.MulByNonResidue3Power2(&z.B2)
	
		return z
	}	

{{- else }}
	// TODO embedding degree {{$.EmbeddingDegree}} not supported
{{- end }}

{{- $d := (div $.EmbeddingDegree 2) }}
{{- range $i, $gammai := .Frobenius }}
	{{- $iplus1 := (add $i 1) }}

	{{- range $j, $gammaij := $gammai }}
		{{- $jplus1 := (add $j 1) }}

		// MulByNonResidue{{$iplus1}}Power{{$jplus1}} set z=x*({{$.Fp6NonResidue}})^({{$jplus1}}*(p^{{$iplus1}}-1)/{{$d}}) and return z
		func (z *{{$.Fp2Name}}) MulByNonResidue{{$iplus1}}Power{{$jplus1}}(x *{{$.Fp2Name}}) *{{$.Fp2Name}} {
			{{- if (eq $gammaij.A1String "0") }}
				// {{$gammaij.A0String}}
				{{- if (eq $gammaij.A0String "1") }}
					// nothing to do
				{{- else }}
					b := fp.Element{
						{{- range $x := $gammaij.A0}}
						{{$x}},{{end}}
					}
					z.A0.Mul(&x.A0, &b)
					z.A1.Mul(&x.A1, &b)
				{{- end }}
			{{- else }}
				// {{ print "(" $gammaij.A0String "," $gammaij.A1String ")" }}
				b := {{$.Fp2Name}}{
					A0: fp.Element{
						{{- range $x := $gammaij.A0}}
						{{$x}},{{end}}
					},
					A1: fp.Element{
						{{- range $x := $gammaij.A1}}
						{{$x}},{{end}}
					},
				}
				z.Mul(x, &b)
			{{- end }}
			return z
		}		
	{{- end}}
{{- end}}
`
