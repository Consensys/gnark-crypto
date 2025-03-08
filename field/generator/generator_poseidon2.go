package generator

import (
	"os"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/asm/amd64"
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
		F31              bool
		Q, QInvNeg       uint64
	}

	data := &poseidon2TemplateData{
		FF:               F.PackageName,
		FieldPackagePath: fieldImportPath,
		F31:              F.F31,
	}

	if data.FF == "koalabear" {
		// note that we can also generate for baby bear if needed, just need to tweak the number of
		// rounds and add the sbox.
		data.Q = F.Q[0]
		data.QInvNeg = F.QInverse[0]
		entries = append(entries, bavard.Entry{File: filepath.Join(outputDir, "poseidon2_amd64.go"), Templates: []string{"poseidon2.amd64.go.tmpl"}, BuildTag: "!purego"})
		entries = append(entries, bavard.Entry{File: filepath.Join(outputDir, "poseidon2_purego.go"), Templates: []string{"poseidon2.purego.go.tmpl"}, BuildTag: "purego || (!amd64)"})

		// generate the assembly file;
		asmFile, err := os.Create(filepath.Join(outputDir, "poseidon2_amd64.s"))
		if err != nil {
			return err
		}

		asmFile.WriteString("//go:build !purego\n")

		if err := amd64.GenerateF31Poseidon2(asmFile, F.NbBits); err != nil {
			asmFile.Close()
			return err
		}
		asmFile.Close()
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
