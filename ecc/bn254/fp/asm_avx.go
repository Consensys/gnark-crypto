//go:build !noavx

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fp

import "golang.org/x/sys/cpu"

var (
	supportAvx512 = supportAdx && cpu.X86.HasAVX512 && cpu.X86.HasAVX512DQ && cpu.X86.HasAVX512VBMI2
	_             = supportAvx512
)
