package fp6

// {{ define }} statements only; this template might appear in multiple packages

const Inline = `
{{- define "fp2InlineMulByNonResidue" }}
	{ // begin inline: set {{.out}} to ({{.in}}) * ({{.all.Fp6NonResidue}})
		{{- if eq .all.Fp6NonResidue "0,1" }}
			buf := ({{.in}}).A0
			{{- template "fpInlineMulByNonResidue" dict "all" .all "out" (print "&(" .out ").A0") "in" (print "&(" .in ").A1") }}
			({{.out}}).A1 = buf
		{{- else if eq .all.Fp6NonResidue "1,1"}}
			var buf {{.all.Fp2Name}}
			buf.Set({{.in}})
			{{.out}}.A1.Add(&buf.A0, &buf.A1)
			{{- template "fpInlineMulByNonResidue" dict "all" .all "out" (print "&(" .out ").A0") "in" "&buf.A1" }}
			{{.out}}.A0.AddAssign(&buf.A0)
		{{- else if eq .all.Fp6NonResidue "9,1"}}
			var buf, buf9 {{.all.Fp2Name}}
			buf.Set({{.in}})
			buf9.Double(&buf).
				Double(&buf9).
				Double(&buf9).
				Add(&buf9, &buf)
			{{.out}}.A1.Add(&buf.A0, &buf9.A1)
			{{- template "fpInlineMulByNonResidue" dict "all" .all "out" (print "&(" .out ").A0") "in" "&buf.A1" }}
			{{.out}}.A0.AddAssign(&buf9.A0)
		{{- else}}
			panic("not implemented")
		{{- end }}
	} // end inline: set {{.out}} to ({{.in}}) * ({{.all.Fp6NonResidue}})
{{- end }}
`
