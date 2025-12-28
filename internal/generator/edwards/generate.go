package edwards

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/edwards/template"
)

func Generate(conf config.TwistedEdwardsCurve, baseDir string, gen *common.Generator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "point.go"), Templates: []string{"point.go.tmpl"}},
		{File: filepath.Join(baseDir, "point_test.go"), Templates: []string{"tests/point.go.tmpl"}},
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "curve.go"), Templates: []string{"curve.go.tmpl"}},
	}

	edwardsGen := common.NewGenerator(template.FS, "Consensys Software Inc.", 2020, "consensys/gnark-crypto")
	return edwardsGen.Generate(conf, conf.Package, "", "", entries...)
}
