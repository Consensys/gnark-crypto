package tower

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/tower/asm/amd64"
	"github.com/consensys/gnark-crypto/internal/generator/tower/template"
)

// Generate generates a tower 2->6->12 over fp
func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	if conf.Equal(config.BW6_761) || conf.Equal(config.BW6_633) || conf.Equal(config.BLS24_315) || conf.Equal(config.BLS24_317) {
		return nil
	}

	towerGen := common.NewDefaultGenerator(template.FS)

	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "e2_amd64.go"), Templates: []string{"fq12over6over2/amd64.fq2.go.tmpl"}},
		{File: filepath.Join(baseDir, "e2_fallback.go"), Templates: []string{"fq12over6over2/fallback.fq2.go.tmpl"}, BuildTag: "!amd64"},
		{File: filepath.Join(baseDir, "asm.go"), Templates: []string{"fq12over6over2/asm.go.tmpl"}, BuildTag: "!noadx"},
		{File: filepath.Join(baseDir, "asm_noadx.go"), Templates: []string{"fq12over6over2/asm_noadx.go.tmpl"}, BuildTag: "noadx"},
	}

	if err := towerGen.Generate(conf, "fptower", "", "", entries...); err != nil {
		return err
	}

	towerConfs := []towerConf{
		{
			Curve:           &conf,
			RecursionDegree: 2,
			TotalDegree:     2,
			BaseName:        "fp.Element",
			BaseElementName: "A",
		},
		{
			Curve:           &conf,
			RecursionDegree: 3,
			TotalDegree:     6,
			BaseName:        "E2",
			BaseElementName: "B",
		},
		{
			Curve:           &conf,
			RecursionDegree: 2,
			TotalDegree:     12,
			BaseName:        "E6",
			BaseElementName: "C",
		},
	}

	for _, towerConf := range towerConfs {

		entries = []bavard.Entry{
			{
				File:      filepath.Join(baseDir, fmt.Sprintf("e%d.go", towerConf.TotalDegree)),
				Templates: []string{fmt.Sprintf("fq12over6over2/fq%d.go.tmpl", towerConf.TotalDegree), "fq12over6over2/base.go.tmpl"},
			},
			{
				File:      filepath.Join(baseDir, fmt.Sprintf("e%d_test.go", towerConf.TotalDegree)),
				Templates: []string{fmt.Sprintf("fq12over6over2/tests/fq%d.go.tmpl", towerConf.TotalDegree), "fq12over6over2/tests/base.go.tmpl"},
			},
		}

		if err := towerGen.Generate(towerConf, "fptower", "", "", entries...); err != nil {
			return err
		}
	}

	{
		// fq2 assembly
		fName := filepath.Join(baseDir, "e2_amd64.s")
		f, err := os.Create(fName)
		if err != nil {
			return err
		}

		Fq2Amd64 := amd64.NewFq2Amd64(f, conf.Fp, conf)
		if err := Fq2Amd64.Generate(true); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()

	}

	if conf.Equal(config.BN254) || conf.Equal(config.BLS12_381) {
		{
			// fq2 assembly
			fName := filepath.Join(baseDir, "e2_adx_amd64.s")
			os.Remove(fName)
		}
	}

	return nil

}

type towerConf struct {
	Curve           *config.Curve
	RecursionDegree int
	TotalDegree     int
	BaseName        string
	BaseElementName string
}
