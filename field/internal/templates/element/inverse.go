package element

const Inverse = `

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

// Inverse z = x^-1 mod q 
// Algorithm 16 in "Efficient Software-Implementation of Finite Fields with Applications to Cryptography"
// if x == 0, sets and returns z = x 
func (z *{{.ElementName}}) Inverse(x *{{.ElementName}}) *{{.ElementName}} {
	if x.IsZero() {
		return z.Set(x)
	}

	// initialize u = q
	var u = {{.ElementName}}{
		{{- range $i := .NbWordsIndexesFull}}
		{{index $.Q $i}},{{end}}
	}

	// initialize s = r^2
	var s = {{.ElementName}}{
		{{- range $i := .RSquare}}
		{{$i}},{{end}}
	}

	// r = 0
	r := {{.ElementName}}{}

	v := *x

	var carry, borrow, t, t2 uint64
	var bigger bool

	for  {
		for v[0]&1 == 0 {
			{{ template "div2" dict "all" . "V" "v"}}
			if s[0]&1 == 1 {
				{{ template "add_q" dict "all" . "V1" "s" }}
			}
			{{ template "div2" dict "all" . "V" "s"}}
		} 
		for u[0]&1 == 0 {
			{{ template "div2" dict "all" . "V" "u"}}
			if r[0]&1 == 1 {
				{{ template "add_q" dict "all" . "V1" "r" }}
			}
			{{ template "div2" dict "all" . "V" "r"}}
		} 
		{{ template "bigger" dict "all" . "V1" "v" "V2" "u"}}
		if bigger  {
			{{ template "sub_noborrow" dict "all" . "V1" "v" "V2" "u" "OmitLast" "true"}}
			{{ template "sub_noborrow" dict "all" . "V1" "s" "V2" "r" "OmitLast" "false"}}
			if borrow == 1 {
				{{ template "add_q" dict "all" . "V1" "s" }}
			}
		} else {
			{{ template "sub_noborrow" dict "all" . "V1" "u" "V2" "v" "OmitLast" "true"}}
			{{ template "sub_noborrow" dict "all" . "V1" "r" "V2" "s" "OmitLast" "false"}}
			if borrow == 1 {
				{{ template "add_q" dict "all" . "V1" "r" }}
			}
		}
		if (u[0] == 1) && ({{- range $i := reverse .NbWordsIndexesNoZero}}u[{{$i}}] {{if eq $i 1}}{{else}} | {{end}}{{end}} ) == 0 {
			return z.Set(&r)
		}
		if (v[0] == 1) && ({{- range $i := reverse .NbWordsIndexesNoZero}}v[{{$i}}] {{if eq $i 1}}{{else}} | {{end}}{{end}} ) == 0 {
			return z.Set(&s)
		}
	}

}

{{ end }}




{{ define "bigger" }}
	// {{$.V1}} >= {{$.V2}}
	bigger = !({{- range $i := reverse $.all.NbWordsIndexesNoZero}} {{$.V1}}[{{$i}}] < {{$.V2}}[{{$i}}] || ( {{$.V1}}[{{$i}}] == {{$.V2}}[{{$i}}] && (
		{{- end}}{{$.V1}}[0] < {{$.V2}}[0] {{- range $i :=  $.all.NbWordsIndexesNoZero}} )) {{- end}} )
{{ end }}

{{ define "add_q" }}
	// {{$.V1}} = {{$.V1}} + q 
	{{$.V1}}[0], carry = bits.Add64({{$.V1}}[0], {{index $.all.Q 0}}, 0)
	{{- range $i := .all.NbWordsIndexesNoZero}}
		{{- if eq $i $.all.NbWordsLastIndex}}
			{{$.V1}}[{{$i}}], _ = bits.Add64({{$.V1}}[{{$i}}], {{index $.all.Q $i}}, carry)
		{{- else}}
			{{$.V1}}[{{$i}}], carry = bits.Add64({{$.V1}}[{{$i}}], {{index $.all.Q $i}}, carry)
		{{- end}}
	{{- end}}
{{ end }}

{{ define "sub_noborrow" }}
	// {{$.V1}} = {{$.V1}} - {{$.V2}}
	{{$.V1}}[0], borrow = bits.Sub64({{$.V1}}[0], {{$.V2}}[0], 0)
	{{- range $i := .all.NbWordsIndexesNoZero}}
		{{- if and (eq $i $.all.NbWordsLastIndex) (eq "true" $.OmitLast)}}
		{{$.V1}}[{{$i}}], _ = bits.Sub64({{$.V1}}[{{$i}}], {{$.V2}}[{{$i}}], borrow)
		{{- else}}
		{{$.V1}}[{{$i}}], borrow = bits.Sub64({{$.V1}}[{{$i}}], {{$.V2}}[{{$i}}], borrow)
		{{- end}}
	{{- end}}
{{ end }}


{{ define "div2" }}
	// {{$.V}} = {{$.V}} >> 1
	{{- range $i :=  reverse .all.NbWordsIndexesNoZero}}
		{{- if eq $i $.all.NbWordsLastIndex}}
			t2 = {{$.V}}[{{$i}}] << 63
			{{$.V}}[{{$i}}] >>= 1
		{{- else}}
			t2 = {{$.V}}[{{$i}}] << 63
			{{$.V}}[{{$i}}] = ({{$.V}}[{{$i}}] >> 1) | t
		{{- end}}
		t = t2
	{{- end}}
	{{$.V}}[0] = ({{$.V}}[0] >> 1) | t
{{ end }}


`
