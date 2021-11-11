package element

const FixedExp = `

{{- if .SqrtQ3Mod4}}
	{{expByAddChain "SqrtExp" .SqrtQ3Mod4ExponentData}}
{{- else if .SqrtAtkin}}
	{{expByAddChain "SqrtExp" .SqrtAtkinExponentData}}
{{- else if .SqrtTonelliShanks}}
	{{expByAddChain "SqrtExp" .SqrtSMinusOneOver2Data}}
{{- end }}

{{expByAddChain "LegendreExp" .LegendreExponentData}}


{{define "expByAddChain name data"}}
	
// expBy{{.name}} is equivalent to z.Exp(x, {{ .data.N }})
// 
// uses {{ .data.Meta.Module }} {{ .data.Meta.ReleaseTag }} to generate a shorter addition chain
func (z *Element) expBy{{.name}}(x *Element) *Element {
	// addition chain:
	//
	{{- range lines_ (format_ .data.Script) }}
	//	{{ . }}
	{{- end }}
	//
	// Operations: {{ .data.Ops.Doubles }} squares {{ .data.Ops.Adds }} multiplies

	// Allocate Temporaries.
	var {{range $i, $e := .data.Program.Temporaries }}{{ $e }} {{- if last_ $i $.data.Program.Temporaries}} Element {{- else }}, {{- end}}{{- end -}}

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
		{{ $i.Output }}.Square({{ ptr_ .X }}{{ $i.Output }})
	}
	{{ end -}}
	{{- end }}
	return z
}

{{end}}





`
