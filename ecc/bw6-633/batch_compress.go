// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bw6633

import (
	"errors"
	"sync/atomic"

	"github.com/consensys/gnark-crypto/ecc/bw6-633/fp"
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
// Note: For BW6-633, G2 is also defined over Fp (not Fp2), so this uses fp.Element.
func BatchCompress2G2(p0, p1 *G2Affine) (z0, z1 fp.Element, flags byte, err error) {
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
		// For G2: ω = thirdRootOneG2, ω² = thirdRootOneG1
		var negY fp.Element
		negY.Neg(&p0.Y)
		// Check k=0: (x0, y0)
		if p0.X.Equal(&p1.X) && p0.Y.Equal(&p1.Y) {
			return
		}
		// Check k=1: (ω·x0, -y0)
		var omegaX fp.Element
		omegaX.Mul(&p0.X, &thirdRootOneG2)
		if omegaX.Equal(&p1.X) && negY.Equal(&p1.Y) {
			flags |= 1
			return
		}
		// Check k=2: (ω²·x0, y0)
		var omega2X fp.Element
		omega2X.Mul(&p0.X, &thirdRootOneG1)
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
	// For G2: uses bTwistCurveCoeff and thirdRootOneG2
	omega := thirdRootOneG2
	var g1, cbrtG1 fp.Element
	g1.Sub(&y1Sq, &bTwistCurveCoeff)
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

// BatchDecompress2G2 decompresses two G2Affine points using only ONE cube root
// extraction (instead of two square roots). Based on https://eprint.iacr.org/2021/1446.pdf
// Note: For BW6-633, G2 is also defined over Fp (not Fp2).
func BatchDecompress2G2(z0, z1 fp.Element, flags byte) (p0, p1 G2Affine, err error) {
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
		ySq.Square(&p1.X).Mul(&ySq, &p1.X).Add(&ySq, &bTwistCurveCoeff)
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
		ySq.Square(&p0.X).Mul(&ySq, &p0.X).Add(&ySq, &bTwistCurveCoeff)
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
		ySq.Square(&p0.X).Mul(&ySq, &p0.X).Add(&ySq, &bTwistCurveCoeff)
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
		ySq.Square(&p0.X).Mul(&ySq, &p0.X).Add(&ySq, &bTwistCurveCoeff)
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
		// For G2: ω = thirdRootOneG2, ω² = thirdRootOneG1
		switch k % 3 {
		case 0:
			p1.X = p0.X
		case 1:
			p1.X.Mul(&p0.X, &thirdRootOneG2)
		case 2:
			p1.X.Mul(&p0.X, &thirdRootOneG1)
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
		g1.Sub(&z1Sq, &bTwistCurveCoeff)

		// Batch inversion: compute 1/z0², 1/z0³, 1/g1 with ONE inversion
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
		g1Prime.Sub(&y1Sq, &bTwistCurveCoeff)

		if x1.Cbrt(&g1Prime) == nil {
			err = errors.New("invalid data: not a cubic residue")
			return
		}

		// For G2: ω = thirdRootOneG2, ω² = thirdRootOneG1
		switch n {
		case 1:
			x1.Mul(&x1, &thirdRootOneG2)
		case 2:
			x1.Mul(&x1, &thirdRootOneG1)
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

// Batch compression sizes for slices
const (
	// SizeOfBatchCompressedG1Pair is the size of two G1 points batch-compressed.
	// Generic case: 160 bytes (z0 + z1, n encoded in z0's high bits)
	// Degenerate case: 161 bytes (z0 + z1 + flags byte)
	SizeOfBatchCompressedG1Pair           = fp.Bytes + fp.Bytes     // 160 bytes (generic)
	SizeOfBatchCompressedG1PairDegenerate = fp.Bytes + fp.Bytes + 1 // 161 bytes (degenerate)

	// SizeOfBatchCompressedG2Pair is the size of two G2 points batch-compressed.
	// Generic case: 160 bytes (z0 + z1, n encoded in z0's high bits)
	// Degenerate case: 161 bytes (z0 + z1 + flags byte)
	SizeOfBatchCompressedG2Pair           = fp.Bytes + fp.Bytes     // 160 bytes (generic)
	SizeOfBatchCompressedG2PairDegenerate = fp.Bytes + fp.Bytes + 1 // 161 bytes (degenerate)
)

// BatchCompressG1Slice compresses a slice of G1Affine points using 2-by-2 batch compression.
// This is more efficient than standard compression for decompression (uses cube roots instead of square roots).
// If the slice has an odd length, the last point is compressed using standard compression.
// Compression is parallelized across pairs (2 points per thread).
//
// Returns the compressed bytes. The format is:
//   - For generic pairs: z0 (80 bytes, n encoded in bits 7-6) + z1 (80 bytes) = 160 bytes
//   - For degenerate pairs: z0 (80 bytes, bits 7-6 = 11) + z1 (80 bytes) + flags (1 byte) = 161 bytes
//   - If odd length: last point in standard compressed form (80 bytes)
func BatchCompressG1Slice(points []G1Affine) ([]byte, error) {
	n := len(points)
	if n == 0 {
		return nil, nil
	}

	nPairs := n / 2
	hasOdd := n%2 == 1

	// Pre-allocate for worst case (all degenerate pairs)
	maxSize := nPairs * SizeOfBatchCompressedG1PairDegenerate
	if hasOdd {
		maxSize += SizeOfG1AffineCompressed
	}

	result := make([]byte, maxSize)

	// Track actual sizes per pair
	pairSizes := make([]int, nPairs)
	var nbErrs uint64
	parallel.Execute(nPairs, func(start, end int) {
		for i := start; i < end; i++ {
			z0, z1, flags, err := BatchCompress2G1(&points[i*2], &points[i*2+1])
			if err != nil {
				atomic.AddUint64(&nbErrs, 1)
				continue
			}

			// Determine if generic (case 0) or degenerate
			caseIndicator := (flags >> 4) & 0x0F
			isGeneric := caseIndicator == 0

			if isGeneric {
				pairSizes[i] = SizeOfBatchCompressedG1Pair // 160 bytes
			} else {
				pairSizes[i] = SizeOfBatchCompressedG1PairDegenerate // 161 bytes
			}

			// Calculate offset based on worst-case positions
			offset := i * SizeOfBatchCompressedG1PairDegenerate

			// Write z0
			z0Bytes := z0.Bytes()
			if isGeneric {
				// Encode n in bits 7-6 of z0's first byte
				cubeRootIdx := (flags >> 2) & 0x3
				z0Bytes[0] |= (cubeRootIdx << 6)
			} else {
				// Set bits 7-6 to 0b11 as degenerate marker
				z0Bytes[0] |= 0xC0
			}
			copy(result[offset:offset+fp.Bytes], z0Bytes[:])
			offset += fp.Bytes

			// Write z1
			z1Bytes := z1.Bytes()
			copy(result[offset:offset+fp.Bytes], z1Bytes[:])
			offset += fp.Bytes

			// Write flags only for degenerate cases
			if !isGeneric {
				result[offset] = flags
			}
		}
	})

	if nbErrs != 0 {
		return nil, errors.New("batch compression failed")
	}

	// Compact the result
	actualSize := 0
	for i := 0; i < nPairs; i++ {
		actualSize += pairSizes[i]
	}
	if hasOdd {
		actualSize += SizeOfG1AffineCompressed
	}

	compacted := make([]byte, actualSize)
	writeOffset := 0
	for i := 0; i < nPairs; i++ {
		srcOffset := i * SizeOfBatchCompressedG1PairDegenerate
		copy(compacted[writeOffset:writeOffset+pairSizes[i]], result[srcOffset:srcOffset+pairSizes[i]])
		writeOffset += pairSizes[i]
	}

	// Handle odd point with standard compression
	if hasOdd {
		lastBytes := points[n-1].Bytes()
		copy(compacted[writeOffset:], lastBytes[:])
	}

	return compacted, nil
}

// BatchDecompressG1Slice decompresses a slice of G1Affine points from batch-compressed form.
// The input must have been created by BatchCompressG1Slice.
// Decompression is parallelized across pairs (2 points per thread), with each thread
// performing its own inversion.
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

	// Minimum size check (all generic pairs)
	minSize := nPairs * SizeOfBatchCompressedG1Pair
	if hasOdd {
		minSize += SizeOfG1AffineCompressed
	}
	if len(data) < minSize {
		return nil, errors.New("insufficient data for batch decompression")
	}

	points := make([]G1Affine, n)

	// First pass: scan data to find pair offsets
	offsets := make([]int, nPairs+1)
	offset := 0
	for i := 0; i < nPairs; i++ {
		offsets[i] = offset
		// Check bits 7-6 of z0's first byte
		highBits := (data[offset] >> 6) & 0x3
		if highBits == 0x3 {
			// Degenerate case: 161 bytes
			offset += SizeOfBatchCompressedG1PairDegenerate
		} else {
			// Generic case: 160 bytes
			offset += SizeOfBatchCompressedG1Pair
		}
	}
	offsets[nPairs] = offset

	// Verify we have enough data
	expectedSize := offset
	if hasOdd {
		expectedSize += SizeOfG1AffineCompressed
	}
	if len(data) < expectedSize {
		return nil, errors.New("insufficient data for batch decompression")
	}

	// Decompress pairs in parallel
	var nbErrs uint64
	parallel.Execute(nPairs, func(start, end int) {
		for i := start; i < end; i++ {
			pairOffset := offsets[i]

			// Read z0's first byte to determine format
			z0FirstByte := data[pairOffset]
			highBits := (z0FirstByte >> 6) & 0x3
			isDegenerate := highBits == 0x3

			// Make a copy of z0 bytes and clear the high bits
			var z0Bytes [fp.Bytes]byte
			copy(z0Bytes[:], data[pairOffset:pairOffset+fp.Bytes])
			z0Bytes[0] &= 0x3F // Clear bits 7-6

			var z0, z1 fp.Element
			if err := z0.SetBytesCanonical(z0Bytes[:]); err != nil {
				atomic.AddUint64(&nbErrs, 1)
				continue
			}

			z1Offset := pairOffset + fp.Bytes
			if err := z1.SetBytesCanonical(data[z1Offset : z1Offset+fp.Bytes]); err != nil {
				atomic.AddUint64(&nbErrs, 1)
				continue
			}

			var flags byte
			if isDegenerate {
				// Read flags byte
				flagsOffset := z1Offset + fp.Bytes
				flags = data[flagsOffset]
			} else {
				// Generic case: reconstruct flags from high bits
				cubeRootIdx := highBits
				flags = (cubeRootIdx & 0x3) << 2
			}

			// Decompress
			p0, p1, err := BatchDecompress2G1(z0, z1, flags)
			if err != nil {
				atomic.AddUint64(&nbErrs, 1)
				continue
			}
			points[i*2] = p0
			points[i*2+1] = p1
		}
	})

	if nbErrs != 0 {
		return nil, errors.New("batch decompression failed")
	}

	// Handle odd point with standard decompression
	if hasOdd {
		oddOffset := offsets[nPairs]
		if _, err := points[n-1].SetBytes(data[oddOffset : oddOffset+SizeOfG1AffineCompressed]); err != nil {
			return nil, err
		}
	}

	return points, nil
}

// BatchCompressG2Slice compresses a slice of G2Affine points using 2-by-2 batch compression.
// This is more efficient than standard compression for decompression.
// If the slice has an odd length, the last point is compressed using standard compression.
// Compression is parallelized across pairs (2 points per thread).
//
// Returns the compressed bytes. The format is:
//   - For generic pairs: z0 (80 bytes, n encoded in bits 7-6) + z1 (80 bytes) = 160 bytes
//   - For degenerate pairs: z0 (80 bytes, bits 7-6 = 11) + z1 (80 bytes) + flags (1 byte) = 161 bytes
//   - If odd length: last point in standard compressed form (80 bytes)
func BatchCompressG2Slice(points []G2Affine) ([]byte, error) {
	n := len(points)
	if n == 0 {
		return nil, nil
	}

	nPairs := n / 2
	hasOdd := n%2 == 1

	// Pre-allocate for worst case (all degenerate pairs)
	maxSize := nPairs * SizeOfBatchCompressedG2PairDegenerate
	if hasOdd {
		maxSize += SizeOfG2AffineCompressed
	}

	result := make([]byte, maxSize)

	// Track actual sizes per pair
	pairSizes := make([]int, nPairs)
	var nbErrs uint64
	parallel.Execute(nPairs, func(start, end int) {
		for i := start; i < end; i++ {
			z0, z1, flags, err := BatchCompress2G2(&points[i*2], &points[i*2+1])
			if err != nil {
				atomic.AddUint64(&nbErrs, 1)
				continue
			}

			// Determine if generic (case 0) or degenerate
			caseIndicator := (flags >> 4) & 0x0F
			isGeneric := caseIndicator == 0

			if isGeneric {
				pairSizes[i] = SizeOfBatchCompressedG2Pair // 160 bytes
			} else {
				pairSizes[i] = SizeOfBatchCompressedG2PairDegenerate // 161 bytes
			}

			// Calculate offset based on worst-case positions
			offset := i * SizeOfBatchCompressedG2PairDegenerate

			// Write z0
			z0Bytes := z0.Bytes()
			if isGeneric {
				// Encode n in bits 7-6 of z0's first byte
				cubeRootIdx := (flags >> 2) & 0x3
				z0Bytes[0] |= (cubeRootIdx << 6)
			} else {
				// Set bits 7-6 to 0b11 as degenerate marker
				z0Bytes[0] |= 0xC0
			}
			copy(result[offset:offset+fp.Bytes], z0Bytes[:])
			offset += fp.Bytes

			// Write z1
			z1Bytes := z1.Bytes()
			copy(result[offset:offset+fp.Bytes], z1Bytes[:])
			offset += fp.Bytes

			// Write flags only for degenerate cases
			if !isGeneric {
				result[offset] = flags
			}
		}
	})

	if nbErrs != 0 {
		return nil, errors.New("batch compression failed")
	}

	// Compact the result
	actualSize := 0
	for i := 0; i < nPairs; i++ {
		actualSize += pairSizes[i]
	}
	if hasOdd {
		actualSize += SizeOfG2AffineCompressed
	}

	compacted := make([]byte, actualSize)
	writeOffset := 0
	for i := 0; i < nPairs; i++ {
		srcOffset := i * SizeOfBatchCompressedG2PairDegenerate
		copy(compacted[writeOffset:writeOffset+pairSizes[i]], result[srcOffset:srcOffset+pairSizes[i]])
		writeOffset += pairSizes[i]
	}

	// Handle odd point with standard compression
	if hasOdd {
		lastBytes := points[n-1].Bytes()
		copy(compacted[writeOffset:], lastBytes[:])
	}

	return compacted, nil
}

// BatchDecompressG2Slice decompresses a slice of G2Affine points from batch-compressed form.
// The input must have been created by BatchCompressG2Slice.
// Decompression is parallelized across pairs (2 points per thread, each thread does its own inversion).
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

	// Minimum size check (all generic pairs)
	minSize := nPairs * SizeOfBatchCompressedG2Pair
	if hasOdd {
		minSize += SizeOfG2AffineCompressed
	}
	if len(data) < minSize {
		return nil, errors.New("insufficient data for batch decompression")
	}

	points := make([]G2Affine, n)

	// First pass: scan data to find pair offsets
	offsets := make([]int, nPairs+1)
	offset := 0
	for i := 0; i < nPairs; i++ {
		offsets[i] = offset
		// Check bits 7-6 of z0's first byte
		highBits := (data[offset] >> 6) & 0x3
		if highBits == 0x3 {
			// Degenerate case: 161 bytes
			offset += SizeOfBatchCompressedG2PairDegenerate
		} else {
			// Generic case: 160 bytes
			offset += SizeOfBatchCompressedG2Pair
		}
	}
	offsets[nPairs] = offset

	// Verify we have enough data
	expectedSize := offset
	if hasOdd {
		expectedSize += SizeOfG2AffineCompressed
	}
	if len(data) < expectedSize {
		return nil, errors.New("insufficient data for batch decompression")
	}

	// Decompress pairs in parallel
	var nbErrs uint64
	parallel.Execute(nPairs, func(start, end int) {
		for i := start; i < end; i++ {
			pairOffset := offsets[i]

			// Read z0's first byte to determine format
			z0FirstByte := data[pairOffset]
			highBits := (z0FirstByte >> 6) & 0x3
			isDegenerate := highBits == 0x3

			// Make a copy of z0 bytes and clear the high bits
			var z0Bytes [fp.Bytes]byte
			copy(z0Bytes[:], data[pairOffset:pairOffset+fp.Bytes])
			z0Bytes[0] &= 0x3F // Clear bits 7-6

			var z0, z1 fp.Element
			if err := z0.SetBytesCanonical(z0Bytes[:]); err != nil {
				atomic.AddUint64(&nbErrs, 1)
				continue
			}

			z1Offset := pairOffset + fp.Bytes
			if err := z1.SetBytesCanonical(data[z1Offset : z1Offset+fp.Bytes]); err != nil {
				atomic.AddUint64(&nbErrs, 1)
				continue
			}

			var flags byte
			if isDegenerate {
				// Read flags byte
				flagsOffset := z1Offset + fp.Bytes
				flags = data[flagsOffset]
			} else {
				// Generic case: reconstruct flags from high bits
				cubeRootIdx := highBits
				flags = (cubeRootIdx & 0x3) << 2
			}

			// Decompress
			p0, p1, err := BatchDecompress2G2(z0, z1, flags)
			if err != nil {
				atomic.AddUint64(&nbErrs, 1)
				continue
			}
			points[i*2] = p0
			points[i*2+1] = p1
		}
	})

	if nbErrs != 0 {
		return nil, errors.New("batch decompression failed")
	}

	// Handle odd point with standard decompression
	if hasOdd {
		oddOffset := offsets[nPairs]
		if _, err := points[n-1].SetBytes(data[oddOffset : oddOffset+SizeOfG2AffineCompressed]); err != nil {
			return nil, err
		}
	}

	return points, nil
}
