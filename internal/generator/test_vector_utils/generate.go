package test_vector_utils

import (
	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/gkr"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial"
	"github.com/consensys/gnark-crypto/internal/generator/sumcheck"
	"path/filepath"
)

type Config struct {
	config.FieldDependency
	RandomizeMissingHashEntries bool
}

func GenerateRationals(bgen *bavard.BatchGenerator) error {
	conf := gkr.Config{
		FieldDependency: config.FieldDependency{
			FieldPackagePath: "github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational",
			FieldPackageName: "small_rational",
			ElementType:      "small_rational.SmallRational",
		},
		GenerateLargeTests:      false,
		TestVectorsRelativePath: "../../../gkr/test_vectors",
	}

	baseDir := "./test_vector_utils/small_rational/"
	if err := polynomial.Generate(conf.FieldDependency, baseDir+"polynomial", false, bgen); err != nil {
		return err
	}
	if err := sumcheck.Generate(conf.FieldDependency, baseDir+"sumcheck", bgen); err != nil {
		return err
	}
	if err := gkr.Generate(conf, baseDir+"gkr", bgen); err != nil {
		return err
	}
	if err := Generate(Config{conf.FieldDependency, true}, baseDir+"test_vector_utils", bgen); err != nil {
		return err
	}

	return nil
}

func Generate(conf Config, baseDir string, bgen *bavard.BatchGenerator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "test_vector_utils.go"), Templates: []string{"map_hash.go.tmpl"}},
		//{File: filepath.Join(baseDir, "gkr_test.go"), Templates: []string{"gkr.test.go.tmpl", "gkr.test.utils.go.tmpl"}},
	}

	return bgen.Generate(conf, "test_vector_utils", "./test_vector_utils/template/", entries...)
}
