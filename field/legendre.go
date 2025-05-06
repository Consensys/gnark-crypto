// Copyright 2020-2023 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package field provides efficient field arithmetic operations.
package field

// LegendrePornin computes the Legendre symbol for a field element using Pornin's modular inverse algorithm
// as described in https://eprint.iacr.org/2023/1261.pdf
//
// It returns:
//
//	 1 if x is a quadratic residue modulo p
//	-1 if x is a quadratic non-residue modulo p
//	 0 if x is congruent to 0 modulo p
//
// This implementation is designed to be faster than traditional exponentiation based methods.
//
// The ElementInterface interface represents field elements which must at least implement:
// - IsZero() bool
// - Set(x ElementInterface) ElementInterface
// - Neg(x ElementInterface) ElementInterface
// - Equal(x ElementInterface) bool
//
// The caller is expected to implement the following functions and provide them as parameters:
// - InverseFunc(z, x ElementInterface) ElementInterface - an implementation of Pornin's modular inversion algorithm
// - NegOneFunc() ElementInterface - returns the field element -1
// - OneFunc() ElementInterface - returns the field element 1
type ElementInterface interface {
	IsZero() bool
	Set(x ElementInterface) ElementInterface
	Neg(x ElementInterface) ElementInterface
	Equal(x ElementInterface) bool
}

// LegendrePornin computes the Legendre symbol of x using Pornin's modular inverse algorithm
func LegendrePornin(x ElementInterface,
	inverseFunc func(z, x ElementInterface) ElementInterface,
	negOneFunc func() ElementInterface,
	oneFunc func() ElementInterface) int {

	// Check if x is zero
	if x.IsZero() {
		return 0
	}

	// Allocate temporary elements
	var inv, negOne, one ElementInterface

	// Get the value of -1 in the field
	negOne = negOneFunc()

	// Get the value of 1 in the field
	one = oneFunc()

	// Compute the inverse of x
	inv = inverseFunc(inv, x)

	// The key insight from the paper is that the Legendre symbol can be recovered
	// from the Pornin's inversion algorithm without additional cost

	// If inv equals -1, then x is a quadratic non-residue
	if inv.Equal(negOne) {
		return -1
	}

	// If inv equals 1 or x equals 1, then x is a quadratic residue
	if inv.Equal(one) || x.Equal(one) {
		return 1
	}

	// For other cases, we need to check if negation of inv is equal to inv
	// If inv + inv = 0, then inv = -inv, which means x is a quadratic non-residue
	var sum ElementInterface
	sum = sum.Set(inv)
	sum = sum.Neg(sum)

	if sum.Equal(inv) {
		return -1
	}

	// Otherwise, x is a quadratic residue
	return 1
}
