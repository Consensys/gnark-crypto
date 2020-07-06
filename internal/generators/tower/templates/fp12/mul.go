package fp12

const Mul = `
// Mul set z=x*y in {{.Fp12Name}} and return z
func (z *{{.Fp12Name}}) Mul(x, y *{{.Fp12Name}}) *{{.Fp12Name}} {
	// Algorithm 20 from https://eprint.iacr.org/2010/354.pdf

	var t0, t1, xSum, ySum E6

	t0.Mul(&x.C0, &y.C0) // step 1
	t1.Mul(&x.C1, &y.C1) // step 2

	// finish processing input in case z==x or y
	xSum.Add(&x.C0, &x.C1)
	ySum.Add(&y.C0, &y.C1)

	// step 3
	{{- template "fp6InlineMulByNonResidue" dict "all" . "out" "z.C0" "in" "&t1" }}
	z.C0.Add(&z.C0, &t0)                             

	// step 4
	z.C1.Mul(&xSum, &ySum).
		Sub(&z.C1, &t0).
		Sub(&z.C1, &t1)

	return z
}

// Square set z=x*x in {{.Fp12Name}} and return z
func (z *{{.Fp12Name}}) Square(x *{{.Fp12Name}}) *{{.Fp12Name}} {
	// TODO implement Algorithm 22 from https://eprint.iacr.org/2010/354.pdf
	// or the complex method from fp2
	// for now do it the dumb way
	var b0, b1 {{.Fp6Name}}

	b0.Square(&x.C0)
	b1.Square(&x.C1)
	{{- template "fp6InlineMulByNonResidue" dict "all" . "out" "b1" "in" "&b1" }}
	b1.Add(&b0, &b1)

	z.C1.Mul(&x.C0, &x.C1).Double(&z.C1)
	z.C0 = b1

	return z
}

// Inverse set z to the inverse of x in {{.Fp12Name}} and return z
func (z *{{.Fp12Name}}) Inverse(x *{{.Fp12Name}}) *{{.Fp12Name}} {
	// Algorithm 23 from https://eprint.iacr.org/2010/354.pdf

	var t [2]{{.Fp6Name}}

	t[0].Square(&x.C0) // step 1
	t[1].Square(&x.C1) // step 2
	{ // step 3
		var buf {{.Fp6Name}}
		{{- template "fp6InlineMulByNonResidue" dict "all" . "out" "buf" "in" "&t[1]" }}
		t[0].Sub(&t[0], &buf)
	}
	t[1].Inverse(&t[0]) // step 4
	z.C0.Mul(&x.C0, &t[1]) // step 5
	z.C1.Mul(&x.C1, &t[1]).Neg(&z.C1) // step 6

	return z
}

// InverseUnitary inverse a unitary element
// TODO deprecate in favour of Conjugate
func (z *{{.Fp12Name}}) InverseUnitary(x *{{.Fp12Name}}) *{{.Fp12Name}} {
	return z.Conjugate(x)
}

// Conjugate set z to (x.C0, -x.C1) and return z
func (z *{{.Fp12Name}}) Conjugate(x *{{.Fp12Name}}) *{{.Fp12Name}} {
	z.Set(x)
	z.C1.Neg(&z.C1)
	return z
}

// MulByNonResidue multiplies a {{.Fp6Name}} by ((0,0),(1,0),(0,0))
// TODO delete this method once you have another way of testing the inlined code
func (z *{{.Fp6Name}}) MulByNonResidue(x *{{.Fp6Name}}) *{{.Fp6Name}} {
	{{- template "fp6InlineMulByNonResidue" dict "all" . "out" "z" "in" "x" }}
	return z
}
`
