# IFMA Vector Multiplication Optimization Plan

## Executive Summary

The current IFMA implementation is slower than the radix-32 AVX-512 implementation due to several design issues. The most critical problem is the **multiply-by-16 correction with 5 binary conditional subtractions**, which adds ~200+ instructions per iteration. This document analyzes the issues and proposes optimizations.

## Current Implementation Analysis

### Instruction Count Comparison (per 8 elements)

| Phase | Current IFMA | Existing radix-32 |
|-------|--------------|-------------------|
| Load data | 4 VMOVDQU64 | 4 VMOVDQU64 |
| Transpose AoS→SoA | ~20 ops | ~20 ops |
| Radix-64→52 conversion | ~20 ops | N/A |
| Montgomery multiply | ~150 ops (5 rounds × 30) | ~200 ops (interleaved) |
| Final normalization | ~15 ops | N/A |
| Conditional subtract (1x q) | ~35 ops | ~12 ops (REDUCE macro) |
| **Multiply by 16** | ~20 ops | N/A |
| **Binary reduction (5x)** | **~175 ops** | N/A |
| Radix-52→64 conversion | ~15 ops | N/A |
| Transpose SoA→AoS | ~20 ops | ~20 ops |
| Store | 4 VMOVDQU64 | 4 VMOVDQU64 |
| **Total per 8 elements** | **~480 ops** | **~270 ops** |

### Root Cause: Radix Mismatch

The fundamental problem is:

1. **Input format**: Elements in radix-64 (4 × 64-bit limbs), Montgomery form with R = 2^256
2. **IFMA radix-52**: 5 limbs × 52 bits = 260-bit representation
3. **Montgomery reduction R**: After 5 IFMA rounds, effective R = 2^(5×52) = 2^260
4. **Mismatch**: Result is A×B×2^{-260}, but we need A×B×2^{-256}
5. **Current fix**: Multiply by 16 = 2^4, then do 5 conditional subtractions

This "fix" is **O(n) conditional subtractions** which is terrible for performance.

## Proposed Optimizations

### Option 1: Almost Montgomery Multiplication (AMM) - RECOMMENDED

**Key insight from OpenSSL/AWS-LC**: Skip the final conditional subtraction entirely.

**Algorithm**:
- Montgomery multiplication naturally produces results in [0, 2q)
- These results are valid inputs for the next Montgomery multiplication
- Only reduce to [0, q) when converting out of Montgomery form

**For the radix correction**:
Instead of: `result = mont_mul(a, b) * 16 mod q`

Use: `result = mont_mul(a, mont_mul(16*R mod q, b))`

This folds the correction into a single extra multiplication per vector, amortized over the entire batch.

**Or better**: Precompute `16*R mod q` and multiply one of the inputs by it before the batch operation:
```
// One-time correction factor (can be precomputed)
correction = 16 * R mod q = 2^4 * 2^256 mod q = 2^260 mod q

// In the loop, just do standard AMM:
for i := 0; i < n; i += 8 {
    result[i:i+8] = AMM(a[i:i+8], b_corrected[i:i+8])  // No conditional subtract!
}
// Final reduction only when leaving Montgomery form
```

**Expected instruction count**: ~300 ops per 8 elements (vs 480 current, 270 radix-32)

### Option 2: Lazy Reduction with Carry Headroom

**Key insight**: IFMA accumulates into 64-bit lanes, leaving 12 bits of headroom per limb.

**Algorithm**:
1. Accumulate multiple operations before normalizing carries
2. Use the 12-bit headroom to avoid overflow
3. Only normalize when headroom is exhausted

**Implementation**:
```go
// Instead of normalizing after each round, defer:
// Round 0: T += A*B[0], T += m*q  (no normalize)
// Round 1: T += A*B[1], T += m*q  (no normalize)
// ...
// Final: Normalize all limbs at once using parallel carry propagation
```

**Parallel carry propagation** (from curve25519-dalek):
```
// Compute all carryouts simultaneously
carry0 = l0 >> 52; l0 &= mask52
carry1 = l1 >> 52; l1 &= mask52
...
// Add carryins simultaneously
l1 += carry0
l2 += carry1
...
```

This reduces sequential dependencies and improves instruction-level parallelism.

### Option 3: Eliminate Radix Conversion Overhead

**Key insight**: The radix-64 ↔ radix-52 conversions are expensive but mandatory with current design.

**Alternative**: Keep data in radix-52 format throughout the computation:
- Store vectors in radix-52 format (5 × 64-bit per element instead of 4 × 64-bit)
- 25% more memory but no conversion overhead
- Only convert at vector boundaries (load from scalar, store to scalar)

**Trade-off**: Memory bandwidth vs. conversion overhead. For large vectors, conversion is amortized, so this may not help significantly.

### Option 4: Hybrid Approach

Use IFMA for the inner Montgomery reduction but keep data in radix-64:

**Algorithm**:
1. Keep A, B in radix-64 format
2. For each multiplication, convert just the necessary parts to radix-52 on-the-fly
3. Use IFMA only for the m*q accumulation (which is the hot path)
4. Combine with MULX/ADX for parts that don't benefit from IFMA

This is complex to implement but may offer the best of both worlds.

## Recommended Implementation Order

### Phase 1: Remove Binary Reduction (Highest Impact)

**Change**: Switch to AMM - remove all 5 conditional subtractions after multiply-by-16.

**Steps**:
1. Compute correction factor `2^260 mod q` as a compile-time constant
2. Modify the API: either pre-multiply one input, or expose AMM results in [0, 2q)
3. Remove `multiplyByConstant16Radix52()` and `conditionalSubtractNQ()` calls
4. Only call `conditionalSubtractQ()` once at the very end (or not at all for AMM)

**Expected improvement**: ~40% faster (remove ~175 instructions)

### Phase 2: Improve Carry Propagation

**Change**: Use parallel carry propagation instead of sequential.

**Steps**:
1. Defer normalization within Montgomery rounds where possible
2. Compute all carries in parallel after each major phase
3. Use `VPSRLQ` + `VPANDQ` + `VPADDQ` pattern simultaneously on all limbs

**Expected improvement**: ~10% faster (reduce critical path)

### Phase 3: Optimize Transpose Operations

**Change**: Investigate if transpose can be reduced or fused with other operations.

**Options**:
- Use gather/scatter instructions (VPGATHERQQ/VPSCATTERQQ) instead of transpose
- Fuse first/last transpose with radix conversion
- Consider processing in AoS format with masking

### Phase 4: Benchmark and Profile

After each change:
1. Run `BenchmarkVectorOps/mul_2097152`
2. Profile with `perf stat` to measure actual instruction counts and IPC
3. Check for memory bandwidth bottlenecks

## Critical Code Changes

### Remove Binary Reduction

In `element_vec_4words_ifma.go`, replace:

```go
func (f *FFAmd64) multiplyByConstant16Radix52() {
    // ... ~20 instructions for x16 ...

    // Binary reduction: conditionally subtract 16q, 8q, 4q, 2q, q
    // ~175 instructions - REMOVE ALL OF THIS
    f.conditionalSubtractNQ("Z5", "Z6", "Z7", "Z8", "Z9") // x4
    // ...
    f.conditionalSubtractQ()  // Keep only this one (or remove for AMM)
}
```

With:

```go
func (f *FFAmd64) multiplyByConstant16Radix52() {
    // Multiply by 16 = 2^4 (left shift with carry) in radix-52
    // ... ~20 instructions ...

    // Single conditional subtraction (result is now in [0, 17q))
    // If using AMM: remove even this subtraction
    f.conditionalSubtractQ()
}
```

Wait - this still has the problem. A better approach:

### Alternative: Eliminate x16 Correction Entirely

**Key insight**: The x16 correction is only needed because R=2^260 ≠ 2^256.

**Solution**: Define the Montgomery radix as R=2^260 for IFMA operations.

1. Pre-convert inputs: `a_260 = a * 2^4 mod q` (done once per vector)
2. IFMA Montgomery multiplication produces correct results in R=2^260 form
3. Post-convert outputs: `a_256 = a_260 * 2^{-4} mod q` (done once per vector)

But 2^{-4} mod q requires another multiplication... unless we precompute it.

**Better**: Accept that IFMA multiplication gives A*B*2^{-260}. If both inputs are in standard R=2^256 form:
- Result = A*R * B*R * R^{-1} (where R=2^260 for IFMA)
- = A*B * 2^{256} * 2^{256} * 2^{-260}
- = A*B * 2^{252}

This is 4 bits short. We need: A*B * 2^{256}

**Trick**: Instead of correcting after, pre-multiply ONE of the inputs by 2^4 = 16:
- Input: A in R=2^256 form, B' = B*16 in R=2^256 form (B' = B*16*2^{256})
- IFMA result = A*B' * 2^{-260} = A * B*16*2^{256} * 2^{-260} = A*B * 2^{-4} * 2^{256} * 2^{-260}

Hmm, this still doesn't work cleanly. Let me think more carefully...

Actually, the correct analysis:
- Input A is in Montgomery form: A_mont = a * 2^256 mod q
- Input B is in Montgomery form: B_mont = b * 2^256 mod q
- Standard Montgomery multiply: A_mont * B_mont * 2^{-256} = a * b * 2^256 mod q = (a*b)_mont

With IFMA R=2^260:
- IFMA multiply: A_mont * B_mont * 2^{-260} = a * b * 2^{256} * 2^{256} * 2^{-260} = a * b * 2^{252}
- This is NOT in Montgomery form with R=2^256

To fix: multiply result by 2^4 = 16 to get a * b * 2^{256} = (a*b)_mont

So the x16 correction IS necessary. The question is how to do it cheaply.

**Best approach**: Fold the x16 into the final conditional subtraction using a special constant.

Since result after IFMA is in [0, 2q) before x16, after x16 it's in [0, 32q).

Instead of 5 conditional subtractions, use:
1. Compare with 16q, conditionally subtract 16q (result in [0, 16q))
2. Compare with 8q, conditionally subtract 8q (result in [0, 8q))
3. Compare with 4q, conditionally subtract 4q (result in [0, 4q))
4. Compare with 2q, conditionally subtract 2q (result in [0, 2q))
5. Compare with q, conditionally subtract q (result in [0, q))

This is exactly what's being done, and it's expensive.

**Alternative - Range Reduction using Barrett**:

For reducing from [0, 32q) to [0, q), use Barrett reduction:
- Compute quotient estimate: k = floor(x / q) using precomputed reciprocal
- Subtract k*q from x

This requires ~15 instructions instead of ~175 for 5 conditional subtractions.

**Concrete implementation**:
```
// Barrett reduction for x in [0, 32q)
// k = floor(x * mu >> shift) where mu ≈ 2^shift / q
// For 5-limb radix-52, shift = 260 + some headroom

// Since x < 32q < 2^262, we can use a 8-bit quotient (since 32 = 2^5)
// Actually k is at most 31

VPMULHQ  mu, x[4], k  // Approximate k from high limb
VPMULUDQ k, q[0..4], sub
VPSUBQ   sub, x, result
// One more conditional subtraction to handle rounding error
```

This is much cheaper than 5 sequential conditional subtractions!

## Summary of Recommendations

1. **Immediate fix (Phase 1)**: Replace 5 conditional subtractions with Barrett reduction (~3x speedup on reduction phase)

2. **Medium-term (Phase 2)**: Use AMM throughout - accept results in [0, 2q), only reduce at API boundaries

3. **Long-term (Phase 3)**: Consider keeping vectors in radix-52 format to amortize conversion cost

4. **Parallel carries (Phase 4)**: Optimize carry propagation within Montgomery rounds

## References

1. [OpenSSL rsaz-avx512.pl](https://github.com/openssl/openssl/blob/master/crypto/bn/asm/rsaz-avx512.pl) - AMM implementation
2. [Intel HEXL](https://github.com/IntelLabs/hexl) - IFMA for homomorphic encryption
3. [curve25519-dalek-ng IFMA](https://github.com/zkcrypto/curve25519-dalek-ng/blob/main/docs/ifma-notes.md) - Parallel carry propagation
4. [Fast Modular Multiplication using 512-bit AVX](https://link.springer.com/article/10.1007/s13389-021-00256-9) - BPS method
