package hash

import (
	"errors"
)

var errStateOverflow = errors.New("the size of the state should not exceed the block size")

type merkleDamgardHasher struct {
	state []byte
	iv    []byte
	f     Compressor
}

// Write implements hash.Write
func (h *merkleDamgardHasher) Write(p []byte) (n int, err error) {
	blockSize := h.f.BlockSize()
	for len(p) != 0 {
		if len(p) < blockSize {
			p = append(make([]byte, blockSize-len(p), blockSize), p...)
		}
		if h.state, err = h.f.Compress(h.state, p[:blockSize]); err != nil {
			return
		}
		n += blockSize
		p = p[blockSize:]
	}
	return
}

func (h *merkleDamgardHasher) Sum(b []byte) []byte {
	if _, err := h.Write(b); err != nil {
		panic(err)
	}
	return h.state
}

func (h *merkleDamgardHasher) Reset() {
	h.state = h.iv
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

func (h *merkleDamgardHasher) SetState(state []byte) error {
	bs := h.BlockSize()
	if len(state) > bs {
		return errStateOverflow
	}
	h.state = make([]byte, bs)
	copy(h.state, state)
	return nil
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
// The value initialState is provided as initial input to the compression
// function. Its preimage should not be known and thus it should be generated
// using a deterministic method.
func NewMerkleDamgardHasher(f Compressor, initialState []byte) StateStorer {
	h := merkleDamgardHasher{
		iv: initialState,
		f:  f,
	}
	bs := h.BlockSize()
	h.state = make([]byte, bs)
	copy(h.state, initialState)
	return &h
}
