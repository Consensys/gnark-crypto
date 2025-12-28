package templates

import "embed"

// FS contains all templates for the field generator
//
//go:embed element/*.go.tmpl element/*.s.tmpl extensions/*.go.tmpl fft/*.go.tmpl fft/tests/*.go.tmpl iop/*.go.tmpl poseidon2/*.go.tmpl sis/*.go.tmpl
var FS embed.FS
