// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate(t *testing.T) {
	packageName := "testfield"
	elementName := "Element"
	modulus := "17"
	outputDir := filepath.Join(t.TempDir(), "testfield")

	err := Generate(packageName, elementName, modulus, outputDir)
	if err != nil {
		t.Fatalf("failed to generate field: %v", err)
	}

	// Check if some files were generated
	if _, err := os.Stat(filepath.Join(outputDir, "element.go")); os.IsNotExist(err) {
		t.Errorf("element.go was not generated")
	}
}
