package pairing

import (
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	if conf.Equal(config.SECP256K1) {
		return nil
	}
	packageName := strings.ReplaceAll(conf.Name, "-", "")
	return bgen.Generate(conf, packageName, "./pairing/template", bavard.Entry{
		File: filepath.Join(baseDir, "pairing_test.go"), Templates: []string{"tests/pairing.go.tmpl"},
	})

}
