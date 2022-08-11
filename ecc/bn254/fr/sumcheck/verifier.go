package sumcheck

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
)

type LazyClaims interface {
	CombinedSum(combinationCoeffs fr.Element) fr.Element
	CombinedEval(combinationCoeffs fr.Element, r []fr.Element) fr.Element
	Degree(i int) int //Degree of the total claim in the i'th variable=
	ClaimsNum() int
	VarsNum() int
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
			fmt.Println("Malformed proof")
			return false
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
