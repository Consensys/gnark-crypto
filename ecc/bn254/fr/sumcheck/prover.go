package sumcheck

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"hash"
	"strconv"
)

// This does not make use of parallelism and represents polynomials as lists of coefficients

// Proof of a multi-sumcheck statement. i.e. one of the form ∑_{0≤i<n} fⱼ(i) = cⱼ for 1 ≤ j ≤ m.
type Proof struct {
	evaluations []polynomial.Polynomial // At iteration 1 ≤ i ≤

}

// what is this data structure?

type LowDegMultivarPolynomial interface {
	HypercubeSum() fr.Element
	Fold() LowDegMultivarPolynomial //TODO: Or should it be in-place?
}

type Claims interface {
	Combine([]fr.Element) Claim
	Serialize() []byte
	VarsNum() int      //number of variables
	SubClaimsNum() int //number of subclaims
}

type Claim struct { //TODO: F doesn't have to be multilin
	F            LowDegMultivarPolynomial
	HypercubeSum fr.Element //@AlexandreBelling: Does Sum have to be provided or should Prove compute it?
}

const linearCombinationCoeffsId = "linear-combination-coeffs"
const polynomialEvaluationId = "polynomial-evaluation"

func newSumcheckTranscript(claims Claims, h hash.Hash, seed []byte) fiatshamir.Transcript {
	challengesId := make([]string, claims.SubClaimsNum()+claims.VarsNum())
	for i := 0; i < claims.SubClaimsNum(); i++ {
		challengesId[i] = linearCombinationCoeffsId + strconv.Itoa(i)
	}
	for i := 0; i < claims.VarsNum(); i++ {
		challengesId[i+claims.SubClaimsNum()] = polynomialEvaluationId + strconv.Itoa(i)
	}
	transcript := fiatshamir.NewTranscript(h, challengesId...)
	if e := transcript.Bind(linearCombinationCoeffsId+"0", seed); e != nil {
		panic("Error in setting up sumcheck challenges: " + e.Error())
	}
	return transcript
}

//TODO: Use Hash To Field functions

// Prove create a non-interactive sumcheck proof
// transcript must have a hash function specified and seeded with a
func Prove(claims Claims, challengeHash hash.Hash, challengeSeed []byte) Proof {
	transcript := newSumcheckTranscript(claims, challengeHash, challengeSeed)

	challenge, err := transcript.ComputeChallenge(linearCombinationCoeffsId)

}
