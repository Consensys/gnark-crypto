package element

const MulCIOS = `
{{ define "mul_cios" }}
	var t [{{add .all.NbWords 1}}]uint64
	var D uint64
	var m, C uint64

	{{- range $j := .all.NbWordsIndexesFull}}
		// -----------------------------------
		// First loop
		{{ if eq $j 0}}
			C, t[0] = bits.Mul64({{$.V2}}[{{$j}}], {{$.V1}}[0])
			{{- range $i := $.all.NbWordsIndexesNoZero}}
				C, t[{{$i}}] = madd1({{$.V2}}[{{$j}}], {{$.V1}}[{{$i}}], C)
			{{- end}}
		{{ else }}
			C, t[0] = madd1({{$.V2}}[{{$j}}], {{$.V1}}[0], t[0])
			{{- range $i := $.all.NbWordsIndexesNoZero}}
				C, t[{{$i}}] = madd2({{$.V2}}[{{$j}}], {{$.V1}}[{{$i}}], t[{{$i}}], C)
			{{- end}}
		{{ end }}
		t[{{$.all.NbWords}}], D = bits.Add64(t[{{$.all.NbWords}}], C, 0)

		// m = t[0]n'[0] mod W
		m = t[0] * qInvNeg

		// -----------------------------------
		// Second loop
		C = madd0(m, q0, t[0])
		{{- range $i := $.all.NbWordsIndexesNoZero}}
				C, t[{{sub $i 1}}] = madd2(m, q{{$i}}, t[{{$i}}], C)
		{{- end}}

		 t[{{sub $.all.NbWords 1}}], C = bits.Add64(t[{{$.all.NbWords}}], C, 0)
		 t[{{$.all.NbWords}}], _ = bits.Add64(0, D, C)
	{{- end}}


	if t[{{$.all.NbWords}}] != 0 {
		// we need to reduce, we have a result on {{add 1 $.all.NbWords}} words
		{{- if gt $.all.NbWords 1}}
		var b uint64
		{{- end}}
		z[0], {{- if gt $.all.NbWords 1}}b{{- else}}_{{- end}} = bits.Sub64(t[0], q0, 0)
		{{- range $i := .all.NbWordsIndexesNoZero}}
			{{-  if eq $i $.all.NbWordsLastIndex}}
				z[{{$i}}], _ = bits.Sub64(t[{{$i}}], q{{$i}}, b)
			{{-  else  }}
				z[{{$i}}], b = bits.Sub64(t[{{$i}}], q{{$i}}, b)
			{{- end}}
		{{- end}}
		return
	}

	// copy t into z 
	{{- range $i := $.all.NbWordsIndexesFull}}
		z[{{$i}}] = t[{{$i}}]
	{{- end}}

{{ end }}

{{ define "mul_cios_one_limb" }}
	// In fact, since the modulus R fits on one register, the CIOS algorithm gets reduced to standard REDC (textbook Montgomery reduction):
	// hi, lo := x * y
	// m := (lo * qInvNeg) mod R
	// (*) r := (hi * R + lo + m * q) / R
	// reduce r if necessary

	// On the emphasized line, we get r = hi + (lo + m * q) / R
	// If we write hi2, lo2 = m * q then R | m * q - lo2 ⇒ R | (lo * qInvNeg) q - lo2 = -lo - lo2
	// This shows lo + lo2 = 0 mod R. i.e. lo + lo2 = 0 if lo = 0 and R otherwise.
	// Which finally gives (lo + m * q) / R = (lo + lo2 + R hi2) / R = hi2 + (lo+lo2) / R = hi2 + (lo != 0)
	// This "optimization" lets us do away with one MUL instruction on ARM architectures and is available for all q < R.

	var r uint64
	hi, lo := bits.Mul64({{$.V1}}[0], {{$.V2}}[0])
	if lo != 0 {
		hi++ // x[0] * y[0] ≤ 2¹²⁸ - 2⁶⁵ + 1, meaning hi ≤ 2⁶⁴ - 2 so no need to worry about overflow
	}
	m := lo * qInvNeg
	hi2, _ := bits.Mul64(m, q)
	r, carry := bits.Add64(hi2, hi, 0)

	if carry != 0 || r >= q  {
		// we need to reduce
		r -= q 
	}
	z[0] = r 
{{ end }}
`
