package mimc

import (
	"fmt"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/crypto/hash/mimc/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	conf.Package = "mimc"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "mimc.go"), Templates: []string{"mimc.go.tmpl"}},
		{File: filepath.Join(baseDir, "options.go"), Templates: []string{"options.go.tmpl"}},
	}
	entriesTest := []bavard.Entry{
		{File: filepath.Join(baseDir, "mimc_test.go"), Templates: []string{"tests/mimc_test.go.tmpl"}},
	}

	mimcGen := common.NewGenerator(template.FS, "Consensys Software Inc.", 2020, "consensys/gnark-crypto")
	if err := mimcGen.Generate(conf, conf.Package, "", "", entries...); err != nil {
		return fmt.Errorf("generate package: %w", err)
	}
	if err := mimcGen.Generate(conf, "mimc_test", "", "", entriesTest...); err != nil {
		return fmt.Errorf("generate tests: %w", err)
	}
	return nil
}
