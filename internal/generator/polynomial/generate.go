package polynomial

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {

	conf.Package = "polynomial"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "polynomial.go"), Templates: []string{"polynomial.go.tmpl"}},
		{File: filepath.Join(baseDir, "multilin.go"), Templates: []string{"multilin.go.tmpl"}},
		{File: filepath.Join(baseDir, "pool.go"), Templates: []string{"pool.go.tmpl"}},
		{File: filepath.Join(baseDir, "polynomial_test.go"), Templates: []string{"polynomial.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "multilin_test.go"), Templates: []string{"multilin.test.go.tmpl"}},
	}
	return bgen.Generate(conf, conf.Package, "./polynomial/template/", entries...)
}
