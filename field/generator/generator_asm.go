package generator

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/consensys/gnark-crypto/field/generator/asm/amd64"
	"github.com/consensys/gnark-crypto/field/generator/asm/arm64"
	"github.com/consensys/gnark-crypto/field/generator/config"
)

var (
	mARM64    sync.Map
	mAMD64    sync.Map
	lockARM64 sync.Mutex
	lockAMD64 sync.Mutex
)

// generateARM64 generates the assembly file for ARM64
// it formats it and returns a hash of the file
// this is safe to run concurrently
func generateARM64(F *config.Field, asm *config.Assembly) (string, error) {
	if !F.GenerateOpsARM64 {
		return "", nil
	}

	err := os.MkdirAll(asm.BuildDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", asm.BuildDir, err)
	}
	pathSrc := filepath.Join(asm.BuildDir, fmt.Sprintf(arm64.ElementASMFileName, F.NbWords))

	hash, ok := mARM64.Load(pathSrc)
	if ok {
		return hash.(string), nil
	}
	lockARM64.Lock()
	defer lockARM64.Unlock()

	fmt.Println("generating", pathSrc)
	f, err := os.Create(pathSrc)
	if err != nil {
		return "", err
	}

	if err := arm64.GenerateCommonASM(f, F.NbWords, F.GenerateVectorOpsARM64); err != nil {
		_ = f.Close()
		return "", err
	}
	_ = f.Close()

	if err := runASMFormatter(pathSrc); err != nil {
		return "", err
	}

	toReturn, err := hashFile(pathSrc)
	if err != nil {
		return "", err
	}

	mARM64.Store(pathSrc, toReturn)
	return toReturn, nil
}

// generateAMD64 generates the assembly file for AMD64
// it formats it and returns a hash of the file
// this is safe to run concurrently
func generateAMD64(F *config.Field, asm *config.Assembly) (string, error) {
	if !F.GenerateOpsAMD64 {
		return "", nil
	}
	err := os.MkdirAll(asm.BuildDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", asm.BuildDir, err)
	}
	pathSrc := filepath.Join(asm.BuildDir, fmt.Sprintf(amd64.ElementASMFileName, F.NbWords))

	hash, ok := mAMD64.Load(pathSrc)
	if ok {
		return hash.(string), nil
	}
	lockAMD64.Lock()
	defer lockAMD64.Unlock()

	fmt.Println("generating", pathSrc)
	f, err := os.Create(pathSrc)
	if err != nil {
		return "", err
	}

	if err := amd64.GenerateCommonASM(f, F.NbWords, F.GenerateVectorOpsAMD64); err != nil {
		_ = f.Close()
		return "", err
	}
	_ = f.Close()

	if err := runASMFormatter(pathSrc); err != nil {
		return "", err
	}

	toReturn, err := hashFile(pathSrc)
	if err != nil {
		return "", err
	}

	mAMD64.Store(pathSrc, toReturn)

	return toReturn, nil
}

type ASMWrapperData struct {
	IncludePath string
	Hash        string
}

func hashFile(filePath string) (string, error) {
	// we hash the file content and include the hash in comment of the generated file
	// to force the Go compiler to recompile the file if the content has changed
	fData, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// hash the file using FNV
	hasher := fnv.New64()
	hasher.Write(fData)
	hash64 := hasher.Sum64()

	return fmt.Sprintf("%d", hash64), nil
}

func newASMWrapperData(hash, asmDirIncludePath, fileName string, nbWords int) (data ASMWrapperData, err error) {
	fileName = fmt.Sprintf(fileName, nbWords)

	includePath := filepath.Join(asmDirIncludePath, fileName)
	// on windows, we replace the "\" by "/"
	if filepath.Separator == '\\' {
		includePath = strings.ReplaceAll(includePath, "\\", "/")
	}

	return ASMWrapperData{
		IncludePath: includePath,
		Hash:        hash,
	}, nil

}
