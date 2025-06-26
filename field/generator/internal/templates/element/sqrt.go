package element

const Sqrt = `

{{ if not .UseAddChain}}
var (
	_bLegendreExponent{{.ElementName}} *big.Int
	_bSqrtExponent{{.ElementName}} *big.Int
)

func init() {
	_bLegendreExponent{{.ElementName}}, _ = new(big.Int).SetString("{{.LegendreExponent}}", 16)
	{{- if .SqrtQ3Mod4}}
		const sqrtExponent{{.ElementName}} = "{{.SqrtQ3Mod4Exponent}}"
	{{- else if .SqrtAtkin}}
		const sqrtExponent{{.ElementName}} = "{{.SqrtAtkinExponent}}"
	{{- else if .SqrtTonelliShanks}}
		const sqrtExponent{{.ElementName}} = "{{.SqrtSMinusOneOver2}}"
	{{- end }}
	_bSqrtExponent{{.ElementName}}, _ = new(big.Int).SetString(sqrtExponent{{.ElementName}}, 16)
}

{{- end }}

{{- $p20 := and  .UsingP20Inverse (not (eq .NbWords 1))}}

// Legendre returns the Legendre symbol of z (either +1, -1, or 0.)
func (z *{{.ElementName}}) Legendre() int {
{{- if $p20}}

	// Adapts "Optimized Binary GCD for Modular Inversion"
	// https://github.com/pornin/bingcd/blob/main/doc/bingcd.pdf
	// For a faithful implementation of Pornin20 see [Inverse].

	// We don't need to account for z being in Montgomery form.
	// (xR|q) = (x|q)(R|q). R is a square (an even power of 2), so (R|q) = 1.
	a := *z
	b := {{.ElementName}} {
		{{- range $i := .NbWordsIndexesFull}}
		q{{$i}},{{end}}
	}	// b := q

	// Update factors: we get [a; b] ← [f₀ g₀; f₁ g₁] [a; b]
	// cᵢ = fᵢ + 2³¹ - 1 + 2³² * (gᵢ + 2³¹ - 1)
	var c0, c1 int64

	var s {{.ElementName}}

	l := 1 // loop invariant: (x|q) = (a|b) . l
	// This means that every time a and b are updated into a' and b',
	// l is updated into l' = (x|q)(a'|b')=(x|q)(a|b)(a|b)(a'|b') = l (a|b)(a'|b')
	// During the algorithm's run, there is no guarantee that b remains prime, or even positive.
	// Therefore, we use the properties of the Kronecker symbol, a generalization of the Legendre symbol to all integers.

	for !a.IsZero() {
		n := max(a.BitLen(), b.BitLen())
		aApprox, bApprox := approximate(&a, n), approximate(&b, n)

		// f₀, g₀, f₁, g₁ = 1, 0, 0, 1
		c0, c1 = updateFactorIdentityMatrixRow0, updateFactorIdentityMatrixRow1

		const nbIterations = k - 3
		// running fewer iterations because we need access to 3 low bits from b, rather than 1 in the inversion algorithm
		for range nbIterations {

			if aApprox&1 == 0 {
				aApprox /= 2

				// update the Kronecker symbol
				//
				// (a/2 | b) (2|b) = (a|b)
				//
				// b is either odd or zero, the latter case implying a non-trivial GCD and an ultimate result of 0,
				// regardless of what value l holds.
				// So in updating l, we may assume that b is odd.
				// Since a is even, we only need to correctly compute l if b is odd.
				// if b is also even, the non-trivial GCD will result in the function returning 0 anyway.
				// so we may here assume b is odd.
				// (2|b) = 1 if b ≡ 1 or 7 (mod 8), and -1 if b ≡ 3 or 5 (mod 8)
				if bMod8 := bApprox & 7; bMod8 == 3 || bMod8 == 5 {
					l = -l
				}

			} else {
				s, borrow := bits.Sub64(aApprox, bApprox, 0)
				if borrow == 1 {
					// Compute (b-a|a)
					// (x-y|z) = (x|z) unless z < 0 and sign(x-y) ≠ sign(x)
					// Pornin20 asserts that at least one of a and b is non-negative.
					// If a is non-negative, we immediately get (b-a|a) = (b|a)
					// If a is negative, b-a > b. But b is already non-negative, so the b-a and b have the same sign.
					// Thus in that case also (b-a|a) = (b|a)
					// Since not both a and b are negative, we get a quadratic reciprocity law
					// like that of the Legendre symbol: (b|a) = (a|b), unless a, b ≡ 3 (mod 4), in which case (b|a) = -(a|b)
					if bApprox&3 == 3 && aApprox&3 == 3 {
						l = -l
					}

					s = bApprox - aApprox
					bApprox = aApprox
					c0, c1 = c1, c0
				}

				aApprox = s / 2
				c0 = c0 - c1

				// update l to reflect halving a, just like in the case where a is even
				if bMod8 := bApprox & 7; bMod8 == 3 || bMod8 == 5 {
					l = -l
				}
			}

			c1 *= 2
		}

		s = a

		var g0 int64
		// from this point on c0 aliases for f0
		c0, g0 = updateFactorsDecompose(c0)
		aHi := a.linearCombNonModular(&s, c0, &b, g0)
		if aHi&signBitSelector != 0 {
			// if aHi < 0
			aHi = negL(&a, aHi)
			// Since a is negative, b is not and hence b ≠ -1
			// So we get (-a|b)=(-1|b)(a|b)
			// b is odd so we get (-1|b) = 1 if b ≡ 1 (mod 4) and -1 otherwise.
			if bApprox&3 == 3 { // we still have two valid lower bits for b
				l = -l
			}
		}
		// right-shift a by k-3 bits
		{{- range $i := .NbWordsIndexesFull}}
			{{-  if eq $i $.NbWordsLastIndex}}
				a[{{$i}}] = (a[{{$i}}] >> nbIterations) | (aHi << (2*k - nbIterations))
			{{-  else  }}
				a[{{$i}}] = (a[{{$i}}] >> nbIterations) | ((a[{{add $i 1}}]) << (2*k - nbIterations))
			{{- end}}
		{{- end}}

		var f1 int64
		// from this point on c1 aliases for g0
		f1, c1 = updateFactorsDecompose(c1)
		bHi := b.linearCombNonModular(&s, f1, &b, c1)
		if bHi&signBitSelector != 0 {
			// if bHi < 0
			bHi = negL(&b, bHi)
			// no need to update l, since we know a ≥ 0
			// (a|-1) = 1 if a ≥ 0
		}
		// right-shift b by k-3 bits
		{{- range $i := .NbWordsIndexesFull}}
			{{-  if eq $i $.NbWordsLastIndex}}
				b[{{$i}}] = (b[{{$i}}] >> nbIterations) | (bHi << (2*k - nbIterations))
			{{-  else  }}
				b[{{$i}}] = (b[{{$i}}] >> nbIterations) | ((b[{{add $i 1}}]) << (2*k - nbIterations))
			{{- end}}
		{{- end}}
	}

	if b[0] == 1 && ({{ range $i := .NbWordsIndexesNoZeroNoLast}}b[{{$i}}]|{{end}}b[{{.NbWordsLastIndex}}]) == 0 {
		return l // (0|1) = 1
	} else {
		return 0 // if b ≠ 1, then (z,q) ≠ 0 ⇒ (z|q) = 0
	}
{{- else}}
	var l {{.ElementName}}
	// z^((q-1)/2)
	{{- if .UseAddChain}}
	l.expByLegendreExp(*z)
	{{- else}}
	l.Exp(*z, _bLegendreExponent{{.ElementName}})
	{{- end}}
	
	if l.IsZero() {
		return 0
	} 

	// if l == 1
	if l.IsOne()  {
		return 1
	}
	return -1
{{- end}}
}

{{- if $p20}}
// approximate a big number x into a single 64 bit word using its uppermost and lowermost bits.
// If x fits in a word as is, no approximation necessary.
// This differs from the standard approximate function in that in the Legendre symbol computation
// we need to access the 3 low bits of b, rather than just one. So lo ≥ n+2 where n is the number of inner iterations.
// The requirement on the high bits is unchanged, hi ≥ n+1.
// Thus we hit a maximum of hi = lo = k and n = k-2 as opposed to n = lo = k-1 and hi = k+1 in the standard approximate function.
// Since we are doing fewer iterations than in the inversion algorithm, all the arguments on bounds for update factors remain valid.
func approximateForLegendre(x *{{.ElementName}}, nBits int) uint64 {

	if nBits <= 64 {
		return x[0]
	}

	const mask = (uint64(1) << k ) - 1 // k ones
	lo := mask & x[0]

	hiWordIndex := (nBits - 1) / 64

	hiWordBitsAvailable := nBits - hiWordIndex * 64
	hiWordBitsUsed := min(hiWordBitsAvailable, k)

	mask_ := uint64(^((1 << (hiWordBitsAvailable - hiWordBitsUsed)) - 1))
	hi := (x[hiWordIndex] & mask_) << (64 - hiWordBitsAvailable)

	mask_ = ^(1<<(k + hiWordBitsUsed) - 1)
	mid := (mask_ & x[hiWordIndex-1]) >> hiWordBitsUsed

	return lo | mid | hi
}
{{- end}}


// Sqrt z = √x (mod q)
// if the square root doesn't exist (x is not a square mod q)
// Sqrt leaves z unchanged and returns nil
func (z *{{.ElementName}}) Sqrt(x *{{.ElementName}}) *{{.ElementName}} {
	{{- if .SqrtQ3Mod4}}
		// q ≡ 3 (mod 4)
		// using  z ≡ ± x^((p+1)/4) (mod q)
		var y, square {{.ElementName}}
		{{- if .UseAddChain}}
		y.expBySqrtExp(*x)
		{{- else}}
		y.Exp(*x, _bSqrtExponent{{.ElementName}})
		{{- end }}
		// as we didn't compute the legendre symbol, ensure we found y such that y * y = x
		square.Square(&y)
		if square.Equal(x) {
			return z.Set(&y)
		} 
		return nil
	{{- else if .SqrtAtkin}}
		// q ≡ 5 (mod 8)
		// see modSqrt5Mod8Prime in math/big/int.go
		var one, alpha, beta, tx, square {{.ElementName}}
		one.SetOne()
		tx.Double(x)
		{{- if .UseAddChain}}
		alpha.expBySqrtExp(tx)
		{{ else }}
		alpha.Exp(tx, _bSqrtExponent{{.ElementName}})
		{{- end }}
		beta.Square(&alpha).
			Mul(&beta, &tx).
			Sub(&beta, &one).
			Mul(&beta, x).
			Mul(&beta, &alpha)
		
		// as we didn't compute the legendre symbol, ensure we found beta such that beta * beta = x
		square.Square(&beta)
		if square.Equal(x) {
			return z.Set(&beta)
		}
		return nil
	{{- else if .SqrtTonelliShanks}}
		// q ≡ 1 (mod 4)
		// see modSqrtTonelliShanks in math/big/int.go
		// using https://www.maa.org/sites/default/files/pdf/upload_library/22/Polya/07468342.di020786.02p0470a.pdf

		var y, b,t, w  {{.ElementName}}
		// w = x^((s-1)/2))
		{{- if .UseAddChain}}
		w.expBySqrtExp(*x)
		{{- else}}
		w.Exp(*x, _bSqrtExponent{{.ElementName}})
		{{- end}}

		// y = x^((s+1)/2)) = w * x
		y.Mul(x, &w)

		// b = xˢ = w * w * x = y * x
		b.Mul(&w, &y)

		// g = nonResidue ^ s
		var g = {{.ElementName}}{
			{{- range $i := .SqrtG}}
			{{$i}},{{end}}
		}
		r := uint64({{.SqrtE}})

		// compute legendre symbol
		// t = x^((q-1)/2) = r-1 squaring of xˢ
		t = b
		for i:=uint64(0); i < r-1; i++ {
			t.Square(&t)
		}
		if t.IsZero() {
			return z.SetZero()
		}
		if !t.IsOne() {
			// t != 1, we don't have a square root
			return nil
		}
		for {
			var m uint64
			t = b 

			// for t != 1
			for !t.IsOne() {
				t.Square(&t)
				m++
			}

			if m == 0 {
				return z.Set(&y)
			}
			// t = g^(2^(r-m-1)) (mod q)
			ge := int(r - m - 1)
			t = g
			for ge > 0 {
				t.Square(&t)
				ge--
			}

			g.Square(&t)
			y.Mul(&y, &t)
			b.Mul(&b, &g)
			r = m
		}

	{{- else}}
		panic("not implemented")	
	{{- end}}
}



`
