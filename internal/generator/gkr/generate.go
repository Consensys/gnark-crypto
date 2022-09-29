package gkr

import (
	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial"
	"github.com/consensys/gnark-crypto/internal/generator/sumcheck"
	"path/filepath"
)

func Generate(conf config.FieldDependency, baseDir string, generateTests bool, bgen *bavard.BatchGenerator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "gkr.go"), Templates: []string{"gkr.go.tmpl"}},
	}

	if generateTests {
		entries = append(entries, bavard.Entry{File: filepath.Join(baseDir, "gkr_test.go"), Templates: []string{"gkr.test.go.tmpl"}})
	}

	return bgen.Generate(conf, "gkr", "./gkr/template/", entries...)
}

func GenerateForRationals(bgen *bavard.BatchGenerator) error {

	conf := config.FieldDependency{
		FieldPackagePath: "github.com/consensys/gnark-crypto/internal/generator/gkr/small_rational",
		FieldPackageName: "small_rational",
		ElementType:      "small_rational.SmallRational",
	}

	baseDir := "./gkr/small_rational/"
	if err := polynomial.Generate(conf, baseDir+"polynomial", false, bgen); err != nil {
		return err
	}

	if err := sumcheck.Generate(conf, baseDir+"sumcheck", bgen); err != nil {
		return err
	}

	if err := Generate(conf, baseDir+"gkr", false, bgen); err != nil {
		return err
	}
	/*conf := config.FieldDependency{
		FieldPackagePath: "./../utils",
		ElementType:      "",
	}
	bgen.Generate()*/return nil
}
