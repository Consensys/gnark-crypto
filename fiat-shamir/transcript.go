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
	errChallengeAlreadyComputed     = errors.New("challenge already computed, cannot be binded to other values")
	errPreviousChallengeNotComputed = errors.New("the previous challenge is needed and has not been computed")
	errChallengeAlreadyExists       = errors.New("this challenge name is already used and recorded")
)

type Transcript struct {
	// hash function that is used.
	h hash.Hash

	challenges         []challenge // the order matters
	nameToChallengePos map[string]int
}

type challenge struct {
	bindings   [][]byte // bindings stores the variables a challenge is binded to.
	name       string
	value      []byte // value stores the computed challenge
	isComputed bool
}

// NewTranscript returns a new transcript.
// Call NewChallenge to attach a challenge to the transcript.
func NewTranscript(h hash.Hash, challengesID ...string) *Transcript {
	t := &Transcript{
		challenges:         make([]challenge, 0, len(challengesID)),
		nameToChallengePos: make(map[string]int, len(challengesID)),
		h:                  h,
	}
	for _, id := range challengesID {
		if _, ok := t.nameToChallengePos[id]; ok {
			panic("duplicate challenge name: " + id)
		}
		t.nameToChallengePos[id] = len(t.challenges)
		t.challenges = append(t.challenges, challenge{name: id})
	}
	return t
}

// Bind binds the challenge to value. A challenge can be binded to an
// arbitrary number of values, but the order in which the binded values
// are added is important. Once a challenge is computed, it cannot be
// binded to other values.
func (t *Transcript) Bind(challengeID string, bValue []byte) error {

	pos, ok := t.nameToChallengePos[challengeID]
	if !ok {
		return errChallengeNotFound
	}

	currentChallenge := t.challenges[pos]
	if currentChallenge.isComputed {
		return errChallengeAlreadyComputed
	}

	bCopy := make([]byte, len(bValue))
	copy(bCopy, bValue)
	currentChallenge.bindings = append(currentChallenge.bindings, bCopy)
	t.challenges[pos] = currentChallenge

	return nil
}

// NewChallenge appends a new challenge to the list of challenges to be computed.
// The newly added challenge is the last on the list
func (t *Transcript) NewChallenge(challengeID string) error {
	if _, ok := t.nameToChallengePos[challengeID]; ok {
		return errChallengeAlreadyExists
	}
	nbChallenges := len(t.challenges)
	challenge := challenge{
		name:       challengeID,
		isComputed: false,
	}
	t.challenges = append(t.challenges, challenge)
	t.nameToChallengePos[challengeID] = nbChallenges
	return nil
}

// ComputeChallenge computes the challenge corresponding to the given name.
// The challenge is:
// * H(name || previous_challenge || binded_values...) if the challenge is not the first one
// * H(name || binded_values... ) if it is the first challenge
func (t *Transcript) ComputeChallenge(challengeID string) ([]byte, error) {

	pos, ok := t.nameToChallengePos[challengeID]
	if !ok {
		return nil, errChallengeNotFound
	}

	// if the challenge was already computed we return it
	challenge := t.challenges[pos]
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
	if pos != 0 {
		if !t.challenges[pos-1].isComputed {
			return nil, errPreviousChallengeNotComputed
		}
		if _, err := t.h.Write(t.challenges[pos-1].value[:]); err != nil {
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

	t.challenges[pos] = challenge

	return res, nil

}
