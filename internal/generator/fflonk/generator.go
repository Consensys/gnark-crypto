package fflonk

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/fflonk/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	// kzg commitment scheme
	conf.Package = "fflonk"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "fflonk.go"), Templates: []string{"fflonk.go.tmpl"}},
		{File: filepath.Join(baseDir, "fflonk_test.go"), Templates: []string{"fflonk.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
		{File: filepath.Join(baseDir, "example_test.go"), Templates: []string{"example_test.go.tmpl"}},
	}
	fflonkGen := common.NewGenerator(template.FS, "Consensys Software Inc.", 2020, "consensys/gnark-crypto")
	return fflonkGen.Generate(conf, conf.Package, "", "", entries...)

}
