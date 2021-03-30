package pairing

import (
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gurvy/ecc"
	"github.com/consensys/gurvy/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	if conf.ID() == ecc.BW6_761 {
		return nil
	}
	packageName := strings.ReplaceAll(conf.Name, "-", "")
	return bgen.GenerateF(conf, packageName, "./pairing/template", bavard.EntryF{
		File: filepath.Join(baseDir, "pairing_test.go"), TemplateF: []string{"tests/pairing.go.tmpl"},
	})

}
