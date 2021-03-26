package pairing

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gurvy/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	if conf.Name == "bw761" {
		return nil
	}
	return bgen.GenerateF(conf, conf.Name, "./pairing/template", bavard.EntryF{
		File: filepath.Join(baseDir, "pairing_test.go"), TemplateF: []string{"tests/pairing.go.tmpl"},
	})

}
