package gkr

import (
	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial"
	"github.com/consensys/gnark-crypto/internal/generator/sumcheck"
	"path/filepath"
)

type Config struct {
	config.FieldDependency
	GenerateLargeTests      bool
	TestVectorsRelativePath string
}

func Generate(config Config, baseDir string, bgen *bavard.BatchGenerator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "gkr.go"), Templates: []string{"gkr.go.tmpl"}},
		{File: filepath.Join(baseDir, "gkr_test.go"), Templates: []string{"gkr.test.go.tmpl", "gkr.test.utils.go.tmpl"}},
	}

	return bgen.Generate(config, "gkr", "./gkr/template/", entries...)
}

func GenerateForRationals(bgen *bavard.BatchGenerator) error {

	conf := Config{
		FieldDependency: config.FieldDependency{
			FieldPackagePath: "github.com/consensys/gnark-crypto/internal/generator/gkr/small_rational",
			FieldPackageName: "small_rational",
			ElementType:      "small_rational.SmallRational",
		},
		GenerateLargeTests:      false,
		TestVectorsRelativePath: "../../rational_cases",
	}

	baseDir := "./gkr/small_rational/"
	if err := polynomial.Generate(conf.FieldDependency, baseDir+"polynomial", false, bgen); err != nil {
		return err
	}

	if err := sumcheck.Generate(conf.FieldDependency, baseDir+"sumcheck", bgen); err != nil {
		return err
	}

	if err := Generate(conf, baseDir+"gkr", bgen); err != nil {
		return err
	}

	return nil
}
