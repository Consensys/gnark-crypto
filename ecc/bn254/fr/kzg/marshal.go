// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package kzg

import (
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"io"
)

// WriteTo writes binary encoding of the SRS
func (srs *SRS) WriteTo(w io.Writer) (int64, error) {
	// encode the SRS
	enc := bn254.NewEncoder(w)

	toEncode := []interface{}{
		&srs.G2[0],
		&srs.G2[1],
		srs.G1,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}

// ReadFrom decodes SRS data from reader.
func (srs *SRS) ReadFrom(r io.Reader) (int64, error) {
	// decode the SRS
	dec := bn254.NewDecoder(r)

	toDecode := []interface{}{
		&srs.G2[0],
		&srs.G2[1],
		&srs.G1,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	return dec.BytesRead(), nil
}

// WriteTo writes binary encoding of a OpeningProof
func (proof *OpeningProof) WriteTo(w io.Writer) (int64, error) {
	enc := bn254.NewEncoder(w)

	toEncode := []interface{}{
		&proof.H,
		&proof.ClaimedValue,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}

// ReadFrom decodes OpeningProof data from reader.
func (proof *OpeningProof) ReadFrom(r io.Reader) (int64, error) {
	dec := bn254.NewDecoder(r)

	toDecode := []interface{}{
		&proof.H,
		&proof.ClaimedValue,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	return dec.BytesRead(), nil
}

// WriteTo writes binary encoding of a BatchOpeningProof
func (proof *BatchOpeningProof) WriteTo(w io.Writer) (int64, error) {
	enc := bn254.NewEncoder(w)

	toEncode := []interface{}{
		&proof.H,
		proof.ClaimedValues,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}

// ReadFrom decodes BatchOpeningProof data from reader.
func (proof *BatchOpeningProof) ReadFrom(r io.Reader) (int64, error) {
	dec := bn254.NewDecoder(r)

	toDecode := []interface{}{
		&proof.H,
		&proof.ClaimedValues,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	return dec.BytesRead(), nil
}
