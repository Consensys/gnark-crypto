package polynomial

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gurvy/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	entries := []bavard.EntryF{
		{File: filepath.Join(baseDir, "polynomial.go"), TemplateF: []string{"polynomial.go.tmpl"}},
	}
	return bgen.GenerateF(conf, "polynomial", "./polynomial/template/", entries...)
}
