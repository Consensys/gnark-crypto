package lattice

import (
	"math/big"
)

// Reconstructor provides efficient rational reconstruction for repeated operations
// on the same modulus. It caches precomputed values like √r to avoid redundant
// computation when processing many scalars.
//
// For one-off reconstructions, use the standalone [RationalReconstruct] function.
// For high-performance scenarios (e.g., SNARK proving with millions of scalars),
// create a Reconstructor once and reuse it.
type Reconstructor struct {
	r           *big.Int // modulus (curve order)
	sqrtR       *big.Int // cached √r for RationalReconstruct (half-GCD threshold)
	cbrtR       *big.Int // cached ∛r for MultiRationalReconstructExt early termination
	cbrtR2      *big.Int // cached r^(2/3) for MultiRationalReconstruct early termination
	fourthRootR *big.Int // cached r^(1/4) for RationalReconstructExt early termination
	lambda      *big.Int // cached λ for extension methods (optional, set via SetLambda)
}

// NewReconstructor creates a new Reconstructor for the given modulus r.
// It precomputes various roots of r which are reused across all reconstruction calls.
//
// For extension methods (RationalReconstructExt, MultiRationalReconstructExt),
// call SetLambda to set the quadratic extension generator.
func NewReconstructor(r *big.Int) *Reconstructor {
	sqrtR := new(big.Int).Sqrt(r)
	fourthRootR := new(big.Int).Sqrt(sqrtR)
	cbrtR := intCbrt(r)
	// r^(2/3) = (r^2)^(1/3)
	r2 := new(big.Int).Mul(r, r)
	cbrtR2 := intCbrt(r2)
	return &Reconstructor{
		r:           new(big.Int).Set(r),
		sqrtR:       sqrtR,
		cbrtR:       cbrtR,
		cbrtR2:      cbrtR2,
		fourthRootR: fourthRootR,
	}
}

// intCbrt computes the integer cube root of n using Newton's method.
// Returns floor(n^(1/3)).
func intCbrt(n *big.Int) *big.Int {
	if n.Sign() <= 0 {
		return big.NewInt(0)
	}

	// Initial guess: 2^((bitLen+2)/3)
	bitLen := n.BitLen()
	x := new(big.Int).Lsh(big.NewInt(1), uint((bitLen+2)/3))

	// Newton iteration: x = (2x + n/x²) / 3
	var x2, q, sum big.Int
	three := big.NewInt(3)
	two := big.NewInt(2)

	for {
		x2.Mul(x, x)
		q.Div(n, &x2)
		sum.Mul(x, two)
		sum.Add(&sum, &q)
		xNew := new(big.Int).Div(&sum, three)

		// Check for convergence
		if xNew.Cmp(x) >= 0 {
			break
		}
		x = xNew
	}

	// Verify and adjust: ensure x³ ≤ n < (x+1)³
	var x3 big.Int
	x3.Mul(x, x)
	x3.Mul(&x3, x)
	if x3.Cmp(n) > 0 {
		x.Sub(x, big.NewInt(1))
	}

	return x
}

// SetLambda sets the quadratic extension generator λ for extension methods.
// This is required before calling RationalReconstructExt or MultiRationalReconstructExt.
// Returns the Reconstructor for method chaining.
func (rc *Reconstructor) SetLambda(lambda *big.Int) *Reconstructor {
	rc.lambda = new(big.Int).Set(lambda)
	return rc
}

// RationalReconstruct finds small (x, z) such that k = x/z mod r.
//
// The bounds on the output are |x|, |z| < √r (more precisely, ≤ γ₂·√r ≈ 1.15·√r
// where γ₂ = 2/√3 ≈ 1.1547 is the 2D Hermite constant).
//
// This method uses the cached √r from the Reconstructor, making it more efficient
// than the standalone function when called repeatedly.
func (rc *Reconstructor) RationalReconstruct(k *big.Int) [2]*big.Int {
	var r0, r1, t0, t1, q, tmp big.Int

	// Initialize: (r0, t0) = (r, 0), (r1, t1) = (k, 1)
	r0.Set(rc.r)
	t0.SetInt64(0)
	r1.Mod(k, rc.r) // Ensure k is reduced mod r
	t1.SetInt64(1)

	// Run extended Euclidean algorithm until r1 < √r
	for r1.Cmp(rc.sqrtR) >= 0 {
		// q = r0 / r1
		q.Div(&r0, &r1)

		// (r0, r1) = (r1, r0 - q*r1)
		tmp.Mul(&q, &r1)
		tmp.Sub(&r0, &tmp)
		r0.Set(&r1)
		r1.Set(&tmp)

		// (t0, t1) = (t1, t0 - q*t1)
		tmp.Mul(&q, &t1)
		tmp.Sub(&t0, &tmp)
		t0.Set(&t1)
		t1.Set(&tmp)
	}

	// x = r1, z = t1
	// We have x ≡ z*k (mod r)
	return [2]*big.Int{
		new(big.Int).Set(&r1),
		new(big.Int).Set(&t1),
	}
}

// MultiRationalReconstruct finds small (x1, x2, z) such that
// k1 = x1/z mod r and k2 = x2/z mod r.
//
// This uses LLL lattice reduction on a 4×3 lattice with early termination
// when a vector with all components ≤ r^(2/3) is found.
// The expected bounds on the output are approximately 1.22*r^(2/3).
//
// This method uses the cached r and r^(2/3) from the Reconstructor.
func (rc *Reconstructor) MultiRationalReconstruct(k1, k2 *big.Int) [3]*big.Int {
	const nRows = 4
	const nCols = 3

	basis := make([][]big.Int, nRows)
	for i := range basis {
		basis[i] = make([]big.Int, nCols)
	}

	// B1 = [r, 0, 0]
	basis[0][0].Set(rc.r)
	// B2 = [0, r, 0]
	basis[1][1].Set(rc.r)
	// B3 = [0, 0, r]
	basis[2][2].Set(rc.r)
	// B4 = [k1 mod r, k2 mod r, 1]
	basis[3][0].Mod(k1, rc.r)
	basis[3][1].Mod(k2, rc.r)
	basis[3][2].SetInt64(1)

	// Denominator column is 2 (z)
	denCols := []int{2}

	// Try early termination with bound = r^(2/3)
	earlyIdx := lllReduceWithBound(basis, nRows, rc.cbrtR2, denCols)
	if earlyIdx >= 0 {
		return [3]*big.Int{
			new(big.Int).Set(&basis[earlyIdx][0]),
			new(big.Int).Set(&basis[earlyIdx][1]),
			new(big.Int).Set(&basis[earlyIdx][2]),
		}
	}

	// Full reduction completed, find best row
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
		panic("lattice: LLL reduction produced no vector with non-zero denominator")
	}

	return [3]*big.Int{
		new(big.Int).Set(&basis[bestIdx][0]),
		new(big.Int).Set(&basis[bestIdx][1]),
		new(big.Int).Set(&basis[bestIdx][2]),
	}
}

// RationalReconstructExt finds small (x, y, z, t) such that k = (x + λy)/(z + λt) mod r.
// This uses LLL lattice reduction on a 7×4 lattice (7 generators, 4 coordinates)
// with early termination when a vector with all components ≤ r^(1/4) is found.
//
// The expected bounds on the output are approximately 1.25*r^(1/4).
//
// This method uses the cached r, r^(1/4), and λ from the Reconstructor.
// Panics if SetLambda was not called.
func (rc *Reconstructor) RationalReconstructExt(k *big.Int) [4]*big.Int {
	if rc.lambda == nil {
		panic("lattice: Reconstructor.RationalReconstructExt requires SetLambda to be called first")
	}

	const nRows = 7
	const nCols = 4

	basis := make([][]big.Int, nRows)
	for i := range basis {
		basis[i] = make([]big.Int, nCols)
	}

	// B1 = [r, 0, 0, 0]
	basis[0][0].Set(rc.r)
	// B2 = [0, r, 0, 0]
	basis[1][1].Set(rc.r)
	// B3 = [0, 0, r, 0]
	basis[2][2].Set(rc.r)
	// B4 = [0, 0, 0, r]
	basis[3][3].Set(rc.r)
	// B5 = [-λ mod r, 1, 0, 0]
	basis[4][0].Mod(rc.lambda, rc.r)
	basis[4][0].Neg(&basis[4][0])
	basis[4][1].SetInt64(1)
	// B6 = [k mod r, 0, 1, 0]
	basis[5][0].Mod(k, rc.r)
	basis[5][2].SetInt64(1)
	// B7 = [0, 0, -λ mod r, 1]
	basis[6][2].Mod(rc.lambda, rc.r)
	basis[6][2].Neg(&basis[6][2])
	basis[6][3].SetInt64(1)

	// Denominator columns are 2 and 3 (z, t)
	denCols := []int{2, 3}

	// Try early termination with bound = r^(1/4)
	earlyIdx := lllReduceWithBound(basis, nRows, rc.fourthRootR, denCols)
	if earlyIdx >= 0 {
		return [4]*big.Int{
			new(big.Int).Set(&basis[earlyIdx][0]),
			new(big.Int).Set(&basis[earlyIdx][1]),
			new(big.Int).Set(&basis[earlyIdx][2]),
			new(big.Int).Set(&basis[earlyIdx][3]),
		}
	}

	// Full reduction completed, find best row
	bestIdx := -1
	var bestNorm big.Int
	for i := 0; i < nRows; i++ {
		if basis[i][2].Sign() != 0 || basis[i][3].Sign() != 0 {
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
		panic("lattice: LLL reduction produced no vector with non-zero denominator")
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
// This uses lattice reduction on a 10×6 lattice (10 generators, 6 coordinates)
// with early termination when a vector with all components ≤ r^(1/3) is found.
// The expected bounds on the output are approximately 1.28*r^(1/3).
//
// This method uses the cached r, r^(1/3), and λ from the Reconstructor.
// Panics if SetLambda was not called.
func (rc *Reconstructor) MultiRationalReconstructExt(k1, k2 *big.Int) [6]*big.Int {
	if rc.lambda == nil {
		panic("lattice: Reconstructor.MultiRationalReconstructExt requires SetLambda to be called first")
	}

	const nRows = 10
	const nCols = 6

	basis := make([][]big.Int, nRows)
	for i := range basis {
		basis[i] = make([]big.Int, nCols)
	}

	// B1-B6 = identity * r (mod r constraints for each coordinate)
	basis[0][0].Set(rc.r)
	basis[1][1].Set(rc.r)
	basis[2][2].Set(rc.r)
	basis[3][3].Set(rc.r)
	basis[4][4].Set(rc.r)
	basis[5][5].Set(rc.r)

	// B7 = [-λ mod r, 1, 0, 0, 0, 0]
	basis[6][0].Mod(rc.lambda, rc.r)
	basis[6][0].Neg(&basis[6][0])
	basis[6][1].SetInt64(1)

	// B8 = [0, 0, -λ mod r, 1, 0, 0]
	basis[7][2].Mod(rc.lambda, rc.r)
	basis[7][2].Neg(&basis[7][2])
	basis[7][3].SetInt64(1)

	// B9 = [k1 mod r, 0, k2 mod r, 0, 1, 0]
	basis[8][0].Mod(k1, rc.r)
	basis[8][2].Mod(k2, rc.r)
	basis[8][4].SetInt64(1)

	// B10 = [0, 0, 0, 0, -λ mod r, 1]
	basis[9][4].Mod(rc.lambda, rc.r)
	basis[9][4].Neg(&basis[9][4])
	basis[9][5].SetInt64(1)

	// Denominator columns are 4 and 5 (z, t)
	denCols := []int{4, 5}

	// Try early termination with bound = r^(1/3)
	earlyIdx := lllReduceWithBound(basis, nRows, rc.cbrtR, denCols)
	if earlyIdx >= 0 {
		return [6]*big.Int{
			new(big.Int).Set(&basis[earlyIdx][0]),
			new(big.Int).Set(&basis[earlyIdx][1]),
			new(big.Int).Set(&basis[earlyIdx][2]),
			new(big.Int).Set(&basis[earlyIdx][3]),
			new(big.Int).Set(&basis[earlyIdx][4]),
			new(big.Int).Set(&basis[earlyIdx][5]),
		}
	}

	// Full reduction completed, find best row
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
		panic("lattice: LLL reduction produced no vector with non-zero denominator")
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

// RationalReconstruct finds small (x, z) such that k = x/z mod r.
//
// The bounds on the output are |x|, |z| < √r (more precisely, ≤ γ₂·√r ≈ 1.15·√r
// where γ₂ = 2/√3 ≈ 1.1547 is the 2D Hermite constant).
//
// # Algorithm
//
// This problem can be solved equivalently by:
//
//  1. Half-GCD: Extended Euclidean algorithm stopped at remainder < √r (used here)
//  2. LLL: Reduce the 2D lattice {(r,0), (k,1)} to find a short vector
//  3. Pornin's Algorithm: Division-free Lagrange reduction using only shifts/adds
//
// We use half-GCD because in Go with math/big:
//
//   - big.Int.Div is highly optimized (assembly-level Karatsuba/Newton)
//   - Half-GCD needs ~40 iterations vs ~98 for Pornin's algorithm
//   - Benchmarks: half-GCD ~5μs vs Pornin ~6μs for 254-bit primes
//
// Pornin's algorithm (eprint 2020/454) is faster on embedded systems where
// division is expensive, but the "no division" benefit is masked by big.Int
// overhead in Go.
//
// Parameters:
//   - k: the scalar to decompose
//   - r: the modulus (curve order)
//
// Returns [x, z] as big.Int pointers.
func RationalReconstruct(k, r *big.Int) [2]*big.Int {
	// The extended Euclidean algorithm maintains the invariant:
	//   r_i = s_i * r + t_i * k
	// which means r_i ≡ t_i * k (mod r).
	//
	// When we stop at r_i < √r, we have:
	//   x = r_i (the remainder, small)
	//   z = t_i (the coefficient, also small by continued fraction theory)
	//
	// This gives x ≡ z*k (mod r), i.e., k ≡ x/z (mod r).

	var r0, r1, t0, t1, q, tmp big.Int

	// Initialize: (r0, t0) = (r, 0), (r1, t1) = (k, 1)
	r0.Set(r)
	t0.SetInt64(0)
	r1.Mod(k, r) // Ensure k is reduced mod r
	t1.SetInt64(1)

	// Compute √r as the stopping threshold
	var sqrtR big.Int
	sqrtR.Sqrt(r)

	// Run extended Euclidean algorithm until r1 < √r
	for r1.Cmp(&sqrtR) >= 0 {
		// q = r0 / r1
		q.Div(&r0, &r1)

		// (r0, r1) = (r1, r0 - q*r1)
		tmp.Mul(&q, &r1)
		tmp.Sub(&r0, &tmp)
		r0.Set(&r1)
		r1.Set(&tmp)

		// (t0, t1) = (t1, t0 - q*t1)
		tmp.Mul(&q, &t1)
		tmp.Sub(&t0, &tmp)
		t0.Set(&t1)
		t1.Set(&tmp)
	}

	// x = r1, z = t1
	// We have x ≡ z*k (mod r)
	return [2]*big.Int{
		new(big.Int).Set(&r1),
		new(big.Int).Set(&t1),
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
	// B4 = [k1 mod r, k2 mod r, 1]
	basis[3][0].Mod(k1, r)
	basis[3][1].Mod(k2, r)
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
		panic("lattice: LLL reduction produced no vector with non-zero denominator")
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
// Note: Unlike the 2D case where half-GCD in Z achieves the same bounds as
// LLL, the 4D problem cannot be solved "as" efficiently (bounds wise) with
// half-GCD in Eisenstein integers Z[ω].
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
	// B5 = [-λ mod r, 1, 0, 0]
	basis[4][0].Mod(lambda, r)
	basis[4][0].Neg(&basis[4][0])
	basis[4][1].SetInt64(1)
	// B6 = [k mod r, 0, 1, 0]
	basis[5][0].Mod(k, r)
	basis[5][2].SetInt64(1)
	// B7 = [0, 0, -λ mod r, 1]
	basis[6][2].Mod(lambda, r)
	basis[6][2].Neg(&basis[6][2])
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
		panic("lattice: LLL reduction produced no vector with non-zero denominator")
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

	// B7 = [-λ mod r, 1, 0, 0, 0, 0]
	basis[6][0].Mod(lambda, r)
	basis[6][0].Neg(&basis[6][0])
	basis[6][1].SetInt64(1)

	// B8 = [0, 0, -λ mod r, 1, 0, 0]
	basis[7][2].Mod(lambda, r)
	basis[7][2].Neg(&basis[7][2])
	basis[7][3].SetInt64(1)

	// B9 = [k1 mod r, 0, k2 mod r, 0, 1, 0]
	basis[8][0].Mod(k1, r)
	basis[8][2].Mod(k2, r)
	basis[8][4].SetInt64(1)

	// B10 = [0, 0, 0, 0, -λ mod r, 1]
	basis[9][4].Mod(lambda, r)
	basis[9][4].Neg(&basis[9][4])
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
		panic("lattice: LLL reduction produced no vector with non-zero denominator")
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

// lazyRat represents a rational number num/den without automatic GCD normalization.
// This avoids the expensive GCD computation in big.Rat while maintaining exact arithmetic.
type lazyRat struct {
	num, den big.Int
}

// setInt sets r = x/1
func (r *lazyRat) setInt(x *big.Int) {
	r.num.Set(x)
	r.den.SetInt64(1)
}

// setInt64 sets r = x/1
func (r *lazyRat) setInt64(x int64) {
	r.num.SetInt64(x)
	r.den.SetInt64(1)
}

// sign returns the sign of r: -1 if r < 0, 0 if r == 0, +1 if r > 0
func (r *lazyRat) sign() int {
	return r.num.Sign() * r.den.Sign()
}

// add sets r = a + b = (a.num*b.den + b.num*a.den) / (a.den*b.den)
func (r *lazyRat) add(a, b *lazyRat) {
	var t1, t2 big.Int
	t1.Mul(&a.num, &b.den)
	t2.Mul(&b.num, &a.den)
	r.num.Add(&t1, &t2)
	r.den.Mul(&a.den, &b.den)
}

// sub sets r = a - b = (a.num*b.den - b.num*a.den) / (a.den*b.den)
func (r *lazyRat) sub(a, b *lazyRat) {
	var t1, t2 big.Int
	t1.Mul(&a.num, &b.den)
	t2.Mul(&b.num, &a.den)
	r.num.Sub(&t1, &t2)
	r.den.Mul(&a.den, &b.den)
}

// mul sets r = a * b = (a.num*b.num) / (a.den*b.den)
func (r *lazyRat) mul(a, b *lazyRat) {
	r.num.Mul(&a.num, &b.num)
	r.den.Mul(&a.den, &b.den)
}

// quo sets r = a / b = (a.num*b.den) / (a.den*b.num)
// Panics if b is zero.
func (r *lazyRat) quo(a, b *lazyRat) {
	if b.num.Sign() == 0 {
		panic("lattice: division by zero in lazyRat.quo")
	}
	var newNum, newDen big.Int
	newNum.Mul(&a.num, &b.den)
	newDen.Mul(&a.den, &b.num)
	// Handle sign: ensure denominator is positive
	if newDen.Sign() < 0 {
		newNum.Neg(&newNum)
		newDen.Neg(&newDen)
	}
	r.num.Set(&newNum)
	r.den.Set(&newDen)
}

// cmp compares r and s: returns -1 if r < s, 0 if r == s, +1 if r > s
// Uses cross-multiplication: r.num*s.den vs s.num*r.den
func (r *lazyRat) cmp(s *lazyRat) int {
	var lhs, rhs big.Int
	lhs.Mul(&r.num, &s.den)
	rhs.Mul(&s.num, &r.den)
	// If denominators have different signs, flip comparison
	if r.den.Sign()*s.den.Sign() < 0 {
		return -lhs.Cmp(&rhs)
	}
	return lhs.Cmp(&rhs)
}

// abs sets r = |a|
func (r *lazyRat) abs(a *lazyRat) {
	r.num.Abs(&a.num)
	r.den.Abs(&a.den)
}

// normalize reduces the fraction using GCD (call sparingly)
func (r *lazyRat) normalize() {
	if r.num.Sign() == 0 {
		r.den.SetInt64(1)
		return
	}
	var g big.Int
	g.GCD(nil, nil, &r.num, &r.den)
	if g.Sign() != 0 && g.Cmp(bigOne) != 0 {
		r.num.Quo(&r.num, &g)
		r.den.Quo(&r.den, &g)
	}
	// Ensure denominator is positive
	if r.den.Sign() < 0 {
		r.num.Neg(&r.num)
		r.den.Neg(&r.den)
	}
}

// roundToInt rounds r to the nearest integer (round half up, toward +∞).
//
// Go's big.Int.DivMod uses Euclidean division where the remainder is always
// non-negative (for positive divisor) and the quotient is the floor.
// For example: -7 DivMod 2 = (-4, 1) because -7 = 2*(-4) + 1.
//
// To round to nearest, we check if the remainder represents >= 0.5:
// if 2*rem >= den, we add 1 to move from floor toward ceiling.
// This works for both positive and negative numbers.
func (r *lazyRat) roundToInt() *big.Int {
	// Make a copy with positive denominator
	num := new(big.Int).Set(&r.num)
	den := new(big.Int).Set(&r.den)
	if den.Sign() < 0 {
		num.Neg(num)
		den.Neg(den)
	}

	q := new(big.Int)
	rem := new(big.Int)
	q.DivMod(num, den, rem)

	// Round to nearest: if 2*rem >= den, add 1 to round up.
	// Since den > 0 and rem >= 0 (Euclidean division), we use Cmp not CmpAbs.
	rem2 := new(big.Int).Mul(rem, big.NewInt(2))
	if rem2.Cmp(den) >= 0 {
		q.Add(q, bigOne)
	}
	return q
}

var bigOne = big.NewInt(1)

// lllReduceWithBound performs LLL reduction with optional early termination.
// If bound is non-nil, it stops early when it finds a row where:
// - At least one of the denominator columns (denCols) is non-zero
// - All components have absolute value <= bound
// Returns the index of the satisfying row, or -1 if none found (or if bound is nil).
func lllReduceWithBound(basis [][]big.Int, m int, bound *big.Int, denCols []int) int {
	if m == 0 {
		return -1
	}
	n := len(basis[0])

	// Helper to check if a row satisfies the bound condition (no-op when bound is nil)
	checkRow := func(row int) bool {
		if bound == nil {
			return false // No early termination when bound is nil
		}
		// Check if at least one denominator column is non-zero
		hasNonZeroDen := false
		for _, col := range denCols {
			if basis[row][col].Sign() != 0 {
				hasNonZeroDen = true
				break
			}
		}
		if !hasNonZeroDen {
			return false
		}
		// Check if all components are within bound
		for j := 0; j < n; j++ {
			var absVal big.Int
			absVal.Abs(&basis[row][j])
			if absVal.Cmp(bound) > 0 {
				return false
			}
		}
		return true
	}

	// Check initial rows before starting LLL (only when bound is set)
	if bound != nil {
		for i := 0; i < m; i++ {
			if checkRow(i) {
				return i
			}
		}
	}

	// delta = 99/100 as lazyRat
	var delta lazyRat
	delta.num.SetInt64(99)
	delta.den.SetInt64(100)

	ortho := make([][]lazyRat, m)
	for i := range ortho {
		ortho[i] = make([]lazyRat, n)
	}

	muCache := make([][]lazyRat, m)
	for i := range muCache {
		muCache[i] = make([]lazyRat, m)
	}

	B := make([]lazyRat, m)
	var term, vi lazyRat

	updateGramSchmidtFrom := func(from int) {
		for i := from; i < m; i++ {
			for j := 0; j < n; j++ {
				ortho[i][j].setInt(&basis[i][j])
			}
			for j := 0; j < i; j++ {
				if B[j].sign() == 0 {
					muCache[i][j].setInt64(0)
					continue
				}
				muCache[i][j].setInt64(0)
				for l := 0; l < n; l++ {
					vi.setInt(&basis[i][l])
					term.mul(&vi, &ortho[j][l])
					muCache[i][j].add(&muCache[i][j], &term)
				}
				muCache[i][j].quo(&muCache[i][j], &B[j])
				muCache[i][j].normalize()

				for l := 0; l < n; l++ {
					term.mul(&muCache[i][j], &ortho[j][l])
					ortho[i][l].sub(&ortho[i][l], &term)
				}
			}
			B[i].setInt64(0)
			for l := 0; l < n; l++ {
				term.mul(&ortho[i][l], &ortho[i][l])
				B[i].add(&B[i], &term)
			}
			B[i].normalize()
			for l := 0; l < n; l++ {
				ortho[i][l].normalize()
			}
		}
	}

	updateGramSchmidtFrom(0)

	k := 1
	var half lazyRat
	half.num.SetInt64(1)
	half.den.SetInt64(2)
	var muSquared, threshold, rhs, absMu lazyRat

	for k < m {
		for {
			reduced := false
			for j := k - 1; j >= 0; j-- {
				if B[j].sign() == 0 {
					continue
				}

				absMu.abs(&muCache[k][j])
				if absMu.cmp(&half) > 0 {
					q := muCache[k][j].roundToInt()

					var tmp big.Int
					for l := 0; l < n; l++ {
						tmp.Mul(q, &basis[j][l])
						basis[k][l].Sub(&basis[k][l], &tmp)
					}

					updateGramSchmidtFrom(k)
					reduced = true

					// Check for early termination after size reduction
					if checkRow(k) {
						return k
					}
				}
			}
			if !reduced {
				break
			}
		}

		if k > 0 && B[k-1].sign() == 0 {
			k++
			continue
		}

		muSquared.mul(&muCache[k][k-1], &muCache[k][k-1])
		threshold.sub(&delta, &muSquared)
		rhs.mul(&threshold, &B[k-1])

		if B[k].cmp(&rhs) >= 0 {
			k++
		} else {
			basis[k], basis[k-1] = basis[k-1], basis[k]
			updateGramSchmidtFrom(k - 1)

			// Check both swapped rows for early termination
			if checkRow(k - 1) {
				return k - 1
			}
			if checkRow(k) {
				return k
			}

			if k > 1 {
				k--
			}
		}
	}

	return -1 // No early termination, full reduction completed
}

// lllReduce performs in-place LLL reduction on an m×n basis matrix (m rows, n columns).
// Uses lazy rational arithmetic (no automatic GCD) for performance.
// Delta is fixed at 99/100 = 0.99 for stronger reduction.
// For non-square matrices (m > n), this finds a reduced basis for the lattice
// generated by the row vectors in R^n.
func lllReduce(basis [][]big.Int, m int) {
	// Delegate to lllReduceWithBound with nil bound (no early termination)
	lllReduceWithBound(basis, m, nil, nil)
}
