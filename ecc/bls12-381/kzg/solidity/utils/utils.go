package utils

import (
	"bytes"
	"io"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	kzg_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/kzg"
)

// OpeningProofSolidity kzg single proof opening
type OpeningProofSolidity struct {
	Point        fr.Element
	ClaimedValue fr.Element
	Commitment   kzg_bls12381.Digest
	Proof        bls12381.G1Affine
}

// ConvertToSolidityFormat converts an opening proof to the format described in EIP4844
func ConvertToSolidityFormat(x fr.Element, pCommitted kzg_bls12381.Digest, proof kzg_bls12381.OpeningProof) OpeningProofSolidity {
	var res OpeningProofSolidity
	res.Point.Set(&x)
	res.ClaimedValue.Set(&proof.ClaimedValue)
	res.Commitment.Set(&pCommitted)
	res.Proof.Set(&proof.H)
	return res
}

// ConvertToKzgFormat converts an opening proof to the format described in EIP4844
func ConvertToKzgFormat(proof OpeningProofSolidity) (fr.Element, kzg_bls12381.Digest, kzg_bls12381.OpeningProof) {
	var resProof kzg_bls12381.OpeningProof
	var commitment kzg_bls12381.Digest
	var point fr.Element
	resProof.ClaimedValue.Set(&proof.ClaimedValue)
	resProof.H.Set(&proof.Proof)
	point.Set(&proof.Point)
	commitment.Set(&proof.Commitment)
	return point, commitment, resProof
}

// WriteTo MarshalSingleProofSolidity serialises a single KZG proof, using EIP4844 format
// [ z || y || commitment || proof ]
func (proof *OpeningProofSolidity) WriteTo(w io.Writer) (int64, error) {

	// var buffer [128]byte

	enc := bls12381.NewEncoder(w, bls12381.RawEncoding())
	toEncode := []interface{}{
		&proof.Point,
		&proof.ClaimedValue,
		&proof.Commitment,
		&proof.Proof,
	}
	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	// encode the digests, using EIP2537 format, that is

	return enc.BytesWritten(), nil
}

// ReadFrom UnmarshalSingleProofSolidity de-serialises a single KZG proof, using EIP4844 format
// [ z || y || commitment || proof ]
func (proof *OpeningProofSolidity) ReadFrom(r io.Reader) (int64, error) {
	dec := bls12381.NewDecoder(r)
	toDecode := []interface{}{
		&proof.Point,
		&proof.ClaimedValue,
		&proof.Commitment,
		&proof.Proof,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}
	return dec.BytesRead(), nil
}

// BatchOpeningProofSolidity kzg proof opening at the same point
type BatchOpeningProofSolidity struct {
	Point         fr.Element
	Digests       []kzg_bls12381.Digest
	ClaimedValues []fr.Element
	Proof         bls12381.G1Affine
}

// ConvertBatchOpeningToSolidityFormat creates a BatchOpeningProofSolidity from a list of commitments and a batch opening proof
func ConvertBatchOpeningToSolidityFormat(point fr.Element, com []kzg_bls12381.Digest, proof kzg_bls12381.BatchOpeningProof) BatchOpeningProofSolidity {
	var res BatchOpeningProofSolidity
	res.Point.Set(&point)
	res.Proof.Set(&proof.H)
	res.Digests = make([]kzg_bls12381.Digest, len(com))
	res.ClaimedValues = make([]fr.Element, len(com))
	copy(res.Digests, com)
	copy(res.ClaimedValues, proof.ClaimedValues)
	return res
}

// ConvertBatchOpeningToKzgFormat creates a BatchOpeningProofSolidity from a list of commitments and a batch opening proof
func ConvertBatchOpeningToKzgFormat(proof BatchOpeningProofSolidity) (fr.Element, []kzg_bls12381.Digest, kzg_bls12381.BatchOpeningProof) {

	var point fr.Element
	point.Set(&proof.Point)

	digests := make([]kzg_bls12381.Digest, len(proof.ClaimedValues))
	copy(digests, proof.Digests)

	var kzgProof kzg_bls12381.BatchOpeningProof
	kzgProof.ClaimedValues = make([]fr.Element, len(proof.ClaimedValues))
	copy(kzgProof.ClaimedValues, proof.ClaimedValues)
	kzgProof.H.Set(&proof.Proof)

	return point, digests, kzgProof
}

// MarshalSolidity raw serialisation to solidity format
func (kzgProof *BatchOpeningProofSolidity) MarshalSolidity() ([]byte, error) {

	var buffer bytes.Buffer

	var padding [16]byte

	// point
	bb := kzgProof.Point.Bytes()
	_, err := buffer.Write(bb[:])
	if err != nil {
		return nil, err
	}

	// digests, on 128 bytes, [0x00, .., 0x00,     x      , 0x00, .., 0x00,     y  ]
	//												<- 16 bytes --><-48 bytes --><- 16 bytes --><-48 bytes -->
	for i := 0; i < len(kzgProof.Digests); i++ {

		// x coordinate
		_, err = buffer.Write(padding[:])
		if err != nil {
			return nil, err
		}
		coordBytes := kzgProof.Digests[i].X.Bytes()
		_, err = buffer.Write(coordBytes[:])
		if err != nil {
			return nil, err
		}

		// y coordinate
		_, err = buffer.Write(padding[:])
		if err != nil {
			return nil, err
		}
		coordBytes = kzgProof.Digests[i].Y.Bytes()
		_, err = buffer.Write(coordBytes[:])
		if err != nil {
			return nil, err
		}
	}

	// claimed values
	for i := 0; i < len(kzgProof.ClaimedValues); i++ {
		bb = kzgProof.ClaimedValues[i].Bytes()
		_, err = buffer.Write(bb[:])
		if err != nil {
			return nil, err
		}
	}

	// quotient
	// x coordinate
	_, err = buffer.Write(padding[:])
	if err != nil {
		return nil, err
	}
	coordBytes := kzgProof.Proof.X.Bytes()
	_, err = buffer.Write(coordBytes[:])
	if err != nil {
		return nil, err
	}

	// y coordinate
	_, err = buffer.Write(padding[:])
	if err != nil {
		return nil, err
	}
	coordBytes = kzgProof.Proof.Y.Bytes()
	_, err = buffer.Write(coordBytes[:])
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil

}

// WriteTo serialises a batchopening proof, following this format:
// [ point || commitments || claimed values || proof ]
func (kzgProof *BatchOpeningProofSolidity) WriteTo(w io.Writer) (int64, error) {

	enc := bls12381.NewEncoder(w, bls12381.RawEncoding())
	toEncode := []interface{}{
		&kzgProof.Point,
		&kzgProof.Digests,
		kzgProof.ClaimedValues,
		&kzgProof.Proof,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}
	return enc.BytesWritten(), nil

}

// ReadFrom Unmarshal a batchopening proof, following this format:
// [ point || digests || claimed values || proof ]
func (kzgProof *BatchOpeningProofSolidity) ReadFrom(r io.Reader) (int64, error) {

	dec := bls12381.NewDecoder(r)
	toDecode := []interface{}{
		&kzgProof.Point,
		&kzgProof.Digests,
		&kzgProof.ClaimedValues,
		&kzgProof.Proof,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	return dec.BytesRead(), nil
}
