package gkr

import (
	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"path/filepath"
)

func Generate(conf config.FieldDependency, baseDir string, bgen *bavard.BatchGenerator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "gkr.go"), Templates: []string{"gkr.go.tmpl"}},
		{File: filepath.Join(baseDir, "gkr_test.go"), Templates: []string{"gkr.test.go.tmpl"}},
	}
	return bgen.Generate(conf, "gkr", "./gkr/template/", entries...)
}

func GenerateForRationals(bgen *bavard.BatchGenerator) error {
	/*conf := config.FieldDependency{
		FieldPackagePath: "./../utils",
		ElementType:      "",
	}
	bgen.Generate()*/return nil
}
