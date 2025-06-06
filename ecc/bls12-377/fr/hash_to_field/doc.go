// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

// Package htf provides hasher based on RFC 9380 Section 5.
//
// The [RFC 9380] defines a method for hashing bytes to elliptic curves. Section
// 5 of the RFC describes a method for uniformly hashing bytes into a field
// using a domain separation. The hashing is implemented in [fp], but this
// package provides a wrapper for the method which implements [hash.Hash] for
// using the method recursively.
//
// [RFC 9380]: https://datatracker.ietf.org/doc/html/rfc9380
package hash_to_field

import (
	_ "hash"

	_ "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)
