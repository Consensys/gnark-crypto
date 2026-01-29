// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
)

// Precomputed operation sequence for CbrtFrobenius
// e = (2p² + 7)/27, decomposed as e = a0 + a1*p
// Each 2-bit value encodes: 0=square only, 1=square+mul(x), 2=square+mul(x̄), 3=square+MulByElement(N)
// 253 operations packed into 8 uint64s (32 ops per uint64, last has 29)
var cbrtFrobeniusOps = [8]uint64{
	0x46eff8f18b734fc1, 0xb05a7333568787f0,
	0xfbf9634fdae3a439, 0x6c79d47169ccca90,
	0x1bb908d39d477e8f, 0x87e908d3e8eff881,
	0x5fa5946ee43f6b97, 0x021fdb7495c470fa,
}

// CbrtFrobenius computes z = ∛x using Frobenius decomposition.
// This is an optimized implementation that decomposes the exponentiation
// x^e into x^a₀ · x̄^a₁ where e = a₀ + a₁·p and x̄ is the conjugate.
//
// The advantages are:
// 1. When both bits are set, we multiply by N(x) ∈ Fp (cheaper MulByElement)
// 2. Uses precomputed operation sequence (no big.Int operations at runtime)
func (z *E2) CbrtFrobenius(x *E2) *E2 {
	// If x is in Fp (i.e., x.A1 == 0), use Fp cube root directly
	if x.A1.IsZero() {
		if z.A0.Cbrt(&x.A0) == nil {
			return nil
		}
		z.A1.SetZero()
		return z
	}

	// Compute y = x^a0 * x̄^a1 using precomputed operation sequence
	var y E2
	y.expByCbrtFrobeniusChain(x)

	// c = y³
	var c E2
	c.Square(&y).Mul(&c, &y)
	if c.Equal(x) {
		return z.Set(&y)
	}

	// Primitive cube roots of unity ω, ω² (in Fp, embedded as (ω, 0))
	omega := fp.Element{
		8183898218631979349,
		12014359695528440611,
		12263358156045030468,
		3187210487005268291,
	}
	omega2 := fp.Element{
		3697675806616062876,
		9065277094688085689,
		6918009208039626314,
		2775033306905974752,
	}

	// Primitive 9th roots of unity ζ, ζ² (in Fp)
	zeta := fp.Element{
		9092840637269024442,
		11284133545212953584,
		7919372827184455520,
		1596114425137527684,
	}
	zeta2 := fp.Element{
		1735008219140503419,
		10465829585049341007,
		6017168831245289042,
		1570250484855163800,
	}

	// Check if c * ω² = x, then y * ζ is the cube root
	var cw2 E2
	cw2.MulByElement(&c, &omega2)
	if cw2.Equal(x) {
		return z.MulByElement(&y, &zeta)
	}

	// Check if c * ω = x, then y * ζ² is the cube root
	var cw E2
	cw.MulByElement(&c, &omega)
	if cw.Equal(x) {
		return z.MulByElement(&y, &zeta2)
	}

	// x is not a cubic residue
	return nil
}

// expByCbrtFrobeniusChain computes z = x^a0 * x̄^a1 using a precomputed operation sequence.
// This is faster than the generic Shamir's trick because:
// 1. No big.Int.Bit() calls (operations are precomputed)
// 2. Better branch prediction (switch on 2-bit value)
// 3. Efficient bit extraction from uint64
func (z *E2) expByCbrtFrobeniusChain(x *E2) *E2 {
	// Precompute: x̄ = conjugate, N(x) = x·x̄ ∈ Fp
	var xConj E2
	xConj.Conjugate(x)

	var norm fp.Element
	x.norm(&norm)

	// Start with z = 1 (but skip leading square, handle first op specially)
	// First operation is always op=1 (mul by x), so start with z = x
	z.Set(x)

	// Process remaining 252 operations from precomputed sequence
	// Operations are packed: 32 ops per uint64, 2 bits each
	// Word 0 bits [63:2] = ops 1-31, word 0 bits [1:0] = op 0 (already handled)

	// Process word 0, ops 1-31 (skip op 0 which we handled above)
	word := cbrtFrobeniusOps[0]
	for j := 1; j < 32; j++ {
		z.Square(z)
		op := (word >> (j * 2)) & 3
		switch op {
		case 1:
			z.Mul(z, x)
		case 2:
			z.Mul(z, &xConj)
		case 3:
			z.MulByElement(z, &norm)
		}
	}

	// Process words 1-6 (full 32 ops each)
	for i := 1; i < 7; i++ {
		word = cbrtFrobeniusOps[i]
		for j := 0; j < 32; j++ {
			z.Square(z)
			op := (word >> (j * 2)) & 3
			switch op {
			case 1:
				z.Mul(z, x)
			case 2:
				z.Mul(z, &xConj)
			case 3:
				z.MulByElement(z, &norm)
			}
		}
	}

	// Process word 7 (29 remaining ops: 253 - 32*7 = 253 - 224 = 29)
	word = cbrtFrobeniusOps[7]
	for j := 0; j < 29; j++ {
		z.Square(z)
		op := (word >> (j * 2)) & 3
		switch op {
		case 1:
			z.Mul(z, x)
		case 2:
			z.Mul(z, &xConj)
		case 3:
			z.MulByElement(z, &norm)
		}
	}

	return z
}
