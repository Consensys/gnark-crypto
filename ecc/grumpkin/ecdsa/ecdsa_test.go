// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package ecdsa

import (
	"crypto/rand"
	"crypto/sha256"
	"github.com/consensys/gnark-crypto/ecc/grumpkin/fr"
	"math/big"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestECDSA(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("[GRUMPKIN] test the signing and verification", prop.ForAll(
		func() bool {

			privKey, _ := GenerateKey(rand.Reader)
			publicKey := privKey.PublicKey

			msg := []byte("testing ECDSA")
			hFunc := sha256.New()
			sig, _ := privKey.Sign(msg, hFunc)
			flag, _ := publicKey.Verify(sig, msg, hFunc)

			return flag
		},
	))

	properties.Property("[GRUMPKIN] test the signing and verification (pre-hashed)", prop.ForAll(
		func() bool {

			privKey, _ := GenerateKey(rand.Reader)
			publicKey := privKey.PublicKey

			msg := []byte("testing ECDSA")
			sig, _ := privKey.Sign(msg, nil)
			flag, _ := publicKey.Verify(sig, msg, nil)

			return flag
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestNonMalleability(t *testing.T) {

	// buffer too big
	t.Run("buffer_overflow", func(t *testing.T) {
		bsig := make([]byte, 2*sizeFr+1)
		var sig Signature
		_, err := sig.SetBytes(bsig)
		if err != errWrongSize {
			t.Fatal("should raise wrong size error")
		}
	})

	// R overflows p_mod
	t.Run("R_overflow", func(t *testing.T) {
		bsig := make([]byte, 2*sizeFr)
		r := big.NewInt(1)
		frMod := fr.Modulus()
		r.Add(r, frMod)
		buf := r.Bytes()
		copy(bsig, buf[:])

		var sig Signature
		_, err := sig.SetBytes(bsig)
		if err != errRBiggerThanRMod {
			t.Fatal("should raise error r >= r_mod")
		}
	})

	// S overflows p_mod
	t.Run("S_overflow", func(t *testing.T) {
		bsig := make([]byte, 2*sizeFr)
		r := big.NewInt(1)
		frMod := fr.Modulus()
		r.Add(r, frMod)
		buf := r.Bytes()
		copy(bsig[sizeFr:], buf[:])
		big.NewInt(1).FillBytes(bsig[:sizeFr])

		var sig Signature
		_, err := sig.SetBytes(bsig)
		if err != errSBiggerThanRMod {
			t.Fatal("should raise error s >= r_mod")
		}
	})

}

func TestNoZeros(t *testing.T) {
	t.Run("R=0", func(t *testing.T) {
		// R is 0
		var sig Signature
		big.NewInt(0).FillBytes(sig.R[:])
		big.NewInt(1).FillBytes(sig.S[:])
		bts := sig.Bytes()
		var newSig Signature
		_, err := newSig.SetBytes(bts)
		if err != errZero {
			t.Fatal("expected error for zero R")
		}
	})
	t.Run("S=0", func(t *testing.T) {
		// S is 0
		var sig Signature
		big.NewInt(1).FillBytes(sig.R[:])
		big.NewInt(0).FillBytes(sig.S[:])
		bts := sig.Bytes()
		var newSig Signature
		_, err := newSig.SetBytes(bts)
		if err != errZero {
			t.Fatal("expected error for zero S")
		}
	})
}

// ------------------------------------------------------------
// benches

func BenchmarkSignECDSA(b *testing.B) {

	privKey, _ := GenerateKey(rand.Reader)

	msg := []byte("benchmarking ECDSA sign()")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		privKey.Sign(msg, nil)
	}
}

func BenchmarkVerifyECDSA(b *testing.B) {

	privKey, _ := GenerateKey(rand.Reader)
	msg := []byte("benchmarking ECDSA sign()")
	sig, _ := privKey.Sign(msg, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		privKey.PublicKey.Verify(sig, msg, nil)
	}
}
