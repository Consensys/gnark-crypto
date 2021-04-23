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

package fiatshamir

import (
	"crypto/sha256"
	"errors"
	"hash"

	gnark_hash "github.com/consensys/gnark-crypto/hash"
)

// errChallengeNotFound is returned when a wrong challenge name is provided.
var (
	errChallengeNotFound            = errors.New("challenge not recorded in the Transcript")
	errChallengeAlreadyComputed     = errors.New("challenge already computed, cannot be binded to other values")
	errPreviousChallengeNotComputed = errors.New("the previous challenge is needed and has not been computed")
)

// HashFS hash function used in Fiat Shamir. Likely snark friendly hash functions will be chosen.
type HashFS uint

// Supported hash functions for Fiat Shamir. Sha256 is arbitrary, we just need something fast non-snark friendly hash.
const (
	SHA256 HashFS = iota
	MIMC_BN254
	MIMC_BLS12_381
	MIMC_BLS12_377
	MIMC_BW6_761
)

// Transcript handles the creation of challenges for Fiat Shamir.
type Transcript struct {

	// stores the current round number. Each time a challenge is generated,
	// the round variable is incremented.
	nbChallenges int

	// challengeOrder maps the challenge's name to a number corresponding to its order.
	challengeOrder map[string]int

	// bindings stores the variables a challenge is binded to.
	// The i-th entry stores the variables to which the i-th challenge is binded to.
	bindings [][]byte

	// challenges stores the computed challenges. The i-th entry stores the i-th computed challenge.
	challenges [][]byte

	// boolean table telling if the i-th challenge has been computed.
	isComputed []bool

	// hash function that is used.
	h hash.Hash
}

// NewTranscript returns a new transcript.
// h is the hash function that is used to compute the challenges.
// challenges are the name of the challenges. The order is important.
func NewTranscript(h HashFS, challenges ...string) Transcript {

	var res Transcript

	res.nbChallenges = len(challenges)

	res.challengeOrder = make(map[string]int)
	for i := 0; i < len(challenges); i++ {
		res.challengeOrder[challenges[i]] = i
	}

	res.bindings = make([][]byte, res.nbChallenges)
	res.challenges = make([][]byte, res.nbChallenges)
	for i := 0; i < res.nbChallenges; i++ {
		res.bindings[i] = make([]byte, 0)
	}

	res.isComputed = make([]bool, res.nbChallenges)

	switch h {
	case SHA256:
		res.h = sha256.New()
	case MIMC_BN254:
		res.h = gnark_hash.MIMC_BN254.New("seed")
	case MIMC_BLS12_381:
		res.h = gnark_hash.MIMC_BLS12_381.New("seed")
	case MIMC_BLS12_377:
		res.h = gnark_hash.MIMC_BLS12_377.New("seed")
	case MIMC_BW6_761:
		res.h = gnark_hash.MIMC_BW6_761.New("seed")
	default:
		panic("the chosen hash function is not available")
	}

	return res
}

// Bind binds the challenge to value. A challenge can be binded to an
// arbitrary number of values, but the order in which the binded values
// are added is important. Once a challenge is computed, it cannot be
// binded to other values.
func (m *Transcript) Bind(challenge string, value []byte) error {

	challengeNumber, ok := m.challengeOrder[challenge]

	if !ok {
		return errChallengeNotFound
	}

	if m.isComputed[challengeNumber] {
		return errChallengeAlreadyComputed
	}
	m.bindings[challengeNumber] = append(m.bindings[challengeNumber], value...)

	return nil

}

// ComputeChallenge computes the challenge corresponding to the given name.
// The challenge is:
// * H(name || previous_challenge || binded_values...) if the challenge is not the first one
// * H(name || binded_values... ) if it's is the first challenge
func (m *Transcript) ComputeChallenge(challenge string) ([]byte, error) {

	challengeNumber, ok := m.challengeOrder[challenge]
	if !ok {
		return nil, errChallengeNotFound
	}

	// if the challenge was already computed we return it
	if m.isComputed[challengeNumber] {
		return m.challenges[challengeNumber], nil
	}

	m.h.Reset()

	// write the challenge name, the purpose is to have a domain separator
	bName := []byte(challenge)
	if _, err := m.h.Write(bName); err != nil {
		return nil, err
	}

	// write the previous challenge if it's not the first challenge
	if challengeNumber != 0 {
		if !m.isComputed[challengeNumber-1] {
			return nil, errPreviousChallengeNotComputed
		}
		bPreviousChallenge := m.challenges[challengeNumber-1]
		if _, err := m.h.Write(bPreviousChallenge[:]); err != nil {
			return nil, err
		}
	}

	// write the binded values in the order they were added
	if _, err := m.h.Write(m.bindings[challengeNumber]); err != nil {
		return nil, err
	}

	// compute the hash of the accumulated values
	res := m.h.Sum(nil)

	m.challenges[challengeNumber] = make([]byte, len(res))
	copy(m.challenges[challengeNumber], res)
	m.isComputed[challengeNumber] = true

	return res, nil

}
