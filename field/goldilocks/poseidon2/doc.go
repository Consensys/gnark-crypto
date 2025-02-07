// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package poseidon2 implements the Poseidon2 permutation
//
// Poseidon2 permutation is a cryptographic permutation for algebraic hashes.
// See the [original paper] by Grassi, Khovratovich and Schofnegger for the full details.
//
// This implementation is based on the [reference implementation] from
// HorizenLabs. See the [specifications] for parameter choices.
//
// [reference implementation]: https://github.com/HorizenLabs/poseidon2/blob/main/plain_implementations/src/poseidon2/poseidon2.rs
// [specifications]: https://github.com/argumentcomputer/neptune/blob/main/spec/poseidon_spec.pdf
// [original paper]: https://eprint.iacr.org/2023/323.pdf
package poseidon2
