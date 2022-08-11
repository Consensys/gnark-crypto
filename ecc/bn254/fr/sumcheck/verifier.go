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
	gJ := make(polynomial.Polynomial, maxDegree+1)
	gJR := claims.CombinedSum(combinationCoeff)

	for j := 0; j < claims.VarsNum(); j++ {
		if len(proof[j]) != claims.Degree(j) {
			fmt.Println("Malformed proof")
			return false
		}
		copy(gJ[1:], proof[j])
		gJ[0].Sub(&gJR, &proof[j][0]) // Requirement that g_j(0) + g_j(1) = gⱼ₋₁(r)

		//Prepare for the next iteration
		r := transcript.NextFromElements(proof[j])
		// This is an extremely inefficient way of interpolating. TODO: Interpolate without symbolically computing a polynomial
		gJCoeffs := polynomial.InterpolateOnRange(gJ[:claims.Degree(j)])
		gJR = gJCoeffs.Eval(&r)
	}

	combinedEval := claims.CombinedEval(combinationCoeff, r)
	return combinedEval.Equal(&gJR)
}
