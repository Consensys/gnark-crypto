package element

const BigNum = `

{{/* Only used for the Pornin Extended GCD Inverse Algorithm*/}}
{{if $.UsingP20Inverse}}

func (z *{{.ElementName}}) neg(x *{{.ElementName}}, xHi uint64) uint64 {
	var b uint64

	z[0], b = bits.Sub64(0, x[0], 0)
	{{- range $i := .NbWordsIndexesNoZero}}
	z[{{$i}}], b = bits.Sub64(0, x[{{$i}}], b)
	{{- end}}
	xHi, _ = bits.Sub64(0, xHi, b)

	return xHi
}

// mulWNonModular multiplies by one word in non-montgomery, without reducing
func (z *{{.ElementName}}) mulWNonModular(x *{{.ElementName}}, y int64) uint64 {

	// w := abs(y)
	m := y >> 63
	w := uint64((y^m)-m)

	var c uint64
	c, z[0] = bits.Mul64(x[0], w)
	{{- range $i := .NbWordsIndexesNoZero }}
	c, z[{{$i}}] = madd1(x[{{$i}}], w, c)
	{{- end}}

	if y < 0 {
		c = z.neg(z, c)
	}

	return c
}

// linearCombNonModular computes a linear combination without modular reduction
func (z *{{.ElementName}}) linearCombNonModular(x *{{.ElementName}}, xC int64, y *{{.ElementName}}, yC int64) uint64 {
	var yTimes {{.ElementName}}

	yHi := yTimes.mulWNonModular(y, yC)
	xHi := z.mulWNonModular(x, xC)

	carry := uint64(0)

	{{- range $i := .NbWordsIndexesFull}}
		z[{{$i}}], carry = bits.Add64(z[{{$i}}], yTimes[{{$i}}], carry)
	{{- end}}

	yHi, _ = bits.Add64(xHi, yHi, carry)

	return yHi
}

{{- end}}
`
