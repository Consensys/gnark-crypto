// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package plookup

import (
	"crypto/sha256"
	"errors"
	"math/big"
	"sort"

	bls24315 "github.com/consensys/gnark-crypto/ecc/bls24-315"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr/permutation"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/kzg"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
)

var (
	ErrIncompatibleSize = errors.New("the tables in f and t are not of the same size")
	ErrFoldedCommitment = errors.New("the folded commitment is malformed")
	ErrNumberDigests    = errors.New("proof.ts and proof.fs are not of the same length")
)

// ProofLookupTables proofs that a list of tables
type ProofLookupTables struct {

	// commitments to the rows f
	fs []kzg.Digest

	// commitments to the rows of t
	ts []kzg.Digest

	// lookup proof for the f and t folded
	foldedProof ProofLookupVector

	// proof that the ts folded correspond to t in the folded proof
	permutationProof permutation.Proof
}

// ProveLookupTables generates a proof that f, seen as a multi dimensional table,
// consists of vectors that are in t. In other words for each i, f[:][i] must be one
// of the t[:][j].
//
// For instance, if t is the truth table of the XOR function, t will be populated such
// that t[:][i] contains the i-th entry of the truth table, so t[0][i] XOR t[1][i] = t[2][i].
//
// The fr.Vector in f and t are supposed to be of the same size constant size.
func ProveLookupTables(pk kzg.ProvingKey, f, t []fr.Vector) (ProofLookupTables, error) {

	// res
	proof := ProofLookupTables{}
	var err error

	// hash function used for Fiat Shamir
	hFunc := sha256.New()

	// transcript to derive the challenge
	fs := fiatshamir.NewTranscript(hFunc, "lambda")

	// check the sizes
	if len(f) != len(t) {
		return proof, ErrIncompatibleSize
	}
	s := len(f[0])
	for i := 1; i < len(f); i++ {
		if len(f[i]) != s {
			return proof, ErrIncompatibleSize
		}
	}
	s = len(t[0])
	for i := 1; i < len(t); i++ {
		if len(t[i]) != s {
			return proof, ErrIncompatibleSize
		}
	}

	// commit to the tables in f and t
	nbRows := len(t)
	proof.fs = make([]kzg.Digest, nbRows)
	proof.ts = make([]kzg.Digest, nbRows)
	_nbColumns := len(f[0]) + 1
	if _nbColumns < len(t[0]) {
		_nbColumns = len(t[0])
	}
	d := fft.NewDomain(uint64(_nbColumns))
	nbColumns := d.Cardinality
	lfs := make([][]fr.Element, nbRows)
	cfs := make([][]fr.Element, nbRows)
	lts := make([][]fr.Element, nbRows)
	cts := make([][]fr.Element, nbRows)

	for i := 0; i < nbRows; i++ {

		cfs[i] = make([]fr.Element, nbColumns)
		lfs[i] = make([]fr.Element, nbColumns)
		copy(cfs[i], f[i])
		copy(lfs[i], f[i])
		for j := len(f[i]); j < int(nbColumns); j++ {
			cfs[i][j] = f[i][len(f[i])-1]
			lfs[i][j] = f[i][len(f[i])-1]
		}
		d.FFTInverse(cfs[i], fft.DIF)
		fft.BitReverse(cfs[i])
		proof.fs[i], err = kzg.Commit(cfs[i], pk)
		if err != nil {
			return proof, err
		}

		cts[i] = make([]fr.Element, nbColumns)
		lts[i] = make([]fr.Element, nbColumns)
		copy(cts[i], t[i])
		copy(lts[i], t[i])
		for j := len(t[i]); j < int(d.Cardinality); j++ {
			cts[i][j] = t[i][len(t[i])-1]
			lts[i][j] = t[i][len(t[i])-1]
		}
		d.FFTInverse(cts[i], fft.DIF)
		fft.BitReverse(cts[i])
		proof.ts[i], err = kzg.Commit(cts[i], pk)
		if err != nil {
			return proof, err
		}
	}

	// fold f and t
	comms := make([]*kzg.Digest, 2*nbRows)
	for i := 0; i < nbRows; i++ {
		comms[i] = new(kzg.Digest)
		comms[i].Set(&proof.fs[i])
		comms[nbRows+i] = new(kzg.Digest)
		comms[nbRows+i].Set(&proof.ts[i])
	}
	lambda, err := deriveRandomness(fs, "lambda", comms...)
	if err != nil {
		return proof, err
	}
	foldedf := make(fr.Vector, nbColumns)
	foldedt := make(fr.Vector, nbColumns)
	for i := 0; i < int(nbColumns); i++ {
		for j := nbRows - 1; j >= 0; j-- {
			foldedf[i].Mul(&foldedf[i], &lambda).
				Add(&foldedf[i], &lfs[j][i])
			foldedt[i].Mul(&foldedt[i], &lambda).
				Add(&foldedt[i], &lts[j][i])
		}
	}

	// generate a proof of permutation of the foldedt and sort(foldedt)
	foldedtSorted := make(fr.Vector, nbColumns)
	copy(foldedtSorted, foldedt)
	sort.Sort(foldedtSorted)
	proof.permutationProof, err = permutation.Prove(pk, foldedt, foldedtSorted)
	if err != nil {
		return proof, err
	}

	// call plookupVector, on foldedf[:len(foldedf)-1] to ensure that the domain size
	// in ProveLookupVector is the same as d's
	proof.foldedProof, err = ProveLookupVector(pk, foldedf[:len(foldedf)-1], foldedt)

	return proof, err
}

// VerifyLookupTables verifies that a ProofLookupTables proof is correct.
func VerifyLookupTables(vk kzg.VerifyingKey, proof ProofLookupTables) error {

	// hash function used for Fiat Shamir
	hFunc := sha256.New()

	// transcript to derive the challenge
	fs := fiatshamir.NewTranscript(hFunc, "lambda")

	// check that the number of digests is the same
	if len(proof.fs) != len(proof.ts) {
		return ErrNumberDigests
	}

	// fold the commitments fs and ts
	nbRows := len(proof.fs)
	comms := make([]*kzg.Digest, 2*nbRows)
	for i := 0; i < nbRows; i++ {
		comms[i] = &proof.fs[i]
		comms[i+nbRows] = &proof.ts[i]
	}
	lambda, err := deriveRandomness(fs, "lambda", comms...)
	if err != nil {
		return err
	}

	// fold the commitments of the rows of t and f
	var comf, comt kzg.Digest
	comf.Set(&proof.fs[nbRows-1])
	comt.Set(&proof.ts[nbRows-1])
	var blambda big.Int
	lambda.BigInt(&blambda)
	for i := nbRows - 2; i >= 0; i-- {
		comf.ScalarMultiplication(&comf, &blambda).
			Add(&comf, &proof.fs[i])
		comt.ScalarMultiplication(&comt, &blambda).
			Add(&comt, &proof.ts[i])
	}

	// check that the folded commitment of the fs correspond to foldedProof.f
	if !comf.Equal(&proof.foldedProof.f) {
		return ErrFoldedCommitment
	}

	// check that the folded commitment of the ts is a permutation of proof.FoldedProof.t
	err = permutation.Verify(vk, proof.permutationProof)
	if err != nil {
		return err
	}

	// verify the inner proof
	return VerifyLookupVector(vk, proof.foldedProof)
}

// TODO put that in fiat-shamir package
func deriveRandomness(fs *fiatshamir.Transcript, challenge string, points ...*bls24315.G1Affine) (fr.Element, error) {

	var buf [bls24315.SizeOfG1AffineUncompressed]byte
	var r fr.Element

	for _, p := range points {
		buf = p.RawBytes()
		if err := fs.Bind(challenge, buf[:]); err != nil {
			return r, err
		}
	}

	b, err := fs.ComputeChallenge(challenge)
	if err != nil {
		return r, err
	}
	r.SetBytes(b)
	return r, nil
}
