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
		D = C

		// m = t[0]n'[0] mod W
		m = t[0] * {{index $.all.QInverse 0}}

		// -----------------------------------
		// Second loop
		C = madd0(m, {{index $.all.Q 0}}, t[0])
		{{- range $i := $.all.NbWordsIndexesNoZero}}
			{{if eq $i $.all.NbWordsLastIndex}}
				C, t[{{sub $i 1}}] = madd3(m, {{index $.all.Q $i}}, t[{{$i}}], C, t[{{$.all.NbWords}}])
			{{else}}
				C, t[{{sub $i 1}}] = madd2(m, {{index $.all.Q $i}}, t[{{$i}}], C)
			{{- end}}
		{{- end}}
		t[{{sub $.all.NbWords 1}}], t[{{$.all.NbWords}}] = bits.Add64(D, C, 0)
	{{- end}}


	if t[{{$.all.NbWords}}] != 0 {
		// we need to reduce, we have a result on {{add 1 $.all.NbWords}} words
		var b uint64
		z[0], b = bits.Sub64(t[0], {{index $.all.Q 0}}, 0)
		{{- range $i := .all.NbWordsIndexesNoZero}}
			{{-  if eq $i $.all.NbWordsLastIndex}}
				z[{{$i}}], _ = bits.Sub64(t[{{$i}}], {{index $.all.Q $i}}, b)
			{{-  else  }}
				z[{{$i}}], b = bits.Sub64(t[{{$i}}], {{index $.all.Q $i}}, b)
			{{- end}}
		{{- end}}
		{{if $.NoReturn }}
		return
		{{else}}
		return z
		{{end}}
	}

	// copy t into z 
	{{- range $i := $.all.NbWordsIndexesFull}}
		z[{{$i}}] = t[{{$i}}]
	{{- end}}

{{ end }}
`
