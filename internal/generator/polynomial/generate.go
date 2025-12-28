package polynomial

import (
	"path/filepath"

	"github.com/consensys/gnark-crypto/field/generator/common"
	"github.com/consensys/gnark-crypto/field/generator/config"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial/template"
)

func Generate(conf config.FieldDependency, baseDir string, generateTests bool, gen *common.Generator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "polynomial.go"), Templates: []string{"polynomial.go.tmpl"}},
		{File: filepath.Join(baseDir, "multilin.go"), Templates: []string{"multilin.go.tmpl"}},
		{File: filepath.Join(baseDir, "pool.go"), Templates: []string{"pool.go.tmpl"}},
	}

	if generateTests {
		entries = append(entries,
			bavard.Entry{File: filepath.Join(baseDir, "polynomial_test.go"), Templates: []string{"polynomial.test.go.tmpl"}},
			bavard.Entry{File: filepath.Join(baseDir, "multilin_test.go"), Templates: []string{"multilin.test.go.tmpl"}},
		)
	}

	polyGen := common.NewGenerator(template.FS, "Consensys Software Inc.", 2020, "consensys/gnark-crypto")
	return polyGen.Generate(conf, "polynomial", "", "", entries...)
}
