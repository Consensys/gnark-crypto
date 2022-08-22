package gkr

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/sumcheck"
)

// The goal is to prove/verify evaluations of many instances of the same circuit

// Gate must be a low-degree polynomial
type Gate interface {
	Evaluate(...fr.Element) fr.Element
	NumInput() int
	Degree() int
	//Degree(i int)   int
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
	inputPreprocessors []polynomial.MultiLin	// P_u in the paper, so that we don't need to pass along all the circuit's evaluations
	eq polynomial.MultiLin	// \sum_i \tau_i eq(x_i, -)
	gate *Gate	// R_v in the paper
	//assignment *WireAssignment
}

func (c *eqTimesGateEvalSumcheckClaims) Combine(a fr.Element) polynomial.Polynomial {

	// TODO: Compute evals in here instead of getting them as input?
	c.combinationCoefficient = a
	evaluationsAsCoefficients := polynomial.Polynomial(c.claimedEvaluations)
	c.combinedEvaluation = evaluationsAsCoefficients.Eval(&a)

	eqLength := 1 << len(c.evaluationPoints[0])
	// initialize the eq tables
	c.eq = polynomial.Make(eqLength)	//TODO: MakeLarge here?
	eqAsPoly := polynomial.Polynomial(c.eq)
	eqAsPoly.SetZero()

	newEq := polynomial.MultiLin(polynomial.Make(eqLength))	//TODO: MakeLarge here?
	newEq[0].SetOne()

	aI := newEq[0]
	for k := 0; k < len(c.eq); k++ {	//TODO: parallelizable?
		// define eq_k = a^k eq(x_k1, ..., x_kn, *, ..., *) where x_ki are the evaluation points
		newEq.Eq(c.evaluationPoints[k])

		if !newEq[0].Equal(&aI) {
			panic("Incorrect assumption: Eq modifies newEq[0]")
		}

		eqAsPoly.Add(eqAsPoly, polynomial.Polynomial(newEq))

		if k+ 1 < len(c.eq) {
			newEq[0].Mul(&newEq[0], &a)	//TODO: Test this. newEq[0] maybe not preserving value?
			aI = newEq[0]
		}
	}

	// from this point on the claim is a rather simple one: g = E(h) \times R_v (P_u0(h), ...) where E and the P_ui are multilinear and R_v is of low-degree

	return c.computeGJ()
}

// computeSumAndStep returns sum = \sum_i m(1, i...) and step = \sum_i m(1, i...) - m(0, i...)
func computeSumAndStep(m polynomial.MultiLin) (sum fr.Element, step fr.Element) {
	sum = m[len(m)/2:].Sum()
	step = m[:len(m)/2].Sum()
	step.Sub(&sum, &step)
	return
}

// computeGJ: gⱼ = ∑_{0≤i<2ⁿ⁻ʲ} g(r₁, r₂, ..., rⱼ₋₁, Xⱼ, i...) = ∑_{0≤i<2ⁿ⁻ʲ} E(r_1, ..., X_j, i...) R_v( P_u0(r_1, ..., X_j, i...), ... ) where  E = \sum eq_k
// the polynomial is represented by the evaluations g_j(1), g_j(2), ..., g_j(deg(g_j)).
func (c *eqTimesGateEvalSumcheckClaims) computeGJ() polynomial.Polynomial {
	degGJ := 1 + c.gate.Degree()	// guaranteed to be no smaller than the actual deg(g_j)

	// Let f \in {E} \cup {P_ui}. It is linear in X_j, so f(n) = n\times(f(1) - f(0)) + f(0), and f(0), f(1) are easily computed from the bookkeeping tables
	ESum, EStep := computeSumAndStep(c.eq)

	puSum := polynomial.Polynomial(polynomial.Make(len(c.inputPreprocessors)))
	puStep := polynomial.Make(len(c.inputPreprocessors))

	for i := 0; i < len(puSum); i++ {
		puSum[i], puStep[i] = computeSumAndStep(c.inputPreprocessors[i])
	}

	evals := make([]fr.Element, degGJ)
	for i := 1; i < degGJ; i++ {
		evals[i] = c.gate.Evaluate(puSum...)
		evals[i].Mul(&evals[i], &ESum)

		if i + 1 < degGJ {
			ESum.Add(&ESum, &EStep)
			puSum.Add(puSum, puStep)
		}
	}
	return evals
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
