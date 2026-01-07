// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// FOO

package secp256k1

import (
	"errors"
	"io"

	"github.com/consensys/gnark-crypto/ecc/secp256k1/fp"
)

// SizeOfG1AffineCompressed represents the size in bytes that a G1Affine need in binary form, compressed
const SizeOfG1AffineCompressed = 32

// SizeOfG1AffineUncompressed represents the size in bytes that a G1Affine need in binary form, uncompressed
const SizeOfG1AffineUncompressed = SizeOfG1AffineCompressed * 2

// RawBytes returns binary representation of p (stores X and Y coordinate)
func (p *G1Affine) RawBytes() (res [SizeOfG1AffineUncompressed]byte) {

	// not compressed
	// we store the Y coordinate
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(res[32:32+fp.Bytes]), p.Y)

	// we store the X coordinate
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(res[0:0+fp.Bytes]), p.X)

	return
}

// SetBytes sets p from binary representation in buf and returns number of consumed bytes
//
// bytes in buf must match RawBytes()
//
// if buf is too short io.ErrShortBuffer is returned
//
// this check if the resulting point is on the curve and in the correct subgroup
func (p *G1Affine) SetBytes(buf []byte) (int, error) {
	return p.setBytes(buf, true)
}

// we store both X and Y and there is no spare bit for flagging
func (p *G1Affine) setBytes(buf []byte, subGroupCheck bool) (int, error) {
	if len(buf) < SizeOfG1AffineCompressed {
		return 0, io.ErrShortBuffer
	}

	// uncompressed point
	// read X and Y coordinates
	if err := p.X.SetBytesCanonical(buf[:fp.Bytes]); err != nil {
		return 0, err
	}
	if err := p.Y.SetBytesCanonical(buf[fp.Bytes : fp.Bytes*2]); err != nil {
		return 0, err
	}

	// subgroup check
	if subGroupCheck && !p.IsInSubGroup() {
		return 0, errors.New("invalid point: subgroup check failed")
	}

	return SizeOfG1AffineUncompressed, nil

}
