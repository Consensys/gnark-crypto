package gkr

import (
	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial"
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

	conf := config.FieldDependency{
		FieldPackagePath: "github.com/consensys/gnark-crypto/internal/generator/gkr/rational_cases/small_rational",
		FieldPackageName: "small_rational",
		ElementType:      "small_rational.SmallRational",
	}

	if err := polynomial.Generate(conf, "./gkr/rational_cases/small_rational/polynomial", false, bgen); err != nil {
		return err
	}

	/*conf := config.FieldDependency{
		FieldPackagePath: "./../utils",
		ElementType:      "",
	}
	bgen.Generate()*/return nil
}
