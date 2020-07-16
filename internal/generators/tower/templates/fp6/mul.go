package fp6

const Mul = `
// Mul sets z to the {{.Fp6Name}}-product of x,y, returns z
func (z *{{.Fp6Name}}) Mul(x, y *{{.Fp6Name}}) *{{.Fp6Name}} {
	{{ template "mul" dict "all" . "V1" "x" "V2" "y"}}
	return z
}

// MulAssign sets z to the {{.Fp6Name}}-product of z,x returns z
func (z *{{.Fp6Name}}) MulAssign(x *{{.Fp6Name}}) *{{.Fp6Name}} {
	{{ template "mul" dict "all" . "V1" "z" "V2" "x"}}
	return z
}

{{- define "mul" }}
	// Algorithm 13 from https://eprint.iacr.org/2010/354.pdf
	var rb0, b0, b1, b2, b3, b4 {{.all.Fp2Name}}
	b0.Mul(&{{.V1}}.B0, &{{.V2}}.B0) // step 1
	b1.Mul(&{{.V1}}.B1, &{{.V2}}.B1) // step 2
	b2.Mul(&{{.V1}}.B2, &{{.V2}}.B2) // step 3

	// step 4
	b3.Add(&{{.V1}}.B1, &{{.V1}}.B2)
	b4.Add(&{{.V2}}.B1, &{{.V2}}.B2)
	rb0.Mul(&b3, &b4).
		SubAssign(&b1).
		SubAssign(&b2)
	{{- template "fp2InlineMulByNonResidue" dict "all" .all "out" "rb0" "in" "&rb0" }}
	rb0.AddAssign(&b0)
	
	// step 5
	b3.Add(&{{.V1}}.B0, &{{.V1}}.B1)
	b4.Add(&{{.V2}}.B0, &{{.V2}}.B1)
	z.B1.Mul(&b3, &b4).
		SubAssign(&b0).
		SubAssign(&b1)
	{{- template "fp2InlineMulByNonResidue" dict "all" .all "out" "b3" "in" "&b2" }}
	z.B1.AddAssign(&b3)
	
	// step 6
	b3.Add(&{{.V1}}.B0, &{{.V1}}.B2)
	b4.Add(&{{.V2}}.B0, &{{.V2}}.B2)
	z.B2.Mul(&b3, &b4).
		SubAssign(&b0).
		SubAssign(&b2).
		AddAssign(&b1)
	z.B0 = rb0
{{- end }}

// Square sets z to the {{.Fp6Name}}-product of x,x, returns z
func (z *{{.Fp6Name}}) Square(x *{{.Fp6Name}}) *{{.Fp6Name}} {
	{{ template "square" dict "all" . "V" "x" }}
	return z
}

// SquareAssign sets z to the {{.Fp6Name}}-product of z,z returns z
func (z *{{.Fp6Name}}) SquareAssign() *{{.Fp6Name}} {
	{{ template "square" dict "all" . "V" "z" }}
	return z
}

{{- define "square" }}
	// Algorithm 16 from https://eprint.iacr.org/2010/354.pdf
	var b0, b1, b2, b3, b4 {{.all.Fp2Name}}
	b3.Mul(&{{.V}}.B0, &{{.V}}.B1).Double(&b3) // step 1
	b4.Square(&{{.V}}.B2) // step 2
	
	// step 3
	{{- template "fp2InlineMulByNonResidue" dict "all" .all "out" "b0" "in" "&b4" }}
	b0.AddAssign(&b3)
	b1.Sub(&b3, &b4) // step 4
	b2.Square(&{{.V}}.B0) // step 5
	b3.Sub(&{{.V}}.B0, &{{.V}}.B1).AddAssign(&{{.V}}.B2).Square(&b3) // steps 6 and 8
	b4.Mul(&{{.V}}.B1, &{{.V}}.B2).Double(&b4) // step 7
	// step 9
	{{- template "fp2InlineMulByNonResidue" dict "all" .all "out" "z.B0" "in" "&b4" }}
	z.B0.AddAssign(&b2)
	
	// step 10
	z.B2.Add(&b1, &b3).
		AddAssign(&b4).
		SubAssign(&b2)
	z.B1 = b0
{{- end }}

{{/* HACK: bw761 is the only curve that needs cyclotomic square in Fp6; all others need it in Fp12 */}}
{{- if (eq .Fpackage "bw761") }}
	// CyclotomicSquare https://eprint.iacr.org/2009/565.pdf, 3.2
	func (z *{{.Fp6Name}}) CyclotomicSquare(x *{{.Fp6Name}}) *{{.Fp6Name}} {

		var res, a {{.Fp6Name}}
		var tmp {{.Fp2Name}}

		// A
		res.B0.Square(&x.B0)
		a.B0.Conjugate(&x.B0)

		// B
		res.B2.A0.Set(&x.B1.A1)
		res.B2.A1.MulByNonResidueInv(&x.B1.A0)
		res.B2.Square(&res.B2).Double(&res.B2).Double(&res.B2).Neg(&res.B2)
		a.B2.Conjugate(&x.B2)

		// C
		tmp.Square(&x.B2)
		res.B1.A0.MulByNonResidue(&tmp.A1)
		res.B1.A1.Set(&tmp.A0)
		a.B1.A0.Neg(&x.B1.A0)
		a.B1.A1.Set(&x.B1.A1)

		z.Sub(&res, &a).Double(z).Add(z, &res)

		return z
	}
{{- end }}

// Inverse an element in {{.Fp6Name}}
func (z *{{.Fp6Name}}) Inverse(x *{{.Fp6Name}}) *{{.Fp6Name}} {
	// Algorithm 17 from https://eprint.iacr.org/2010/354.pdf
	// step 9 is wrong in the paper!
	// memalloc
	var t [7]{{.Fp2Name}}
	var c [3]{{.Fp2Name}}
	var buf {{.Fp2Name}}
	t[0].Square(&x.B0) // step 1
	t[1].Square(&x.B1) // step 2
	t[2].Square(&x.B2) // step 3
	t[3].Mul(&x.B0, &x.B1) // step 4
	t[4].Mul(&x.B0, &x.B2) // step 5
	t[5].Mul(&x.B1, &x.B2) // step 6
	// step 7
	{{- template "fp2InlineMulByNonResidue" dict "all" . "out" "c[0]" "in" "&t[5]" }}
	c[0].Neg(&c[0]).AddAssign(&t[0])
	// step 8
	{{- template "fp2InlineMulByNonResidue" dict "all" . "out" "c[1]" "in" "&t[2]" }}
	c[1].SubAssign(&t[3])
	c[2].Sub(&t[1], &t[4]) // step 9 is wrong in 2010/354!
	// steps 10, 11, 12
	t[6].Mul(&x.B2, &c[1])
	buf.Mul(&x.B1, &c[2])
	t[6].AddAssign(&buf)
	{{- template "fp2InlineMulByNonResidue" dict "all" . "out" "t[6]" "in" "&t[6]" }}
	buf.Mul(&x.B0, &c[0])
	t[6].AddAssign(&buf)
	
	t[6].Inverse(&t[6]) // step 13
	z.B0.Mul(&c[0], &t[6]) // step 14
	z.B1.Mul(&c[1], &t[6]) // step 15
	z.B2.Mul(&c[2], &t[6]) // step 16
	return z
}
`
