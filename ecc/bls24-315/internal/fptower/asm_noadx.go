//go:build noadx
// +build noadx

// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

// note: this is needed for test purposes, as dynamically changing supportAdx doesn't flag
// certain errors (like fatal error: missing stackmap)
// this ensures we test all asm path.
var supportAdx = false
