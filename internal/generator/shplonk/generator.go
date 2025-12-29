package shplonk

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/shplonk/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	// kzg commitment scheme
	conf.Package = "shplonk"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "shplonk.go"), Templates: []string{"shplonk.go.tmpl"}},
		{File: filepath.Join(baseDir, "shplonk_test.go"), Templates: []string{"shplonk.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
		{File: filepath.Join(baseDir, "example_test.go"), Templates: []string{"example_test.go.tmpl"}},
		// {File: filepath.Join(baseDir, "utils.go"), Templates: []string{"utils.go.tmpl"}},
	}
	shplonkGen := common.NewDefaultGenerator(template.FS)
	return shplonkGen.Generate(conf, conf.Package, "", "", entries...)

}
