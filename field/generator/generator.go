package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/consensys/gnark-crypto/field/generator/config"
)

func GenerateFF(F *config.Field, outputDir string, options ...Option) error {

	// default config
	cfg := generatorOptions(options...)

	// generate asm
	// note: we need to do that before the fields, as the fields will include a hash of the (shared)
	// asm files to force a recompile of the field package if the asm files have changed
	var hashArm64, hashAMD64 string
	var err error
	if cfg.HasArm64() {
		hashArm64, err = generateARM64(F, cfg.asmConfig)
		if err != nil {
			return err
		}
	}

	if cfg.HasAMD64() {
		hashAMD64, err = generateAMD64(F, cfg.asmConfig)
		if err != nil {
			return err
		}
	}

	// generate field
	asmIncludeDir := ""
	if cfg.HasArm64() || cfg.HasAMD64() {
		asmIncludeDir = cfg.asmConfig.IncludeDir
	}
	if err := generateField(F, outputDir, asmIncludeDir, hashArm64, hashAMD64); err != nil {
		return err
	}

	// generate fft
	if cfg.HasFFT() {
		if err := generateFFT(F, cfg.fftConfig, outputDir); err != nil {
			return err
		}
	}

	// generate SIS
	if cfg.HasSIS() {
		if err := generateSIS(F, outputDir); err != nil {
			return err
		}
	}

	return runFormatters(outputDir)
}

func runFormatters(outputDir string) error {
	var out strings.Builder
	{
		// run go fmt on whole directory
		cmd := exec.Command("gofmt", "-s", "-w", outputDir)
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("gofmt failed: %v\n%s", err, out.String())
		}
	}
	{
		// run asmfmt on whole directory
		cmd := exec.Command("asmfmt", "-w", outputDir)
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("asmfmt failed: %v\n%s", err, out.String())
		}
	}
	return nil
}

func runASMFormatter(file string) error {
	var out strings.Builder
	{
		// run asmfmt on whole directory
		cmd := exec.Command("asmfmt", "-w", file)
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("asmfmt failed: %v\n%s", err, out.String())
		}
	}
	return nil
}

func getImportPath(dir string) (string, error) {
	// get absolute path for dir
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %w", err)
	}

	modDir, err := findGoMod(dir)
	if err != nil {
		return "", fmt.Errorf("error finding go.mod: %w", err)
	}

	modulePath, err := getModulePath(modDir)
	if err != nil {
		return "", fmt.Errorf("error reading module path: %w", err)
	}

	relPath, err := filepath.Rel(modDir, dir)
	if err != nil {
		return "", fmt.Errorf("error computing relative path: %w", err)
	}

	// Handle the case where the directory is the module root
	if relPath == "." {
		return modulePath, nil
	}
	return modulePath + "/" + filepath.ToSlash(relPath), nil
}

// findGoMod ascends the directory tree to locate the go.mod file.
func findGoMod(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		newDir := filepath.Dir(dir)
		if newDir == dir {
			return "", fmt.Errorf("no go.mod found up to root")
		}
		dir = newDir
	}
}

// getModulePath extracts the module path from the go.mod file.
func getModulePath(modDir string) (string, error) {
	content, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("error reading go.mod: %w", err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}
	return "", fmt.Errorf("module declaration not found in go.mod")
}
