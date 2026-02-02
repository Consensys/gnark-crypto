// Package lattice provides lattice-based algorithms for cryptographic
// applications.
//
// This package implements rational reconstruction algorithms used in elliptic
// curve cryptography, particularly for (fake) GLV/GLS scalar multiplication
// optimizations in gnark circuits.
//
// # Algorithm Selection
//
// The 2D rational reconstruction problem (finding small x, z with k ≡ x/z mod r)
// can be solved by three equivalent approaches, all achieving bounds |x|,|z| < √r:
//
//  1. Half-GCD: Extended Euclidean algorithm stopped when remainder < √r.
//     Complexity O(n²) with ~n/2 iterations and one division per iteration.
//
//  2. LLL: Build lattice {(r,0), (k,1)} and reduce. Gives proven bounds via
//     Hermite constant γ₂ = 2/√3 ≈ 1.15, but has significant overhead from
//     Gram-Schmidt orthogonalization.
//
//  3. Pornin's Algorithm: Division-free variant of Lagrange's 2D reduction
//     using only shifts, additions, and subtractions. See eprint 2020/454.
//     Complexity O(n²) with ~n iterations (2.5x more than half-GCD).
//
// This package uses half-GCD for [RationalReconstruct] because in Go:
//
//   - big.Int.Div uses optimized assembly (Karatsuba/Newton division)
//   - Half-GCD needs ~40 iterations vs ~98 for Pornin's algorithm
//   - Measured: half-GCD ~5μs vs Pornin ~6μs for 254-bit primes on Macbook Air M1.
//
// Pornin's algorithm is faster on embedded systems (ARM Cortex-M0/M4) where
// division is 10-100x slower than addition and fixed-width arithmetic avoids
// big.Int overhead.
//
// # Rational Reconstruction
//
// The [RationalReconstruct] function finds small integers (x, z) such that:
//
//	k ≡ x/z (mod r)
//
// where k is a scalar and r is the modulus. The outputs satisfy |x|,|z| < √r.
//
// # Multi Rational Reconstruction
//
// The [MultiRationalReconstruct] function finds small integers (x1, x2, z)
// such that:
//
//	k1 ≡ x1/z (mod r)
//	k2 ≡ x2/z (mod r)
//
// with a shared denominator z. The outputs satisfy bounds of approximately
// 1.22·r^(2/3). This uses LLL on a 4×3 lattice.
//
// # Rational Reconstruction (Quadratic Extension)
//
// The [RationalReconstructExt] function finds small integers (x, y, z, t) such
// that:
//
//	k ≡ (x + λy)/(z + λt) (mod r)
//
// where λ is a quadratic extension generator (e.g., a primitive cube root of
// unity mod r). The outputs satisfy bounds of approximately 1.25·r^(1/4).
// This uses LLL on a 7×4 lattice.
//
// Note: The Eisenstein half-GCD (in Z[ω]) cannot cannot achieve the same
// LLL bounds as explained in ePrint 2025/933.
//
// # Multi Rational Reconstruction (Quadratic Extension)
//
// The [MultiRationalReconstructExt] function extends reconstruction to two
// scalars simultaneously, finding (x1, y1, x2, y2, z, t) such that:
//
//	k1 ≡ (x1 + λy1)/(z + λt) (mod r)
//	k2 ≡ (x2 + λy2)/(z + λt) (mod r)
//
// with bounds of approximately 1.28·r^(1/3). This uses LLL on a 10×6 lattice.
//
// # LLL Implementation
//
// The LLL functions use δ = 0.99 and exact rational arithmetic (lazy evaluation
// to minimize GCD computations) for numerical stability. The implementation
// handles non-square lattices (more generators than coordinates) which arise
// naturally in these cryptographic constructions.
//
// # References
//
//   - Lagrange (1773): Original 2D lattice basis reduction
//   - Lenstra, Lenstra, Lovász (1982): LLL algorithm for higher dimensions
//   - Pornin (2020): Division-free 2D reduction, https://eprint.iacr.org/2020/454
//   - Eagen, El Housni, Masson, Piellard (2025): "Fast elliptic curve scalar
//     multiplications in SN(T)ARK circuits", https://eprint.iacr.org/2025/933
package lattice
