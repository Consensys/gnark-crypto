package sumcheck

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// This is an implementation of Fiat-Shamir optimized for in-circuit verifiers.
// i.e. the hashes used operate on and return field elements.

type ArithmeticTranscript interface {
	NextFromElements(_ []fr.Element) fr.Element    // NextFromElements does not directly incorporate the number of elements in the hash. The size must be fixed and checked by the verifier.
	NextFromBytes(_ []byte, size int) []fr.Element // NextFromElements does not directly incorporate the size of input in the hash. The size must be fixed and checked by the verifier.
}

/*const linearCombinationCoeffsId = "linear-combination-coeffs"
const polynomialEvaluationId = "polynomial-evaluation"

func newSumcheckTranscript(claims Claims, h hash.Hash, seed []byte) fiatshamir.Transcript {
	challengesId := make([]string, claims.ClaimsNum()+claims.VarsNum())
	for i := 0; i < claims.ClaimsNum(); i++ {
		challengesId[i] = linearCombinationCoeffsId + strconv.Itoa(i)
	}
	for i := 0; i < claims.VarsNum(); i++ {
		challengesId[i+claims.ClaimsNum()] = polynomialEvaluationId + strconv.Itoa(i)
	}
	transcript := fiatshamir.NewTranscript(h, challengesId...)
	if e := transcript.Bind(linearCombinationCoeffsId+"0", seed); e != nil {
		panic("Error in setting up sumcheck challenges: " + e.Error())
	}
	return transcript
}
*/
