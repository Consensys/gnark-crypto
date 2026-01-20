# Tate-Based G1 Subgroup Membership Testing for BLS12-381

This document describes the implementation of Tate pairing-based subgroup membership tests for the G1 group of the BLS12-381 curve, based on the paper ["Revisiting subgroup membership testing on pairing-friendly curves via the Tate pairing"](https://eprint.iacr.org/2024/1790.pdf) by Y. Dai et al.

## Background

### The Problem

Given a point P on the BLS12-381 curve E(F_p), we want to efficiently determine whether P belongs to the prime-order subgroup G1 (the r-torsion subgroup).

### Existing Approach: Scott's Method

The standard approach in gnark-crypto uses Scott's method based on the GLV endomorphism:

```
φ(P) = (ωx, y)  where ω is a cube root of unity
```

For P ∈ G1, we have φ(P) = λP where λ is the eigenvalue of the endomorphism. Scott's method computes both sides and compares.

**Cost**: ~41μs (scalar multiplication by λ ≈ 64-bit scalar)

### Tate Pairing Approach

The Tate-based method uses the bilinearity of the Tate pairing to test membership. If P ∈ G1, then the Tate pairing T_r(P, Q) is trivial for any Q in the torsion group.

For BLS12-381, we decompose the test using:
- e2 = |z - 1| where z = -0xd201000000010000 (the curve parameter)
- Two checks using T_{e2}(P, Q) and T_{e2}(φ̂(P), Q)

where φ̂(x,y) = (ω²x, y) is the dual endomorphism.

## Implementation Methods

### 1. Original Tate Method (`IsInSubGroupTate`)

**File**: `g1_subgroup_tate.go`

The baseline implementation following the paper directly:
- NAF-based Miller loop with precomputed table
- Two separate final exponentiations: f1^exp1 and f2^exp2
- exp1 = (p-1)/e2 (~317 bits, 311 squares + 70 multiplies)
- exp2 = |z^5-z^4-z^3+z^2+z+2| (polynomial in z, ~320 squares via expBySeed)

**Performance**: ~55μs

### 2. Chain-Based Method (`IsInSubGroupTateChain`)

**File**: `g1_subgroup_tate_chain.go`

Optimizes the Miller loop by unrolling it into a straight-line addition chain:
- Eliminates loop branching overhead
- Exploits the fixed NAF structure of e2-1
- Same final exponentiations as Original

**Chain Structure** (33 iterations):
```
1.  SQPL + SSUB  (bits 63→62→61)    6 table entries
2.  SQPL + SADD  (bits 61→60→59)    6 table entries
3.  SQPL         (bits 59→58→57)    4 table entries
4.  SDADD        (bit 57)           4 table entries
5-8.  4× SQPL    (bits 56→48)      16 table entries
9.  SDADD        (bit 48)           4 table entries
10-24. 15× SQPL  (bits 47→17)      60 table entries
25. SQPL + SADD  (bits 17→16→15)    6 table entries
26-33. 8× SQPL   (bits 15→-1)      32 table entries
                           Total: 138 table entries
```

**Operations**:
- SQPL: Quadrupling (two doublings) - stores λ_T, x_{2T}, y_{2T}, λ_{2T}
- SADD: Addition after SQPL - stores λ_{T,P}, x_T
- SSUB: Subtraction after SQPL - stores λ_{P,-T}, x_{T-P}
- SDADD: Combined doubling + addition - stores x_T, y_T, A, B

**Performance**: ~52μs (~5% faster than Original)

### 3. Fast Method (`IsInSubGroupTateFast`)

**File**: `g1_subgroup_tate_fast.go`

Optimizes the check order for potential early exit:
- Uses chain-based Miller loop
- Checks f2 first (cheaper), then f1
- For non-members failing the f2 check, saves the f1 computation

**Performance**: ~46μs (~12% slower than Scott)

### 4. Combined Method (`IsInSubGroupTateCombined`)

**File**: `g1_subgroup_tate_fast.go`

Uses a combined product check instead of two separate checks:
- Computes r1 = f1^exp1 and r2 = f2^exp2
- Checks if r1 × r2 = 1

**Soundness**: For P ∈ G1, both r1 = 1 and r2 = 1, so the product is 1. For P ∉ G1, the probability that r1 = r2^{-1} (causing a false positive) is negligible (~2^{-381}).

**Performance**: ~53μs

### 5. Probabilistic Method (`IsInSubGroupTateProbabilistic`)

**File**: `g1_subgroup_tate_fast.go`

The fastest variant that skips the expensive exp1 computation:
- Uses chain-based Miller loop
- Only performs the f2 check (exp2 exponentiation)
- Skips the f1 check entirely

**Soundness**:
- For P ∈ G1: Always returns true (no false negatives)
- For P ∉ G1: Returns false with probability 1 - 2^{-64}

The false positive probability comes from non-members that happen to satisfy the f2 check but fail the f1 check. This is bounded by ~2^{-64}, which is negligible for most applications.

**Performance**: ~34μs (**~17% faster than Scott**)

## Performance Comparison

| Method | Time | vs Scott | Deterministic |
|--------|------|----------|---------------|
| **Probabilistic** | ~34μs | **17% faster** | No (2^{-64} FP rate) |
| Scott | ~41μs | baseline | Yes |
| Fast | ~46μs | 12% slower | Yes |
| Chain | ~52μs | 27% slower | Yes |
| Combined | ~53μs | 29% slower | Yes |
| Original | ~55μs | 34% slower | Yes |

## When to Use Each Method

### Use `IsInSubGroup()` (Scott's method) when:
- You need a deterministic test with no precomputation
- Single-point verification without setup phase
- Memory is constrained (no precomputed table needed)

### Use `IsInSubGroupTateChain()` when:
- You need deterministic correctness
- You can afford the precomputation table (~138 field elements)
- You want the Tate-based approach for consistency with other code

### Use `IsInSubGroupTateFast()` when:
- You need deterministic correctness
- You expect many non-member inputs (benefits from early exit)
- Similar use case to Chain but with slight optimization

### Use `IsInSubGroupTateProbabilistic()` when:
- **Maximum performance is critical**
- You can tolerate a 2^{-64} false positive rate
- Validating points from semi-trusted sources
- Batch verification where occasional false positives are acceptable
- The 17% speedup over Scott justifies the negligible error rate

## Mathematical Details

### The Tate Pairing Test

The membership test exploits the fact that for P ∈ G1:
```
T_{e2}(P, Q)^{(p-1)/e2} = 1
T_{e2}(φ̂(P), Q)^{exp2} = 1
```

where Q is a fixed torsion point precomputed for the test.

### Why Two Checks?

The e2-torsion subgroup E[e2] has two generators over F_p. The two checks ensure P is orthogonal to both generators:
- First check (f1^exp1 = 1): Tests orthogonality to generator G1
- Second check (f2^exp2 = 1): Tests orthogonal to generator G2

A non-member could potentially pass one check but not both.

### The exp2 Optimization

The exponent exp2 = |z^5 - z^4 - z^3 + z^2 + z + 2| has polynomial structure in z. Instead of a generic 320-bit exponentiation, we compute:

```go
u0 = x^2 * x^{z^2} * x^{z^3}
u1 = x^z * x^{z^4} * x^{z^5}
check: u0 == u1
```

Each x^{z^k} is computed via `expBySeed` which exploits the sparse binary representation of z.

### Precomputation Table

The precomputed table stores intermediate values from the Miller loop computation on a fixed torsion point Q. This allows the actual membership test to avoid expensive point operations, replacing them with field multiplications using the precomputed values.

## Files

- `g1_subgroup_tate.go` - Original NAF-based implementation
- `g1_subgroup_tate_precompute.go` - Table generation for Original
- `g1_subgroup_tate_chain.go` - Chain-based Miller loop
- `g1_subgroup_tate_chain_precompute.go` - Table generation for Chain
- `g1_subgroup_tate_fast.go` - Fast, Combined, and Probabilistic variants
- `g1_subgroup_tate_chain_test.go` - Tests and benchmarks

## References

1. Y. Dai et al., "Revisiting subgroup membership testing on pairing-friendly curves via the Tate pairing", 2024. https://eprint.iacr.org/2024/1790.pdf

2. M. Scott, "A note on group membership tests for G1, G2 and GT on BLS pairing-friendly curves", 2021. https://eprint.iacr.org/2021/1130.pdf
