package pedersen

// TODO: Better name

import (
	"crypto/rand"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"math/big"
)

// Key for proof and verification
type Key struct {
	g             bn254.G2Affine // TODO @tabaie: does this really have to be randomized?
	gRootSigmaNeg bn254.G2Affine //gRootSigmaNeg = g^{-1/σ}
	basis         []*bn254.G1Affine
	basisExpSigma []bn254.G1Affine
}

func randomOnG2() (bn254.G2Affine, error) { // TODO: Add to G2.go?
	gBytes := make([]byte, fr.Bytes)
	if _, err := rand.Read(gBytes); err != nil {
		return bn254.G2Affine{}, err
	}
	return bn254.HashToG2(gBytes, []byte("random on g2"))
}

func Setup(basis []*bn254.G1Affine) (Key, error) {
	var (
		k   Key
		err error
	)

	if k.g, err = randomOnG2(); err != nil {
		return k, err
	}

	var modMinusOne big.Int
	modMinusOne.Sub(fr.Modulus(), big.NewInt(1))
	var sigma *big.Int
	if sigma, err = rand.Int(rand.Reader, &modMinusOne); err != nil {
		return k, err
	}
	sigma.Add(sigma, big.NewInt(1))

	var sigmaInvNeg big.Int
	sigmaInvNeg.ModInverse(sigma, fr.Modulus())
	sigmaInvNeg.Sub(fr.Modulus(), &sigmaInvNeg)
	k.gRootSigmaNeg.ScalarMultiplication(&k.g, &sigmaInvNeg)

	k.basisExpSigma = make([]bn254.G1Affine, len(basis))
	for i, gᵢ := range basis {
		k.basisExpSigma[i].ScalarMultiplication(gᵢ, sigma)
	}

	k.basis = basis
	return k, err
}

// TODO: If this takes too long in practice, edit MultiExp to accept pointers too
func ptrSliceToSlice[T any](ptrSlice []*T) []T {
	slice := make([]T, len(ptrSlice))
	for i, p := range ptrSlice {
		slice[i] = *p
	}
	return slice
}

func (k *Key) Commit(values []*fr.Element) (commitment bn254.G1Affine, knowledgeProof bn254.G1Affine, err error) {

	if len(values) != len(k.basis) {
		err = fmt.Errorf("unexpected number of values")
		return
	}

	valuesNoPtr := ptrSliceToSlice(values)
	config := ecc.MultiExpConfig{
		NbTasks:     1, // TODO Experiment
		ScalarsMont: true,
	}

	if _, err = commitment.MultiExp(ptrSliceToSlice(k.basis), valuesNoPtr, config); err != nil {
		return
	}

	_, err = knowledgeProof.MultiExp(k.basisExpSigma, valuesNoPtr, config)

	return
}

// VerifyKnowledgeProof checks if the proof of knowledge is valid
func (k *Key) VerifyKnowledgeProof(commitment bn254.G1Affine, knowledgeProof bn254.G1Affine) error {

	if !commitment.IsInSubGroup() || !knowledgeProof.IsInSubGroup() {
		return fmt.Errorf("subgroup check failed")
	}

	product, err := bn254.Pair([]bn254.G1Affine{commitment, knowledgeProof}, []bn254.G2Affine{k.g, k.gRootSigmaNeg})
	if err != nil {
		return err
	}
	if product.C0.B0.A0.IsOne() && product.C0.B0.A1.IsZero() && product.C0.B1.IsZero() && product.C1.IsZero() {
		return nil
	}
	return fmt.Errorf("proof rejected")
}
