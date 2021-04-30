/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package signature defines interfaces for a Signer and a PublicKey similarly to go/crypto standard package.
package signature

import (
	"hash"
	"io"
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

	// Equal compares the public key to other.
	Equal(other PublicKey) bool
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

type SignatureScheme uint

const maxSignatures = 4

const (
	EDDSA_BN254 SignatureScheme = iota
	EDDSA_BLS12_381
	EDDSA_BLS12_377
	EDDSA_BW6_761
)

var signatures = make([]func(io.Reader) (Signer, error), maxSignatures)

// Register registers a key pair generating function for a given signature scheme.
func Register(ss SignatureScheme, f func(io.Reader) (Signer, error)) {
	// we cannot import the corresponding constructors directly due to import cycles.
	signatures[ss] = f
}

// New takes a source of randomness and returns a new key pair
func (ss SignatureScheme) New(r io.Reader) (Signer, error) {
	f := signatures[ss]
	return f(r)
}
