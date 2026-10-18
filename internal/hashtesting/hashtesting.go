// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package hashtesting provides test utilities for the hashers of package hash.
package hashtesting

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/consensys/gnark-crypto/hash"
	"github.com/stretchr/testify/require"
)

// Hasher is a hasher of field elements of type F, with digests of type D, that
// also hashes bytes.
type Hasher[F, D any] interface {
	hash.FieldHasher[F, D]
	hash.StateStorer
}

// Call is one write to a hasher. If Bytes is nil, Elements are written with
// WriteField. Otherwise Bytes are written with Write, and Elements are ignored.
type Call[F any] struct {
	Bytes    []byte
	Elements []F
}

// Marshal returns the concatenation of the marshalled elements.
func Marshal[F any, PF interface {
	*F
	Marshal() []byte
}](e []F) []byte {
	var b []byte
	for i := range e {
		b = append(b, PF(&e[i]).Marshal()...)
	}
	return b
}

// CheckMatchesByteHasher checks a hasher against a reference hasher over bytes.
// The reference pads the last block of each write, and WriteField does not, so
// if blocks have several elements, two calls of elements must not be
// consecutive.
//
// It checks that:
//   - Reset discards everything written before.
//   - After each call, Sum, State and the marshalled SumField of the hasher
//     equal the Sum and State of the reference, and Write reports the length of
//     its input.
//   - SumField does not change the state.
//   - A hasher restored by SetState from State continues as the original does
//     after PadBlock.
func CheckMatchesByteHasher[F any, PF interface {
	*F
	Marshal() []byte
}, D any](t *testing.T, newHasher func() Hasher[F, D], newRef func() hash.StateStorer, digestBytes func(D) []byte, calls []Call[F]) {
	t.Helper()

	h := newHasher()
	writeCalls[F, PF](t, calls, false, h)
	h.Reset()

	ref := newRef()
	for i, c := range calls {
		if c.Bytes != nil {
			write(t, ref, c.Bytes)
			write(t, h, c.Bytes)
		} else {
			write(t, ref, Marshal[F, PF](c.Elements))
			h.WriteField(c.Elements[:len(c.Elements)/2]...)
			h.SumField()
			h.WriteField(c.Elements[len(c.Elements)/2:]...)
		}

		msg := []any{"call %d of %d", i, len(calls)}
		want := ref.Sum(nil)
		require.Equal(t, want, h.Sum(nil), msg...)
		require.Equal(t, want, digestBytes(h.SumField()), msg...)
		require.Equal(t, ref.State(), h.State(), msg...)
	}

	restored := newHasher()
	// writing the calls twice puts restored in a state other than h's, so that
	// SetState must change it
	writeCalls[F, PF](t, calls, false, restored)
	writeCalls[F, PF](t, calls, false, restored)
	require.NoError(t, restored.SetState(h.State()))

	h.PadBlock()
	require.Equal(t, ref.State(), h.State())

	writeCalls[F, PF](t, calls, true, h, restored)
	require.Equal(t, h.Sum(nil), restored.Sum(nil))
}

// CheckMerkleDamgard runs CheckMatchesByteHasher on calls that write full and
// partial blocks of blockLen elements, as bytes, or alternately as bytes and
// field elements, including last blocks that do not end on an element boundary.
func CheckMerkleDamgard[F any, PF interface {
	*F
	Marshal() []byte
}, D any](t *testing.T, newHasher func() Hasher[F, D], newRef func() hash.StateStorer, digestBytes func(D) []byte, blockLen int, randomElements func(n int) []F) {
	t.Helper()

	var zero F
	elemSize := len(PF(&zero).Marshal())

	// A call writes k bytes of value 1, then n elements. With k != 0 it is written
	// as bytes, and its last block does not end on an element boundary.
	type call struct{ k, n int }
	cases := [][]call{
		{{n: 0}},
		{{n: 1}},
		{{n: blockLen}},
		{{n: blockLen - 1}, {n: 3}},
		{{n: 2*blockLen + 1}, {n: 5}, {n: 0}, {n: blockLen}},
		{{n: 37}},
	}
	for k := 1; k < elemSize; k++ {
		for n := range blockLen {
			cases = append(cases, []call{{n: blockLen}, {k: k, n: n}})
		}
	}

	for _, alternate := range []bool{false, true} {
		name := "bytes"
		if alternate {
			name = "alternating"
		}
		t.Run(name, func(t *testing.T) {
			for _, cs := range cases {
				t.Run(fmt.Sprint(cs), func(t *testing.T) {
					calls := make([]Call[F], len(cs))
					for i, c := range cs {
						elems := randomElements(c.n)
						if c.k == 0 && alternate && i%2 == 0 {
							calls[i] = Call[F]{Elements: elems}
						} else {
							calls[i] = Call[F]{Bytes: append(bytes.Repeat([]byte{1}, c.k), Marshal[F, PF](elems)...)}
						}
					}
					CheckMatchesByteHasher[F, PF](t, newHasher, newRef, digestBytes, calls)
				})
			}
		})
	}
}

// write writes b to w, and requires no error and that all of b is reported written.
func write(t *testing.T, w io.Writer, b []byte) {
	t.Helper()
	n, err := w.Write(b)
	require.NoError(t, err)
	require.Equal(t, len(b), n)
}

// writeCalls writes calls to each of hs, as given, or all as bytes if asBytes.
func writeCalls[F any, PF interface {
	*F
	Marshal() []byte
}, D any](t *testing.T, calls []Call[F], asBytes bool, hs ...Hasher[F, D]) {
	t.Helper()
	for _, c := range calls {
		for _, h := range hs {
			if c.Bytes == nil && !asBytes {
				h.WriteField(c.Elements...)
				continue
			}
			b := c.Bytes
			if b == nil {
				b = Marshal[F, PF](c.Elements)
			}
			write(t, h, b)
		}
	}
}
