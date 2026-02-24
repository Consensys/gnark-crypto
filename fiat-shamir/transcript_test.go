// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fiatshamir

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func initTranscript() *Transcript {

	fs := NewTranscript(sha256.New(), "alpha", "beta", "gamma")

	values := [][]byte{[]byte("v1"), []byte("v2"), []byte("v3"), []byte("v4"), []byte("v5"), []byte("v6")}
	if err := fs.Bind("alpha", values[0]); err != nil {
		panic(err)
	}
	if err := fs.Bind("alpha", values[1]); err != nil {
		panic(err)
	}
	if err := fs.Bind("beta", values[2]); err != nil {
		panic(err)
	}
	if err := fs.Bind("beta", values[3]); err != nil {
		panic(err)
	}
	if err := fs.Bind("gamma", values[4]); err != nil {
		panic(err)
	}
	if err := fs.Bind("gamma", values[5]); err != nil {
		panic(err)
	}

	return fs
}

func initTranscriptWithNewChallenge() *Transcript {

	fs := NewTranscript(sha256.New())

	values := [][]byte{[]byte("v1"), []byte("v2"), []byte("v3"), []byte("v4"), []byte("v5"), []byte("v6")}
	err := fs.NewChallenge("alpha")
	if err != nil {
		panic(err)
	}
	if err := fs.Bind("alpha", values[0]); err != nil {
		panic(err)
	}
	if err := fs.Bind("alpha", values[1]); err != nil {
		panic(err)
	}
	err = fs.NewChallenge("beta")
	if err != nil {
		panic(err)
	}
	if err := fs.Bind("beta", values[2]); err != nil {
		panic(err)
	}
	if err := fs.Bind("beta", values[3]); err != nil {
		panic(err)
	}
	err = fs.NewChallenge("gamma")
	if err != nil {
		panic(err)
	}
	if err := fs.Bind("gamma", values[4]); err != nil {
		panic(err)
	}
	if err := fs.Bind("gamma", values[5]); err != nil {
		panic(err)
	}

	return fs
}

func TestDuplicateNamesInit(t *testing.T) {

	// no error is raised here
	fs := NewTranscript(sha256.New(), "a", "b", "a")

	values := [][]byte{[]byte("v1"), []byte("v2"), []byte("v3"), []byte("v4"), []byte("v5"), []byte("v6")}
	fs.Bind("a", values[0])
	fs.Bind("a", values[1])
	fs.Bind("b", values[2])
	fs.Bind("b", values[3])
	fs.Bind("a", values[4])
	fs.Bind("a", values[5])

	_, err := fs.ComputeChallenge("a")
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.ComputeChallenge("b")
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.ComputeChallenge("a")
	if err == nil {
		t.Fatal(err)
	}

}

func TestTranscript(t *testing.T) {
	t.Parallel()

	fs := initTranscript()

	// test when everything is fine
	alpha, err := fs.ComputeChallenge("alpha")
	if err != nil {
		t.Fatal(err)
	}
	beta, err := fs.ComputeChallenge("beta")
	if err != nil {
		t.Fatal(err)
	}
	gamma, err := fs.ComputeChallenge("gamma")
	if err != nil {
		t.Fatal(err)
	}

	if len(alpha) == 0 || len(beta) == 0 || len(gamma) == 0 {
		t.Fatal("one of the challenge result is empty")
	}

	// re compute the challenges to verify they are the same
	alphaBis, err := fs.ComputeChallenge("alpha")
	if err != nil {
		t.Fatal(err)
	}
	betaBis, err := fs.ComputeChallenge("beta")
	if err != nil {
		t.Fatal(err)
	}
	gammaBis, err := fs.ComputeChallenge("gamma")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(alpha, alphaBis) {
		t.Fatal("computing the same challenge twice should return the same value")
	}
	if !bytes.Equal(beta, betaBis) {
		t.Fatal("computing the same challenge twice should return the same value")
	}
	if !bytes.Equal(gamma, gammaBis) {
		t.Fatal("computing the same challenge twice should return the same value")
	}

}

func TestNewChallenge(t *testing.T) {

	fs := initTranscript()

	// test when everything is fine
	alpha, err := fs.ComputeChallenge("alpha")
	if err != nil {
		t.Fatal(err)
	}
	beta, err := fs.ComputeChallenge("beta")
	if err != nil {
		t.Fatal(err)
	}
	gamma, err := fs.ComputeChallenge("gamma")
	if err != nil {
		t.Fatal(err)
	}

	fsBis := initTranscriptWithNewChallenge()
	alphaBis, err := fsBis.ComputeChallenge("alpha")
	if err != nil {
		t.Fatal(err)
	}
	betaBis, err := fsBis.ComputeChallenge("beta")
	if err != nil {
		t.Fatal(err)
	}
	gammaBis, err := fsBis.ComputeChallenge("gamma")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(alpha, alphaBis) {
		t.Fatal("New(<challenge>) api not consistent with NewTranscript(<challenge...>)")
	}
	if !bytes.Equal(beta, betaBis) {
		t.Fatal("New(<challenge>) api not consistent with NewTranscript(<challenge...>)")
	}
	if !bytes.Equal(gamma, gammaBis) {
		t.Fatal("New(<challenge>) api not consistent with NewTranscript(<challenge...>)")
	}

}

func TestNonExistingChallenge(t *testing.T) {
	t.Parallel()

	fs := initTranscript()

	// query inexisting challenges
	_, err := fs.ComputeChallenge("delta")
	if err == nil {
		t.Fatal(err)
	}

}

func TestWrongOrder(t *testing.T) {
	t.Parallel()

	fs := initTranscript()

	// query inexisting challenges
	_, err := fs.ComputeChallenge("beta")
	if err == nil {
		t.Fatal(err)
	}

}

func TestBindToComputedChallenge(t *testing.T) {
	t.Parallel()

	fs := initTranscript()

	_, err := fs.ComputeChallenge("alpha")
	if err != nil {
		t.Fatal(err)
	}

	// bind value to an already computed challenge
	err = fs.Bind("alpha", []byte("test"))
	if err == nil {
		t.Fatal(err)
	}

}
