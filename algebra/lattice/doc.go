// Package lattice provides lattice-based algorithms for cryptographic
// applications.
//
// This package implements LLL (Lenstra-Lenstra-Lovász) lattice reduction and
// rational reconstruction algorithms used in elliptic curve cryptography,
// particularly for (fake) GLV/GLS scalar multiplication optimizations in gnark
// circuits.
//
// # Rational Reconstruction
//
// The [RationalReconstruct] function finds small integers (x, z) such that:
//
//	k ≡ x/z (mod r)
//
// where k is a scalar and r is the modulus. The outputs satisfy bounds of
// approximately 1.16·r^(1/2). This uses a 3×2 lattice.
//
// # Multi Rational Reconstruction
//
// The [MultiRationalReconstruct] function finds small integers (x1, x2, z)
// such that:
//
//	k1 ≡ x1/z (mod r) k2 ≡ x2/z (mod r)
//
// with a shared denominator z. The outputs satisfy bounds of approximately
// 2·r^(2/3). This uses a 4×3 lattice.
//
// # Rational Reconstruction (Quadratic Extension)
//
// The [RationalReconstructExt] function finds small integers (x, y, z, t) such
// that:
//
//	k ≡ (x + λy)/(z + λt) (mod r)
//
// where λ is a quadratic extension generator (e.g., a primitive cube root of
// unity mod r).  The outputs satisfy bounds of approximately 1.25·r^(1/4).
// This uses a 7×4 lattice.
//
// # Multi Rational Reconstruction (Quadratic Extension)
//
// The [MultiRationalReconstructExt] function extends reconstruction to two
// scalars simultaneously, finding (x1, y1, x2, y2, z, t) such that:
//
//	k1 ≡ (x1 + λy1)/(z + λt) (mod r) k2 ≡ (x2 + λy2)/(z + λt) (mod r)
//
// with bounds of approximately 1.28·r^(1/3). This uses a 10×6 lattice.
//
// # LLL Algorithm
//
// All functions use the LLL lattice reduction algorithm with δ = 0.99 and
// exact rational arithmetic for numerical stability. The implementation
// handles non-square lattices (more generators than coordinates) which arise
// naturally in these cryptographic constructions.
//
// # References
//
//   - Eagen, El Housni, Masson, Piellard: "Fast elliptic curve scalar
//     multiplications in SN(T)ARK circuits" (LatinCrypt 2025)
//     https://eprint.iacr.org/2025/933.pdf
package lattice
