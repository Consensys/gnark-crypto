package element

// TODO this should  be changed to match ASM constant time version
const Reduce = `
{{ define "reduce" }}
// if z > q --> z -= q
// note: this is NOT constant time
if !({{- range $i := reverse .NbWordsIndexesNoZero}} z[{{$i}}] < {{index $.Q $i}} || ( z[{{$i}}] == {{index $.Q $i}} && (
{{- end}}z[0] < {{index $.Q 0}} {{- range $i :=  .NbWordsIndexesNoZero}} )) {{- end}} ){
	var b uint64
	z[0], b = bits.Sub64(z[0], {{index $.Q 0}}, 0)
	{{- range $i := .NbWordsIndexesNoZero}}
		{{-  if eq $i $.NbWordsLastIndex}}
			z[{{$i}}], _ = bits.Sub64(z[{{$i}}], {{index $.Q $i}}, b)
		{{-  else  }}
			z[{{$i}}], b = bits.Sub64(z[{{$i}}], {{index $.Q $i}}, b)
		{{- end}}
	{{- end}}
}
{{-  end }}

`
