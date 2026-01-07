// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package signature defines interfaces for a Signer and a PublicKey similarly to go/crypto standard package.
package signature

import (
	"hash"
)

// PublicKey public key interface.
// The public key has a Verify function to check signatures.
type PublicKey interface {

	// Verify verifies a signature of a message
	// If hFunc is not provided, implementation may consider the message
	// to be pre-hashed, else, will use hFunc to hash the message.
	Verify(sigBin, message []byte, hFunc hash.Hash) (bool, error)

	// SetBytes sets p from binary representation in buf.
	// buf represents a public key as x||y where x, y are
	// interpreted as big endian binary numbers corresponding
	// to the coordinates of a point on the twisted Edwards.
	// It returns the number of bytes read from the buffer.
	SetBytes(buf []byte) (int, error)

	// Bytes returns the binary representation of pk
	// as x||y where x, y are the coordinates of the point
	// on the twisted Edwards as big endian integers.
	Bytes() []byte

	Equal(PublicKey) bool
}

// Signer signer interface.
type Signer interface {

	// Public returns the public key associated to
	// the signer's private key.
	Public() PublicKey

	// Sign signs a message. If hFunc is not provided, implementation may consider the message
	// to be pre-hashed, else, will use hFunc to hash the message.
	// Returns Signature or error
	Sign(message []byte, hFunc hash.Hash) ([]byte, error)

	// Bytes returns the binary representation of pk,
	// as byte array publicKey||scalar||randSrc
	// where publicKey is as publicKey.Bytes(), and
	// scalar is in big endian, of size sizeFr.
	Bytes() []byte

	// SetBytes sets pk from buf, where buf is interpreted
	// as  publicKey||scalar||randSrc
	// where publicKey is as publicKey.Bytes(), and
	// scalar is in big endian, of size sizeFr.
	// It returns the number byte read.
	SetBytes(buf []byte) (int, error)
}
