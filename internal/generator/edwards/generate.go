package edwards

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	return bgen.GenerateF(conf, "twistededwards", "./edwards/template", bavard.EntryF{
		File: filepath.Join(baseDir, "point.go"), TemplateF: []string{"pointtwistededwards.go.tmpl"},
	})

}
