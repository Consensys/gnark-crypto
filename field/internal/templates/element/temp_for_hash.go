package element

const TempForHash = `

// Sgn0 is an algebraic substitute for the notion of sign in ordered fields
// Namely, every non-zero quadratic residue in a finite field of characteristic =/= 2 has exactly two square roots, one of each sign
// Taken from https://datatracker.ietf.org/doc/draft-irtf-cfrg-hash-to-curve/ section 4.1
// The sign of an element is not obviously related to that of its Montgomery form
func (z *{{.ElementName}}) Sgn0() bool {
	nonMont := *z
	nonMont.FromMont()

	return nonMont[0]%2 == 1
}

func (z *{{.ElementName}}) SetHex(hex string) {
	var i big.Int
	i.SetString(hex, 16)
	if _, b := i.SetString(hex, 16); !b {
		panic("SetString failed")
	}
	z.SetBigInt(&i)
}

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
