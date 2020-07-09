package fp2

// {{ define }} statements only; this template might appear in multiple packages

const Inline = `
{{- define "fpInlineMulByNonResidue" }}
	{ // begin inline: set {{.out}} to ({{.in}}) * ({{.all.Fp2NonResidue}})
		{{- if eq .all.Fp2NonResidue "5" }}
			buf := *({{.in}})
			({{.out}}).Double(&buf).Double({{$.out}}).AddAssign(&buf)
		{{- else if eq .all.Fp2NonResidue "-1" }}
			({{.out}}).Neg({{.in}})
		{{- else if eq .all.Fp2NonResidue "3" }}
			buf := *({{.in}})
			({{.out}}).Double(&buf).AddAssign(&buf)
		{{- else if eq .all.Fp2NonResidue "-4" }}
			buf := *({{.in}})
			({{.out}}).Double(&buf).Double({{.out}}).Neg({{.out}})
		{{- else }}
			panic("not implemented")
		{{- end }}
	} // end inline: set {{.out}} to ({{.in}}) * ({{.all.Fp2NonResidue}})
{{- end }}
`
