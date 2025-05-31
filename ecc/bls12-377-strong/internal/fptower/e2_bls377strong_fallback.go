//go:build !amd64
// +build !amd64

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import "github.com/consensys/gnark-crypto/ecc/bls12-377-strong/fp"

// MulByNonResidue multiplies a E2 by (2,1)
func (z *E2) MulByNonResidue(x *E2) *E2 {
	var a, b fp.Element
	a.Sub(&x.A0, &x.A1).Add(&a, &x.A0)
	b.Add(&x.A0, &x.A1).Add(&b, &x.A1)
	z.A0.Set(&a)
	z.A1.Set(&b)
	return z
}

// Mul sets z to the E2-product of x,y, returns z
func (z *E2) Mul(x, y *E2) *E2 {
	mulGenericE2(z, x, y)
	return z
}

// Square sets z to the E2-product of x,x returns z
func (z *E2) Square(x *E2) *E2 {
	squareGenericE2(z, x)
	return z
}
