package plookup

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/plookup/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	// kzg commitment scheme
	conf.Package = "plookup"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "vector.go"), Templates: []string{"vector.go.tmpl"}},
		{File: filepath.Join(baseDir, "table.go"), Templates: []string{"table.go.tmpl"}},
		{File: filepath.Join(baseDir, "plookup_test.go"), Templates: []string{"plookup.test.go.tmpl"}},
	}
	plookupGen := common.NewGenerator(template.FS, "Consensys Software Inc.", 2020, "consensys/gnark-crypto")
	return plookupGen.Generate(conf, conf.Package, "", "", entries...)

}
