package lattice

import (
	"math/big"
)

// RationalReconstruct finds small (x, z) such that k = x/z mod r.
// This uses LLL lattice reduction on a 3×2 lattice.
//
// The expected bounds on the output are approximately 1.16*r^(1/2).
//
// Parameters:
//   - k: the scalar to decompose
//   - r: the modulus (curve order)
//
// Returns [x, z] as big.Int pointers.
func RationalReconstruct(k, r *big.Int) [2]*big.Int {
	// Build a 3x2 basis for the lattice.
	// The lattice consists of vectors (x, z) ∈ Z^2 such that x ≡ k*z (mod r).
	//
	// The 3x2 basis is:
	//   B1 = [r, 0]    (mod r in x)
	//   B2 = [0, r]    (mod r in z)
	//   B3 = [k, 1]    (x ≡ k*z)

	const nRows = 3
	const nCols = 2

	basis := make([][]big.Int, nRows)
	for i := range basis {
		basis[i] = make([]big.Int, nCols)
	}

	// B1 = [r, 0]
	basis[0][0].Set(r)
	// B2 = [0, r]
	basis[1][1].Set(r)
	// B3 = [k, 1]
	basis[2][0].Set(k)
	basis[2][1].SetInt64(1)

	// Run LLL reduction
	lllReduce(basis, nRows)

	// Find the shortest row with non-zero z component
	bestIdx := -1
	var bestNorm big.Int
	for i := 0; i < nRows; i++ {
		if basis[i][1].Sign() != 0 {
			var norm big.Int
			for j := 0; j < nCols; j++ {
				var absVal big.Int
				absVal.Abs(&basis[i][j])
				if absVal.Cmp(&norm) > 0 {
					norm.Set(&absVal)
				}
			}
			if bestIdx == -1 || norm.Cmp(&bestNorm) < 0 {
				bestIdx = i
				bestNorm.Set(&norm)
			}
		}
	}

	if bestIdx == -1 {
		bestIdx = 0
	}

	return [2]*big.Int{
		new(big.Int).Set(&basis[bestIdx][0]),
		new(big.Int).Set(&basis[bestIdx][1]),
	}
}

// MultiRationalReconstruct finds small (x1, x2, z) such that
// k1 = x1/z mod r and k2 = x2/z mod r.
//
// This uses LLL lattice reduction on a 4×3 lattice.
// The expected bounds on the output are approximately 1.22*r^(2/3).
//
// Parameters:
//   - k1, k2: the scalars to decompose
//   - r: the modulus (curve order)
//
// Returns [x1, x2, z] as big.Int pointers.
func MultiRationalReconstruct(k1, k2, r *big.Int) [3]*big.Int {
	// Build a 4x3 basis for the lattice.
	// The lattice consists of vectors (x1, x2, z) ∈ Z^3 such that
	// x1 ≡ k1*z (mod r) and x2 ≡ k2*z (mod r).
	//
	// The 4x3 basis is:
	//   B1 = [r, 0, 0]     (mod r in x1)
	//   B2 = [0, r, 0]     (mod r in x2)
	//   B3 = [0, 0, r]     (mod r in z)
	//   B4 = [k1, k2, 1]   (x1 ≡ k1*z, x2 ≡ k2*z)

	const nRows = 4
	const nCols = 3

	basis := make([][]big.Int, nRows)
	for i := range basis {
		basis[i] = make([]big.Int, nCols)
	}

	// B1 = [r, 0, 0]
	basis[0][0].Set(r)
	// B2 = [0, r, 0]
	basis[1][1].Set(r)
	// B3 = [0, 0, r]
	basis[2][2].Set(r)
	// B4 = [k1, k2, 1]
	basis[3][0].Set(k1)
	basis[3][1].Set(k2)
	basis[3][2].SetInt64(1)

	// Run LLL reduction
	lllReduce(basis, nRows)

	// Find the shortest row with non-zero z component
	bestIdx := -1
	var bestNorm big.Int
	for i := 0; i < nRows; i++ {
		if basis[i][2].Sign() != 0 {
			var norm big.Int
			for j := 0; j < nCols; j++ {
				var absVal big.Int
				absVal.Abs(&basis[i][j])
				if absVal.Cmp(&norm) > 0 {
					norm.Set(&absVal)
				}
			}
			if bestIdx == -1 || norm.Cmp(&bestNorm) < 0 {
				bestIdx = i
				bestNorm.Set(&norm)
			}
		}
	}

	if bestIdx == -1 {
		bestIdx = 0
	}

	return [3]*big.Int{
		new(big.Int).Set(&basis[bestIdx][0]),
		new(big.Int).Set(&basis[bestIdx][1]),
		new(big.Int).Set(&basis[bestIdx][2]),
	}
}

// RationalReconstructExt finds small (x, y, z, t) such that k = (x + λy)/(z + λt) mod r.
// This uses LLL lattice reduction on a 7×4 lattice (7 generators, 4 coordinates).
//
// The expected bounds on the output are approximately 1.25*r^(1/4).
//
// Parameters:
//   - k: the scalar to decompose
//   - r: the modulus (curve order)
//   - lambda: a quadratic extension generator (e.g., primitive cube root of unity mod r)
//
// Returns [x, y, z, t] as big.Int pointers.
func RationalReconstructExt(k, r, lambda *big.Int) [4]*big.Int {
	// Build a 7x4 basis for the lattice.
	// The lattice consists of vectors (x, y, z, t) ∈ Z^4 such that
	// x + λy ≡ k(z + λt) (mod r).
	//
	// The 7x4 basis is (as per Sage reference):
	//   B1 = [r, 0, 0, 0]       (mod r in x)
	//   B2 = [0, r, 0, 0]       (mod r in y)
	//   B3 = [0, 0, r, 0]       (mod r in z)
	//   B4 = [0, 0, 0, r]       (mod r in t)
	//   B5 = [-λ, 1, 0, 0]      (x + λy relation)
	//   B6 = [k, 0, 1, 0]       (x ≡ k*z)
	//   B7 = [0, 0, -λ, 1]      (z + λt relation in denominator)

	const nRows = 7
	const nCols = 4

	basis := make([][]big.Int, nRows)
	for i := range basis {
		basis[i] = make([]big.Int, nCols)
	}

	// B1 = [r, 0, 0, 0]
	basis[0][0].Set(r)
	// B2 = [0, r, 0, 0]
	basis[1][1].Set(r)
	// B3 = [0, 0, r, 0]
	basis[2][2].Set(r)
	// B4 = [0, 0, 0, r]
	basis[3][3].Set(r)
	// B5 = [-λ, 1, 0, 0]
	basis[4][0].Neg(lambda)
	basis[4][1].SetInt64(1)
	// B6 = [k, 0, 1, 0]
	basis[5][0].Set(k)
	basis[5][2].SetInt64(1)
	// B7 = [0, 0, -λ, 1]
	basis[6][2].Neg(lambda)
	basis[6][3].SetInt64(1)

	// Run LLL reduction on the 7x4 matrix
	lllReduce(basis, nRows)

	// Find the shortest row with non-zero (z, t) component
	bestIdx := -1
	var bestNorm big.Int
	for i := 0; i < nRows; i++ {
		// Check if z or t is non-zero (columns 2 and 3)
		if basis[i][2].Sign() != 0 || basis[i][3].Sign() != 0 {
			// Compute infinity norm: max(|x|, |y|, |z|, |t|)
			var norm big.Int
			for j := 0; j < nCols; j++ {
				var absVal big.Int
				absVal.Abs(&basis[i][j])
				if absVal.Cmp(&norm) > 0 {
					norm.Set(&absVal)
				}
			}
			if bestIdx == -1 || norm.Cmp(&bestNorm) < 0 {
				bestIdx = i
				bestNorm.Set(&norm)
			}
		}
	}

	if bestIdx == -1 {
		bestIdx = 0
	}

	return [4]*big.Int{
		new(big.Int).Set(&basis[bestIdx][0]),
		new(big.Int).Set(&basis[bestIdx][1]),
		new(big.Int).Set(&basis[bestIdx][2]),
		new(big.Int).Set(&basis[bestIdx][3]),
	}
}

// MultiRationalReconstructExt finds small (x1, y1, x2, y2, z, t) such that
// k1 = (x1 + λy1)/(z + λt) mod r and k2 = (x2 + λy2)/(z + λt) mod r.
//
// This uses lattice reduction on a 10×6 lattice (10 generators, 6 coordinates).
// The expected bounds on the output are approximately 1.28*r^(1/3).
//
// Parameters:
//   - k1, k2: the scalars to decompose
//   - r: the modulus (curve order)
//   - lambda: a quadratic extension generator
//
// Returns [x1, y1, x2, y2, z, t] as big.Int pointers.
func MultiRationalReconstructExt(k1, k2, r, lambda *big.Int) [6]*big.Int {
	// Build a 10x6 basis for the lattice.
	// The lattice consists of vectors (x1, y1, x2, y2, z, t) ∈ Z^6 such that
	// x1 + λy1 ≡ k1(z + λt) (mod r) and x2 + λy2 ≡ k2(z + λt) (mod r).
	//
	// The 10x6 basis is (as per Sage reference):
	//   B1  = [r, 0, 0, 0, 0, 0]       (mod r in x1)
	//   B2  = [0, r, 0, 0, 0, 0]       (mod r in y1)
	//   B3  = [0, 0, r, 0, 0, 0]       (mod r in x2)
	//   B4  = [0, 0, 0, r, 0, 0]       (mod r in y2)
	//   B5  = [0, 0, 0, 0, r, 0]       (mod r in z)
	//   B6  = [0, 0, 0, 0, 0, r]       (mod r in t)
	//   B7  = [-λ, 1, 0, 0, 0, 0]      (x1 + λy1 relation)
	//   B8  = [0, 0, -λ, 1, 0, 0]      (x2 + λy2 relation)
	//   B9  = [k1, 0, k2, 0, 1, 0]     (x1 ≡ k1*z, x2 ≡ k2*z)
	//   B10 = [0, 0, 0, 0, -λ, 1]      (z + λt relation in denominator)

	const nRows = 10
	const nCols = 6

	basis := make([][]big.Int, nRows)
	for i := range basis {
		basis[i] = make([]big.Int, nCols)
	}

	// B1-B6 = identity * r (mod r constraints for each coordinate)
	basis[0][0].Set(r)
	basis[1][1].Set(r)
	basis[2][2].Set(r)
	basis[3][3].Set(r)
	basis[4][4].Set(r)
	basis[5][5].Set(r)

	// B7 = [-λ, 1, 0, 0, 0, 0]
	basis[6][0].Neg(lambda)
	basis[6][1].SetInt64(1)

	// B8 = [0, 0, -λ, 1, 0, 0]
	basis[7][2].Neg(lambda)
	basis[7][3].SetInt64(1)

	// B9 = [k1, 0, k2, 0, 1, 0]
	basis[8][0].Set(k1)
	basis[8][2].Set(k2)
	basis[8][4].SetInt64(1)

	// B10 = [0, 0, 0, 0, -λ, 1]
	basis[9][4].Neg(lambda)
	basis[9][5].SetInt64(1)

	// Run LLL reduction on the 10x6 matrix
	lllReduce(basis, nRows)

	// Find the shortest row with non-zero (z, t) component
	bestIdx := -1
	var bestNorm big.Int
	for i := 0; i < nRows; i++ {
		if basis[i][4].Sign() != 0 || basis[i][5].Sign() != 0 {
			var norm big.Int
			for j := 0; j < nCols; j++ {
				var absVal big.Int
				absVal.Abs(&basis[i][j])
				if absVal.Cmp(&norm) > 0 {
					norm.Set(&absVal)
				}
			}
			if bestIdx == -1 || norm.Cmp(&bestNorm) < 0 {
				bestIdx = i
				bestNorm.Set(&norm)
			}
		}
	}

	if bestIdx == -1 {
		bestIdx = 0
	}

	return [6]*big.Int{
		new(big.Int).Set(&basis[bestIdx][0]),
		new(big.Int).Set(&basis[bestIdx][1]),
		new(big.Int).Set(&basis[bestIdx][2]),
		new(big.Int).Set(&basis[bestIdx][3]),
		new(big.Int).Set(&basis[bestIdx][4]),
		new(big.Int).Set(&basis[bestIdx][5]),
	}
}

// lllReduce performs in-place LLL reduction on an m×n basis matrix (m rows, n columns).
// Uses rational arithmetic with big.Rat for correctness.
// Delta is fixed at 99/100 = 0.99 for stronger reduction.
// For non-square matrices (m > n), this finds a reduced basis for the lattice
// generated by the row vectors in R^n.
func lllReduce(basis [][]big.Int, m int) {
	if m == 0 {
		return
	}
	n := len(basis[0]) // number of columns

	// delta = 99/100
	delta := big.NewRat(99, 100)

	// ortho stores the Gram-Schmidt orthogonalized vectors as rationals
	ortho := make([][]big.Rat, m)
	for i := range ortho {
		ortho[i] = make([]big.Rat, n)
	}

	// muCache[i][j] stores the Gram-Schmidt coefficient μ[i][j] = <basis[i], ortho[j]> / <ortho[j], ortho[j]>
	// Only valid for j < i. Updated incrementally with Gram-Schmidt.
	muCache := make([][]big.Rat, m)
	for i := range muCache {
		muCache[i] = make([]big.Rat, m)
	}

	// B[i] stores ||ortho[i]||² (squared norm of i-th orthogonalized vector)
	B := make([]big.Rat, m)

	// Temporary variables for Gram-Schmidt computation
	var term, vi big.Rat

	// updateGramSchmidtFrom recomputes Gram-Schmidt orthogonalization starting from index 'from'.
	// Also updates muCache[i][j] for i >= from and B[i] for i >= from.
	updateGramSchmidtFrom := func(from int) {
		for i := from; i < m; i++ {
			// Start with ortho[i] = basis[i]
			for j := 0; j < n; j++ {
				ortho[i][j].SetInt(&basis[i][j])
			}
			// Subtract projections onto previous orthogonalized vectors
			for j := 0; j < i; j++ {
				// Skip zero vectors (B[j] == 0)
				if B[j].Sign() == 0 {
					muCache[i][j].SetInt64(0)
					continue
				}
				// Compute μ[i][j] = <basis[i], ortho[j]> / B[j]
				muCache[i][j].SetInt64(0)
				for l := 0; l < n; l++ {
					vi.SetInt(&basis[i][l])
					term.Mul(&vi, &ortho[j][l])
					muCache[i][j].Add(&muCache[i][j], &term)
				}
				muCache[i][j].Quo(&muCache[i][j], &B[j])

				// ortho[i] -= μ[i][j] * ortho[j]
				for l := 0; l < n; l++ {
					term.Mul(&muCache[i][j], &ortho[j][l])
					ortho[i][l].Sub(&ortho[i][l], &term)
				}
			}
			// Compute B[i] = ||ortho[i]||²
			B[i].SetInt64(0)
			for l := 0; l < n; l++ {
				term.Mul(&ortho[i][l], &ortho[i][l])
				B[i].Add(&B[i], &term)
			}
		}
	}

	// Initial full Gram-Schmidt
	updateGramSchmidtFrom(0)

	k := 1
	half := big.NewRat(1, 2)
	var muSquared, threshold, rhs, absMu big.Rat

	for k < m {
		// Size reduction: repeat until all |μ[k][j]| <= 1/2
		for {
			reduced := false
			for j := k - 1; j >= 0; j-- {
				if B[j].Sign() == 0 {
					continue
				}

				// Check if |μ[k][j]| > 1/2 using cached value
				absMu.Abs(&muCache[k][j])
				if absMu.Cmp(half) > 0 {
					// q = round(μ[k][j])
					q := roundRat(&muCache[k][j])

					// basis[k] -= q * basis[j]
					var tmp big.Int
					for l := 0; l < n; l++ {
						tmp.Mul(q, &basis[j][l])
						basis[k][l].Sub(&basis[k][l], &tmp)
					}

					// Only recompute Gram-Schmidt from k onwards
					updateGramSchmidtFrom(k)
					reduced = true
				}
			}
			if !reduced {
				break
			}
		}

		// Check for zero vector at k-1
		if k > 0 && B[k-1].Sign() == 0 {
			k++
			continue
		}

		// Lovász condition: B[k] >= (δ - μ[k][k-1]²) * B[k-1]
		muSquared.Mul(&muCache[k][k-1], &muCache[k][k-1])
		threshold.Sub(delta, &muSquared)
		rhs.Mul(&threshold, &B[k-1])

		if B[k].Cmp(&rhs) >= 0 {
			k++
		} else {
			// Swap basis[k] and basis[k-1]
			basis[k], basis[k-1] = basis[k-1], basis[k]
			// Only recompute Gram-Schmidt from k-1 onwards
			updateGramSchmidtFrom(k - 1)
			if k > 1 {
				k--
			}
		}
	}
}

// roundRat rounds a rational to the nearest integer
func roundRat(r *big.Rat) *big.Int {
	// Get numerator and denominator
	num := r.Num()
	den := r.Denom()

	// Compute quotient and remainder
	q := new(big.Int)
	rem := new(big.Int)
	q.DivMod(num, den, rem)

	// Round to nearest: if |rem| * 2 >= |den|, adjust
	rem2 := new(big.Int).Mul(rem, big.NewInt(2))
	if rem2.CmpAbs(den) >= 0 {
		if num.Sign() >= 0 {
			q.Add(q, big.NewInt(1))
		} else {
			q.Sub(q, big.NewInt(1))
		}
	}

	return q
}
