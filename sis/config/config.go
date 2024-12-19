package config

import "strings"

// SIS struct containing the necessary information for code generation
type SIS struct {

	// FieldPackagePath path to the finite field package
	FieldPackagePath string

	// FFTPackagePath path to the fft package
	FFTPackagePath string

	// OutputDir path of the fodler where the files are generated
	OutputDir string

	// Package name of the generated package
	Package string

	// FF the name of the package corresponding to the finite field
	FF string
}

// NewFFTConfig returns a data structure with needed information to generate apis for the FFT
func NewConfig(fieldPackagePath,
	fftPackagePath,
	outputDir string) SIS {
	splittedPath := strings.Split(fieldPackagePath, "/")
	finiteFieldPackage := splittedPath[len(splittedPath)-1]
	return SIS{
		FieldPackagePath: fieldPackagePath,
		FFTPackagePath:   fftPackagePath,
		OutputDir:        outputDir,
		FF:               finiteFieldPackage,
	}
}
