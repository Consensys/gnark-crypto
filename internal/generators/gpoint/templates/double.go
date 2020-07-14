package gpoint

const Double = `

// DoubleAssign doubles a point in Jacobian coordinates
// https://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-3.html#doubling-dbl-2007-bl
func (p *{{.PName}}Jac) DoubleAssign() *{{.PName}}Jac {
	
	// get some Element from our pool
	var XX, YY, YYYY, ZZ, S, M, T {{.PName}}CoordType

	XX.Square(&p.X)
	YY.Square(&p.Y)
	YYYY.Square(&YY)
	ZZ.Square(&p.Z)
	S.Add(&p.X, &YY)
	S.Square(&S).
		SubAssign(&XX).
		SubAssign(&YYYY).
		Double(&S)
	M.Double(&XX).AddAssign(&XX)
	p.Z.AddAssign(&p.Y).
		Square(&p.Z).
		SubAssign( &YY).
		SubAssign( &ZZ)
	T.Square(&M)
	p.X = T
	T.Double(&S)
	p.X.SubAssign(&T)
	p.Y.Sub(&S, &p.X).
		MulAssign(&M)
	YYYY.Double(&YYYY).Double(&YYYY).Double(&YYYY)
	p.Y.SubAssign(&YYYY)

	return p
}
`
