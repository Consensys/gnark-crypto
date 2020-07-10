package fp12

// {{ define }} statements only; this template might appear in multiple packages

const Inline = `
{{- define "fp6InlineMulByNonResidue" }}
	{ // begin inline: set {{.out}} to ({{.in}}) * ((0,0),(1,0),(0,0))
		var result {{.all.Fp6Name}}
		result.B1.Set(&({{.in}}).B0)
		result.B2.Set(&({{.in}}).B1)
		{{- template "fp2InlineMulByNonResidue" dict "all" .all "out" "result.B0" "in" (print "&(" .in ").B2") }}
		{{.out}}.Set(&result)
	} // end inline: set {{.out}} to ({{.in}}) * ((0,0),(1,0),(0,0))
{{- end }}
`
