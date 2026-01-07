// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package generator_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/consensys/gnark-crypto/field/generator"
)

func ExampleGenerate() {
	// Define the field parameters
	packageName := "goldilocks"
	elementName := "Element"
	modulus := "0xffffffff00000001"
	outputDir := filepath.Join(os.TempDir(), "gnark-crypto-field-gen")

	// Ensure output directory exists
	_ = os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(outputDir)

	// Generate the field arithmetic code
	err := generator.Generate(packageName, elementName, modulus, outputDir,
		generator.WithASM(filepath.Join(outputDir, "asm"), "asm"),
	)
	if err != nil {
		fmt.Printf("failed to generate field: %v\n", err)
		return
	}

	fmt.Println("successfully generated field arithmetic code")
}
