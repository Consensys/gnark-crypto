package generator

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/consensys/gnark-crypto/field/generator/config"
)

var globalLock sync.Mutex

func GenerateFF(F *config.Field, outputDir string, options ...Option) error {

	// default config
	cfg := generatorOptions(options...)

	// generate asm
	// note: we need to do that before the fields, as the fields will include a hash of the (shared)
	// asm files to force a recompile of the field package if the asm files have changed
	globalLock.Lock() // TODO @gbotrel temporary need to handle shared files hashes.
	if cfg.HasArm64() {
		if err := generateARM64(F, cfg.asmConfig); err != nil {
			return err
		}
	}

	if cfg.HasAMD64() {
		if err := generateAMD64(F, cfg.asmConfig); err != nil {
			return err
		}
	}

	// generate field
	if err := generateField(F, outputDir, cfg.asmConfig.BuildDir, cfg.asmConfig.IncludeDir); err != nil {
		return err
	}

	globalLock.Unlock()

	// generate fft
	if cfg.HasFFT() {
		if err := generateFFT(F, cfg.fftConfig, outputDir); err != nil {
			return err
		}
	}

	return nil
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
