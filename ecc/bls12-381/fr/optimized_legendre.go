// Copyright 2020-2023 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fr

import (
	"github.com/consensys/gnark-crypto/field"
)

// elementAdapter is an adapter that implements field.ElementInterface for fr.Element
type elementAdapter struct {
	Element
}

// Set sets z to x and returns z
func (z *elementAdapter) Set(x field.ElementInterface) field.ElementInterface {
	// Type assertion to extract the underlying Element
	xAdapter, ok := x.(*elementAdapter)
	if !ok {
		panic("set: not an elementAdapter")
	}
	z.Element.Set(&xAdapter.Element)
	return z
}

// IsZero returns true if z equals 0
func (z *elementAdapter) IsZero() bool {
	return z.Element.IsZero()
}

// Neg sets z to -x (mod p) and returns z
func (z *elementAdapter) Neg(x field.ElementInterface) field.ElementInterface {
	// Type assertion to extract the underlying Element
	xAdapter, ok := x.(*elementAdapter)
	if !ok {
		panic("neg: not an elementAdapter")
	}
	z.Element.Neg(&xAdapter.Element)
	return z
}

// Equal returns true if z equals x
func (z *elementAdapter) Equal(x field.ElementInterface) bool {
	// Type assertion to extract the underlying Element
	xAdapter, ok := x.(*elementAdapter)
	if !ok {
		panic("equal: not an elementAdapter")
	}
	return z.Element.Equal(&xAdapter.Element)
}

// inverseAdapter adapts fr.Element.Inverse to the signature expected by LegendrePornin
func inverseAdapter(z, x field.ElementInterface) field.ElementInterface {
	// Type assertions to extract the underlying Element
	zAdapter, ok := z.(*elementAdapter)
	if !ok {
		panic("inverseAdapter: z not an elementAdapter")
	}

	xAdapter, ok := x.(*elementAdapter)
	if !ok {
		panic("inverseAdapter: x not an elementAdapter")
	}

	zAdapter.Element.Inverse(&xAdapter.Element)
	return zAdapter
}

// negOneAdapter returns -1 in the field
func negOneAdapter() field.ElementInterface {
	var e Element
	e.SetOne()
	e.Neg(&e)
	return &elementAdapter{e}
}

// oneAdapter returns 1 in the field
func oneAdapter() field.ElementInterface {
	var e Element
	e.SetOne()
	return &elementAdapter{e}
}

// LegendreOptimized computes the Legendre symbol using the optimized algorithm
// based on the paper https://eprint.iacr.org/2023/1261
//
// Returns:
//
//	 1 if x is a quadratic residue modulo p
//	-1 if x is a quadratic non-residue modulo p
//	 0 if x is congruent to 0 modulo p
func (z *Element) LegendreOptimized() int {
	adapter := &elementAdapter{*z}
	return field.LegendrePornin(adapter, inverseAdapter, negOneAdapter, oneAdapter)
}
