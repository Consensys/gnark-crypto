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

// Joint Sparse Form (JSF) representation of (exp1, exp2)
// Each entry is a pair (d1, d2) where d1, d2 ∈ {-1, 0, 1}
// Stored as two slices: jsf1[i] and jsf2[i] give the i-th digits
var jsf1 []int8
var jsf2 []int8
var jsfLen int

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

	// exp2 = 2 + z² + z³ - z - z⁴ - z⁵
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

	// Compute Joint Sparse Form
	jsf1, jsf2 = computeJSF(exp1Big, exp2Big)
	jsfLen = len(jsf1)
}

// computeJSF computes the Joint Sparse Form of two integers.
// JSF minimizes the total number of non-zero digits across both representations.
// Reference: "Improved Algorithms for Arithmetic on Anomalous Binary Curves" by Solinas
func computeJSF(k1, k2 *big.Int) ([]int8, []int8) {
	// Work with copies
	d1 := new(big.Int).Set(k1)
	d2 := new(big.Int).Set(k2)

	// Estimate max length (bit length + 1 for potential carry)
	maxLen := d1.BitLen()
	if d2.BitLen() > maxLen {
		maxLen = d2.BitLen()
	}
	maxLen += 1

	jsf1 := make([]int8, 0, maxLen)
	jsf2 := make([]int8, 0, maxLen)

	zero := big.NewInt(0)
	one := big.NewInt(1)
	two := big.NewInt(2)
	four := big.NewInt(4)
	eight := big.NewInt(8)

	for d1.Cmp(zero) > 0 || d2.Cmp(zero) > 0 {
		// Get low 3 bits of each
		m1 := int(new(big.Int).And(d1, big.NewInt(7)).Int64())
		m2 := int(new(big.Int).And(d2, big.NewInt(7)).Int64())

		var u1, u2 int8

		// Determine u1
		if m1&1 == 1 { // d1 is odd
			u1 = 2 - int8(m1&3)
			if (m1 == 3 || m1 == 5) && (m2&3) == 2 {
				u1 = -u1
			}
		}

		// Determine u2
		if m2&1 == 1 { // d2 is odd
			u2 = 2 - int8(m2&3)
			if (m2 == 3 || m2 == 5) && (m1&3) == 2 {
				u2 = -u2
			}
		}

		jsf1 = append(jsf1, u1)
		jsf2 = append(jsf2, u2)

		// d1 = (d1 - u1) / 2
		if u1 == 1 {
			d1.Sub(d1, one)
		} else if u1 == -1 {
			d1.Add(d1, one)
		}
		d1.Rsh(d1, 1)

		// d2 = (d2 - u2) / 2
		if u2 == 1 {
			d2.Sub(d2, one)
		} else if u2 == -1 {
			d2.Add(d2, one)
		}
		d2.Rsh(d2, 1)

		// Suppress unused variable warnings
		_ = two
		_ = four
		_ = eight
	}

	return jsf1, jsf2
}

// jointExpIsOne computes f1^exp1 * f2^exp2 and checks if the result is 1.
// This uses Joint Sparse Form (JSF) to minimize the number of multiplications.
//
// JSF guarantees that the joint Hamming weight is at most n/2 + 1 on average,
// compared to ~n for standard binary representation.
//
// To avoid expensive inversions, we track numerator and denominator separately:
// result = num/den, and we check if num == den at the end.
func jointExpIsOne(f1, f2 *fp.Element) bool {
	// Precompute products for positive digits
	var f12 fp.Element
	f12.Mul(f1, f2)

	// Track numerator and denominator separately
	// result = num / den
	// Positive JSF digits multiply num, negative digits multiply den
	var num, den fp.Element
	num.SetOne()
	den.SetOne()

	// Process from MSB to LSB
	for i := jsfLen - 1; i >= 0; i-- {
		// Square both
		num.Square(&num)
		den.Square(&den)

		d1 := jsf1[i]
		d2 := jsf2[i]

		// Handle each JSF digit pair
		// Positive digits multiply numerator, negative multiply denominator
		switch {
		case d1 == 1 && d2 == 1:
			num.Mul(&num, &f12)
		case d1 == 1 && d2 == 0:
			num.Mul(&num, f1)
		case d1 == 1 && d2 == -1:
			num.Mul(&num, f1)
			den.Mul(&den, f2)
		case d1 == 0 && d2 == 1:
			num.Mul(&num, f2)
		case d1 == 0 && d2 == -1:
			den.Mul(&den, f2)
		case d1 == -1 && d2 == 1:
			den.Mul(&den, f1)
			num.Mul(&num, f2)
		case d1 == -1 && d2 == 0:
			den.Mul(&den, f1)
		case d1 == -1 && d2 == -1:
			den.Mul(&den, &f12)
			// case d1 == 0 && d2 == 0: do nothing
		}
	}

	// Check if num == den (i.e., num/den == 1)
	return num.Equal(&den)
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
