package eddsa

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {

	// eddsa
	entriesF := []bavard.EntryF{
		{File: filepath.Join(baseDir, "eddsa.go"), TemplateF: []string{"eddsa.go.tmpl"}},
		{File: filepath.Join(baseDir, "eddsa_test.go"), TemplateF: []string{"eddsa.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), TemplateF: []string{"marshal.go.tmpl"}},
	}
	return bgen.GenerateF(conf, "eddsa", "./crypto/signature/eddsa/template", entriesF...)

}
