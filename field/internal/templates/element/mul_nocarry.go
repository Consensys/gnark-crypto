package element

// MulNoCarry see https://hackmd.io/@gnark/modular_multiplication for more info on the algorithm
const MulNoCarry = `
{{ define "mul_nocarry" }}
var t [{{.all.NbWords}}]uint64
var c [3]uint64
{{- range $j := .all.NbWordsIndexesFull}}
{
	// round {{$j}}
	v := {{$.V1}}[{{$j}}]
	{{- if eq $j $.all.NbWordsLastIndex}}
		c[1], c[0] = madd1(v, {{$.V2}}[0], t[0])
		m := c[0] * qInvNeg
		c[2] = madd0(m, q0, c[0])
		{{- if eq $.all.NbWords 1}}
			z[0], _ = madd3(m, q0, c[0], c[2], c[1])
		{{- else}}
			{{- range $i := $.all.NbWordsIndexesNoZero}}
				c[1], c[0] = madd2(v, {{$.V2}}[{{$i}}],  c[1], t[{{$i}}])
				{{- if eq $i $.all.NbWordsLastIndex}}
					z[{{sub $.all.NbWords 1}}], z[{{sub $i 1}}] = madd3(m, q{{$i}}, c[0], c[2], c[1])
				{{- else}}
					c[2], z[{{sub $i 1}}] = madd2(m, q{{$i}},  c[2], c[0])
				{{- end}}
			{{- end}}
		{{- end}}
	{{- else if eq $j 0}}
		c[1], c[0] = bits.Mul64(v, {{$.V2}}[0])
		m := c[0] * qInvNeg
		c[2] = madd0(m, q0, c[0])
		{{- range $i := $.all.NbWordsIndexesNoZero}}
			c[1], c[0] = madd1(v, {{$.V2}}[{{$i}}], c[1])
			{{- if eq $i $.all.NbWordsLastIndex}}
				t[{{sub $.all.NbWords 1}}], t[{{sub $i 1}}]  = madd3(m, q{{$i}}, c[0], c[2], c[1])
			{{- else}}
				c[2], t[{{sub $i 1}}] = madd2(m, q{{$i}}, c[2], c[0])
			{{- end}}
		{{- end}}
	{{- else}}
		c[1], c[0] = madd1(v, {{$.V2}}[0], t[0])
		m := c[0] * qInvNeg
		c[2] = madd0(m, q0, c[0])
		{{- range $i := $.all.NbWordsIndexesNoZero}}
			c[1], c[0] = madd2(v, {{$.V2}}[{{$i}}], c[1], t[{{$i}}])
			{{- if eq $i $.all.NbWordsLastIndex}}
				t[{{sub $.all.NbWords 1}}], t[{{sub $i 1}}] = madd3(m, q{{$i}}, c[0], c[2], c[1])
			{{- else}}
				c[2], t[{{sub $i 1}}] = madd2(m, q{{$i}}, c[2], c[0])
			{{- end}}
		{{- end}}
	{{-  end }}
}
{{- end}}
{{ end }}

`
