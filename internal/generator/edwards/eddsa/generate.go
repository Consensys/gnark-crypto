package eddsa

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/edwards/eddsa/template"
)

func Generate(conf config.TwistedEdwardsCurve, baseDir string, gen *common.Generator) error {
	// eddsa
	conf.Package = "eddsa"
	baseDir = filepath.Join(baseDir, conf.Package)

	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "eddsa.go"), Templates: []string{"eddsa.go.tmpl"}},
		{File: filepath.Join(baseDir, "eddsa_test.go"), Templates: []string{"eddsa.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
	}
	eddsaGen := common.NewGenerator(template.FS, "Consensys Software Inc.", 2020, "consensys/gnark-crypto")
	return eddsaGen.Generate(conf, conf.Package, "", "", entries...)

}
