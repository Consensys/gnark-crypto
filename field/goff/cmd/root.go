// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package cmd is the CLI interface for goff
package cmd

import (
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/gnark-crypto/field/generator"
	"github.com/consensys/gnark-crypto/field/generator/config"
	field "github.com/consensys/gnark-crypto/field/generator/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "goff",
	Short:   "goff generates arithmetic operations for any moduli",
	Run:     cmdGenerate,
	Version: Version,
}

// flags
var (
	fModulus     string
	fOutputDir   string
	fPackageName string
	fElementName string
)

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVarP(&fElementName, "element", "e", "", "name of the generated struct and file")
	rootCmd.PersistentFlags().StringVarP(&fModulus, "modulus", "m", "", "field modulus (base 10)")
	rootCmd.PersistentFlags().StringVarP(&fOutputDir, "output", "o", "", "destination path to create output files")
	rootCmd.PersistentFlags().StringVarP(&fPackageName, "package", "p", "", "package name in generated files")
	if bits.UintSize != 64 {
		panic("goff only supports 64bits architectures")
	}
}

func cmdGenerate(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println("running goff version", Version)
	fmt.Println()

	// parse flags
	if err := parseFlags(cmd); err != nil {
		_ = cmd.Usage()
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(-1)
	}

	// generate code
	F, err := field.NewFieldConfig(fPackageName, fElementName, fModulus, false)
	if err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(-1)
	}

	asmDir := filepath.Join(fOutputDir, "asm")

	if err := generator.GenerateFF(F, fOutputDir, generator.WithASM(&config.Assembly{BuildDir: asmDir, IncludeDir: "asm"})); err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(-1)
	}
}

func parseFlags(cmd *cobra.Command) error {
	if fModulus == "" ||
		fOutputDir == "" ||
		fPackageName == "" ||
		fElementName == "" {
		return errMissingArgument
	}

	// clean inputs
	fOutputDir = filepath.Clean(fOutputDir)
	fPackageName = strings.ToLower(fPackageName)

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
