package ecdsa

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/ecdsa/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	// ecdsa
	conf.Package = "ecdsa"
	baseDir = filepath.Join(baseDir, conf.Package)

	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "ecdsa.go"), Templates: []string{"ecdsa.go.tmpl"}},
		{File: filepath.Join(baseDir, "ecdsa_test.go"), Templates: []string{"ecdsa.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal_test.go"), Templates: []string{"marshal.test.go.tmpl"}},
	}
	ecdsaGen := common.NewDefaultGenerator(template.FS)
	return ecdsaGen.Generate(conf, conf.Package, "", "", entries...)

}
