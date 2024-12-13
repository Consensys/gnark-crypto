// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fr

import (
	"fmt"
	"math/big"
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc"
)

// Generator returns a generator for Z/2^(log(m))Z
// or an error if m is too big (required root of unity doesn't exist)
func Generator(m uint64) (Element, error) {
	x := ecc.NextPowerOfTwo(m)

	var rootOfUnity Element
	rootOfUnity.SetString("8065159656716812877374967518403273466521432693661810619979959746626482506078")
	const maxOrderRoot uint64 = 47

	// find generator for Z/2^(log(m))Z
	logx := uint64(bits.TrailingZeros64(x))
	if logx > maxOrderRoot {
		return Element{}, fmt.Errorf("m (%d) is too big: the required root of unity does not exist", m)
	}

	expo := uint64(1 << (maxOrderRoot - logx))
	var generator Element
	generator.Exp(rootOfUnity, big.NewInt(int64(expo))) // order x
	return generator, nil
}
