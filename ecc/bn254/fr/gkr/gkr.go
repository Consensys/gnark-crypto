package gkr

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/sumcheck"
)

// The goal is to prove/verify evaluations of many instances of the same circuit

// Gate must be a low-degree polynomial
type Gate interface {
	Combine(...fr.Element) fr.Element
	NumInput() int
	Degree(i int)   int
}

type Wire struct {
	Gate   *Gate
	Inputs []*Wire
	NumOutputs int	// number of other wires using it as input
}

type CircuitLayer []Wire

// TODO: Constructor so that user doesn't have to give layers explicitly.
type Circuit []CircuitLayer

// WireAssignment is assignment of values to the same wire across many instances of the circuit
type WireAssignment map[*Wire]polynomial.MultiLin

type Proof struct {
}

// A claim about the value of a wire
type claim struct {
	in  []fr.Element
	out fr.Element
}

type eqTimesGateEvalSumcheckClaims struct {
	claimedEvaluations []fr.Element	// y in the paper
	combinedEvaluation fr.Element
	combinationCoefficient fr.Element
	evaluationPoints [][]fr.Element	// x in the paper
	inputPreprocessors []polynomial.MultiLin	// P_u in the paper
	eq []polynomial.MultiLin	// \tau_i eq(x_i, -)
	gate *Gate	// R_v in the paper
	//assignment *WireAssignment
}

func (c *eqTimesGateEvalSumcheckClaims) Combine(a fr.Element) polynomial.Polynomial {

	// TODO: Compute evals in here instead of getting them as input?
	c.combinationCoefficient = a
	evaluationsAsCoefficients := polynomial.Polynomial(c.claimedEvaluations)
	c.combinedEvaluation = evaluationsAsCoefficients.Eval(&a)

	// initialize the eq tables
	c.eq = make([]polynomial.MultiLin, c.ClaimsNum())

	var aI fr.Element	// this is being recomputed, already computed in the combinedEvaluation step. TODO: Big deal?
	aI.SetOne()
	for k := 0; k < len(c.eq); k++ {
		// define eq_k = a^k eq(x_k1, ..., x_kn, *, ..., *) where x_ki are the evaluation points
		c.eq[k] = polynomial.Make(1 << len(c.evaluationPoints[k]))
		c.eq[k][0] = aI
		c.eq[k].Eq(c.evaluationPoints[k])

		if k+ 1 < len(c.eq) {
			aI.Mul(&aI, &a)
		}
	}

	return c.computeGJ()
}

// computeGJ: gⱼ = ∑_{0≤i<2ⁿ⁻ʲ} g(r₁, r₂, ..., rⱼ₋₁, Xⱼ, i...) = ∑_{0≤i<2ⁿ⁻ʲ} (\sum eq_k (r_1, ..., X_j, i...) ) Gate( P_u0(r_1, ..., X_j, i...), ... )
func (c *eqTimesGateEvalSumcheckClaims) computeGJ() polynomial.Polynomial {
	eqSum := make([]fr.Element, len(c.eq))	// TODO: Use pool
	eqStep := make([]fr.Element, len(c.eq))

	for k := 0; k < len(c.eq); k++ {
		eq := c.eq[k]
		eqSum[k] = eq[len(eq)/2:].Sum()	// initially to hold \sum eq_k(r_1, ..., 1, i...)
		eqStep[k] = eq[:len(eq)/2].Sum()
		eqStep[k].Sub(&eqStep[k], &eqSum[k])	//holds \sum eq_k(r_1, ..., 1, i..) - \sum eq_k(r_1, ..., 0, i...)
	}
}

func (c *eqTimesGateEvalSumcheckClaims) Next(element fr.Element) polynomial.Polynomial {
	panic("implement me")
}

func (c *eqTimesGateEvalSumcheckClaims) VarsNum() int {
	return len(c.evaluationPoints[0])
}

func (c *eqTimesGateEvalSumcheckClaims) ClaimsNum() int {
	return len(c.claimedEvaluations)
}

func (c eqTimesGateEvalSumcheckClaims) ProveFinalEval(r []fr.Element) interface{} {
	panic("implement me")
}

// Prove consistency of the claimed assignment
func Prove(c Circuit, assignment WireAssignment, transcript sumcheck.ArithmeticTranscript, firstChallenge []fr.Element) Proof {
	var claims map[*Wire]eqTimesGateEvalSumcheckClaims

	// firstChallenge called rho in the paper

	outLayer := c[0]
	inLayer := c[len(c)-1]
	for i := 0; i < len(outLayer); i++ {
		//claims[&outLayer[i]] = []claim{{in: firstChallenge}} //TODO: Just directly prove it?
		claim := eqTimesGateEvalSumcheckClaims{
			wire: &outLayer[i],

		}
		sumcheck.Prove(, transcript, nil)
	}

	for layerI := 1; layerI+1 < len(c); layerI++ {
		layer := c[layerI]
		for wireI := 0; wireI < len(layer); wireI++ {
			wire := &layer[wireI]

		}

	}
}
