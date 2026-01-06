// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package generator provides a wrapper around the internal field generator.
package generator

import (
	"github.com/consensys/gnark-crypto/internal/generator/field"
	"github.com/consensys/gnark-crypto/internal/generator/field/config"
)

// Option is a functional option for the generator.
type Option func(*options)

type options struct {
	internal    []field.Option
	useAddChain bool
}

// WithASM enables assembly code generation.
// buildDir is the destination path to create output files.
// includeDir is the directory name to include in the generated files.
func WithASM(buildDir, includeDir string) Option {
	return func(o *options) {
		o.internal = append(o.internal, field.WithASM(&config.Assembly{
			BuildDir:   buildDir,
			IncludeDir: includeDir,
		}))
	}
}

// WithFFT enables FFT code generation.
func WithFFT(generatorFullMultiplicativeGroup uint64, generatorMaxTwoAdicSubgroup string, logTwoOrderMaxTwoAdicSubgroup string) Option {
	return func(o *options) {
		o.internal = append(o.internal, field.WithFFT(&config.FFT{
			GeneratorFullMultiplicativeGroup: generatorFullMultiplicativeGroup,
			GeneratorMaxTwoAdicSubgroup:      generatorMaxTwoAdicSubgroup,
			LogTwoOrderMaxTwoAdicSubgroup:    logTwoOrderMaxTwoAdicSubgroup,
		}))
	}
}

// WithSIS enables SIS code generation.
func WithSIS() Option {
	return func(o *options) {
		o.internal = append(o.internal, field.WithSIS())
	}
}

// WithPoseidon2 enables Poseidon2 code generation.
func WithPoseidon2() Option {
	return func(o *options) {
		o.internal = append(o.internal, field.WithPoseidon2())
	}
}

// WithExtensions enables field extensions code generation.
func WithExtensions() Option {
	return func(o *options) {
		o.internal = append(o.internal, field.WithExtensions())
	}
}

// WithIOP enables IOP code generation.
func WithIOP() Option {
	return func(o *options) {
		o.internal = append(o.internal, field.WithIOP())
	}
}

// WithAddChain enables the use of addition chains for inversion.
func WithAddChain() Option {
	return func(o *options) {
		o.useAddChain = true
	}
}

// Generate generates arithmetic operations for a given modulus.
func Generate(packageName, elementName, modulus string, outputDir string, opts ...Option) error {
	var cfg options
	for _, opt := range opts {
		opt(&cfg)
	}

	F, err := config.NewFieldConfig(packageName, elementName, modulus, cfg.useAddChain)
	if err != nil {
		return err
	}

	return field.GenerateFF(F, outputDir, cfg.internal...)
}
