package fp2

// {{ define }} statements only; this template might appear in multiple packages

const Inline = `
{{- define "fpInlineMulByNonResidue" }}
	{ // begin inline: set {{.out}} to ({{.in}}) * ({{.all.Fp2NonResidue}})
		{{- if eq $.all.Fp2NonResidue "5" }}
			buf := *({{.in}})
			({{.out}}).Double(&buf).Double({{$.out}}).AddAssign(&buf)
		{{- else if eq $.all.Fp2NonResidue "-1" }}
			({{.out}}).Neg({{.in}})
		{{- else if eq $.all.Fp2NonResidue "3" }}
			buf := *({{.in}})
			({{.out}}).Double(&buf).AddAssign(&buf)
		{{- else if eq .all.Fp2NonResidue "-4" }}
			buf := *({{.in}})
			({{.out}}).Double(&buf).Double({{.out}}).Neg({{.out}})
		{{- else }}
			// TODO not implemented
		{{- end }}
	} // end inline: set {{.out}} to ({{.in}}) * ({{.all.Fp2NonResidue}})
{{- end }}

{{- define "fpInlineMulByNonResidueInv" }}
	{ // begin inline: set {{.out}} to ({{.in}}) * ({{.all.Fp2NonResidue}})^{-1}
		{{- if eq $.all.Fp2NonResidue "5" }}
			nrinv := fp.Element{
				330620507644336508,
				9878087358076053079,
				11461392860540703536,
				6973035786057818995,
				8846909097162646007,
				104838758629667239,
			}
			({{$.out}}).Mul({{$.in}}, &nrinv)
		{{- else if eq $.all.Fp2NonResidue "-1" }}
			// TODO this should be a no-op when {{$.out}}=={{$.in}}
			// TODO uh, why is -1 inverse equal to +1???
			({{$.out}}).Set({{$.in}})
		{{- else if eq $.all.Fp2NonResidue "3" }}
			nrinv := fp.Element{
				12669921578670009932,
				16188407930212075331,
				13036317521149659693,
				1499583668832556317,
			}
			({{$.out}}).Mul(({{$.in}}), &nrinv)
		{{- else }}
			// TODO not implemented
		{{- end }}
	} // end inline: set {{.out}} to ({{.in}}) * ({{.all.Fp2NonResidue}})^{-1}
{{- end }}
`
