package sumcheck

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
)

// This does not make use of parallelism and represents polynomials as lists of coefficients
// It is currently geared towards arithmetic hashes. Once we have a more unified hash function interface, this can be generified.

// Claims to a multi-sumcheck statement. i.e. one of the form ∑_{0≤i<2ⁿ} fⱼ(i) = cⱼ for 1 ≤ j ≤ m.
// Later evolving into a claim of the form gⱼ = ∑_{0≤i<2ⁿ⁻ʲ} g(r₁, r₂, ..., rⱼ₋₁, Xⱼ, i...)
type Claims interface {
	Combine(a fr.Element) polynomial.Polynomial // Combine into the 0ᵗʰ sumcheck subclaim. Create g := ∑_{1≤j≤m} aʲ⁻¹fⱼ for which now we seek to prove ∑_{0≤i<2ⁿ} g(i) = c := ∑_{1≤j≤m} aʲ⁻¹cⱼ. Return g₁.
	Next(fr.Element) polynomial.Polynomial      // Return the evaluations gⱼ(k) for 1 ≤ k < degⱼ(g). Update the claim to gⱼ₊₁ for the input value as rⱼ
	VarsNum() int                               //number of variables
	ClaimsNum() int                             //number of claims
	ProveFinalEval(r []fr.Element) interface{}  //in case it is difficult for the verifier to compute g(r₁, ..., rₙ) on its own, the prover can provide the value and a proof
}

// LazyClaims is the Claims data structure on the verifier side. It is "lazy" in that it has to compute fewer things.
type LazyClaims interface {
	ClaimsNum() int                      // ClaimsNum = m
	VarsNum() int                        // VarsNum = n
	CombinedSum(a fr.Element) fr.Element // CombinedSum returns c = ∑_{1≤j≤m} aʲ⁻¹cⱼ
	Degree(i int) int                    //Degree of the total claim in the i'th variable
	VerifyFinalEval(r []fr.Element, combinationCoeff fr.Element, purportedValue fr.Element, proof interface{}) bool
}

// Proof of a multi-sumcheck statement.
type Proof struct {
	partialSumPolys []polynomial.Polynomial
	finalEvalProof  interface{} //in case it is difficult for the verifier to compute g(r₁, ..., rₙ) on its own, the prover can provide the value and a proof
}

// Prove create a non-interactive sumcheck proof
// transcript must have a hash function specified and seeded with a
func Prove(claims Claims, transcript ArithmeticTranscript, challengeSeed interface{}) Proof {
	// TODO: Are claims supposed to already be incorporated in the challengeSeed? Given the business with the commitments

	var combinationCoeff fr.Element
	if claims.ClaimsNum() >= 2 {
		combinationCoeff = NextChallenge(transcript, challengeSeed)
	}

	var proof Proof
	proof.partialSumPolys = make([]polynomial.Polynomial, claims.VarsNum())
	proof.partialSumPolys[0] = claims.Combine(combinationCoeff)

	for j := 1; j < len(proof.partialSumPolys); j++ {
		r := transcript.NextFromElements(proof.partialSumPolys[j-1])
		proof.partialSumPolys[j] = claims.Next(r)
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
		if len(proof.partialSumPolys[j]) != claims.Degree(j) {
			return false //Malformed proof
		}
		copy(gJ[1:], proof.partialSumPolys[j])
		gJ[0].Sub(&gJR, &proof.partialSumPolys[j][0]) // Requirement that gⱼ(0) + gⱼ(1) = gⱼ₋₁(r)
		// gJ is ready

		//Prepare for the next iteration
		r[j] = transcript.NextFromElements(proof.partialSumPolys[j])
		// This is an extremely inefficient way of interpolating. TODO: Interpolate without symbolically computing a polynomial
		gJCoeffs := polynomial.InterpolateOnRange(gJ[:(claims.Degree(j) + 1)])
		gJR = gJCoeffs.Eval(&r[j])
	}

	return claims.VerifyFinalEval(r, combinationCoeff, gJR, proof.finalEvalProof)
}
