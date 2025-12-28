package pairing

import (
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/pairing/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	packageName := strings.ReplaceAll(conf.Name, "-", "")
	pairingGen := common.NewGenerator(template.FS, "Consensys Software Inc.", 2020, "consensys/gnark-crypto")
	return pairingGen.Generate(conf, packageName, "", "", bavard.Entry{
		File: filepath.Join(baseDir, "pairing_test.go"), Templates: []string{"tests/pairing.go.tmpl"},
	})

}
