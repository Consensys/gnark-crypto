package generator

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/config"
)

func generateExtensions(F *config.Field, outputDir string) error {

	fieldImportPath, err := getImportPath(outputDir)
	if err != nil {
		return err
	}

	outputDir = filepath.Join(outputDir, "extensions")

	entries_ext2 := []bavard.Entry{
		{File: filepath.Join(outputDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(outputDir, "utils.go"), Templates: []string{"utils.go.tmpl"}},
		{File: filepath.Join(outputDir, "e2.go"), Templates: []string{"e2.go.tmpl"}},
		{File: filepath.Join(outputDir, "e2_test.go"), Templates: []string{"e2_test.go.tmpl"}},
	}
	entries_ext4 := []bavard.Entry{
		{File: filepath.Join(outputDir, "e4.go"), Templates: []string{"e4.go.tmpl"}},
		{File: filepath.Join(outputDir, "e4_test.go"), Templates: []string{"e4_test.go.tmpl"}},
	}

	type extensionsTemplateData struct {
		FF               string
		FieldPackagePath string
	}

	data := &extensionsTemplateData{
		FF:               F.PackageName,
		FieldPackagePath: fieldImportPath,
	}

	bgen := bavard.NewBatchGenerator("Consensys Software Inc.", 2020, "consensys/gnark-crypto")

	extensionsTemplatesRootDir, err := findTemplatesRootDir()
	if err != nil {
		return err
	}
	extensionsTemplatesRootDir = filepath.Join(extensionsTemplatesRootDir, "extensions")

	if err := bgen.GenerateWithOptions(data, "extensions", extensionsTemplatesRootDir, nil, entries_ext2...); err != nil {
		return err
	}
	if F.F31 {
		if err := bgen.GenerateWithOptions(data, "extensions", extensionsTemplatesRootDir, nil, entries_ext4...); err != nil {
			return err
		}
	}

	return runFormatters(outputDir)
}
