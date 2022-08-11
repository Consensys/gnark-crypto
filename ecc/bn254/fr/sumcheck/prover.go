package sumcheck

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
)

// This does not make use of parallelism and represents polynomials as lists of coefficients
// It is currently geared towards arithmetic hashes. Once we have a more unified hash function interface, this can be generified.

// Claims to a multi-sumcheck statement. i.e. one of the form ∑_{0≤i<2ⁿ} fⱼ(i) = cⱼ for 1 ≤ j ≤ m.
type Claims interface {
	Combine(fr.Element) (SubClaim, polynomial.Polynomial) // Combine into the 0ᵗʰ sumcheck subclaim. Create g := ∑_{1≤j≤m} rʲ⁻¹fⱼ for which now we seek to prove ∑_{0≤i<2ⁿ} g(i) = c := ∑_{1≤j≤m} rʲ⁻¹cⱼ. Return a SubClaim for the first step and g₁.
	VarsNum() int                                         //number of variables
	ClaimsNum() int                                       //number of claims
	//Serialize() []fr.Element                     //An expression of the claims, to be incorporated into Fiat-Shamir hashes
}

// Proof of a multi-sumcheck statement.
type Proof []polynomial.Polynomial

/*type Proof struct {
	evaluations []polynomial.Polynomial // At iteration 1 ≤ i ≤
}*/

// SubClaim is a claim of the form gⱼ = ∑_{0≤i<2ⁿ⁻ʲ} g(r₁, r₂, ..., rⱼ₋₁, Xⱼ, i...)
type SubClaim interface {
	Next(fr.Element) polynomial.Polynomial // Return the evaluations gⱼ(k) for 1 ≤ k < degⱼ(g).
	//Update the subclaim to gⱼ₊₁ for the input value as rⱼ
	// Verifier should know and check len(Next()) so it's not incorporated in the fiat shamir hash. TODO Is this true?

	//Evaluations() []fr.Element // Evaluations gⱼ(k) for 0 ≤ k < deg(gⱼ)
	//Next(fr.Element) SubClaim  // Compute gⱼ₊₁ for the given value as rⱼ
}

//TODO: Use Hash To Field functions

// Prove create a non-interactive sumcheck proof
// transcript must have a hash function specified and seeded with a
func Prove(claims Claims, transcript ArithmeticTranscript, challengeSeed []byte /*challengeHash hash.Hash*/) Proof {
	//transcript := newSumcheckTranscript(claims, challengeHash, challengeSeed)

	// Are claims supposed to already be incorporated in the challengeSeed? Given the business with the commitments

	//challenge, err := transcript.ComputeChallenge(linearCombinationCoeffsId)
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
