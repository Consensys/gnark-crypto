// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package hashtesting provides test utilities for the hashers of package hash.
package hashtesting

import (
	"io"
	"testing"

	"github.com/consensys/gnark-crypto/hash"
	"github.com/stretchr/testify/require"
)

// Hasher is a Merkle-Damgard hasher over field elements F with digests D.
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

	write := func(w io.Writer, b []byte) {
		t.Helper()
		n, err := w.Write(b)
		require.NoError(t, err)
		require.Equal(t, len(b), n)
	}

	h := newHasher()
	writeCalls(t, h, calls)
	h.Reset()

	ref := newRef()
	for i, c := range calls {
		if c.Bytes != nil {
			write(ref, c.Bytes)
			write(h, c.Bytes)
		} else {
			write(ref, Marshal[F, PF](c.Elements))
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
	writeCalls(t, restored, calls)
	require.NoError(t, restored.SetState(h.State()))

	h.PadBlock()
	require.Equal(t, ref.State(), h.State())

	for _, c := range calls {
		b := c.Bytes
		if b == nil {
			b = Marshal[F, PF](c.Elements)
		}
		write(h, b)
		write(restored, b)
	}
	require.Equal(t, h.Sum(nil), restored.Sum(nil))
}

func writeCalls[F, D any](t *testing.T, h Hasher[F, D], calls []Call[F]) {
	t.Helper()
	for _, c := range calls {
		if c.Bytes == nil {
			h.WriteField(c.Elements...)
			continue
		}
		_, err := h.Write(c.Bytes)
		require.NoError(t, err)
	}
}
