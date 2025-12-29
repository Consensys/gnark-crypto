package kzg

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/kzg/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	// kzg commitment scheme
	conf.Package = "kzg"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "kzg.go"), Templates: []string{"kzg.go.tmpl"}},
		{File: filepath.Join(baseDir, "kzg_test.go"), Templates: []string{"kzg.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
		{File: filepath.Join(baseDir, "utils.go"), Templates: []string{"utils.go.tmpl"}},
		{File: filepath.Join(baseDir, "mpcsetup.go"), Templates: []string{"mpcsetup.go.tmpl"}},
	}
	kzgGen := common.NewDefaultGenerator(template.FS)
	return kzgGen.Generate(conf, conf.Package, "", "", entries...)

}
