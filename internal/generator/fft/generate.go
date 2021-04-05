package fft

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	doc := "provides in-place discrete Fourier transform"
	entries := []bavard.EntryF{
		{File: filepath.Join(baseDir, "domain_test.go"), TemplateF: []string{"tests/domain.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(baseDir, "domain.go"), TemplateF: []string{"domain.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(baseDir, "fft_test.go"), TemplateF: []string{"tests/fft.go.tmpl", "imports.go.tmpl"}},
		{File: filepath.Join(baseDir, "fft.go"), TemplateF: []string{"fft.go.tmpl", "imports.go.tmpl"}, PackageDoc: doc},
	}
	return bgen.GenerateF(conf, "fft", "./fft/template/", entries...)
}
