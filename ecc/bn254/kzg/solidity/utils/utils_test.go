package utils

import (
	"bytes"
	"crypto/sha256"
	"os"
	"reflect"
	"testing"

	kzg_bn254 "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
)

func checkError(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestSerialisation(t *testing.T) {

	t.Skip("generate polys, etc")

	// serialialise, deserialise
	fproof, err := os.Open("../data/singleProof")
	checkError(err, t)
	var proof, testProof OpeningProofSolidity
	_, err = proof.ReadFrom(fproof)
	checkError(err, t)
	var buffer bytes.Buffer
	_, err = proof.WriteTo(&buffer)
	checkError(err, t)
	_, err = testProof.ReadFrom(&buffer)
	checkError(err, t)
	if !reflect.DeepEqual(testProof, proof) {
		t.Fatal("error serialising proof")
	}

	point, commitment, kzgProof := ConvertToKzgFormat(proof)
	fsrs, err := os.Open("../data/vk")
	checkError(err, t)
	defer fsrs.Close()
	var srs kzg_bn254.VerifyingKey
	srs.ReadFrom(fsrs)

	err = kzg_bn254.Verify(&commitment, &kzgProof, point, srs)
	checkError(err, t)
}

func TestSerialisationBatchOpening(t *testing.T) {

	t.Skip("generate polys, etc")

	// serialialise, deserialise
	fproof, err := os.Open("../data/batchopeningproof")
	checkError(err, t)
	var proof, testProof BatchOpeningProofSolidity
	_, err = proof.ReadFrom(fproof)
	checkError(err, t)
	var buffer bytes.Buffer
	_, err = proof.WriteTo(&buffer)
	checkError(err, t)
	_, err = testProof.ReadFrom(&buffer)
	checkError(err, t)
	if !reflect.DeepEqual(testProof, proof) {
		t.Fatal("error serialising proof")
	}

	point, commitments, kzgProof := ConvertBatchOpeningToKzgFormat(proof)
	fsrs, err := os.Open("../data/vk")
	checkError(err, t)
	defer fsrs.Close()
	var srs kzg_bn254.VerifyingKey
	srs.ReadFrom(fsrs)

	err = kzg_bn254.BatchVerifySinglePoint(commitments, &kzgProof, point, sha256.New(), srs)
	checkError(err, t)

}
