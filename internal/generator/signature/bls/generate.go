package bls

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	// bls
	conf.Package = "bls"
	baseDir = filepath.Join(baseDir, conf.Package)

	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "bls.go"), Templates: []string{"bls.go.tmpl"}},
		{File: filepath.Join(baseDir, "bls_test.go"), Templates: []string{"bls.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal_test.go"), Templates: []string{"marshal.test.go.tmpl"}},
	}
	return bgen.Generate(conf, conf.Package, "./signature/bls/template", entries...)

}
