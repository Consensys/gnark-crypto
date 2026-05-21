// Package multisethash implements y-increment elliptic-curve multiset
// hashing (ECMSH) over octobear.
//
// The package exposes three variants:
//
//   - Classical one-point ECMSH (Accumulator, Hash, Map): 16-bit messages.
//     Each message m is mapped by scanning k in [0, 256) and setting
//     y = m*256 + k in the base subfield of Fp^8. The first resulting point
//     (x, y) on octobear is used as the image. Security is ~124 classical bits
//     (no post-quantum security).
//
//   - Linear-separator vector ECMSH (LinearAccumulator, HashLinear,
//     MapLinear): a digest of N = 23 curve points. Coordinate i uses
//     y_i(m, k) = T*(m + i*M) + k with T = 128, M = 2^18.
//
//   - Poseidon2-sponge vector ECMSH (Poseidon2Accumulator, HashPoseidon2,
//     MapPoseidon2): a digest of N = 23 curve points. The N ordinates are
//     derived by absorbing (domain tag, msg) into a width-16 Poseidon2
//     sponge (rate 8, 3 squeeze permutations) and range-reducing each
//     output into [0, floor(p/(2T))) with T = 256.
//
// The two vector variants are post-quantum candidates: under Shor's
// algorithm a collision becomes a bounded modular linear relation, i.e. a
// SIS-shaped problem with modulus r ~ 2^248 and dimension N = 23
// (5704 SIS-volume bits, matching the Linea KoalaBear LtHash baseline).
package multisethash
