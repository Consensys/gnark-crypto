//go:build purego || !amd64

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package extensions

import (
	fr "github.com/consensys/gnark-crypto/field/babybear"
)

func mulAccE4_avx512(alpha *E4, scale *fr.Element, res *E4, N uint64) {
	panic("mulAccE4_avx512 is not implemented")
}
