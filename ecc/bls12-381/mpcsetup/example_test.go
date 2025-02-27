// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mpcsetup_test

import (
	"crypto/sha256"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/mpcsetup"
	"github.com/consensys/gnark-crypto/field/hash"
)

// Custom deterministic hash function for testing
type deterministicHashFunction struct{}

func (d deterministicHashFunction) Hash(msg, dst []byte, lenInBytes int) ([]byte, error) {
	// Simple deterministic hash for demonstration
	h := sha256.New()
	h.Write(msg)
	h.Write(dst)
	result := h.Sum(nil)

	// Repeat the hash to reach the required length
	output := make([]byte, lenInBytes)
	for i := 0; i < lenInBytes; i++ {
		output[i] = result[i%len(result)]
	}
	return output, nil
}

// Custom deterministic randomness source for testing
type deterministicRandomnessSource struct {
	counter int
}

func (d *deterministicRandomnessSource) GetRandomness(b []byte) (int, error) {
	// Fill the byte slice with a deterministic pattern
	for i := range b {
		b[i] = byte((i + d.counter) % 256)
	}
	d.counter++
	return len(b), nil
}

// ExampleCustomHashAndRandomness demonstrates how to use custom hash functions and randomness sources
func Example_customHashAndRandomness() {
	// Save the original hash function and randomness source
	originalHashFunc := hash.GlobalHashFunction
	originalRandomnessSource := hash.GlobalRandomnessSource

	// Restore them after the example
	defer func() {
		hash.GlobalHashFunction = originalHashFunc
		hash.GlobalRandomnessSource = originalRandomnessSource
	}()

	// Set custom implementations
	hash.GlobalHashFunction = deterministicHashFunction{}
	hash.GlobalRandomnessSource = &deterministicRandomnessSource{}

	// Use the MPC setup with custom implementations
	// For demonstration, we'll just create some random field elements
	var element fr.Element
	element.SetRandomWithSource(hash.GlobalRandomnessSource)
	fmt.Printf("Random element generated: %v\n", !element.IsZero())

	// Generate beacon contributions with custom hash function
	contributions := mpcsetup.BeaconContributions(
		[]byte("test hash"),
		[]byte("domain separation tag"),
		[]byte("beacon challenge"),
		1,
	)
	fmt.Printf("Beacon contributions generated: %v\n", len(contributions) > 0)

	// Output:
	// Random element generated: true
	// Beacon contributions generated: true
}
