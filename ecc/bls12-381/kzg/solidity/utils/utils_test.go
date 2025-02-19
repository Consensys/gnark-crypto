package utils

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"reflect"
	"testing"

	kzg_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/kzg"
)

func checkError(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestMarshalSolidity(t *testing.T) {

	fproof, err := os.Open("../data/batchopeningproof")
	checkError(err, t)
	var proof BatchOpeningProofSolidity
	_, err = proof.ReadFrom(fproof)
	checkError(err, t)

	solidityProof, err := proof.MarshalSolidity()
	checkError(err, t)
	for i := 0; i < len(solidityProof); i++ {
		tmp := int(solidityProof[i])
		if tmp < 16 {
			fmt.Printf("0%x", solidityProof[i])
		} else {
			fmt.Printf("%x", solidityProof[i])
		}
	}

}
func TestSerialisation(t *testing.T) {

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
	var srs kzg_bls12381.VerifyingKey
	srs.ReadFrom(fsrs)

	err = kzg_bls12381.Verify(&commitment, &kzgProof, point, srs)
	checkError(err, t)
}

func TestSerialisationBatchOpening(t *testing.T) {

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
	var srs kzg_bls12381.VerifyingKey
	srs.ReadFrom(fsrs)

	err = kzg_bls12381.BatchVerifySinglePoint(commitments, &kzgProof, point, sha256.New(), srs)
	checkError(err, t)

}
