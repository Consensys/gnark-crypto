package sumcheck

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
)

// This does not make use of parallelism and represents polynomials as lists of coefficients
// It is currently geared towards arithmetic hashes. Once we have a more unified hash function interface, this can be generified.

// Claims to a multi-sumcheck statement. i.e. one of the form ∑_{0≤i<2ⁿ} fⱼ(i) = cⱼ for 1 ≤ j ≤ m.
type Claims interface {
	Combine(a fr.Element) (SubClaim, polynomial.Polynomial) // Combine into the 0ᵗʰ sumcheck subclaim. Create g := ∑_{1≤j≤m} aʲ⁻¹fⱼ for which now we seek to prove ∑_{0≤i<2ⁿ} g(i) = c := ∑_{1≤j≤m} aʲ⁻¹cⱼ. Return a SubClaim for the first step and g₁.
	VarsNum() int                                           //number of variables
	ClaimsNum() int                                         //number of claims
}

// SubClaim is a claim of the form gⱼ = ∑_{0≤i<2ⁿ⁻ʲ} g(r₁, r₂, ..., rⱼ₋₁, Xⱼ, i...)
type SubClaim interface {
	Next(fr.Element) polynomial.Polynomial // Return the evaluations gⱼ(k) for 1 ≤ k < degⱼ(g).
	//Update the subclaim to gⱼ₊₁ for the input value as rⱼ
}

// LazyClaims is the Claims data structure on the verifier side. It is "lazy" in that it has to compute fewer things.
type LazyClaims interface {
	ClaimsNum() int                                           // ClaimsNum = m
	VarsNum() int                                             // VarsNum = n
	CombinedSum(a fr.Element) fr.Element                      // CombinedSum returns c = ∑_{1≤j≤m} aʲ⁻¹cⱼ
	CombinedEval(coeff fr.Element, r []fr.Element) fr.Element // CombinedEval returns returns g(r₁, ..., rₙ) = ∑_{1≤j≤m} aʲ⁻¹fⱼ(r₁, ..., rₙ)
	Degree(i int) int                                         //Degree of the total claim in the i'th variable
}

// Proof of a multi-sumcheck statement.
type Proof []polynomial.Polynomial

// Prove create a non-interactive sumcheck proof
// transcript must have a hash function specified and seeded with a
func Prove(claims Claims, transcript ArithmeticTranscript, challengeSeed []byte) Proof {
	// TODO: Are claims supposed to already be incorporated in the challengeSeed? Given the business with the commitments

	var combinationCoeff fr.Element
	if claims.ClaimsNum() >= 2 {
		combinationCoeff = transcript.NextFromBytes(challengeSeed)
	}

	var claim SubClaim
	proof := make(Proof, claims.VarsNum())
	claim, proof[0] = claims.Combine(combinationCoeff)

	for j := 1; j < len(proof); j++ {
		r := transcript.NextFromElements(proof[j-1])
		proof[j] = claim.Next(r)
	}

	return proof
}

func Verify(claims LazyClaims, proof Proof, transcript ArithmeticTranscript, challengeSeed []byte) bool {
	var combinationCoeff fr.Element

	if claims.ClaimsNum() >= 2 {
		combinationCoeff = transcript.NextFromBytes(challengeSeed)
	}

	r := make([]fr.Element, claims.VarsNum())

	// Just so that there is enough room for gJ to be reused
	maxDegree := claims.Degree(0)
	for j := 1; j < claims.VarsNum(); j++ {
		if d := claims.Degree(j); d > maxDegree {
			maxDegree = d
		}
	}
	gJ := make(polynomial.Polynomial, maxDegree+1) //At the end of iteration j, gJ = ∑_{i < 2ⁿ⁻ʲ⁻¹} g(X₁, ..., Xⱼ₊₁, i...)		NOTE: n is shorthand for claims.VarsNum()
	gJR := claims.CombinedSum(combinationCoeff)    // At the beginning of iteration j, gJR = ∑_{i < 2ⁿ⁻ʲ} g(r₁, ..., rⱼ, i...)

	for j := 0; j < claims.VarsNum(); j++ {
		if len(proof[j]) != claims.Degree(j) {
			return false //Malformed proof
		}
		copy(gJ[1:], proof[j])
		gJ[0].Sub(&gJR, &proof[j][0]) // Requirement that gⱼ(0) + gⱼ(1) = gⱼ₋₁(r)
		// gJ is ready

		//Prepare for the next iteration
		r[j] = transcript.NextFromElements(proof[j])
		// This is an extremely inefficient way of interpolating. TODO: Interpolate without symbolically computing a polynomial
		gJCoeffs := polynomial.InterpolateOnRange(gJ[:(claims.Degree(j) + 1)])
		gJR = gJCoeffs.Eval(&r[j])
	}

	combinedEval := claims.CombinedEval(combinationCoeff, r)
	return combinedEval.Equal(&gJR)
}
