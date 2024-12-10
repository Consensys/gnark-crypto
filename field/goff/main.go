// Copyright 2020 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package goff (go finite field) is a library that generates fast field arithmetic code for a given modulus.
//
// Generated code is optimized for x86 (amd64) targets, and most methods do not allocate memory on the heap.
//
// Example usage:
//
//	goff -m 0xffffffff00000001 -o ./goldilocks/ -p goldilocks -e Element
//
// # Warning
//
// The generated code has not been audited for all moduli (only bn254 and bls12-381) and is provided as-is. In particular, there is no security guarantees such as constant time implementation or side-channel attack resistance.
package main

import "github.com/consensys/gnark-crypto/field/goff/cmd"

func main() {
	cmd.Execute()
}
