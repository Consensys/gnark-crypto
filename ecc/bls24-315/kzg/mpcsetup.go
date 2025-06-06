// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package kzg

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	curve "github.com/consensys/gnark-crypto/ecc/bls24-315"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/mpcsetup"
	"github.com/consensys/gnark-crypto/utils"
	"io"
	"math/big"
)

type MpcSetup struct {
	srs       SRS
	proof     mpcsetup.UpdateProof
	challenge []byte
}

func InitializeSetup(N int) MpcSetup {
	var res MpcSetup
	_, _, g1, g2 := curve.Generators()

	res.srs.Pk.G1 = make([]curve.G1Affine, N)
	for i := range N {
		res.srs.Pk.G1[i] = g1
	}
	res.srs.Vk.G1 = g1
	res.srs.Vk.G2[0] = g2
	res.srs.Vk.G2[1] = g2

	return res
}

// WriteTo implements io.WriterTo
func (s *MpcSetup) WriteTo(w io.Writer) (int64, error) {
	n, err := s.proof.WriteTo(w)
	if err != nil {
		return n, err
	}
	if err = binary.Write(w, binary.BigEndian, uint64(len(s.srs.Pk.G1))); err != nil {
		return -1, err // binary.Write doesn't return the number written in case of failure
	}
	n += 8
	enc := curve.NewEncoder(w)
	for i := range len(s.srs.Pk.G1) - 1 {
		if err = enc.Encode(&s.srs.Pk.G1[i+1]); err != nil {
			return n + enc.BytesWritten(), err
		}
	}
	if err = enc.Encode(&s.srs.Vk.G2[1]); err != nil {
		return n + enc.BytesWritten(), err
	}
	err = enc.Encode(s.challenge)
	return n + enc.BytesWritten(), err
}

// ReadFrom implements io.ReaderFrom
func (s *MpcSetup) ReadFrom(r io.Reader) (int64, error) {
	n, err := s.proof.ReadFrom(r)
	if err != nil {
		return n, err
	}
	var N uint64
	if err = binary.Read(r, binary.BigEndian, &N); err != nil {
		return -1, err
	}
	_, _, g1, g2 := curve.Generators()
	n += 8
	dec := curve.NewDecoder(r)
	s.srs.Pk.G1 = make([]curve.G1Affine, N)
	s.srs.Pk.G1[0] = g1
	s.srs.Vk.G2[0] = g2
	for i := range N - 1 {
		if err = dec.Decode(&s.srs.Pk.G1[i+1]); err != nil {
			return n + dec.BytesRead(), err
		}
	}
	if err = dec.Decode(&s.srs.Vk.G2[1]); err != nil {
		return n + dec.BytesRead(), err
	}
	if len(s.challenge) != 32 {
		s.challenge = make([]byte, 32)
	}
	err = dec.Decode(&s.challenge)
	return n + dec.BytesRead(), err
}

func (s *MpcSetup) hash() []byte {
	hsh := sha256.New()
	if _, err := s.WriteTo(hsh); err != nil {
		panic(err)
	}
	return hsh.Sum(nil)
}

func (s *MpcSetup) Contribute() {
	s.challenge = s.hash()
	var contribution fr.Element

	s.proof = mpcsetup.UpdateValues(&contribution, append([]byte("KZG Setup"), s.challenge...), 0, &s.srs.Vk.G2[1])
	mpcsetup.UpdateMonomialsG1(s.srs.Pk.G1, &contribution)
}

func (s *MpcSetup) Verify(next *MpcSetup) error {
	challenge := s.hash()
	if len(next.challenge) != 0 && !bytes.Equal(next.challenge, challenge) {
		return errors.New("the challenge does not match the previous contribution's hash")
	}
	next.challenge = challenge

	if len(s.srs.Pk.G1) != len(next.srs.Pk.G1) {
		return errors.New("different domain sizes")
	}

	if !next.srs.Vk.G2[1].IsInSubGroup() {
		return errors.New("[x]₂ representation not in subgroup")
	}

	// TODO @Tabaie replace with batch subgroup check
	n := len(next.srs.Pk.G1) - 1
	wp := utils.NewWorkerPool()
	defer wp.Stop()
	fail := make(chan error, wp.NbWorkers())

	wp.Submit(n, func(start, end int) {
		for i := start; i < end; i++ {
			if !next.srs.Pk.G1[i+1].IsInSubGroup() {
				fail <- fmt.Errorf("[x^%d]₁ representation not in subgroup", i+1)
				break
			}
		}
	}, n/wp.NbWorkers()+1).Wait()
	close(fail)
	for err := range fail {
		if err != nil {
			return err
		}
	}

	if err := next.proof.Verify(append([]byte("KZG Setup"), challenge...), 0, mpcsetup.ValueUpdate{
		Previous: s.srs.Vk.G2[1],
		Next:     next.srs.Vk.G2[1],
	}); err != nil {
		return err
	}

	return mpcsetup.SameRatioMany(s.srs.Pk.G1, s.srs.Vk.G2[:])
}

func (s *MpcSetup) Seal(beaconChallenge []byte) SRS {
	contributions := mpcsetup.BeaconContributions(s.hash(), []byte("KZG Setup"), beaconChallenge, 1)
	var I big.Int
	contributions[0].BigInt(&I)
	s.srs.Vk.G2[1].ScalarMultiplication(&s.srs.Vk.G2[1], &I)
	mpcsetup.UpdateMonomialsG1(s.srs.Pk.G1, &contributions[0])

	s.srs.Vk.Lines[0] = curve.PrecomputeLines(s.srs.Vk.G2[0])
	s.srs.Vk.Lines[1] = curve.PrecomputeLines(s.srs.Vk.G2[1])

	return s.srs
}
