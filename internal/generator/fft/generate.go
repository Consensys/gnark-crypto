package fft

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {

	conf.Package = "fft"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "domain_test.go"), Templates: []string{"tests/domain.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(baseDir, "domain.go"), Templates: []string{"domain.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(baseDir, "fft_test.go"), Templates: []string{"tests/fft.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(baseDir, "fft.go"), Templates: []string{"fft.go.tmpl", "imports.go.tmpl"}},
	}
	return bgen.Generate(conf, conf.Package, "./fft/template/", entries...)
}
