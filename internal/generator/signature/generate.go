package signature

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	conf.Package = "signature"

	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "utils.go"), Templates: []string{"utils.go.tmpl"}},
	}

	return bgen.Generate(conf, conf.Package, "./signature/template", entries...)
}
