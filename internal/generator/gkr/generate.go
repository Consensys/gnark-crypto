package gkr

import (
	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/sumcheck"
	"path/filepath"
)

func Generate(conf sumcheck.FieldInfo, baseDir string, bgen *bavard.BatchGenerator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "gkr.go"), Templates: []string{"gkr.go.tmpl"}},
		{File: filepath.Join(baseDir, "gkr_test.go"), Templates: []string{"gkr.test.go.tmpl"}},
	}
	return bgen.Generate(conf, "gkr", "./gkr/template/", entries...)
}
