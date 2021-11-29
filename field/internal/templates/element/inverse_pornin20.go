package element

const InversePornin = `

{{/* We use big.Int for Inverse for these type of moduli */}}
{{if eq .NoCarry false}}

// Inverse z = x^-1 mod q 
// note: allocates a big.Int (math/big)
func (z *{{.ElementName}}) Inverse( x *{{.ElementName}}) *{{.ElementName}} {
	var _xNonMont big.Int
	x.ToBigIntRegular( &_xNonMont)
	_xNonMont.ModInverse(&_xNonMont, Modulus())
	z.SetBigInt(&_xNonMont)
	return z
}

{{ else }}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

//Though we're defining k as a constant, this code "profoundly" assumes that the processor is 64 bit
const k = 32 // word size / 2
const signBitSelector = uint64(1) << 63
const approxLowBitsN = k - 1
const approxHighBitsN = k + 1

func approximate(x *{{.ElementName}}, n int) uint64 {

	if n <= 64 {
		return x[0]
	}

	const mask = (uint64(1) << (k - 1)) - 1 //k-1 ones
	lo := mask & x[0]

	hiWordIndex := (n - 1) / 64

	hiWordBitsAvailable := n - hiWordIndex * 64
	hiWordBitsUsed := min(hiWordBitsAvailable, approxHighBitsN)

	mask_ := uint64(^((1 << (hiWordBitsAvailable - hiWordBitsUsed)) - 1))
	hi := (x[hiWordIndex] & mask_) << (64 - hiWordBitsAvailable)

	mask_ = ^(1<<(approxLowBitsN + hiWordBitsUsed) - 1)
	mid := (mask_ & x[hiWordIndex-1]) >> hiWordBitsUsed

	return lo | mid | hi
}

//TODO: Work out formula for correction factor
var inversionCorrectionFactor = {{.ElementName}}{5743661648749932980, 12551916556084744593, 23273105902916091, 802172129993363311}

func (z *{{.ElementName}}) Inverse(x *{{.ElementName}}) *{{.ElementName}} {
	if x.IsZero() {
		z.SetZero()
		return z
	}

	var a = *x
	var b = qElement
	var u = {{.ElementName}}{1}

	//Update factors: we get [u; v]:= [f0 g0; f1 g1] [u; v]
	var f0, g0, f1, g1 int64

	//Saved update factors to reduce the number of field multiplications
	var pf0, pg0, pf1, pg1 int64

	var i uint

	var v, s {{.ElementName}}

	const iterationN =   2 * ( (2 * Bits - 2) / (2*k) + 1) // 2  ⌈ (2 * field size - 1) / 2k ⌉

	//Since u,v are updated every other iteration, we must make sure we terminate after evenly many iterations
	//This also lets us get away with half as many updates to u,v
	//To make this constant-time-ish, replace the condition with i < iterationN
	for i = 0; i&1 == 1 || !a.IsZero(); i++ {
		n := max(a.BitLen(), b.BitLen())
		aApprox, bApprox := approximate(&a, n), approximate(&b, n)

		// After 0 iterations, we have f₀ ≤ 2⁰ and f₁ < 2⁰
		f0, g0, f1, g1 = 1, 0, 0, 1

		for j := 0; j < approxLowBitsN; j++ {

			if aApprox&1 == 0 {
				aApprox /= 2
			} else {
				s, borrow := bits.Sub64(aApprox, bApprox, 0)
				if borrow == 1 {
					s = bApprox - aApprox
					bApprox = aApprox
					f0, f1 = f1, f0
					g0, g1 = g1, g0
				}

				aApprox = s / 2
				f0 -= f1
				g0 -= g1

				//Now |f₀| < 2ʲ + 2ʲ = 2ʲ⁺¹
				//|f₁| ≤ 2ʲ still

			}

			f1 *= 2
			g1 *= 2
			//|f₁| ≤ 2ʲ⁺¹

		}

		s = a
		aHi := a.linearCombNonModular(&s, f0, &b, g0)
		if aHi & signBitSelector != 0 {
			// if aHi < 0
			f0, g0 = -f0, -g0
			aHi = a.neg(&a, aHi)
		}
		//right-shift a by k-1 bits

		{{- range $i := .NbWordsIndexesFull}}
			{{-  if eq $i $.NbWordsLastIndex}}
				a[{{$i}}] = (a[{{$i}}] >> approxLowBitsN) | (aHi << approxHighBitsN)
			{{-  else  }}
				//TODO: Make sure the +1 thing is working
				a[{{$i}}] = (a[{{$i}}] >> approxLowBitsN) | ((a[{{add $i 1}}]) << approxHighBitsN)
			{{- end}}
		{{- end}}

		bHi := b.linearCombNonModular(&s, f1, &b, g1)
		if bHi & signBitSelector != 0 {
			// if bHi < 0
			f1, g1 = -f1, -g1
			bHi = b.neg(&b, bHi)
		}
		//right-shift b by k-1 bits
		/*b[0] = (b[0] >> approxLowBitsN) | ((b[1]) << approxHighBitsN)
		b[1] = (b[1] >> approxLowBitsN) | ((b[2]) << approxHighBitsN)
		b[2] = (b[2] >> approxLowBitsN) | ((b[3]) << approxHighBitsN)
		b[3] = (b[3] >> approxLowBitsN) | ((bHi) << approxHighBitsN)*/

		{{- range $i := .NbWordsIndexesFull}}
			{{-  if eq $i $.NbWordsLastIndex}}
				b[{{$i}}] = (b[{{$i}}] >> approxLowBitsN) | (bHi << approxHighBitsN)
			{{-  else  }}
				//TODO: Make sure the +1 thing is working
				b[{{$i}}] = (b[{{$i}}] >> approxLowBitsN) | ((b[{{add $i 1}}]) << approxHighBitsN)
			{{- end}}
		{{- end}}

		if i&1 == 1 {
			//Combine current update factors with previously stored ones
			// [f₀, g₀; f₁, g₁] ← [f₀, g₀; f₁, g₀] [pf₀, pg₀; pf₀, pg₀]
			// We have |f₀|, |g₀|, |pf₀|, |pf₁| ≤ 2ᵏ⁻¹, and that |pf_i| < 2ᵏ⁻¹ for i ∈ {0, 1}
			// Then for the new value we get |f₀| < 2ᵏ⁻¹ × 2ᵏ⁻¹ + 2ᵏ⁻¹ × 2ᵏ⁻¹ = 2²ᵏ⁻¹
			// Which leaves us with an extra bit for the sign
			f0, g0, f1, g1 = f0*pf0+g0*pf1,
				f0*pg0+g0*pg1,
				f1*pf0+g1*pf1,
				f1*pg0+g1*pg1

			s = u
			u.linearCombSosSigned(&u, f0, &v, g0)
			v.linearCombSosSigned(&s, f1, &v, g1)

		} else {
			//Save update factors
			pf0, pg0, pf1, pg1 = f0, g0, f1, g1
		}

	}

	//For every iteration that we miss, v is not being multiplied by 2²ᵏ⁻²
	const pSq int64 = 1 << (2 * (k - 1))
	//If the function is constant-time ish, this loop will not run (probably no need to take it out explicitly)
	for ; i < iterationN; i += 2 {
		v.mulWSigned(&v, pSq)
	}

	z.Mul(&v, &inversionCorrectionFactor)
	return z
}

func (z *{{.ElementName}}) linearCombSosSigned(x *{{.ElementName}}, xC int64, y *{{.ElementName}}, yC int64) {
	hi := z.linearCombNonModular(x, xC, y, yC)
	z.montReduceSigned(z, hi)
}

//montReduceSigned SOS algorithm; xHi must be at most 63 bits long. Last bit of xHi may be used as a sign bit
func (z *{{.ElementName}}) montReduceSigned(x *{{.ElementName}}, xHi uint64) {

	const qInvNegLsw uint64 = 0x87d20782e4866389
	const signBitRemover = ^signBitSelector
	neg := xHi & signBitSelector != 0
	//the SOS implementation requires that most significant bit is 0
	// Let X be xHi*r + x
	// note that if X is negative we would have initially stored it as 2⁶⁴ r + X
	xHi &= signBitRemover
	// with this a negative X is now represented as 2⁶³ r + X

	var t [Limbs + 3]uint64
	var C uint64

	m := x[0] * qInvNegLsw

	C = madd0(m, qElement[0], x[0])
	{{- range $i := .NbWordsIndexesNoZero}}
	C, t[{{$i}}] = madd2(m, qElement[{{$i}}], x[{{$i}}], C)
	{{- end}}

	// the high word of m * qElement[{{.NbWordsLastIndex}}] is at most 62 bits
	// x[{{.NbWordsLastIndex}}] + C is at most 65 bits (high word at most 1 bit)
	// Thus the resulting C will be at most 63 bits
	t[{{.NbWords}}] = xHi + C
	// xHi and C are 63 bits, therefore no overflow

	{{ $NbWordsIndexesNoZeroInnerLoop := .NbWordsIndexesNoZero}}
	{{- range $i := .NbWordsIndexesNoZeroNoLast}}
	{
		const i = {{$i}}
		m = t[i] * qInvNegLsw

		C = madd0(m, qElement[0], t[i+0])

		{{- range $j := $NbWordsIndexesNoZeroInnerLoop}}
		C, t[i + {{$j}}] = madd2(m, qElement[{{$j}}], t[i +  {{$j}}], C)
		{{- end}}

		t[i + Limbs] += C
	}
	{{- end}}
	{
		const i = {{.NbWordsLastIndex}}
		m := t[i] * qInvNegLsw

		C = madd0(m, qElement[0], t[i+0])
		{{- range $j := $.NbWordsIndexesNoZeroNoLast}}
		C, z[{{sub $j 1}}] = madd2(m, qElement[{{$j}}], t[i+{{$j}}], C)
		{{- end}}
		z[{{.NbWordsLastIndex}}], z[{{sub .NbWordsLastIndex 1}}] = madd2(m, qElement[{{.NbWordsLastIndex}}], t[i+{{.NbWordsLastIndex}}], C)
	}

    {{ template "reduce" . }}
	if neg {
		//We have computed ( 2⁶³ r + X ) r⁻¹ = 2⁶³ + X r⁻¹ instead
		var b uint64
		z[0], b = bits.Sub64(z[0], signBitSelector, 0)

		{{- range $i := .NbWordsIndexesNoZero}}
		z[{{$i}}], b = bits.Sub64(z[{{$i}}], 0, b)
		{{- end}}

		//very unlikely
		if b != 0 {
			// z[{{.NbWordsLastIndex}}] = -1
			//negative: add q
			const neg1 = 0xFFFFFFFFFFFFFFFF

			b = 0
			{{- range $i := .NbWordsIndexesNoLast}}
			z[{{$i}}], b = bits.Add64(z[{{$i}}], {{index $.Q $i}}, b)
			{{- end}}
			z[{{.NbWordsLastIndex}}], _ = bits.Add64(neg1, {{index $.Q $.NbWordsLastIndex}}, b)
		}

	}

}

// mulWSigned mul word signed (w/ montgomery reduction)
func (z *{{.ElementName}}) mulWSigned(x *{{.ElementName}}, y int64) {
	m := y >> 63
	_mulWGeneric(z, x, uint64((y ^ m) - m))
	//multiply by abs(y)
	if y < 0 {
		z.Neg(z)
	}
}

{{ end }}
`
