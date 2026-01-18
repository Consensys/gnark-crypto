// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package amd64

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestGenerateIFMAAssembly generates the IFMA assembly code and writes it to a file
// for manual inspection and verification.
//
// Run with: go test -v -run TestGenerateIFMAAssembly ./internal/generator/field/asm/amd64/
func TestGenerateIFMAAssembly(t *testing.T) {
	var buf bytes.Buffer

	// Create FFAmd64 for 4-word (256-bit) field
	f := NewFFAmd64(&buf, 4)

	// Write header
	f.Comment("AVX-512 IFMA Vector Multiplication for 4-word Fields")
	f.Comment("Generated for prototype testing - DO NOT USE IN PRODUCTION")
	f.WriteLn("")
	f.WriteLn("#include \"textflag.h\"")
	f.WriteLn("#include \"funcdata.h\"")
	f.WriteLn("#include \"go_asm.h\"")
	f.WriteLn("")

	// Generate the IFMA multiplication function
	f.generateMulVecIFMA()

	// Output to file for inspection
	outputDir := filepath.Join(os.TempDir(), "gnark-crypto-ifma-test")
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	outputPath := filepath.Join(outputDir, "element_ifma_amd64.s")
	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("failed to write assembly: %v", err)
	}

	t.Logf("Generated IFMA assembly written to: %s", outputPath)
	t.Logf("Assembly size: %d bytes", buf.Len())

	// Also write to the project directory for easier access
	projectOutputPath := filepath.Join("testdata", "element_ifma_amd64.s")
	os.MkdirAll("testdata", 0755)
	if err := os.WriteFile(projectOutputPath, buf.Bytes(), 0644); err != nil {
		t.Logf("Note: Could not write to testdata/: %v", err)
	} else {
		t.Logf("Also written to: %s", projectOutputPath)
	}

	// Print the first 100 lines for quick inspection
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	t.Log("First 100 lines of generated assembly:")
	for i, line := range lines {
		if i >= 100 {
			t.Logf("... (%d more lines)", len(lines)-100)
			break
		}
		t.Logf("%4d: %s", i+1, string(line))
	}
}

// TestIFMAAssemblyStructure verifies the structure of the generated IFMA assembly
func TestIFMAAssemblyStructure(t *testing.T) {
	var buf bytes.Buffer

	f := NewFFAmd64(&buf, 4)
	f.WriteLn("#include \"textflag.h\"")
	f.WriteLn("")
	f.generateMulVecIFMA()

	asm := buf.String()

	// Check for expected elements in the assembly
	expectedElements := []string{
		"mulVecIFMA",       // Function name
		"VPMADD52LUQ",      // IFMA low instruction
		"VPMADD52HUQ",      // IFMA high instruction
		"VPBROADCASTQ",     // Broadcast for constants
		"Z31",              // mask52 register
		"Z30",              // qInvNeg52 register
		"$0xFFFFFFFFFFFFF", // 52-bit mask
		"VMOVDQU64",        // Vector load/store
		"Montgomery",       // Montgomery reduction comment
	}

	for _, expected := range expectedElements {
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Errorf("Missing expected element in assembly: %s", expected)
		}
	}

	t.Logf("Generated assembly length: %d bytes", len(asm))
}
