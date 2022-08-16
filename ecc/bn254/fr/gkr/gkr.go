package gkr

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/sumcheck"
)

// This implements the GKR protocol more or less as described in the "Libra" paper

// GateWiring stores the input and output indexes of wires
type GateWiring struct {
	out      int
	inL, inR int // Fan-in = 2
}

// CircuitLayerGate currently has fan-in 2
type CircuitLayerGate struct {
	Degree int // Degree as a polynomial
	Eval   func(l, r fr.Element) fr.Element
	Wiring []GateWiring // Because it is assumed to be sparse, the Wiring is not described as a polynomial
}

type CircuitLayer struct {
	Values polynomial.MultiLin // Consisting of 2ⁿ values
	Gates  []CircuitLayerGate  // For each kind of computation
	Input  *CircuitLayer       // Previous layer of the circuit. nil if the current layer is itself the input
}

type Claim struct {
	OutputValues       polynomial.MultiLin //The claimed OutputValues of the circuit
	InputValues        polynomial.MultiLin //The agreed-upon InputValues of the circuit
	CircuitOutputLayer CircuitLayer        //The final layer of the circuit, producing the OutputValues
}

//LazyClaim is for the verifier
type LazyClaim struct {
}

type Proof sumcheck.Proof

type gkrOutputSumcheckClaim struct{}
type gkrOutputSumcheckSubclaim struct{
	layer *CircuitLayer
}

func (c *gkrOutputSumcheckSubclaim) Next(element fr.Element) polynomial.Polynomial {
	panic("implement me")
}

func (c gkrOutputSumcheckClaim) Combine(a fr.Element) (sumcheck.SubClaim, polynomial.Polynomial) {
	//panic("implement me")
	//TODO Enable multi-sumcheck
	var subclaim gkrOutputSumcheckSubclaim

	subclaim.layer = 

	return &subclaim, polynomial.Polynomial{}
}

func (c gkrOutputSumcheckClaim) VarsNum() int {
	panic("implement me")
}

func (c gkrOutputSumcheckClaim) ClaimsNum() int {
	panic("implement me")
}

func (c gkrOutputSumcheckClaim) ProveFinalEval(r []fr.Element) interface{} {
	panic("implement me")
}

type gkrMiddleSumcheckClaim struct{}
type gkrInputSumcheckClaim struct{}

// Prove converts the GKR claim into a number of sumcheck claims. Most of the heavy-lifting all delegated to sumchecks
func (c Claim) Prove(transcript sumcheck.ArithmeticTranscript, challengeSeed []byte) Proof {
	g0 := sumcheck.NextFromBytes(transcript, challengeSeed, c.OutputValues.NumVars())

	v0G := c.OutputValues.Evaluate(g0)
	s1 := c.CircuitOutputLayer.Input.Values.NumVars()
	// We must now prove the claim v0G = ∑_{x,y ∈ {0,1}^s1 } ∑_{G ∈ fan-in-2-gates} GWiring(g0, x, y) × G(V₁(x),V₁(y))

	//TODO Fold sparse wiring predicates with g0
	var foldedGateWirings []polynomial.MultiLin

	var outputClaim gkrOutputSumcheckClaim
	return Proof(sumcheck.Prove(outputClaim, transcript, challengeSeed))
}
