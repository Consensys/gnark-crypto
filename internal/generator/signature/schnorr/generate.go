package schnorr

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	// schnorr
	conf.Package = "schnorr"
	baseDir = filepath.Join(baseDir, conf.Package)

	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "schnorr.go"), Templates: []string{"schnorr.go.tmpl"}},
		{File: filepath.Join(baseDir, "schnorr_test.go"), Templates: []string{"schnorr.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal_test.go"), Templates: []string{"marshal.test.go.tmpl"}},
	}
	return bgen.Generate(conf, conf.Package, "./signature/schnorr/template", entries...)

}
