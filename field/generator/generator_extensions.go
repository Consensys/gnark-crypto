package generator

import (
	"os"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/asm/amd64"
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

	type extensionsTemplateData struct {
		FF               string
		FieldPackagePath string
		F31              bool
		Q, QInvNeg       uint64
		IsKoalaBear      bool
		IsBabyBear       bool
	}

	isKoalaBear := F.Q[0] == 2130706433
	isBabyBear := F.Q[0] == 2013265921
	data := &extensionsTemplateData{
		FF:               F.PackageName,
		FieldPackagePath: fieldImportPath,
		F31:              F.F31,
		IsKoalaBear:      isKoalaBear,
		IsBabyBear:       isBabyBear,
		Q:                F.Q[0],
		QInvNeg:          F.QInverse[0],
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
		entries_ext4 := []bavard.Entry{
			{File: filepath.Join(outputDir, "e4.go"), Templates: []string{"e4.go.tmpl"}},
			{File: filepath.Join(outputDir, "vector.go"), Templates: []string{"vector.go.tmpl"}},
			{File: filepath.Join(outputDir, "e4_test.go"), Templates: []string{"e4_test.go.tmpl"}},
		}

		if isKoalaBear {
			entries_ext4 = append(entries_ext4, bavard.Entry{File: filepath.Join(outputDir, "e4_amd64.go"), Templates: []string{"e4.amd64.go.tmpl"}, BuildTag: "!purego"})
			entries_ext4 = append(entries_ext4, bavard.Entry{File: filepath.Join(outputDir, "e4_purego.go"), Templates: []string{"e4.purego.go.tmpl"}, BuildTag: "purego || (!amd64)"})
		}

		if err := bgen.GenerateWithOptions(data, "extensions", extensionsTemplatesRootDir, nil, entries_ext4...); err != nil {
			return err
		}

		if isKoalaBear {
			// generate the assembly file;
			asmFile, err := os.Create(filepath.Join(outputDir, "e4_amd64.s"))
			if err != nil {
				return err
			}

			asmFile.WriteString("//go:build !purego\n")

			if err := amd64.GenerateF31E4(asmFile); err != nil {
				asmFile.Close()
				return err
			}
			asmFile.Close()
		}
	}

	return runFormatters(outputDir)
}
