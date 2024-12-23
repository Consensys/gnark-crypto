// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package mpcsetup

import (
	"bytes"
	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
)

// BeaconContributions provides the last
func BeaconContributions(hash, dst, beaconChallenge []byte, n int) []fr.Element {
	var (
		bb  bytes.Buffer
		err error
	)
	bb.Grow(len(hash) + len(beaconChallenge))
	bb.Write(hash)
	bb.Write(beaconChallenge)

	res := make([]fr.Element, 1)

	allNonZero := func() bool {
		for i := range res {
			if res[i].IsZero() {
				return false
			}
		}
		return true
	}

	// cryptographically unlikely for this to be run more than once
	for !allNonZero() {
		if res, err = fr.Hash(bb.Bytes(), dst, n); err != nil {
			panic(err)
		}
		bb.WriteByte('=') // padding just so that the hash is different next time
	}

	return res
}
