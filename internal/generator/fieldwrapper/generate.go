package fieldwrapper

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/fieldwrapper/template"
)

func Generate(conf config.Curve, baseDir string) error {
	entry := bavard.Entry{
		File:      filepath.Join(baseDir, "fp.go"),
		Templates: []string{"fp.go.tmpl"},
	}
	gen := common.NewDefaultGenerator(template.FS)
	return gen.Generate(conf, "fp", "", "", entry)
}
