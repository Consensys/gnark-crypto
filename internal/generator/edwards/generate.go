package edwards

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/addchain"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/edwards/template"
)

func Generate(conf config.TwistedEdwardsCurve, baseDir string, gen *common.Generator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "point.go"), Templates: []string{"point.go.tmpl"}},
		{File: filepath.Join(baseDir, "subgroup.go"), Templates: []string{"subgroup.go.tmpl"}},
		{File: filepath.Join(baseDir, "point_test.go"), Templates: []string{"tests/point.go.tmpl"}},
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "curve.go"), Templates: []string{"curve.go.tmpl"}},
	}

	edwardsGen := common.NewDefaultGenerator(template.FS)
	funcs := common.Funcs()
	for _, f := range addchain.Functions {
		funcs[f.Name] = f.Func
	}
	return edwardsGen.GenerateWithOptions(conf, conf.Package, "", "", []func(*bavard.Bavard) error{
		bavard.Funcs(funcs),
	}, entries...)
}
