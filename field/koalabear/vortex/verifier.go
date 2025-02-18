package vortex

import "github.com/consensys/gnark-crypto/field/koalabear"

// VerifierInput collects all the inputs to the verifier
// of a vortex opening.
type VerifierInput struct {

	// ClaimedValue value of the leaf. This field is exported
	ClaimedValues [][4]koalabear.Element

	// EvaluationPoint is the evaluation point
	EvaluationPoint [4]koalabear.Element

	// SelectedColumns are the positions of the columns sampled by
	// the verifier
	SelectedColumns []int

	// Alpha is the coin sampled by the verifier to compute the
	// linear combination UAlpha
	Alpha [4]koalabear.Element

	// Proof is the opening proof
	Proof Proof
}
