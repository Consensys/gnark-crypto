package element

const HashUtils = `
func (z *{{.ElementName}}) EvalPolynomial(monic bool, coefficients []{{.ElementName}}, x *{{.ElementName}}) {
    dst := coefficients[len(coefficients) - 1]

    if monic {
        dst.Add(&dst, x)
    }

    for i := len(coefficients) - 2; i >= 0; i-- {
        dst.Mul(&dst, x)
        dst.Add(&dst, &coefficients[i])
    }

    *z = dst
}
`
