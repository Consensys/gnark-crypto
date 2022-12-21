package permutation

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	if conf.Equal(config.SECP256K1) || conf.Equal(config.SECQ256K1) {
		return nil
	}

	// permutation data
	conf.Package = "permutation"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "permutation.go"), Templates: []string{"permutation.go.tmpl"}},
		{File: filepath.Join(baseDir, "permutation_test.go"), Templates: []string{"permutation.test.go.tmpl"}},
	}
	return bgen.Generate(conf, conf.Package, "./permutation/template/", entries...)

}
