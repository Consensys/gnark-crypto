// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fiatshamir

import (
	"errors"
	"fmt"
	"hash"
)

// errChallengeNotFound is returned when a wrong challenge name is provided.
var (
	errChallengeNotFound            = errors.New("challenge not recorded in the transcript")
	errChallengeAlreadyComputed     = errors.New("challenge already computed, cannot be bound to other values")
	errPreviousChallengeNotComputed = errors.New("the previous challenge is needed and has not been computed")
)

// Transcript handles the creation of challenges for Fiat Shamir.
type Transcript struct {
	// hash function that is used.
	h hash.Hash

	challenges map[string]challenge
	previous   *challenge
}

type challenge struct {
	position   int      // position of the challenge in the Transcript. order matters.
	bindings   [][]byte // bindings stores the variables a challenge is bound to.
	value      []byte   // value stores the computed challenge
	isComputed bool
}

// NewTranscript returns a new transcript.
// h is the hash function that is used to compute the challenges.
// challenges are the name of the challenges. The order of the challenges IDs matters.
func NewTranscript(h hash.Hash, challengesID ...string) *Transcript {
	challenges := make(map[string]challenge)
	for i := range challengesID {
		challenges[challengesID[i]] = challenge{position: i}
	}
	t := &Transcript{
		challenges: challenges,
		h:          h,
	}
	return t
}

// Bind binds the challenge to value. A challenge can be bound to an
// arbitrary number of values, but the order in which the bound values
// are added is important. Once a challenge is computed, it cannot be
// bound to other values.
func (t *Transcript) Bind(challengeID string, bValue []byte) error {

	currentChallenge, ok := t.challenges[challengeID]
	if !ok {
		return errChallengeNotFound
	}

	if currentChallenge.isComputed {
		return errChallengeAlreadyComputed
	}

	bCopy := make([]byte, len(bValue))
	copy(bCopy, bValue)
	currentChallenge.bindings = append(currentChallenge.bindings, bCopy)
	t.challenges[challengeID] = currentChallenge

	return nil

}

// ComputeChallenge computes the challenge corresponding to the given name.
// The challenge is:
// * H(name || previous_challenge || binded_values...) if the challenge is not the first one
// * H(name || binded_values... ) if it is the first challenge
func (t *Transcript) ComputeChallenge(challengeID string) ([]byte, error) {

	challenge, ok := t.challenges[challengeID]
	if !ok {
		return nil, errChallengeNotFound
	}

	// if the challenge was already computed we return it
	if challenge.isComputed {
		return challenge.value, nil
	}

	// reset before populating the internal state
	t.h.Reset()
	defer t.h.Reset()

	if _, err := t.h.Write([]byte(challengeID)); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	// write the previous challenge if it's not the first challenge
	if challenge.position != 0 {
		if t.previous == nil || (t.previous.position != challenge.position-1) {
			return nil, errPreviousChallengeNotComputed
		}
		if _, err := t.h.Write(t.previous.value[:]); err != nil {
			return nil, err
		}
	}

	// write the binded values in the order they were added
	for _, b := range challenge.bindings {
		if _, err := t.h.Write(b); err != nil {
			return nil, err
		}
	}

	// compute the hash of the accumulated values
	res := t.h.Sum(nil)

	challenge.value = make([]byte, len(res))
	copy(challenge.value, res)
	challenge.isComputed = true

	t.challenges[challengeID] = challenge
	t.previous = &challenge

	return res, nil

}
