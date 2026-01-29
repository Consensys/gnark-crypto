// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

// Precomputed operation sequence for CbrtFrobenius
// e = (p² + 8)/27, decomposed as e = a0 + a1*p
// Each 2-bit value encodes: 0=square only, 1=square+mul(x), 2=square+mul(x̄), 3=square+MulByElement(N)
// 380 operations packed into 12 uint64s (32 ops per uint64, last has 28)
var cbrtFrobeniusOps = [12]uint64{
	0xf6716f2e4368bb41, 0x1fefdb9b99744bef,
	0x54be6c0d874d3f92, 0x9b010934d8c7fbf6,
	0x338216f906f25fcb, 0x95bc042dcc59a036,
	0x5da2c9e27a434f78, 0xa436a4109a9042f2,
	0x56449a2ed97d5026, 0xa65d1268bb6e48e2,
	0xc0fc0fc0fc9744b7, 0x00b0abf41bc0fc0f,
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
		14772873186050699377,
		6749526151121446354,
		6372666795664677781,
		10283423008382700446,
		286397964926079186,
		1796971870900422465,
	}
	omega2 := fp.Element{
		3526659474838938856,
		17562030475567847978,
		1632777218702014455,
		14009062335050482331,
		3906511377122991214,
		368068849512964448,
	}

	// Primitive 9th roots of unity ζ, ζ² (in Fp)
	zeta := fp.Element{
		13616190144799058984,
		9227582506135211912,
		4426607408274926740,
		7455198167498346307,
		10794825842164118204,
		335101026345095675,
	}
	zeta2 := fp.Element{
		3828863564860874189,
		5918733612565202776,
		16843310164143221096,
		16127847466718491017,
		17435063908385505950,
		407112797415018074,
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

	// Process remaining 379 operations from precomputed sequence
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

	// Process words 1-10 (full 32 ops each)
	for i := 1; i < 11; i++ {
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

	// Process word 11 (28 remaining ops: 380 - 32*11 = 380 - 352 = 28)
	word = cbrtFrobeniusOps[11]
	for j := 0; j < 28; j++ {
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
