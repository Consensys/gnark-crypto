package edwards

import (
	"fmt"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	aka := " "
	if conf.ID() == ecc.BN254 {
		aka += " (a.k.a Baby JubJub) "
	} else if conf.ID() == ecc.BLS12_381 {
		aka += " (a.k.a JubJub) "
	}
	doc := fmt.Sprintf("provides %s twisted edwards \"companion curve\"%sdefined on fr.", conf.Name, aka)

	return bgen.GenerateF(conf, "twistededwards", "./edwards/template", bavard.EntryF{
		File: filepath.Join(baseDir, "point.go"), TemplateF: []string{"pointtwistededwards.go.tmpl"}, PackageDoc: doc,
	})

}
