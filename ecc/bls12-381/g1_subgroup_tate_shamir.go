package bls12381

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

// exp1 = (p-1)/e2, the exponent used in f1IsOne
// Value: 0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd809536aad8a973fff0aaaaaa5555aaaa
var exp1Big *big.Int

// exp2 = |2 - z + z² + z³ - z⁴ - z⁵| where z = -x₀ (seed)
// This is the effective exponent used in f2IsOne/exp2Compute
var exp2Big *big.Int

// Precomputed bit arrays for fast access (stored as uint64 words, LSB first)
var exp1Words []uint64
var exp2Words []uint64

// exp1Len and exp2Len store the bit lengths
var exp1Len, exp2Len int

// maxExpLen is the maximum bit length between exp1 and exp2
var maxExpLen int

func init() {
	// exp1 = (p-1)/e2
	exp1Big = new(big.Int)
	exp1Big.SetString("1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd809536aad8a973fff0aaaaaa5555aaaa", 16)

	// Compute exp2 = 2 - z + z² + z³ - z⁴ - z⁵ where z = -x₀
	// x₀ = -15132376222941642752, so z = 15132376222941642752
	z := new(big.Int)
	z.SetString("15132376222941642752", 10) // |x₀|

	z2 := new(big.Int).Mul(z, z)
	z3 := new(big.Int).Mul(z2, z)
	z4 := new(big.Int).Mul(z3, z)
	z5 := new(big.Int).Mul(z4, z)

	// exp2 = 2 - z + z² + z³ - z⁴ - z⁵
	// But we need to be careful: the actual exponent in f2IsOne is computed as:
	// u0 = x^(2 + z² + z³), u1 = x^(z + z⁴ + z⁵)
	// Then checks u0 == u1, which means x^(2 + z² + z³ - z - z⁴ - z⁵) == 1
	//
	// With z = |x₀| (positive), the BLS12-381 seed x₀ is NEGATIVE.
	// So expBySeed computes x^(-x₀) = x^z (since x₀ < 0, -x₀ = z > 0)
	// Actually looking at expBySeed more carefully, it computes x^|z|
	//
	// The sequential exponentiations in f2IsOne:
	// u1 = x^z, u2 = u1^z = x^(z²), u3 = u2^z = x^(z³)
	// etc.
	//
	// So exp2 = 2 + z² + z³ - z - z⁴ - z⁵
	// where z = |x₀| = 15132376222941642752

	exp2Big = new(big.Int)
	exp2Big.SetInt64(2)
	exp2Big.Sub(exp2Big, z)  // 2 - z
	exp2Big.Add(exp2Big, z2) // 2 - z + z²
	exp2Big.Add(exp2Big, z3) // 2 - z + z² + z³
	exp2Big.Sub(exp2Big, z4) // 2 - z + z² + z³ - z⁴
	exp2Big.Sub(exp2Big, z5) // 2 - z + z² + z³ - z⁴ - z⁵

	// Take absolute value (the exponent should be positive for exponentiation)
	if exp2Big.Sign() < 0 {
		exp2Big.Neg(exp2Big)
	}

	exp1Len = exp1Big.BitLen()
	exp2Len = exp2Big.BitLen()

	maxExpLen = exp1Len
	if exp2Len > maxExpLen {
		maxExpLen = exp2Len
	}

	// Convert to uint64 word arrays for fast bit access
	// big.Int.Bits() returns little-endian words
	exp1Words = bigIntToWords(exp1Big)
	exp2Words = bigIntToWords(exp2Big)
}

// bigIntToWords converts a big.Int to a slice of uint64 words (little-endian)
func bigIntToWords(n *big.Int) []uint64 {
	bits := n.Bits()
	words := make([]uint64, len(bits))
	for i, w := range bits {
		words[i] = uint64(w)
	}
	return words
}

// jointExpIsOne computes f1^exp1 * f2^exp2 and checks if the result is 1.
// This uses Shamir's trick to share squarings between the two exponentiations.
//
// Cost: max(bitlen(exp1), bitlen(exp2)) squarings + (HW(exp1) + HW(exp2) - joint_bits) multiplications
// where HW is the Hamming weight and joint_bits counts positions where both bits are 1.
func jointExpIsOne(f1, f2 *fp.Element) bool {
	// Precompute f1*f2 for when both bits are 1
	var f12 fp.Element
	f12.Mul(f1, f2)

	var acc fp.Element
	acc.SetOne()

	// Process from MSB to LSB using precomputed word arrays
	for i := maxExpLen - 1; i >= 0; i-- {
		// Square
		acc.Square(&acc)

		// Get bits at position i using fast word access
		bit1 := getBitFromWords(exp1Words, i)
		bit2 := getBitFromWords(exp2Words, i)

		// Multiply based on bits
		if bit1 && bit2 {
			acc.Mul(&acc, &f12)
		} else if bit1 {
			acc.Mul(&acc, f1)
		} else if bit2 {
			acc.Mul(&acc, f2)
		}
	}

	return acc.IsOne()
}

// getBitFromWords returns true if the bit at position i (0-indexed from LSB) is 1
// words is a little-endian slice of uint64
func getBitFromWords(words []uint64, i int) bool {
	wordIdx := i / 64
	if wordIdx >= len(words) {
		return false
	}
	bitIdx := uint(i % 64)
	return (words[wordIdx] & (1 << bitIdx)) != 0
}

// membershipTestShamir performs the Tate-based membership test using Shamir's trick
// for the joint f1^exp1 * f2^exp2 computation.
func membershipTestShamir(p, q *G1Affine, tab []fp.Element) bool {
	// Step 1: Compute φ̂(p) where φ̂(x,y) = (ω²x, y)
	var p2 G1Affine
	p2.X.Mul(&p.X, &thirdRootOneG1)
	p2.Y.Set(&p.Y)

	var n1, d1, n2, d2 fp.Element
	n1.Sub(&p.X, &q.X)
	n2.Sub(&p2.X, &q.X)
	d1.SetOne()
	d2.SetOne()

	// Chain-based shared Miller loop
	sharedMillerloopChain(tab, &n1, &d1, &n2, &d2, q, p, &p2)

	if n1.IsZero() || d1.IsZero() || n2.IsZero() || d2.IsZero() {
		return false
	}

	// Simultaneous inversion
	n1.Mul(&n1, &d2)
	n2.Mul(&n2, &d1)
	d1.Mul(&d1, &d2)
	d1.Inverse(&d1)
	n1.Mul(&n1, &d1)
	n2.Mul(&n2, &d1)

	// Use Shamir's trick for joint exponentiation
	return jointExpIsOne(&n1, &n2)
}

// IsInSubGroupTateShamir checks whether p is in the correct subgroup using
// the Tate-based test with Shamir's trick for joint exponentiation.
//
// This version computes f1^exp1 * f2^exp2 in a single pass, sharing all
// squarings between the two exponentiations.
func (p *G1Affine) IsInSubGroupTateShamir(tab chainTable) bool {
	if p.IsInfinity() {
		return false
	}
	if !p.IsOnCurve() {
		return false
	}
	return membershipTestShamir(p, &tab.q, tab.tab)
}
