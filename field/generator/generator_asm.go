package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/consensys/gnark-crypto/field/generator/asm/amd64"
	"github.com/consensys/gnark-crypto/field/generator/asm/arm64"
	"github.com/consensys/gnark-crypto/field/generator/config"
)

func generateARM64(F *config.Field, asm *config.Assembly) error {
	if !F.GenerateOpsARM64 {
		return nil
	}

	os.MkdirAll(asm.BuildDir, 0755)
	pathSrc := filepath.Join(asm.BuildDir, fmt.Sprintf(arm64.ElementASMFileName, F.NbWords))

	fmt.Println("generating", pathSrc)
	f, err := os.Create(pathSrc)
	if err != nil {
		return err
	}

	if err := arm64.GenerateCommonASM(f, F.NbWords, F.GenerateVectorOpsARM64); err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Close()

	return runFormatters(asm.BuildDir)
}

func generateAMD64(F *config.Field, asm *config.Assembly) error {
	if !F.GenerateOpsAMD64 {
		return nil
	}
	os.MkdirAll(asm.BuildDir, 0755)
	pathSrc := filepath.Join(asm.BuildDir, fmt.Sprintf(amd64.ElementASMFileName, F.NbWords))

	fmt.Println("generating", pathSrc)
	f, err := os.Create(pathSrc)
	if err != nil {
		return err
	}

	if err := amd64.GenerateCommonASM(f, F.NbWords, F.GenerateVectorOpsAMD64); err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Close()

	return runFormatters(asm.BuildDir)
}
