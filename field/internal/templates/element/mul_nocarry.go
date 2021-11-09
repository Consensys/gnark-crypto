package element

// MulNoCarry see https://hackmd.io/@zkteam/modular_multiplication for more info on the algorithm
const MulNoCarry = `
{{ define "mul_nocarry" }}
var t [{{.all.NbWords}}]uint64
var c [3]uint64
{{- range $j := .all.NbWordsIndexesFull}}
{
	// round {{$j}}
	v := {{$.V1}}[{{$j}}]
	{{- if eq $j 0}}
		c[1], c[0] = bits.Mul64(v, {{$.V2}}[0])
		m := c[0] * {{index $.all.QInverse 0}}
		c[2] = madd0(m, {{index $.all.Q 0}}, c[0])
		{{- range $i := $.all.NbWordsIndexesNoZero}}
			c[1], c[0] = madd1(v, {{$.V2}}[{{$i}}], c[1])
			{{- if eq $i $.all.NbWordsLastIndex}}
				t[{{sub $.all.NbWords 1}}], t[{{sub $i 1}}]  = madd3(m, {{index $.all.Q $i}}, c[0], c[2], c[1])
			{{- else}}
				c[2], t[{{sub $i 1}}] = madd2(m, {{index $.all.Q $i}}, c[2], c[0])
			{{- end}}
		{{- end}}
	{{- else if eq $j $.all.NbWordsLastIndex}}
		c[1], c[0] = madd1(v, {{$.V2}}[0], t[0])
		m := c[0] * {{index $.all.QInverse 0}}
		c[2] = madd0(m, {{index $.all.Q 0}}, c[0])
		{{- range $i := $.all.NbWordsIndexesNoZero}}
			c[1], c[0] = madd2(v, {{$.V2}}[{{$i}}],  c[1], t[{{$i}}])
			{{- if eq $i $.all.NbWordsLastIndex}}
				z[{{sub $.all.NbWords 1}}], z[{{sub $i 1}}] = madd3(m, {{index $.all.Q $i}}, c[0], c[2], c[1])
			{{- else}}
				c[2], z[{{sub $i 1}}] = madd2(m, {{index $.all.Q $i}},  c[2], c[0])
			{{- end}}
		{{- end}}
	{{- else}}
		c[1], c[0] = madd1(v, {{$.V2}}[0], t[0])
		m := c[0] * {{index $.all.QInverse 0}}
		c[2] = madd0(m, {{index $.all.Q 0}}, c[0])
		{{- range $i := $.all.NbWordsIndexesNoZero}}
			c[1], c[0] = madd2(v, {{$.V2}}[{{$i}}], c[1], t[{{$i}}])
			{{- if eq $i $.all.NbWordsLastIndex}}
				t[{{sub $.all.NbWords 1}}], t[{{sub $i 1}}] = madd3(m, {{index $.all.Q $i}}, c[0], c[2], c[1])
			{{- else}}
				c[2], t[{{sub $i 1}}] = madd2(m, {{index $.all.Q $i}}, c[2], c[0])
			{{- end}}
		{{- end}}
	{{-  end }}
}
{{- end}}
{{ end }}




{{ define "mul_nocarry_v2" }}
var t [{{.all.NbWords}}]uint64

{{- range $j := .all.NbWordsIndexesFull}}
{
	// round {{$j}}
	
	{{- if eq $j 0}}
		c1, c0 := bits.Mul64(y, {{$.V2}}[0])
		m := c0 * {{index $.all.QInverse 0}}
		c2 := madd0(m, {{index $.all.Q 0}}, c0)
		{{- range $i := $.all.NbWordsIndexesNoZero}}
			c1, c0 = madd1(y, {{$.V2}}[{{$i}}], c1)
			{{- if eq $i $.all.NbWordsLastIndex}}
				t[{{sub $.all.NbWords 1}}], t[{{sub $i 1}}]  = madd3(m, {{index $.all.Q $i}}, c0, c2, c1)
			{{- else}}
				c2, t[{{sub $i 1}}] = madd2(m, {{index $.all.Q $i}}, c2, c0)
			{{- end}}
		{{- end}}
		{{- else if eq $j $.all.NbWordsLastIndex}}
			m := t[0] * {{index $.all.QInverse 0}}
			c2 := madd0(m, {{index $.all.Q 0}}, t[0])
			{{- range $i := $.all.NbWordsIndexesNoZero}}
				{{- if eq $i $.all.NbWordsLastIndex}}
					z[{{sub $.all.NbWords 1}}], z[{{sub $i 1}}] = madd2(m, {{index $.all.Q $i}}, t[{{$i}}], c2)
				{{- else}}
					c2, z[{{sub $i 1}}] = madd2(m, {{index $.all.Q $i}},  c2, t[{{$i}}])
				{{- end}}
			{{- end}}
		{{- else}}
			m := t[0] * {{index $.all.QInverse 0}}
			c2 := madd0(m, {{index $.all.Q 0}}, t[0])
			{{- range $i := $.all.NbWordsIndexesNoZero}}
				{{- if eq $i $.all.NbWordsLastIndex}}
					t[{{sub $.all.NbWords 1}}], t[{{sub $i 1}}] = madd2(m, {{index $.all.Q $i}}, t[{{$i}}], c2)
				{{- else}}
					c2, t[{{sub $i 1}}] = madd2(m, {{index $.all.Q $i}}, c2, t[{{$i}}])
				{{- end}}
			{{- end}}
		{{-  end }}
}
{{- end}}
{{ end }}
`
