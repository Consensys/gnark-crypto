package element

const Exp = `
// Exp z = xᵏ (mod q)
func (z *{{.ElementName}}) Exp(x {{.ElementName}}, k *big.Int) *{{.ElementName}} {
	{{- if .F31}}
		if k.IsInt64() {
			return z.ExpInt64(x, k.Int64())
		}	
	{{- else}}
		if k.IsUint64() && k.Uint64() == 0 {
			return z.SetOne()
		}
	{{- end}}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q) == (x⁻¹)ᵏ (mod q)
		x.Inverse(&x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = pool.BigInt.Get()
		defer pool.BigInt.Put(e)
		e.Neg(k)
	}

	z.Set(&x)

	for i := e.BitLen() - 2; i >= 0; i-- {
		z.Square(z)
		if e.Bit(i) == 1 {
			z.Mul(z, &x)
		}
	}

	return z
}

{{- if .F31}}
// ExpInt64 z = xᵏ (mod q)
func (z *{{.ElementName}}) ExpInt64(x {{.ElementName}}, k int64) *{{.ElementName}} {
	if k == 0 {
		return z.SetOne()
	}

	if k < 0 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q) == (x⁻¹)⁻ᵏ (mod q)
		x.Inverse(&x)
		k = -k // if k == math.MinInt64, -k overflows, but uint64(-k) is correct
	}
	e := uint64(k)

	z.Set(&x)

	for i := int(bits.Len64(e)) - 2; i >= 0; i-- {
		z.Square(z)
		if (e>>i)&1 == 1 {
			z.Mul(z, &x)
		}
	}

	return z
}
{{- end}}

`
