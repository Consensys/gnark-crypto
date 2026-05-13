// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fiatshamir

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// This example demonstrates how to use a Fiat-Shamir transcript to derive
// verifier challenges non-interactively.
//
// The prover registers named challenges, binds public values (commitments,
// evaluations, etc.) to them, and then computes each challenge sequentially.
// Each challenge is the hash of its name, the previous challenge, and its
// bound values.
func Example() {
	// Create a transcript with two challenges.
	fs := NewTranscript(sha256.New(), "alpha", "beta")

	// Bind public data to the first challenge.
	fs.Bind("alpha", []byte("commitment_0"))
	fs.Bind("alpha", []byte("commitment_1"))

	// Compute the first challenge: H("alpha" || bindings...).
	alpha, err := fs.ComputeChallenge("alpha")
	if err != nil {
		panic(err)
	}
	fmt.Println("alpha:", hex.EncodeToString(alpha))

	// Bind public data to the second challenge.
	fs.Bind("beta", []byte("evaluation_0"))

	// Compute the second challenge: H("beta" || alpha || bindings...).
	beta, err := fs.ComputeChallenge("beta")
	if err != nil {
		panic(err)
	}
	fmt.Println("beta:", hex.EncodeToString(beta))

	// Challenges can also be added after construction with NewChallenge.
	if err := fs.NewChallenge("gamma"); err != nil {
		panic(err)
	}
	fs.Bind("gamma", []byte("proof_element"))

	gamma, err := fs.ComputeChallenge("gamma")
	if err != nil {
		panic(err)
	}
	fmt.Println("gamma:", hex.EncodeToString(gamma))

	// Output:
	// alpha: 4fd93a75df92f59a0b4b02bf342fdb74164f5618aed33aea2d09cee35fe90c3e
	// beta: 8c07b11b832e18b4ed3c9ff5dca7b199d6fafbf0e3ca6bff866e4843a01df137
	// gamma: 35bba866c55c055872ad35f3d1613ad8d74f4935a351b9f1df57f2b8f3f468ff
}
