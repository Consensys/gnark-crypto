package generator

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/config"
)

func generatePoseidon2(F *config.Field, outputDir string) error {

	fieldImportPath, err := getImportPath(outputDir)
	if err != nil {
		return err
	}

	outputDir = filepath.Join(outputDir, "poseidon2")

	entries := []bavard.Entry{
		{File: filepath.Join(outputDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(outputDir, "poseidon2.go"), Templates: []string{"poseidon2.go.tmpl"}},
		{File: filepath.Join(outputDir, "hash.go"), Templates: []string{"hash.go.tmpl"}},
	}

	type poseidon2TemplateData struct {
		FF               string
		FieldPackagePath string
	}

	data := &poseidon2TemplateData{
		FF:               F.PackageName,
		FieldPackagePath: fieldImportPath,
	}

	bgen := bavard.NewBatchGenerator("Consensys Software Inc.", 2020, "consensys/gnark-crypto")

	poseidon2TemplatesRootDir, err := findTemplatesRootDir()
	if err != nil {
		return err
	}
	poseidon2TemplatesRootDir = filepath.Join(poseidon2TemplatesRootDir, "poseidon2")

	if err := bgen.GenerateWithOptions(data, "poseidon2", poseidon2TemplatesRootDir, nil, entries...); err != nil {
		return err
	}

	return runFormatters(outputDir)
}
