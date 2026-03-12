// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fiatshamir

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestNewTranscriptDuplicateChallenge(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		NewTranscript(sha256.New(), "alpha", "beta", "alpha")
	}, "NewTranscript should panic on duplicate challenge names")
}

func TestNewChallenge(t *testing.T) {
	t.Parallel()

	fs := NewTranscript(sha256.New(), "alpha")

	// adding a new challenge after construction should work
	require.NoError(t, fs.NewChallenge("beta"))

	// adding a duplicate should return an error
	require.Error(t, fs.NewChallenge("alpha"))

	// the transcript should work normally
	require.NoError(t, fs.Bind("alpha", []byte("v1")))
	_, err := fs.ComputeChallenge("alpha")
	require.NoError(t, err)
	_, err = fs.ComputeChallenge("beta")
	require.NoError(t, err)
}

// TestNewTranscriptVsNewChallenge verifies that challenges computed from a
// transcript created with NewTranscript(h, ids...) are identical to those
// from a transcript built with NewTranscript(h) + NewChallenge(id) calls.
func TestNewTranscriptVsNewChallenge(t *testing.T) {
	t.Parallel()

	names := []string{"alpha", "beta", "gamma"}
	values := [][]byte{[]byte("v1"), []byte("v2"), []byte("v3"), []byte("v4"), []byte("v5"), []byte("v6")}

	// transcript 1: challenges defined at construction
	fs1 := NewTranscript(sha256.New(), names...)
	require.NoError(t, fs1.Bind("alpha", values[0]))
	require.NoError(t, fs1.Bind("alpha", values[1]))
	require.NoError(t, fs1.Bind("beta", values[2]))
	require.NoError(t, fs1.Bind("beta", values[3]))
	require.NoError(t, fs1.Bind("gamma", values[4]))
	require.NoError(t, fs1.Bind("gamma", values[5]))

	// transcript 2: challenges added via NewChallenge
	fs2 := NewTranscript(sha256.New())
	for _, name := range names {
		require.NoError(t, fs2.NewChallenge(name))
	}
	require.NoError(t, fs2.Bind("alpha", values[0]))
	require.NoError(t, fs2.Bind("alpha", values[1]))
	require.NoError(t, fs2.Bind("beta", values[2]))
	require.NoError(t, fs2.Bind("beta", values[3]))
	require.NoError(t, fs2.Bind("gamma", values[4]))
	require.NoError(t, fs2.Bind("gamma", values[5]))

	for _, name := range names {
		c1, err := fs1.ComputeChallenge(name)
		require.NoError(t, err)
		c2, err := fs2.ComputeChallenge(name)
		require.NoError(t, err)
		require.Equal(t, c1, c2, "challenge %s should be identical regardless of registration method", name)
	}
}
