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
		{File: filepath.Join(outputDir, "poseidon2_test.go"), Templates: []string{"poseidon2_test.go.tmpl"}},
		{File: filepath.Join(outputDir, "hash.go"), Templates: []string{"hash.go.tmpl"}},
	}

	type poseidon2TemplateData struct {
		FF                string
		FieldPackagePath  string
		F31               bool
		Q, QInvNeg        uint64
		ParamsCompression amd64.Poseidon2Parameters
		ParamsSponge      amd64.Poseidon2Parameters
		Params            []amd64.Poseidon2Parameters
	}

	data := &poseidon2TemplateData{
		FF:               F.PackageName,
		FieldPackagePath: fieldImportPath,
		F31:              F.F31,
	}
	switch data.FF {
	case "koalabear":
		data.ParamsCompression = amd64.Poseidon2Parameters{
			Width:         16,
			FullRounds:    6,
			PartialRounds: 21,
			SBoxDegree:    3,
			DiagInternal:  []uint64{2130706431, 1, 2, 1065353217, 3, 4, 1065353216, 2130706430, 2130706429, 2122383361, 1864368129, 2130706306, 8323072, 266338304, 133169152, 127},
		}

		data.ParamsSponge = amd64.Poseidon2Parameters{
			Width:         24,
			FullRounds:    6,
			PartialRounds: 21,
			SBoxDegree:    3,
			DiagInternal:  []uint64{2130706431, 1, 2, 1065353217, 3, 4, 1065353216, 2130706430, 2130706429, 2122383361, 1598029825, 1864368129, 1997537281, 2064121857, 2097414145, 2130706306, 8323072, 266338304, 133169152, 66584576, 33292288, 16646144, 4161536, 127},
		}

		data.Params = []amd64.Poseidon2Parameters{
			data.ParamsSponge,
			data.ParamsCompression,
		}
	case "babybear":
		data.ParamsCompression = amd64.Poseidon2Parameters{
			Width:         16,
			FullRounds:    8,
			PartialRounds: 13,
			SBoxDegree:    7,
			DiagInternal:  []uint64{2013265919, 1, 2, 1006632961, 3, 4, 1006632960, 2013265918, 2013265917, 2005401601, 1509949441, 1761607681, 2013265906, 7864320, 125829120, 15},
		}
		data.ParamsSponge = amd64.Poseidon2Parameters{
			Width:         24,
			FullRounds:    8,
			PartialRounds: 21,
			SBoxDegree:    7,
			DiagInternal:  []uint64{2013265919, 1, 2, 1006632961, 3, 4, 1006632960, 2013265918, 2013265917, 2005401601, 1509949441, 1761607681, 1887436801, 1997537281, 2009333761, 2013265906, 7864320, 503316480, 251658240, 125829120, 62914560, 31457280, 15728640, 15},
		}
		data.Params = []amd64.Poseidon2Parameters{
			data.ParamsSponge,
			data.ParamsCompression,
		}
	case "goldilocks":
		data.ParamsCompression = amd64.Poseidon2Parameters{
			Width:         8,
			FullRounds:    6,
			PartialRounds: 17,
			SBoxDegree:    7,
			// same as https://github.com/Plonky3/Plonky3/blob/f91c76545cf5c4ae9182897bcc557715817bcbdc/goldilocks/src/poseidon2.rs#L54
			DiagInternal: []uint64{0xa98811a1fed4e3a5, 0x1cc48b54f377e2a0, 0xe40cd4f6c5609a26, 0x11de79ebca97a4a3, 0x9177c73d8b7e929c, 0x2a6fe8085797e791, 0x3de6e93329f8d5ad, 0x3f7af9125da962fe},
		}
		data.ParamsSponge = amd64.Poseidon2Parameters{
			Width:         12,
			FullRounds:    6,
			PartialRounds: 17,
			SBoxDegree:    7,
			// same as https://github.com/Plonky3/Plonky3/blob/f91c76545cf5c4ae9182897bcc557715817bcbdc/goldilocks/src/poseidon2.rs#L65
			DiagInternal: []uint64{0xc3b6c08e23ba9300, 0xd84b5de94a324fb6, 0x0d0c371c5b35b84f, 0x7964f570e7188037, 0x5daf18bbd996604b, 0x6743bc47b9595257, 0x5528b9362c59bb70, 0xac45e25b7127b68b, 0xa2077d7dfbb606b5, 0xf3faac6faee378ae, 0x0c6388b51545e883, 0xd27dbb6944917b60},
		}
		data.Params = []amd64.Poseidon2Parameters{
			data.ParamsSponge,
			data.ParamsCompression,
		}
	default:
		panic("unknown field")
	}

	if data.F31 {
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

		if err := amd64.GenerateF31Poseidon2(asmFile, F.NbBits, data.Params); err != nil {
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
