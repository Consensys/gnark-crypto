package element

const FixedExp = `

{{- if .SqrtQ3Mod4}}
	{{expByAddChain "SqrtPp1o4" .SqrtQ3Mod4ExponentData .ElementName}}
	{{expByAddChain "SqrtPm3o4" .SqrtQ3Mod4ExponentData2 .ElementName}}
{{- else if .SqrtAtkin}}
	{{expByAddChain "SqrtPm5o8" .SqrtAtkinExponentData .ElementName}}
{{- else if .SqrtTonelliShanks}}
	{{expByAddChain "SqrtExp" .SqrtSMinusOneOver2Data .ElementName}}
{{- end }}

{{- if and (not .UsingP20Inverse) (not (eq .NbWords 1))}}
	{{expByAddChain "LegendreExp" .LegendreExponentData .ElementName}}
{{- end}}

{{define "expByAddChain name data eName"}}

// ExpBy{{.name}} is equivalent to z.Exp(x, {{ .data.N }}).
{{- if eq .name "SqrtPp1o4"}}
// It raises x to the (p+1)/4 power using a shorter addition chain.
{{- else if eq .name "SqrtPm3o4"}}
// It raises x to the (p-3)/4 power using a shorter addition chain.
{{- else if eq .name "SqrtPm5o8"}}
// It raises x to the (p-5)/8 power using a shorter addition chain.
{{- else if eq .name "SqrtExp"}}
// It raises x to the (p-2^s-1)/2^(s+1) power using a shorter addition chain,
// where s the 2-adic valuation of p-1.
{{- end }}
//
// uses {{ .data.Meta.Module }} {{ .data.Meta.ReleaseTag }} to generate a shorter addition chain
func (z *{{.eName}}) ExpBy{{$.name}}(x {{.eName}}) *{{.eName}} {
	// addition chain:
	//
	{{- range lines_ (format_ .data.Script) }}
	//	{{ . }}
	{{- end }}
	//
	// Operations: {{ .data.Ops.Doubles }} squares {{ .data.Ops.Adds }} multiplies

	// Allocate Temporaries.
	var (
		{{- range .data.Program.Temporaries }}
		{{ . }} = new({{$.eName}})
		{{- end -}}
	)

	// var {{range $i, $e := .data.Program.Temporaries }}{{ $e }} {{- if last_ $i $.data.Program.Temporaries}} {{$.eName}} {{- else }}, {{- end}}{{- end -}}

	{{ range $i := .data.Program.Instructions }}
	// {{ printf "Step %d: %s = x^%#x" $i.Output.Index $i.Output (index $.data.Chain $i.Output.Index) }}
	{{- with add_ $i.Op }}
	{{ $i.Output }}.Mul({{ ptr_ .X }}{{ .X }}, {{ ptr_ .Y }}{{ .Y }})
	{{ end -}}

	{{- with double_ $i.Op }}
	{{ $i.Output }}.Square({{ ptr_ .X }}{{ .X }})
	{{ end -}}

	{{- with shift_ $i.Op -}}
	{{- $first := 0 -}}
	{{- if ne $i.Output.Identifier .X.Identifier }}
	{{ $i.Output }}.Square({{ ptr_ .X }}{{ .X }})
	{{- $first = 1 -}}
	{{- end }}
	for s := {{ $first }}; s < {{ .S }}; s++ {
		{{ $i.Output }}.Square({{ ptr_ $i.Output }}{{ $i.Output }})
	}
	{{ end -}}
	{{- end }}
	return z
}

{{end}}





`
