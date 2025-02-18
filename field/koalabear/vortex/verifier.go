package vortex

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

// VerifierInput collects all the inputs to the verifier
// of a vortex opening.
type VerifierInput struct {

	// MerkleRoot is the commitment to the input matrix
	MerkleRoot Hash

	// ClaimedValue value of the leaf. This field is exported
	ClaimedValues []fext.E4

	// EvaluationPoint is the evaluation point
	EvaluationPoint fext.E4

	// SelectedColumns are the positions of the columns sampled by
	// the verifier
	SelectedColumns []int

	// Alpha is the coin sampled by the verifier to compute the
	// linear combination UAlpha
	Alpha fext.E4

	// Proof is the opening proof
	Proof *Proof
}

// Verify implements the verification algorithm for a Vortex opening proof.
func (p *Params) Verify(input VerifierInput) error {

	var (
		proof = input.Proof
		root  = input.MerkleRoot
	)

	// This checks the consistency between uAlpha and the claimed value
	var (
		uAlphaAtX     = evalHornerExt(input.Proof.UAlpha, input.EvaluationPoint)
		claimsAtAlpha = evalHornerExt(input.ClaimedValues, input.Alpha)
	)

	if uAlphaAtX != claimsAtAlpha {
		return errors.New("invalid proof: ualpha and the claim do not match")
	}

	// This checks the reed-solomon member ship of UAlpha
	if err := p.IsReedSolomonCodewords(proof.UAlpha); err != nil {
		return fmt.Errorf("invalid proof: uAlpha is not a reed-solomon codeword: %w", err)
	}

	// This checks the consistency between the proof and the selected columns
	// to the input matrix.
	for i, c := range input.SelectedColumns {

		sisHash := make([]koalabear.Element, p.Key.Degree)
		if err := p.Key.Hash(proof.OpenedColumns[i], sisHash); err != nil {
			return fmt.Errorf("invalid proof: could not hash the column: %w", err)
		}

		leaf := HashPoseidon2(sisHash)

		if err := proof.MerkleProofOpenedColumns[i].Verify(c, leaf, root); err != nil {
			return fmt.Errorf("invalid proof: merkle proof verification failed: %w", err)
		}
	}

	return nil

}

// evalHorner evaluates p at point x using the Horner algorithm
// where p is a polynomial of degree len(p) - 1
func evalHornerExt(p []fext.E4, x fext.E4) fext.E4 {
	res := fext.E4{}
	for i := len(p) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.Add(&res, &p[i])
	}
	return res
}
