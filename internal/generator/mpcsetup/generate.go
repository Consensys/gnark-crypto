package mpcsetup

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/mpcsetup/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "mpcsetup.go"), Templates: []string{"mpcsetup.go.tmpl"}},
		{File: filepath.Join(baseDir, "mpcsetup_test.go"), Templates: []string{"tests/mpcsetup.go.tmpl"}},
	}
	mpcGen := common.NewGenerator(template.FS, "Consensys Software Inc.", 2020, "consensys/gnark-crypto")
	return mpcGen.Generate(conf, "mpcsetup", "", "", entries...)
}
