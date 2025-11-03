package hash

import (
	"bytes"
	"fmt"
)

type merkleDamgardHasher struct {
	state  []byte
	iv     []byte
	buffer []byte // invariant: len(buffer) < cap(buffer) = block len
	f      Compressor
}

// Write implements hash.Write.
// The input is buffered until it hits the hash block size.
func (h *merkleDamgardHasher) Write(p []byte) (n int, err error) {
	blockSize := h.f.BlockSize()

	l := min(blockSize-len(h.buffer), len(p))
	h.buffer = append(h.buffer, p[:l]...)
	p = p[l:]
	n = l

	for len(h.buffer) == blockSize {
		if h.state, err = h.f.Compress(h.state, p[:blockSize]); err != nil {
			h.buffer = h.buffer[:len(h.buffer)-l]
			return n - l, err
		}

		l = min(blockSize, len(p))
		h.buffer = append(h.buffer[:0], p[:l]...)
		p = p[l:]
		n += l
	}

	return
}

// Sum returns the computed hash appended by the input b.
// If the written input's size is not a multiple of the block size, it will be right padded
// with zeros.
func (h *merkleDamgardHasher) Sum(b []byte) []byte {
	if len(h.buffer) == 0 {
		return append(b, h.state...)
	}

	// save the state
	state, buffer := bytes.Clone(h.state), bytes.Clone(h.buffer)

	// compute the sum
	if _, err := h.Write(make([]byte, h.BlockSize()-len(h.buffer))); err != nil {
		panic(err)
	}
	res := append(b, h.state...)

	// rewind the state
	h.state, h.buffer = state, buffer

	return res
}

func (h *merkleDamgardHasher) Reset() {
	h.state = bytes.Clone(h.iv)
}

func (h *merkleDamgardHasher) Size() int {
	return h.f.BlockSize()
}

func (h *merkleDamgardHasher) BlockSize() int {
	return h.f.BlockSize()
}

func (h *merkleDamgardHasher) State() []byte {
	return h.state
}

// SetState sets h's state to state. If len(state) > BlockSize, an error is thrown.
// If len(state) < BlockSize, h's state is set to state, and left padded with zeroes.
// len(state) > BlockSize will result in an error.
func (h *merkleDamgardHasher) SetState(state []byte) error {
	var err error
	h.state, err = cloneLeftPadded(state, h.BlockSize())
	return err
}

// NewMerkleDamgardHasher transforms a 2-1 one-way compression function into a
// hash function using a Merkle-Damgard construction. The resulting hash
// function has a block size equal to the block size of compression function.
//
// NB! The construction does not perform explicit padding on the input data. The
// last block of input data is zero-padded to full block size. This means that
// the construction is not collision resistant for generic data as the digest of
// input and input concatenated with zeros (up to the same number of total
// blocks) is same. For collision resistance the caller should perform explicit
// padding on the input data.
//
// - initialState is provided as initial input to the compression
// function. Its preimage should not be known and thus it should be generated
// using a deterministic method.
// If the given initialState is shorter than the hash block size, it will be zero-padded
// on the left. An oversized initialState will cause a panic.
func NewMerkleDamgardHasher(f Compressor, initialState []byte) StateStorer {
	iv, err := cloneLeftPadded(initialState, f.BlockSize())
	if err != nil {
		panic(err)
	}
	return &merkleDamgardHasher{
		f:      f,
		iv:     iv,
		state:  bytes.Clone(iv),
		buffer: make([]byte, 0, f.BlockSize()),
	}
}

// cloneLeftPadded copies b into a new byte slice of size n.
// If len(b) < n, it will be padded on the left.
// len(b) > n will result in an error.
func cloneLeftPadded(b []byte, n int) ([]byte, error) {
	if len(b) > n {
		return nil, fmt.Errorf("state/iv must not exceed the hash block size: %d > %d", len(b), n)
	}
	res := make([]byte, n)
	copy(res[n-len(b):], b)
	return res, nil
}
