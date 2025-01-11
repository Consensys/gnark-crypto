package generator

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/config"
)

func generateSIS(F *config.Field, outputDir string) error {

	fieldImportPath, err := getImportPath(outputDir)
	if err != nil {
		return err
	}

	outputDir = filepath.Join(outputDir, "sis")

	entries := []bavard.Entry{
		{File: filepath.Join(outputDir, "sis_fft.go"), Templates: []string{"fft.go.tmpl"}},
		{File: filepath.Join(outputDir, "sis.go"), Templates: []string{"sis.go.tmpl"}},
		{File: filepath.Join(outputDir, "sis_test.go"), Templates: []string{"sis.test.go.tmpl"}},
	}

	funcs := make(map[string]interface{})
	funcs["bitReverse"] = bitReverse

	bavardOpts := []func(*bavard.Bavard) error{bavard.Funcs(funcs)}

	type sisTemplateData struct {
		FF               string
		FieldPackagePath string
	}

	data := &sisTemplateData{
		FF:               F.PackageName,
		FieldPackagePath: fieldImportPath,
	}

	bgen := bavard.NewBatchGenerator("Consensys Software Inc.", 2020, "consensys/gnark-crypto")

	sisTemplatesRootDir, err := findTemplatesRootDir()
	if err != nil {
		return err
	}
	sisTemplatesRootDir = filepath.Join(sisTemplatesRootDir, "sis")

	if err := bgen.GenerateWithOptions(data, "sis", sisTemplatesRootDir, bavardOpts, entries...); err != nil {
		return err
	}

	return runFormatters(outputDir)
}
