package tower

import (
	"io"
	"os"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/tower/asm/amd64"
)

// Generate generates a tower 2->6->12 over fp
func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	if conf.ID() == ecc.BW6_761 {
		return nil
	}

	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "e2.go"), Templates: []string{"fq2.go.tmpl"}},
		{File: filepath.Join(baseDir, "e6.go"), Templates: []string{"fq6.go.tmpl"}},
		{File: filepath.Join(baseDir, "e12.go"), Templates: []string{"fq12.go.tmpl"}},
		{File: filepath.Join(baseDir, "e2_amd64.go"), Templates: []string{"amd64.fq2.go.tmpl"}},
		{File: filepath.Join(baseDir, "e2_fallback.go"), Templates: []string{"fallback.fq2.go.tmpl"}, BuildTag: "!amd64"},
		{File: filepath.Join(baseDir, "e2_test.go"), Templates: []string{"tests/fq2.go.tmpl"}},
		{File: filepath.Join(baseDir, "e6_test.go"), Templates: []string{"tests/fq6.go.tmpl"}},
		{File: filepath.Join(baseDir, "e12_test.go"), Templates: []string{"tests/fq12.go.tmpl"}},
		{File: filepath.Join(baseDir, "asm.go"), Templates: []string{"asm.go.tmpl"}, BuildTag: "!noadx"},
		{File: filepath.Join(baseDir, "asm_noadx.go"), Templates: []string{"asm_noadx.go.tmpl"}, BuildTag: "noadx"},
	}

	if err := bgen.Generate(conf, "fptower", "./tower/template/fq12over6over2", entries...); err != nil {
		return err
	}

	{
		// fq2 assembly
		fName := filepath.Join(baseDir, "e2_amd64.s")
		f, err := os.Create(fName)
		if err != nil {
			return err
		}

		if conf.ID() == ecc.BN254 || conf.ID() == ecc.BLS12_381 {
			_, _ = io.WriteString(f, "// +build !amd64_adx\n")
		}
		Fq2Amd64 := amd64.NewFq2Amd64(f, conf.Fp, conf)
		if err := Fq2Amd64.Generate(true); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()

	}

	if conf.ID() == ecc.BN254 || conf.ID() == ecc.BLS12_381 {
		{
			// fq2 assembly
			fName := filepath.Join(baseDir, "e2_adx_amd64.s")
			f, err := os.Create(fName)
			if err != nil {
				return err
			}

			_, _ = io.WriteString(f, "// +build amd64_adx\n")
			Fq2Amd64 := amd64.NewFq2Amd64(f, conf.Fp, conf)
			if err := Fq2Amd64.Generate(false); err != nil {
				_ = f.Close()
				return err
			}
			_ = f.Close()

		}
	}

	return nil

}
