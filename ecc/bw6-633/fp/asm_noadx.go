//go:build noadx

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fp

// note: this is needed for test purposes, as dynamically changing supportAdx doesn't flag
// certain errors (like fatal error: missing stackmap)
// this ensures we test all asm path.
var (
	supportAdx = false
	_          = supportAdx
)
