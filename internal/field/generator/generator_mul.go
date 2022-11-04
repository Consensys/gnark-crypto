package generator

// There are couple of variations to the multiplication (and squaring) algorithms.
//
// All versions are derived from the Montgomery CIOS algorithm: see
// section 2.3.2 of Tolga Acar's thesis
// https://www.microsoft.com/en-us/research/wp-content/uploads/1998/06/97Acar.pdf
//
// For 1-word modulus, the generator will call mul_cios_one_limb (standard REDC)
//
// For 13-word+ modulus, the generator will output a unoptimized textbook CIOS code, in plain Go.
//
// For all other modulus, we look at the available bits in the last limb.
// If they are none (like secp256k1) we generate a unoptimized textbook CIOS code, in plain Go, for all architectures.
// If there is at least one we can ommit a carry propagation in the CIOS algorithm.
// If there is at least two we can use the same technique for the CIOS Squaring.
// See appendix in https://eprint.iacr.org/2022/1400.pdf for the exact condition.
//
// In practice, we have 3 differents targets in mind: x86(amd64), arm64 and wasm.
//
// For amd64, we can leverage (when available) the BMI2 and ADX instructions to have 2-carry-chains in parallel.
// This make the use of assembly worth it as it results in a significant perf improvment; most CPUs since 2016 support these
// instructions, and we assume it to be the "default path"; in case the CPU has no support, we fall back to a slow, unoptimized version.
//
// On amd64, the Squaring algorithm always call the Multiplication (assembly) implementation.
//
// For arm64, we unroll the loops in the CIOS (+nocarry optimization) algorithm, such that the instructions generated
// by the Go compiler closely match what we would hand-write. Hence, there is no assembly needed for arm64 target.
//
// Additionally, if 2-bits+ are available on the last limb, we have a template to generate a dedicated Squaring algorithm
// This is not activated by default, to minimize the codebase size.
// On M1, AWS Graviton3 it results in a 5-10% speedup. On some mobile devices, speed up observed was more important (~20%).
//
// The same (arm64) unrolled Go code produce satisfying perfomrance for WASM (compiled using TinyGo).
func generateMulAndSquare() {

}
