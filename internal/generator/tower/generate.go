package tower

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gurvy/internal/asm/amd64"
	"github.com/consensys/gurvy/internal/generator/config"
)

// Generate generates a tower 2->6->12 over fp
func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	if conf.Name == "bw761" {
		return nil
	}

	entries := []bavard.EntryF{
		{File: filepath.Join(baseDir, "e2.go"), TemplateF: []string{"fq2.go.tmpl"}},
		{File: filepath.Join(baseDir, "e6.go"), TemplateF: []string{"fq6.go.tmpl"}},
		{File: filepath.Join(baseDir, "e12.go"), TemplateF: []string{"fq12.go.tmpl"}},
		{File: filepath.Join(baseDir, "e2_amd64.go"), TemplateF: []string{"amd64.fq2.go.tmpl"}},
		{File: filepath.Join(baseDir, "e2_fallback.go"), TemplateF: []string{"fallback.fq2.go.tmpl"}, BuildTag: "!amd64"},
		{File: filepath.Join(baseDir, "e2_test.go"), TemplateF: []string{"tests/fq2.go.tmpl"}},
		{File: filepath.Join(baseDir, "e6_test.go"), TemplateF: []string{"tests/fq6.go.tmpl"}},
		{File: filepath.Join(baseDir, "e12_test.go"), TemplateF: []string{"tests/fq12.go.tmpl"}},
		{File: filepath.Join(baseDir, "asm.go"), TemplateF: []string{"asm.go.tmpl"}, BuildTag: "!noadx"},
		{File: filepath.Join(baseDir, "asm_noadx.go"), TemplateF: []string{"asm_noadx.go.tmpl"}, BuildTag: "noadx"},
	}

	if err := bgen.GenerateF(conf, "fptower", "./tower/template/fq12over6over2", entries...); err != nil {
		return err
	}

	{
		// fq2 assembly
		fName := filepath.Join(baseDir, "e2_amd64.s")
		f, err := os.Create(fName)
		if err != nil {
			return err
		}

		if conf.Name == "bn256" || conf.Name == "bls381" {
			_, _ = io.WriteString(f, "// +build !amd64_adx\n")
		}
		Fq2Amd64 := amd64.NewFq2Amd64(f, conf.Fp, conf.Name)
		if err := Fq2Amd64.Generate(true); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()

		cmd := exec.Command("asmfmt", "-w", fName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if conf.Name == "bn256" || conf.Name == "bls381" {
		{
			// fq2 assembly
			fName := filepath.Join(baseDir, "e2_adx_amd64.s")
			f, err := os.Create(fName)
			if err != nil {
				return err
			}

			_, _ = io.WriteString(f, "// +build amd64_adx\n")
			Fq2Amd64 := amd64.NewFq2Amd64(f, conf.Fp, conf.Name)
			if err := Fq2Amd64.Generate(false); err != nil {
				_ = f.Close()
				return err
			}
			_ = f.Close()

			cmd := exec.Command("asmfmt", "-w", fName)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	return nil

}
