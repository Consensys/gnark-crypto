package generator

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/config"
)

func generateIOP(F *config.Field, outputDir string) error {

	fieldImportPath, err := getImportPath(outputDir)
	if err != nil {
		return err
	}

	outputDir = filepath.Join(outputDir, "iop")

	entries := []bavard.Entry{
		{File: filepath.Join(outputDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(outputDir, "polynomial.go"), Templates: []string{"polynomial.go.tmpl"}},
		{File: filepath.Join(outputDir, "polynomial_test.go"), Templates: []string{"polynomial.test.go.tmpl"}},
		{File: filepath.Join(outputDir, "ratios.go"), Templates: []string{"ratios.go.tmpl"}},
		{File: filepath.Join(outputDir, "ratios_test.go"), Templates: []string{"ratios.test.go.tmpl"}},
		{File: filepath.Join(outputDir, "quotient.go"), Templates: []string{"quotient.go.tmpl"}},
		{File: filepath.Join(outputDir, "quotient_test.go"), Templates: []string{"quotient.test.go.tmpl"}},
		{File: filepath.Join(outputDir, "expressions.go"), Templates: []string{"expressions.go.tmpl"}},
		{File: filepath.Join(outputDir, "expressions_test.go"), Templates: []string{"expressions.test.go.tmpl"}},
		{File: filepath.Join(outputDir, "utils.go"), Templates: []string{"utils.go.tmpl"}},
	}

	bgen := bavard.NewBatchGenerator("Consensys Software Inc.", 2020, "consensys/gnark-crypto")

	iopTemplatesRootDir, err := findTemplatesRootDir()
	if err != nil {
		return err
	}
	iopTemplatesRootDir = filepath.Join(iopTemplatesRootDir, "iop")

	fieldInfo := config.FieldDependency{
		FieldPackagePath: fieldImportPath,
		FieldPackageName: F.PackageName,
		ElementType:      F.PackageName + ".Element",
	}

	if err := bgen.Generate(fieldInfo, "iop", iopTemplatesRootDir, entries...); err != nil {
		return err
	}

	return runFormatters(outputDir)
}
