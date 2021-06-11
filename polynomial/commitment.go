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

// Package polynomial provides interfaces for polynomial and polynomial commitment schemes defined in gnark-crypto/ecc/.../fr.
package polynomial

import (
	"errors"
	"io"
	"math/big"
)

var (
	ErrVerifyOpeningProof            = errors.New("error verifying opening proof")
	ErrVerifyBatchOpeningSinglePoint = errors.New("error verifying batch opening proof at single point")
)

// Digest interface that an additively homomorphic
// polynomial commitment should implement.
type Digest interface {
	Clone() Digest
	Add(d1, d2 Digest) Digest
	Sub(d1, d2 Digest) Digest
	ScalarMul(d Digest, s big.Int) Digest
	Marshal() []byte
	Unmarshal(buf []byte) error
}

// OpeningProof interface that an opening proof
// should implement.
type OpeningProof interface {
	Marshal() []byte
	GetClaimedValue() []byte
}

// BatchOpeningProofSinglePoint interface that a bacth opening proof (single point)
// should implement.
type BatchOpeningProofSinglePoint interface {
	Marshal() []byte

	// GetClaimedValues get the claimed values, in the order in which the polynomials
	// are given in the opening proof.
	GetClaimedValues() [][]byte
}

// CommitmentScheme interface for an additively homomorphic
// polynomial commitment scheme.
// The function BatchOpenSinglePoint is proper to an additively
// homomorphic commitment scheme.
type CommitmentScheme interface {
	io.WriterTo
	io.ReaderFrom

	Commit(p Polynomial) (Digest, error)

	Open(point interface{}, p Polynomial) (OpeningProof, error)

	// Verify verifies an opening proof of commitment at point
	Verify(commitment Digest, proof OpeningProof) error

	// BatchOpenSinglePoint creates a batch opening proof at _val of a list of polynomials.
	// It's an interactive protocol, made non interactive using Fiat Shamir.
	// point is the point at which the polynomials are opened.
	// digests is the list of committed polynomials to open, need to derive the challenge using Fiat Shamir.
	// polynomials is the list of polynomials to open.
	BatchOpenSinglePoint(point interface{}, digests []Digest, polynomials []Polynomial) (BatchOpeningProofSinglePoint, error)

	// BatchVerifySinglePoint verifies a batched opening proof at a single point of a list of polynomials.
	// point: point at which the polynomials are evaluated
	// claimedValues: claimed values of the polynomials at _val
	// commitments: list of commitments to the polynomials which are opened
	// batchOpeningProof: the batched opening proof at a single point of the polynomials.
	BatchVerifySinglePoint(digests []Digest, batchOpeningProof BatchOpeningProofSinglePoint) error
}
