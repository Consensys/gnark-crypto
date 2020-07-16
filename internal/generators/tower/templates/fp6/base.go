package fp6

const Base = `
// {{.Fp6Name}} is a degree-three finite field extension of fp2:
// B0 + B1v + B2v^2 where v^3-{{.Fp6NonResidue}} is irrep in fp2
type {{.Fp6Name}} struct {
	B0, B1, B2 {{.Fp2Name}}
}

// Equal returns true if z equals x, fasle otherwise
// TODO can this be deleted?  Should be able to use == operator instead
func (z *{{.Fp6Name}}) Equal(x *{{.Fp6Name}}) bool {
	return z.B0.Equal(&x.B0) && z.B1.Equal(&x.B1) && z.B2.Equal(&x.B2)
}

// SetString sets a {{.Fp6Name}} elmt from stringf
func (z *{{.Fp6Name}}) SetString(s1, s2, s3, s4, s5, s6 string) *{{.Fp6Name}} {
	z.B0.SetString(s1, s2)
	z.B1.SetString(s3, s4)
	z.B2.SetString(s5, s6)
	return z
}

// Set Sets a {{.Fp6Name}} elmt form another {{.Fp6Name}} elmt
func (z *{{.Fp6Name}}) Set(x *{{.Fp6Name}}) *{{.Fp6Name}} {
	z.B0 = x.B0
	z.B1 = x.B1
	z.B2 = x.B2
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *{{.Fp6Name}}) SetOne() *{{.Fp6Name}} {
	z.B0.A0.SetOne()
	z.B0.A1.SetZero()
	z.B1.A0.SetZero()
	z.B1.A1.SetZero()
	z.B2.A0.SetZero()
	z.B2.A1.SetZero()
	return z
}

// SetRandom set z to a random elmt
func (z *{{.Fp6Name}}) SetRandom() *{{.Fp6Name}} {
	z.B0.SetRandom()
	z.B1.SetRandom()
	z.B2.SetRandom()
	return z
}

// ToMont converts to Mont form
func (z *{{.Fp6Name}}) ToMont() *{{.Fp6Name}} {
	z.B0.ToMont()
	z.B1.ToMont()
	z.B2.ToMont()
	return z
}

// FromMont converts from Mont form
func (z *{{.Fp6Name}}) FromMont() *{{.Fp6Name}} {
	z.B0.FromMont()
	z.B1.FromMont()
	z.B2.FromMont()
	return z
}

// Add adds two elements of {{.Fp6Name}}
func (z *{{.Fp6Name}}) Add(x, y *{{.Fp6Name}}) *{{.Fp6Name}} {
	z.B0.Add(&x.B0, &y.B0)
	z.B1.Add(&x.B1, &y.B1)
	z.B2.Add(&x.B2, &y.B2)
	return z
}

// Neg negates the {{.Fp6Name}} number
func (z *{{.Fp6Name}}) Neg(x *{{.Fp6Name}}) *{{.Fp6Name}} {
	z.B0.Neg(&z.B0)
	z.B1.Neg(&z.B1)
	z.B2.Neg(&z.B2)
	return z
}

// Sub two elements of {{.Fp6Name}}
func (z *{{.Fp6Name}}) Sub(x, y *{{.Fp6Name}}) *{{.Fp6Name}} {
	z.B0.Sub(&x.B0, &y.B0)
	z.B1.Sub(&x.B1, &y.B1)
	z.B2.Sub(&x.B2, &y.B2)
	return z
}

// Double doubles an element in {{.Fp6Name}}
func (z *{{.Fp6Name}}) Double(x *{{.Fp6Name}}) *{{.Fp6Name}} {
	z.B0.Double(&x.B0)
	z.B1.Double(&x.B1)
	z.B2.Double(&x.B2)
	return z
}

// String puts {{.Fp6Name}} elmt in string form
func (z *{{.Fp6Name}}) String() string {
	return (z.B0.String() + "+(" + z.B1.String() + ")*v+(" + z.B2.String() + ")*v**2")
}
`
