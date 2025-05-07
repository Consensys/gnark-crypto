// Copyright 2020-2023 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fr

// SqrtOptimized computes the square root of x using the optimized Legendre symbol implementation
// If x is not a quadratic residue, the function returns nil.
// This provides an optimization by avoiding unnecessary Legendre symbol
// computation in cases where we already know the result from the inversion algorithm.
func (z *Element) SqrtOptimized(x *Element) *Element {
	// Check if x is zero
	if x.IsZero() {
		return z.SetZero()
	}

	// Check if x is a quadratic residue using the optimized Legendre
	// algorithm from https://eprint.iacr.org/2023/1261
	if x.LegendreOptimized() != 1 {
		// If x is not a quadratic residue, return nil
		return nil
	}

	// If x is a quadratic residue, compute the square root
	// using the existing Sqrt algorithm
	// This function assumes x is already known to be a quadratic residue

	// For BLS12-381, q ≡ 1 (mod 4), so we use the Tonelli-Shanks algorithm
	// The implementation is identical to the existing Sqrt function, but
	// without the Legendre symbol check which we've already performed

	var y, b, t, w Element
	// w = x^((s-1)/2))
	w.expBySqrtExp(*x)

	// y = x^((s+1)/2)) = w * x
	y.Mul(x, &w)

	// b = xˢ = w * w * x = y * x
	b.Mul(&w, &y)

	// g = nonResidue ^ s
	var g = Element{
		15230403791020821917,
		9241180428717820143,
		14190631206722888944,
		1328741580841222485,
	}
	r := uint64(32)

	for {
		var m uint64
		t = b

		// for t != 1
		for !t.IsOne() {
			t.Square(&t)
			m++
		}

		if m == 0 {
			return z.Set(&y)
		}
		// t = g^(2^(r-m-1)) (mod q)
		ge := int(r - m - 1)
		t = g
		for ge > 0 {
			t.Square(&t)
			ge--
		}

		g.Square(&t)
		y.Mul(&y, &t)
		b.Mul(&b, &g)
		r = m
	}
}
