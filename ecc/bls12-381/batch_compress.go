// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls12381

import (
	"errors"
	"sync/atomic"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/internal/fptower"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

// BatchCompress2G1 compresses two G1Affine points into (z0, z1, flags) using
// the birational map χ₂,₃ from https://eprint.iacr.org/2021/1446.pdf (Section 3).
// Decompression requires only ONE cube root extraction instead of two square roots.
//
// Flags encoding: [case indicator (4 bits) | cube root index n (2 bits) | sign bits (2 bits)]
func BatchCompress2G1(p0, p1 *G1Affine) (z0, z1 fp.Element, flags byte, err error) {
	p0Inf := p0.IsInfinity()
	p1Inf := p1.IsInfinity()

	if p0Inf && p1Inf {
		flags = 0b11 << 4
		return
	}

	if p0Inf {
		z0 = p1.X
		flags = 0b10 << 4
		if p1.Y.LexicographicallyLargest() {
			flags |= 0b01
		}
		return
	}

	if p1Inf {
		z0 = p0.X
		flags = 0b01 << 4
		if p0.Y.LexicographicallyLargest() {
			flags |= 0b10
		}
		return
	}

	// Degenerate case: x0 == x1
	if p0.X.Equal(&p1.X) {
		z0 = p0.X
		if p0.Y.Equal(&p1.Y) {
			flags = 0b0100 << 4 // same point
		} else {
			flags = 0b0101 << 4 // negation
		}
		if p0.Y.LexicographicallyLargest() {
			flags |= 0b10
		}
		return
	}

	// Degenerate case: y0² == y1² means P1 = [-ω]^k(P0) for some k ∈ {0,...,5}
	var y0Sq, y1Sq fp.Element
	y0Sq.Square(&p0.Y)
	y1Sq.Square(&p1.Y)
	if y0Sq.Equal(&y1Sq) {
		z0 = p0.X
		z1 = p0.Y
		flags = 0b0110 << 4
		// Find k: [-ω]^k(x,y) = (ω^k·x, (-1)^k·y)
		// ω = thirdRootOneG1, ω² = thirdRootOneG2
		var negY fp.Element
		negY.Neg(&p0.Y)
		// Check k=0: (x0, y0)
		if p0.X.Equal(&p1.X) && p0.Y.Equal(&p1.Y) {
			return
		}
		// Check k=1: (ω·x0, -y0)
		var omegaX fp.Element
		omegaX.Mul(&p0.X, &thirdRootOneG1)
		if omegaX.Equal(&p1.X) && negY.Equal(&p1.Y) {
			flags |= 1
			return
		}
		// Check k=2: (ω²·x0, y0)
		var omega2X fp.Element
		omega2X.Mul(&p0.X, &thirdRootOneG2)
		if omega2X.Equal(&p1.X) && p0.Y.Equal(&p1.Y) {
			flags |= 2
			return
		}
		// Check k=3: (x0, -y0) - but this is x0==x1 case handled above
		// Check k=4: (ω·x0, y0)
		if omegaX.Equal(&p1.X) && p0.Y.Equal(&p1.Y) {
			flags |= 4
			return
		}
		// Check k=5: (ω²·x0, -y0)
		if omega2X.Equal(&p1.X) && negY.Equal(&p1.Y) {
			flags |= 5
			return
		}
		err = errors.New("failed to find automorphism k")
		return
	}

	// Generic case: compute χ₂,₃(P0, P1) = (z0, z1)
	// z0 = x1·(2x0²y1 - x0x1(y0 - y1) - 2y0x1²) / (y0² - y1²)
	// z1 = (x0³y1 + 2x0x1(x0y1 - y0x1) - y0x1³) / (y0² - y1²)
	x0, y0, x1, y1 := p0.X, p0.Y, p1.X, p1.Y

	var x0Sq, x1Sq, x0Cubed, x1Cubed fp.Element
	x0Sq.Square(&x0)
	x1Sq.Square(&x1)
	x0Cubed.Mul(&x0Sq, &x0)
	x1Cubed.Mul(&x1Sq, &x1)

	var denInv fp.Element
	denInv.Sub(&y0Sq, &y1Sq)
	denInv.Inverse(&denInv)

	// z0 numerator
	var term1, term2, term3, numZ0 fp.Element
	term1.Mul(&x0Sq, &y1)
	term1.Double(&term1) // 2x0²y1

	var y0MinusY1 fp.Element
	y0MinusY1.Sub(&y0, &y1)
	term2.Mul(&x0, &x1)
	term2.Mul(&term2, &y0MinusY1) // x0x1(y0 - y1)

	term3.Mul(&y0, &x1Sq)
	term3.Double(&term3) // 2y0x1²

	numZ0.Sub(&term1, &term2)
	numZ0.Sub(&numZ0, &term3)
	numZ0.Mul(&numZ0, &x1)
	z0.Mul(&numZ0, &denInv)

	// z1 numerator
	var numZ1, innerTerm, temp fp.Element
	numZ1.Mul(&x0Cubed, &y1) // x0³y1

	innerTerm.Mul(&x0, &y1)
	temp.Mul(&y0, &x1)
	innerTerm.Sub(&innerTerm, &temp) // x0y1 - y0x1
	innerTerm.Mul(&innerTerm, &x0)
	innerTerm.Mul(&innerTerm, &x1)
	innerTerm.Double(&innerTerm) // 2x0x1(x0y1 - y0x1)

	numZ1.Add(&numZ1, &innerTerm)
	temp.Mul(&y0, &x1Cubed)
	numZ1.Sub(&numZ1, &temp)
	z1.Mul(&numZ1, &denInv)

	// Determine cube root index n: x1 = ω^n · cbrt(y1² - b)
	omega := thirdRootOneG1
	var g1, cbrtG1 fp.Element
	g1.Sub(&y1Sq, &bCurveCoeff)
	if cbrtG1.Cbrt(&g1) == nil {
		err = errors.New("g1 is not a cubic residue")
		return
	}

	var n byte
	if cbrtG1.Equal(&x1) {
		n = 0
	} else {
		var cbrtG1Omega fp.Element
		cbrtG1Omega.Mul(&cbrtG1, &omega)
		if cbrtG1Omega.Equal(&x1) {
			n = 1
		} else {
			n = 2
		}
	}

	flags = (n & 0x3) << 2
	return
}

// BatchDecompress2G1 decompresses two G1Affine points using only ONE cube root
// extraction (instead of two square roots). Based on https://eprint.iacr.org/2021/1446.pdf
func BatchDecompress2G1(z0, z1 fp.Element, flags byte) (p0, p1 G1Affine, err error) {
	caseIndicator := (flags >> 4) & 0x0F

	switch caseIndicator {
	case 0b11: // Both infinity
		p0.SetInfinity()
		p1.SetInfinity()
		return

	case 0b10: // P0 infinity
		p0.SetInfinity()
		p1.X = z0
		var ySq fp.Element
		ySq.Square(&p1.X).Mul(&ySq, &p1.X).Add(&ySq, &bCurveCoeff)
		if p1.Y.Sqrt(&ySq) == nil {
			err = errors.New("invalid point: not on curve")
			return
		}
		if p1.Y.LexicographicallyLargest() != ((flags & 0b01) != 0) {
			p1.Y.Neg(&p1.Y)
		}
		return

	case 0b01: // P1 infinity
		p1.SetInfinity()
		p0.X = z0
		var ySq fp.Element
		ySq.Square(&p0.X).Mul(&ySq, &p0.X).Add(&ySq, &bCurveCoeff)
		if p0.Y.Sqrt(&ySq) == nil {
			err = errors.New("invalid point: not on curve")
			return
		}
		if p0.Y.LexicographicallyLargest() != ((flags & 0b10) != 0) {
			p0.Y.Neg(&p0.Y)
		}
		return

	case 0b0100: // P0 == P1
		p0.X = z0
		var ySq fp.Element
		ySq.Square(&p0.X).Mul(&ySq, &p0.X).Add(&ySq, &bCurveCoeff)
		if p0.Y.Sqrt(&ySq) == nil {
			err = errors.New("invalid point: not on curve")
			return
		}
		if p0.Y.LexicographicallyLargest() != ((flags & 0b10) != 0) {
			p0.Y.Neg(&p0.Y)
		}
		p1 = p0
		return

	case 0b0101: // P0 == -P1
		p0.X = z0
		var ySq fp.Element
		ySq.Square(&p0.X).Mul(&ySq, &p0.X).Add(&ySq, &bCurveCoeff)
		if p0.Y.Sqrt(&ySq) == nil {
			err = errors.New("invalid point: not on curve")
			return
		}
		if p0.Y.LexicographicallyLargest() != ((flags & 0b10) != 0) {
			p0.Y.Neg(&p0.Y)
		}
		p1.X = p0.X
		p1.Y.Neg(&p0.Y)
		return

	case 0b0110: // P1 = [-ω]^k(P0)
		k := flags & 0x07
		p0.X, p0.Y = z0, z1
		// Apply ω^(k%3) to x and (-1)^k to y
		switch k % 3 {
		case 0:
			p1.X = p0.X
		case 1:
			p1.X.Mul(&p0.X, &thirdRootOneG1)
		case 2:
			p1.X.Mul(&p0.X, &thirdRootOneG2) // ω² = thirdRootOneG1²
		}
		if k%2 == 0 {
			p1.Y = p0.Y
		} else {
			p1.Y.Neg(&p0.Y)
		}
		return

	case 0b00: // Generic case: apply ψ⁻¹₂,₃ then cube root
		n := (flags >> 2) & 0x3

		var z0Sq, z0Cubed, z1Sq fp.Element
		z0Sq.Square(&z0)
		z0Cubed.Mul(&z0Sq, &z0)
		z1Sq.Square(&z1)

		var g1 fp.Element
		g1.Sub(&z1Sq, &bCurveCoeff)

		// Batch inversion: compute 1/z0², 1/z0³, 1/g1 with ONE inversion
		// Using Montgomery's trick: inv(a·b·c) then multiply back
		var z0Fifth, z0FifthG1, invAll fp.Element
		z0Fifth.Mul(&z0Sq, &z0Cubed) // z0⁵
		z0FifthG1.Mul(&z0Fifth, &g1) // z0⁵·g1
		invAll.Inverse(&z0FifthG1)   // 1/(z0⁵·g1)

		var g1Inv, invZ0Fifth, z0CubedInv, z0SqInv fp.Element
		g1Inv.Mul(&invAll, &z0Fifth)       // 1/g1
		invZ0Fifth.Mul(&invAll, &g1)       // 1/z0⁵
		z0CubedInv.Mul(&invZ0Fifth, &z0Sq) // 1/z0³
		z0SqInv.Mul(&invZ0Fifth, &z0Cubed) // 1/z0²

		// t = g1 / z0²
		var t fp.Element
		t.Mul(&g1, &z0SqInv)

		// y0 = (z0³·z1 - 2·z0·(z0 - z1)·g1 - g1²) / z0³
		var term1, term2, g1Sq, y0 fp.Element
		term1.Mul(&z0Cubed, &z1)

		var z0MinusZ1 fp.Element
		z0MinusZ1.Sub(&z0, &z1)
		term2.Double(&z0)
		term2.Mul(&term2, &z0MinusZ1)
		term2.Mul(&term2, &g1)

		g1Sq.Square(&g1)

		y0.Sub(&term1, &term2)
		y0.Sub(&y0, &g1Sq)
		y0.Mul(&y0, &z0CubedInv)

		// y1 = -(z0²·(z0 - 2·z1) + (2·z0 - z1)·g1) / g1
		var twoZ1, z0Minus2z1, twoZ0, twoZ0MinusZ1 fp.Element
		twoZ1.Double(&z1)
		z0Minus2z1.Sub(&z0, &twoZ1)
		twoZ0.Double(&z0)
		twoZ0MinusZ1.Sub(&twoZ0, &z1)

		var part1, part2, y1 fp.Element
		part1.Mul(&z0Sq, &z0Minus2z1)
		part2.Mul(&twoZ0MinusZ1, &g1)
		y1.Add(&part1, &part2)
		y1.Mul(&y1, &g1Inv)
		y1.Neg(&y1)

		// x1 = ω^n · ∛(y1² - b)
		var y1Sq, g1Prime, x1 fp.Element
		y1Sq.Square(&y1)
		g1Prime.Sub(&y1Sq, &bCurveCoeff)

		if x1.Cbrt(&g1Prime) == nil {
			err = errors.New("invalid data: not a cubic residue")
			return
		}

		switch n {
		case 1:
			x1.Mul(&x1, &thirdRootOneG1)
		case 2:
			x1.Mul(&x1, &thirdRootOneG2) // ω² = thirdRootOneG1²
		}

		// x0 = t·x1
		var x0 fp.Element
		x0.Mul(&t, &x1)

		p0.X, p0.Y = x0, y0
		p1.X, p1.Y = x1, y1
		return

	default:
		err = errors.New("invalid flags")
		return
	}
}

// BatchCompress2G2 compresses two G2Affine points into (z0, z1, flags) using
// the birational map χ₂,₃ from https://eprint.iacr.org/2021/1446.pdf (Section 3).
// Decompression requires only ONE cube root extraction instead of two square roots.
func BatchCompress2G2(p0, p1 *G2Affine) (z0, z1 fptower.E2, flags byte, err error) {
	p0Inf := p0.IsInfinity()
	p1Inf := p1.IsInfinity()

	if p0Inf && p1Inf {
		flags = 0b11 << 4
		return
	}

	if p0Inf {
		z0 = p1.X
		flags = 0b10 << 4
		if p1.Y.LexicographicallyLargest() {
			flags |= 0b01
		}
		return
	}

	if p1Inf {
		z0 = p0.X
		flags = 0b01 << 4
		if p0.Y.LexicographicallyLargest() {
			flags |= 0b10
		}
		return
	}

	// Degenerate case: x0 == x1
	if p0.X.Equal(&p1.X) {
		z0 = p0.X
		if p0.Y.Equal(&p1.Y) {
			flags = 0b0100 << 4 // same point
		} else {
			flags = 0b0101 << 4 // negation
		}
		if p0.Y.LexicographicallyLargest() {
			flags |= 0b10
		}
		return
	}

	// Degenerate case: y0² == y1² means P1 = [-ω]^k(P0) for some k ∈ {0,...,5}
	var y0Sq, y1Sq fptower.E2
	y0Sq.Square(&p0.Y)
	y1Sq.Square(&p1.Y)
	if y0Sq.Equal(&y1Sq) {
		z0 = p0.X
		z1 = p0.Y
		flags = 0b0110 << 4
		// Find k: [-ω]^k(x,y) = (ω^k·x, (-1)^k·y)
		// ω = thirdRootOneG2, ω² = thirdRootOneG1 (embedded in Fp2)
		var omega, omega2 fptower.E2
		omega.A0 = thirdRootOneG2
		omega2.A0 = thirdRootOneG1
		var negY fptower.E2
		negY.Neg(&p0.Y)
		// Check k=0: (x0, y0)
		if p0.X.Equal(&p1.X) && p0.Y.Equal(&p1.Y) {
			return
		}
		// Check k=1: (ω·x0, -y0)
		var omegaX fptower.E2
		omegaX.Mul(&p0.X, &omega)
		if omegaX.Equal(&p1.X) && negY.Equal(&p1.Y) {
			flags |= 1
			return
		}
		// Check k=2: (ω²·x0, y0)
		var omega2X fptower.E2
		omega2X.Mul(&p0.X, &omega2)
		if omega2X.Equal(&p1.X) && p0.Y.Equal(&p1.Y) {
			flags |= 2
			return
		}
		// Check k=4: (ω·x0, y0)
		if omegaX.Equal(&p1.X) && p0.Y.Equal(&p1.Y) {
			flags |= 4
			return
		}
		// Check k=5: (ω²·x0, -y0)
		if omega2X.Equal(&p1.X) && negY.Equal(&p1.Y) {
			flags |= 5
			return
		}
		err = errors.New("failed to find automorphism k")
		return
	}

	// Generic case: compute χ₂,₃(P0, P1) = (z0, z1)
	x0, y0, x1, y1 := p0.X, p0.Y, p1.X, p1.Y

	var x0Sq, x1Sq, x0Cubed, x1Cubed fptower.E2
	x0Sq.Square(&x0)
	x1Sq.Square(&x1)
	x0Cubed.Mul(&x0Sq, &x0)
	x1Cubed.Mul(&x1Sq, &x1)

	var denInv fptower.E2
	denInv.Sub(&y0Sq, &y1Sq)
	denInv.Inverse(&denInv)

	// z0 numerator
	var term1, term2, term3, numZ0 fptower.E2
	term1.Mul(&x0Sq, &y1)
	term1.Double(&term1) // 2x0²y1

	var y0MinusY1 fptower.E2
	y0MinusY1.Sub(&y0, &y1)
	term2.Mul(&x0, &x1)
	term2.Mul(&term2, &y0MinusY1) // x0x1(y0 - y1)

	term3.Mul(&y0, &x1Sq)
	term3.Double(&term3) // 2y0x1²

	numZ0.Sub(&term1, &term2)
	numZ0.Sub(&numZ0, &term3)
	numZ0.Mul(&numZ0, &x1)
	z0.Mul(&numZ0, &denInv)

	// z1 numerator
	var numZ1, innerTerm, temp fptower.E2
	numZ1.Mul(&x0Cubed, &y1) // x0³y1

	innerTerm.Mul(&x0, &y1)
	temp.Mul(&y0, &x1)
	innerTerm.Sub(&innerTerm, &temp) // x0y1 - y0x1
	innerTerm.Mul(&innerTerm, &x0)
	innerTerm.Mul(&innerTerm, &x1)
	innerTerm.Double(&innerTerm) // 2x0x1(x0y1 - y0x1)

	numZ1.Add(&numZ1, &innerTerm)
	temp.Mul(&y0, &x1Cubed)
	numZ1.Sub(&numZ1, &temp)
	z1.Mul(&numZ1, &denInv)

	// Determine cube root index n: x1 = ω^n · cbrt(y1² - b)
	var omega fptower.E2
	omega.A0 = thirdRootOneG2
	var g1, cbrtG1 fptower.E2
	g1.Sub(&y1Sq, &bTwistCurveCoeff)
	if cbrtG1.Cbrt(&g1) == nil {
		err = errors.New("g1 is not a cubic residue")
		return
	}

	var n byte
	if cbrtG1.Equal(&x1) {
		n = 0
	} else {
		var cbrtG1Omega fptower.E2
		cbrtG1Omega.Mul(&cbrtG1, &omega)
		if cbrtG1Omega.Equal(&x1) {
			n = 1
		} else {
			n = 2
		}
	}

	flags = (n & 0x3) << 2
	return
}

// BatchDecompress2G2 decompresses two G2Affine points using only ONE cube root
// extraction (instead of two square roots). Based on https://eprint.iacr.org/2021/1446.pdf
func BatchDecompress2G2(z0, z1 fptower.E2, flags byte) (p0, p1 G2Affine, err error) {
	caseIndicator := (flags >> 4) & 0x0F

	switch caseIndicator {
	case 0b11: // Both infinity
		p0.SetInfinity()
		p1.SetInfinity()
		return

	case 0b10: // P0 infinity
		p0.SetInfinity()
		p1.X = z0
		var ySq fptower.E2
		ySq.Square(&p1.X).Mul(&ySq, &p1.X).Add(&ySq, &bTwistCurveCoeff)
		if ySq.Legendre() == -1 {
			err = errors.New("invalid point: not on curve")
			return
		}
		p1.Y.Sqrt(&ySq)
		if p1.Y.LexicographicallyLargest() != ((flags & 0b01) != 0) {
			p1.Y.Neg(&p1.Y)
		}
		return

	case 0b01: // P1 infinity
		p1.SetInfinity()
		p0.X = z0
		var ySq fptower.E2
		ySq.Square(&p0.X).Mul(&ySq, &p0.X).Add(&ySq, &bTwistCurveCoeff)
		if ySq.Legendre() == -1 {
			err = errors.New("invalid point: not on curve")
			return
		}
		p0.Y.Sqrt(&ySq)
		if p0.Y.LexicographicallyLargest() != ((flags & 0b10) != 0) {
			p0.Y.Neg(&p0.Y)
		}
		return

	case 0b0100: // P0 == P1
		p0.X = z0
		var ySq fptower.E2
		ySq.Square(&p0.X).Mul(&ySq, &p0.X).Add(&ySq, &bTwistCurveCoeff)
		if ySq.Legendre() == -1 {
			err = errors.New("invalid point: not on curve")
			return
		}
		p0.Y.Sqrt(&ySq)
		if p0.Y.LexicographicallyLargest() != ((flags & 0b10) != 0) {
			p0.Y.Neg(&p0.Y)
		}
		p1 = p0
		return

	case 0b0101: // P0 == -P1
		p0.X = z0
		var ySq fptower.E2
		ySq.Square(&p0.X).Mul(&ySq, &p0.X).Add(&ySq, &bTwistCurveCoeff)
		if ySq.Legendre() == -1 {
			err = errors.New("invalid point: not on curve")
			return
		}
		p0.Y.Sqrt(&ySq)
		if p0.Y.LexicographicallyLargest() != ((flags & 0b10) != 0) {
			p0.Y.Neg(&p0.Y)
		}
		p1.X = p0.X
		p1.Y.Neg(&p0.Y)
		return

	case 0b0110: // P1 = [-ω]^k(P0)
		k := flags & 0x07
		p0.X, p0.Y = z0, z1
		// Apply ω^(k%3) to x and (-1)^k to y
		// ω = thirdRootOneG2, ω² = thirdRootOneG1
		var omega, omega2 fptower.E2
		omega.A0 = thirdRootOneG2
		omega2.A0 = thirdRootOneG1
		switch k % 3 {
		case 0:
			p1.X = p0.X
		case 1:
			p1.X.Mul(&p0.X, &omega)
		case 2:
			p1.X.Mul(&p0.X, &omega2)
		}
		if k%2 == 0 {
			p1.Y = p0.Y
		} else {
			p1.Y.Neg(&p0.Y)
		}
		return

	case 0b00: // Generic case: apply ψ⁻¹₂,₃ then cube root
		n := (flags >> 2) & 0x3

		var z0Sq, z0Cubed, z1Sq fptower.E2
		z0Sq.Square(&z0)
		z0Cubed.Mul(&z0Sq, &z0)
		z1Sq.Square(&z1)

		var g1 fptower.E2
		g1.Sub(&z1Sq, &bTwistCurveCoeff)

		// Batch inversion: compute 1/z0², 1/z0³, 1/g1 with ONE inversion
		var z0Fifth, z0FifthG1, invAll fptower.E2
		z0Fifth.Mul(&z0Sq, &z0Cubed)
		z0FifthG1.Mul(&z0Fifth, &g1)
		invAll.Inverse(&z0FifthG1)

		var g1Inv, invZ0Fifth, z0CubedInv, z0SqInv fptower.E2
		g1Inv.Mul(&invAll, &z0Fifth)
		invZ0Fifth.Mul(&invAll, &g1)
		z0CubedInv.Mul(&invZ0Fifth, &z0Sq)
		z0SqInv.Mul(&invZ0Fifth, &z0Cubed)

		// t = g1 / z0²
		var t fptower.E2
		t.Mul(&g1, &z0SqInv)

		// y0 = (z0³·z1 - 2·z0·(z0 - z1)·g1 - g1²) / z0³
		var term1, term2, g1Sq, y0 fptower.E2
		term1.Mul(&z0Cubed, &z1)

		var z0MinusZ1 fptower.E2
		z0MinusZ1.Sub(&z0, &z1)
		term2.Double(&z0)
		term2.Mul(&term2, &z0MinusZ1)
		term2.Mul(&term2, &g1)

		g1Sq.Square(&g1)

		y0.Sub(&term1, &term2)
		y0.Sub(&y0, &g1Sq)
		y0.Mul(&y0, &z0CubedInv)

		// y1 = -(z0²·(z0 - 2·z1) + (2·z0 - z1)·g1) / g1
		var twoZ1, z0Minus2z1, twoZ0, twoZ0MinusZ1 fptower.E2
		twoZ1.Double(&z1)
		z0Minus2z1.Sub(&z0, &twoZ1)
		twoZ0.Double(&z0)
		twoZ0MinusZ1.Sub(&twoZ0, &z1)

		var part1, part2, y1 fptower.E2
		part1.Mul(&z0Sq, &z0Minus2z1)
		part2.Mul(&twoZ0MinusZ1, &g1)
		y1.Add(&part1, &part2)
		y1.Mul(&y1, &g1Inv)
		y1.Neg(&y1)

		// x1 = ω^n · ∛(y1² - b)
		var y1Sq, g1Prime, x1 fptower.E2
		y1Sq.Square(&y1)
		g1Prime.Sub(&y1Sq, &bTwistCurveCoeff)

		if x1.CbrtFrobenius(&g1Prime) == nil {
			err = errors.New("invalid data: not a cubic residue")
			return
		}

		// ω = thirdRootOneG2 (in Fp, embedded as (ω,0) in Fp2)
		// ω² = thirdRootOneG1 (since thirdRootOneG2 = thirdRootOneG1²)
		var omega, omega2 fptower.E2
		omega.A0 = thirdRootOneG2
		omega2.A0 = thirdRootOneG1
		switch n {
		case 1:
			x1.Mul(&x1, &omega)
		case 2:
			x1.Mul(&x1, &omega2)
		}

		// x0 = t·x1
		var x0 fptower.E2
		x0.Mul(&t, &x1)

		p0.X, p0.Y = x0, y0
		p1.X, p1.Y = x1, y1
		return

	default:
		err = errors.New("invalid flags")
		return
	}
}

// Batch compression sizes for slices
const (
	// SizeOfBatchCompressedG1Pair is the size of two G1 points batch-compressed: z0 + z1 + flags
	SizeOfBatchCompressedG1Pair = fp.Bytes + fp.Bytes + 1 // 97 bytes

	// SizeOfBatchCompressedG2Pair is the size of two G2 points batch-compressed: z0 + z1 + flags
	SizeOfBatchCompressedG2Pair = 2*fp.Bytes + 2*fp.Bytes + 1 // 193 bytes
)

// BatchCompressG1Slice compresses a slice of G1Affine points using 2-by-2 batch compression.
// This is more efficient than standard compression for decompression (uses cube roots instead of square roots).
// If the slice has an odd length, the last point is compressed using standard compression.
//
// Returns the compressed bytes. The format is:
//   - For each pair of points: z0 (48 bytes) + z1 (48 bytes) + flags (1 byte)
//   - If odd length: last point in standard compressed form (48 bytes)
func BatchCompressG1Slice(points []G1Affine) ([]byte, error) {
	n := len(points)
	if n == 0 {
		return nil, nil
	}

	nPairs := n / 2
	hasOdd := n%2 == 1

	// Calculate total size
	totalSize := nPairs * SizeOfBatchCompressedG1Pair
	if hasOdd {
		totalSize += SizeOfG1AffineCompressed
	}

	result := make([]byte, totalSize)
	offset := 0

	// Compress pairs
	for i := 0; i < nPairs; i++ {
		z0, z1, flags, err := BatchCompress2G1(&points[i*2], &points[i*2+1])
		if err != nil {
			return nil, err
		}

		// Write z0
		z0Bytes := z0.Bytes()
		copy(result[offset:offset+fp.Bytes], z0Bytes[:])
		offset += fp.Bytes

		// Write z1
		z1Bytes := z1.Bytes()
		copy(result[offset:offset+fp.Bytes], z1Bytes[:])
		offset += fp.Bytes

		// Write flags
		result[offset] = flags
		offset++
	}

	// Handle odd point with standard compression
	if hasOdd {
		lastBytes := points[n-1].Bytes()
		copy(result[offset:], lastBytes[:])
	}

	return result, nil
}

// BatchDecompressG1Slice decompresses a slice of G1Affine points from batch-compressed form.
// The input must have been created by BatchCompressG1Slice.
// Decompression is parallelized across pairs for better performance.
//
// This implementation uses batch inversion (Montgomery's trick) to compute all
// fp inversions in one shot, reducing the cost.
//
// Parameters:
//   - data: the compressed bytes
//   - n: the number of points to decompress
func BatchDecompressG1Slice(data []byte, n int) ([]G1Affine, error) {
	if n == 0 {
		return nil, nil
	}

	nPairs := n / 2
	hasOdd := n%2 == 1

	// Verify data size
	expectedSize := nPairs * SizeOfBatchCompressedG1Pair
	if hasOdd {
		expectedSize += SizeOfG1AffineCompressed
	}
	if len(data) < expectedSize {
		return nil, errors.New("insufficient data for batch decompression")
	}

	points := make([]G1Affine, n)

	// Parse compressed data
	// Store z0,z1 in points[i*2].X and points[i*2+1].X
	// Store flags in points[i*2].Y[0]
	offset := 0
	for i := 0; i < nPairs; i++ {
		if err := points[i*2].X.SetBytesCanonical(data[offset : offset+fp.Bytes]); err != nil {
			return nil, err
		}
		offset += fp.Bytes

		if err := points[i*2+1].X.SetBytesCanonical(data[offset : offset+fp.Bytes]); err != nil {
			return nil, err
		}
		offset += fp.Bytes

		// Store flags in Y[0] as scratch space
		points[i*2].Y[0] = uint64(data[offset])
		offset++
	}

	// ============================================================
	// PHASE 1: Compute values to invert for generic cases
	// ============================================================
	// For generic case (flags >> 4 == 0b00), we need to invert z0FifthG1 = z0⁵ * g1

	genericMask := make([]bool, nPairs)
	flagsArray := make([]byte, nPairs)
	toInvert := make([]fp.Element, 0, nPairs)
	invertIdx := make([]int, 0, nPairs)

	for i := 0; i < nPairs; i++ {
		flags := byte(points[i*2].Y[0])
		flagsArray[i] = flags
		caseType := flags >> 4

		if caseType == 0b0000 { // Generic case
			genericMask[i] = true
			z0 := points[i*2].X
			z1 := points[i*2+1].X

			var z0Sq, z0Cubed, z1Sq, g1, z0Fifth, z0FifthG1 fp.Element
			z0Sq.Square(&z0)
			z0Cubed.Mul(&z0Sq, &z0)
			z1Sq.Square(&z1)
			g1.Sub(&z1Sq, &bCurveCoeff)
			z0Fifth.Mul(&z0Sq, &z0Cubed)
			z0FifthG1.Mul(&z0Fifth, &g1)

			// Store intermediate values for later use:
			// points[i*2].Y = z0FifthG1 (will be replaced by inverted value)
			// points[i*2+1].Y = z0Sq
			points[i*2].Y = z0FifthG1
			points[i*2+1].Y = z0Sq

			toInvert = append(toInvert, z0FifthG1)
			invertIdx = append(invertIdx, i)
		}
	}

	// ============================================================
	// PHASE 2: Batch invert all generic case values
	// ============================================================
	var inverted []fp.Element
	if len(toInvert) > 0 {
		inverted = fp.BatchInvert(toInvert)

		// Store inverted values back
		for j, i := range invertIdx {
			points[i*2].Y = inverted[j] // Now contains invAll = 1/(z0FifthG1)
		}
	}

	// ============================================================
	// PHASE 3: Complete decompression in parallel
	// ============================================================
	var nbErrs uint64
	parallel.Execute(nPairs, func(start, end int) {
		for i := start; i < end; i++ {
			z0 := points[i*2].X
			z1 := points[i*2+1].X
			flags := flagsArray[i]

			if genericMask[i] {
				// Generic case: use pre-computed inverse
				invAll := points[i*2].Y // 1/(z0FifthG1)
				z0Sq := points[i*2+1].Y // stored earlier

				p0, p1, err := batchDecompress2G1WithInv(z0, z1, flags, invAll, z0Sq)
				if err != nil {
					atomic.AddUint64(&nbErrs, 1)
					continue
				}
				points[i*2] = p0
				points[i*2+1] = p1
			} else {
				// Special case: use original function
				p0, p1, err := BatchDecompress2G1(z0, z1, flags)
				if err != nil {
					atomic.AddUint64(&nbErrs, 1)
					continue
				}
				points[i*2] = p0
				points[i*2+1] = p1
			}
		}
	})

	if nbErrs != 0 {
		return nil, errors.New("batch decompression failed")
	}

	// Handle odd point with standard decompression
	if hasOdd {
		if _, err := points[n-1].SetBytes(data[offset : offset+SizeOfG1AffineCompressed]); err != nil {
			return nil, err
		}
	}

	return points, nil
}

// batchDecompress2G1WithInv is the generic case decompression with pre-computed inverse.
// This avoids the expensive fp.Inverse call by using batch inversion.
func batchDecompress2G1WithInv(z0, z1 fp.Element, flags byte, invAll, z0Sq fp.Element) (p0, p1 G1Affine, err error) {
	n := (flags >> 2) & 0x3

	// Recompute needed values (z0Cubed, g1, z0Fifth)
	var z0Cubed, z1Sq, g1, z0Fifth fp.Element
	z0Cubed.Mul(&z0Sq, &z0)
	z1Sq.Square(&z1)
	g1.Sub(&z1Sq, &bCurveCoeff)
	z0Fifth.Mul(&z0Sq, &z0Cubed)

	// Compute individual inverses from batch inverse
	// invAll = 1/(z0Fifth * g1)
	// g1Inv = invAll * z0Fifth = 1/g1
	// invZ0Fifth = invAll * g1 = 1/z0Fifth
	// z0CubedInv = invZ0Fifth * z0Sq = 1/z0³
	// z0SqInv = invZ0Fifth * z0Cubed = 1/z0²
	var g1Inv, invZ0Fifth, z0CubedInv, z0SqInv fp.Element
	g1Inv.Mul(&invAll, &z0Fifth)
	invZ0Fifth.Mul(&invAll, &g1)
	z0CubedInv.Mul(&invZ0Fifth, &z0Sq)
	z0SqInv.Mul(&invZ0Fifth, &z0Cubed)

	// t = g1 / z0²
	var t fp.Element
	t.Mul(&g1, &z0SqInv)

	// y0 = (z0³·z1 - 2·z0·(z0 - z1)·g1 - g1²) / z0³
	var term1, term2, g1Sq, y0 fp.Element
	term1.Mul(&z0Cubed, &z1)

	var z0MinusZ1 fp.Element
	z0MinusZ1.Sub(&z0, &z1)
	term2.Double(&z0)
	term2.Mul(&term2, &z0MinusZ1)
	term2.Mul(&term2, &g1)

	g1Sq.Square(&g1)

	y0.Sub(&term1, &term2)
	y0.Sub(&y0, &g1Sq)
	y0.Mul(&y0, &z0CubedInv)

	// y1 = -(z0²·(z0 - 2·z1) + (2·z0 - z1)·g1) / g1
	var twoZ1, z0Minus2z1, twoZ0, twoZ0MinusZ1 fp.Element
	twoZ1.Double(&z1)
	z0Minus2z1.Sub(&z0, &twoZ1)
	twoZ0.Double(&z0)
	twoZ0MinusZ1.Sub(&twoZ0, &z1)

	var part1, part2, y1 fp.Element
	part1.Mul(&z0Sq, &z0Minus2z1)
	part2.Mul(&twoZ0MinusZ1, &g1)
	y1.Add(&part1, &part2)
	y1.Mul(&y1, &g1Inv)
	y1.Neg(&y1)

	// x1 = ω^n · ∛(y1² - b)
	var y1Sq, g1Prime, x1 fp.Element
	y1Sq.Square(&y1)
	g1Prime.Sub(&y1Sq, &bCurveCoeff)

	if x1.Cbrt(&g1Prime) == nil {
		err = errors.New("invalid data: not a cubic residue")
		return
	}

	switch n {
	case 1:
		x1.Mul(&x1, &thirdRootOneG1)
	case 2:
		x1.Mul(&x1, &thirdRootOneG2)
	}

	// x0 = t·x1
	var x0 fp.Element
	x0.Mul(&t, &x1)

	p0.X, p0.Y = x0, y0
	p1.X, p1.Y = x1, y1
	return
}

// BatchCompressG2Slice compresses a slice of G2Affine points using 2-by-2 batch compression.
// This is more efficient than standard compression for decompression.
// If the slice has an odd length, the last point is compressed using standard compression.
//
// Returns the compressed bytes. The format is:
//   - For each pair of points: z0 (96 bytes) + z1 (96 bytes) + flags (1 byte)
//   - If odd length: last point in standard compressed form (96 bytes)
func BatchCompressG2Slice(points []G2Affine) ([]byte, error) {
	n := len(points)
	if n == 0 {
		return nil, nil
	}

	nPairs := n / 2
	hasOdd := n%2 == 1

	// Calculate total size
	totalSize := nPairs * SizeOfBatchCompressedG2Pair
	if hasOdd {
		totalSize += SizeOfG2AffineCompressed
	}

	result := make([]byte, totalSize)
	offset := 0

	// Compress pairs
	for i := 0; i < nPairs; i++ {
		z0, z1, flags, err := BatchCompress2G2(&points[i*2], &points[i*2+1])
		if err != nil {
			return nil, err
		}

		// Write z0 (A1 | A0)
		z0A1Bytes := z0.A1.Bytes()
		copy(result[offset:offset+fp.Bytes], z0A1Bytes[:])
		offset += fp.Bytes
		z0A0Bytes := z0.A0.Bytes()
		copy(result[offset:offset+fp.Bytes], z0A0Bytes[:])
		offset += fp.Bytes

		// Write z1 (A1 | A0)
		z1A1Bytes := z1.A1.Bytes()
		copy(result[offset:offset+fp.Bytes], z1A1Bytes[:])
		offset += fp.Bytes
		z1A0Bytes := z1.A0.Bytes()
		copy(result[offset:offset+fp.Bytes], z1A0Bytes[:])
		offset += fp.Bytes

		// Write flags
		result[offset] = flags
		offset++
	}

	// Handle odd point with standard compression
	if hasOdd {
		lastBytes := points[n-1].Bytes()
		copy(result[offset:], lastBytes[:])
	}

	return result, nil
}

// BatchDecompressG2Slice decompresses a slice of G2Affine points from batch-compressed form.
// The input must have been created by BatchCompressG2Slice.
// Decompression is parallelized across pairs for better performance.
//
// This implementation uses batch inversion (Montgomery's trick) to compute all
// E2 inversions in one shot, significantly reducing the cost.
//
// Parameters:
//   - data: the compressed bytes
//   - n: the number of points to decompress
func BatchDecompressG2Slice(data []byte, n int) ([]G2Affine, error) {
	if n == 0 {
		return nil, nil
	}

	nPairs := n / 2
	hasOdd := n%2 == 1

	// Verify data size
	expectedSize := nPairs * SizeOfBatchCompressedG2Pair
	if hasOdd {
		expectedSize += SizeOfG2AffineCompressed
	}
	if len(data) < expectedSize {
		return nil, errors.New("insufficient data for batch decompression")
	}

	points := make([]G2Affine, n)

	// Parse compressed data
	// Store z0,z1 in points[i*2].X and points[i*2+1].X
	// Store flags in points[i*2].Y.A0[0]
	offset := 0
	for i := 0; i < nPairs; i++ {
		// Read z0 (A1 | A0) into points[i*2].X
		if err := points[i*2].X.A1.SetBytesCanonical(data[offset : offset+fp.Bytes]); err != nil {
			return nil, err
		}
		offset += fp.Bytes
		if err := points[i*2].X.A0.SetBytesCanonical(data[offset : offset+fp.Bytes]); err != nil {
			return nil, err
		}
		offset += fp.Bytes

		// Read z1 (A1 | A0) into points[i*2+1].X
		if err := points[i*2+1].X.A1.SetBytesCanonical(data[offset : offset+fp.Bytes]); err != nil {
			return nil, err
		}
		offset += fp.Bytes
		if err := points[i*2+1].X.A0.SetBytesCanonical(data[offset : offset+fp.Bytes]); err != nil {
			return nil, err
		}
		offset += fp.Bytes

		// Store flags in Y.A0[0] as scratch space
		points[i*2].Y.A0[0] = uint64(data[offset])
		offset++
	}

	// ============================================================
	// PHASE 1: Compute values to invert for generic cases
	// ============================================================
	// For generic case (flags >> 4 == 0b00), we need to invert z0FifthG1 = z0⁵ * g1

	genericMask := make([]bool, nPairs)
	flagsArray := make([]byte, nPairs)
	toInvert := make([]fptower.E2, 0, nPairs)
	invertIdx := make([]int, 0, nPairs)

	for i := 0; i < nPairs; i++ {
		flags := byte(points[i*2].Y.A0[0])
		flagsArray[i] = flags
		caseType := flags >> 4

		if caseType == 0b0000 { // Generic case
			genericMask[i] = true
			z0 := points[i*2].X
			z1 := points[i*2+1].X

			var z0Sq, z0Cubed, z1Sq, g1, z0Fifth, z0FifthG1 fptower.E2
			z0Sq.Square(&z0)
			z0Cubed.Mul(&z0Sq, &z0)
			z1Sq.Square(&z1)
			g1.Sub(&z1Sq, &bTwistCurveCoeff)
			z0Fifth.Mul(&z0Sq, &z0Cubed)
			z0FifthG1.Mul(&z0Fifth, &g1)

			// Store intermediate values for later use:
			// points[i*2].Y = z0FifthG1 (will be replaced by inverted value)
			// points[i*2+1].Y = z0Sq
			points[i*2].Y = z0FifthG1
			points[i*2+1].Y = z0Sq

			toInvert = append(toInvert, z0FifthG1)
			invertIdx = append(invertIdx, i)
		}
	}

	// ============================================================
	// PHASE 2: Batch invert all generic case values
	// ============================================================
	var inverted []fptower.E2
	if len(toInvert) > 0 {
		inverted = fptower.BatchInvertE2(toInvert)

		// Store inverted values back
		for j, i := range invertIdx {
			points[i*2].Y = inverted[j] // Now contains invAll = 1/(z0FifthG1)
		}
	}

	// ============================================================
	// PHASE 3: Complete decompression in parallel
	// ============================================================
	var nbErrs uint64
	parallel.Execute(nPairs, func(start, end int) {
		for i := start; i < end; i++ {
			z0 := points[i*2].X
			z1 := points[i*2+1].X
			flags := flagsArray[i]

			if genericMask[i] {
				// Generic case: use pre-computed inverse
				invAll := points[i*2].Y // 1/(z0FifthG1)
				z0Sq := points[i*2+1].Y // stored earlier

				p0, p1, err := batchDecompress2G2WithInv(z0, z1, flags, invAll, z0Sq)
				if err != nil {
					atomic.AddUint64(&nbErrs, 1)
					continue
				}
				points[i*2] = p0
				points[i*2+1] = p1
			} else {
				// Special case: use original function
				p0, p1, err := BatchDecompress2G2(z0, z1, flags)
				if err != nil {
					atomic.AddUint64(&nbErrs, 1)
					continue
				}
				points[i*2] = p0
				points[i*2+1] = p1
			}
		}
	})

	if nbErrs != 0 {
		return nil, errors.New("batch decompression failed")
	}

	// Handle odd point with standard decompression
	if hasOdd {
		if _, err := points[n-1].SetBytes(data[offset : offset+SizeOfG2AffineCompressed]); err != nil {
			return nil, err
		}
	}

	return points, nil
}

// batchDecompress2G2WithInv is the generic case decompression with pre-computed inverse.
// This avoids the expensive E2.Inverse call by using batch inversion.
func batchDecompress2G2WithInv(z0, z1 fptower.E2, flags byte, invAll, z0Sq fptower.E2) (p0, p1 G2Affine, err error) {
	n := (flags >> 2) & 0x3

	// Recompute needed values (z0Cubed, g1, z0Fifth)
	var z0Cubed, z1Sq, g1, z0Fifth fptower.E2
	z0Cubed.Mul(&z0Sq, &z0)
	z1Sq.Square(&z1)
	g1.Sub(&z1Sq, &bTwistCurveCoeff)
	z0Fifth.Mul(&z0Sq, &z0Cubed)

	// Compute individual inverses from batch inverse
	// invAll = 1/(z0Fifth * g1)
	// g1Inv = invAll * z0Fifth = 1/g1
	// invZ0Fifth = invAll * g1 = 1/z0Fifth
	// z0CubedInv = invZ0Fifth * z0Sq = 1/z0³
	// z0SqInv = invZ0Fifth * z0Cubed = 1/z0²
	var g1Inv, invZ0Fifth, z0CubedInv, z0SqInv fptower.E2
	g1Inv.Mul(&invAll, &z0Fifth)
	invZ0Fifth.Mul(&invAll, &g1)
	z0CubedInv.Mul(&invZ0Fifth, &z0Sq)
	z0SqInv.Mul(&invZ0Fifth, &z0Cubed)

	// t = g1 / z0²
	var t fptower.E2
	t.Mul(&g1, &z0SqInv)

	// y0 = (z0³·z1 - 2·z0·(z0 - z1)·g1 - g1²) / z0³
	var term1, term2, g1Sq, y0 fptower.E2
	term1.Mul(&z0Cubed, &z1)

	var z0MinusZ1 fptower.E2
	z0MinusZ1.Sub(&z0, &z1)
	term2.Double(&z0)
	term2.Mul(&term2, &z0MinusZ1)
	term2.Mul(&term2, &g1)

	g1Sq.Square(&g1)

	y0.Sub(&term1, &term2)
	y0.Sub(&y0, &g1Sq)
	y0.Mul(&y0, &z0CubedInv)

	// y1 = -(z0²·(z0 - 2·z1) + (2·z0 - z1)·g1) / g1
	var twoZ1, z0Minus2z1, twoZ0, twoZ0MinusZ1 fptower.E2
	twoZ1.Double(&z1)
	z0Minus2z1.Sub(&z0, &twoZ1)
	twoZ0.Double(&z0)
	twoZ0MinusZ1.Sub(&twoZ0, &z1)

	var part1, part2, y1 fptower.E2
	part1.Mul(&z0Sq, &z0Minus2z1)
	part2.Mul(&twoZ0MinusZ1, &g1)
	y1.Add(&part1, &part2)
	y1.Mul(&y1, &g1Inv)
	y1.Neg(&y1)

	// x1 = ω^n · ∛(y1² - b)
	var y1Sq, g1Prime, x1 fptower.E2
	y1Sq.Square(&y1)
	g1Prime.Sub(&y1Sq, &bTwistCurveCoeff)

	if x1.CbrtFrobenius(&g1Prime) == nil {
		err = errors.New("invalid data: not a cubic residue")
		return
	}

	// ω = thirdRootOneG2 (in Fp, embedded as (ω,0) in Fp2)
	// ω² = thirdRootOneG1 (since thirdRootOneG2 = thirdRootOneG1²)
	var omega, omega2 fptower.E2
	omega.A0 = thirdRootOneG2
	omega2.A0 = thirdRootOneG1
	switch n {
	case 1:
		x1.Mul(&x1, &omega)
	case 2:
		x1.Mul(&x1, &omega2)
	}

	// x0 = t·x1
	var x0 fptower.E2
	x0.Mul(&t, &x1)

	p0.X, p0.Y = x0, y0
	p1.X, p1.Y = x1, y1
	return
}
