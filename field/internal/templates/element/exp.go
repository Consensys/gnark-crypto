package element

const Exp = `
// Exp z = x^exponent mod q
func (z *{{.ElementName}}) Exp(x {{.ElementName}}, exponent *big.Int) *{{.ElementName}} {
	var bZero big.Int
	if exponent.Cmp(&bZero) == 0 {
		return z.SetOne()
	}

	z.Set(&x)

	for i := exponent.BitLen() - 2; i >= 0; i-- {
		z.Square(z)
		if exponent.Bit(i) == 1 {
			z.Mul(z, &x)
		}
	}

	return z
}

`
